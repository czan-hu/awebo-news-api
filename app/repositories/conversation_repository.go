package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// FindOrCreateDirect returns the existing 1:1 conversation between the two
// users, or creates one (+ its two participant rows) if none exists yet.
// v1 only ever creates two-participant conversations, so "the conversation
// with exactly these two participants" is unambiguous.
func (r *ConversationRepository) FindOrCreateDirect(userA, userB uuid.UUID) (*models.Conversation, error) {
	// .Row().Scan (raw database/sql scanning) rather than GORM's own .Scan —
	// uuid.UUID is a [16]byte under the hood, and GORM's reflection-based
	// Scan treats array-kind destinations as "one row per element" instead
	// of respecting uuid.UUID's sql.Scanner implementation, which fails with
	// "converting driver.Value type string ... to a uint8".
	var existingID uuid.UUID
	row := r.db.Table("conversation_participants AS cp1").
		Select("cp1.conversation_id").
		Joins("JOIN conversation_participants AS cp2 ON cp2.conversation_id = cp1.conversation_id AND cp2.user_id = ?", userB).
		Where("cp1.user_id = ?", userA).
		Limit(1).
		Row()
	switch err := row.Scan(&existingID); {
	case err == nil:
		return r.FindByID(existingID)
	case !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}

	conversation := &models.Conversation{ID: uuid.New()}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(conversation).Error; err != nil {
			return err
		}
		participants := []models.ConversationParticipant{
			{ConversationID: conversation.ID, UserID: userA},
			{ConversationID: conversation.ID, UserID: userB},
		}
		return tx.Create(&participants).Error
	})
	if err != nil {
		return nil, err
	}
	return conversation, nil
}

func (r *ConversationRepository) FindByID(id uuid.UUID) (*models.Conversation, error) {
	var conversation models.Conversation
	if err := r.db.First(&conversation, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conversation, nil
}

// ListForUser returns the conversations userID participates in, most
// recently active first. GORM has no clean builder for "order by the max
// timestamp in a related table", so this uses a scoped raw subquery — the
// one deliberate exception to the codebase's general raw-SQL avoidance.
func (r *ConversationRepository) ListForUser(userID uuid.UUID, page, limit int) ([]models.Conversation, int64, error) {
	base := r.db.Model(&models.Conversation{}).
		Joins("JOIN conversation_participants ON conversation_participants.conversation_id = conversations.id").
		Where("conversation_participants.user_id = ?", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var conversations []models.Conversation
	err := base.
		Order("(SELECT MAX(m.created_at) FROM messages m WHERE m.conversation_id = conversations.id) DESC NULLS LAST").
		Order("conversations.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}
	return conversations, total, nil
}

func (r *ConversationRepository) IsParticipant(conversationID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Count(&count).Error
	return count > 0, err
}

// OtherParticipant returns the id of whichever participant isn't userID —
// used to target a WebSocket push after a message is sent. v1 conversations
// always have exactly two participants.
func (r *ConversationRepository) OtherParticipant(conversationID, userID uuid.UUID) (uuid.UUID, error) {
	var otherID uuid.UUID
	row := r.db.Model(&models.ConversationParticipant{}).
		Select("user_id").
		Where("conversation_id = ? AND user_id <> ?", conversationID, userID).
		Limit(1).
		Row()
	err := row.Scan(&otherID)
	return otherID, err
}

func (r *ConversationRepository) AddMessage(conversationID, senderID uuid.UUID, body string) (*models.Message, error) {
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Body:           body,
	}
	if err := r.db.Create(message).Error; err != nil {
		return nil, err
	}
	return r.FindMessageByID(message.ID)
}

func (r *ConversationRepository) FindMessageByID(id uuid.UUID) (*models.Message, error) {
	var message models.Message
	if err := r.db.Preload("Sender").First(&message, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// ListMessages returns the newest page first (consistent pagination
// semantics with every other list endpoint); callers rendering a chat
// window reverse the page for oldest-first reading order.
func (r *ConversationRepository) ListMessages(conversationID uuid.UUID, page, limit int) ([]models.Message, int64, error) {
	query := r.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var messages []models.Message
	if err := query.Preload("Sender").
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

func (r *ConversationRepository) MarkRead(conversationID, userID uuid.UUID) error {
	return r.db.Model(&models.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_at", time.Now()).Error
}

func (r *ConversationRepository) LastReadAt(conversationID, userID uuid.UUID) (time.Time, error) {
	var participant models.ConversationParticipant
	err := r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		First(&participant).Error
	if err != nil {
		return time.Time{}, err
	}
	return participant.LastReadAt, nil
}

func (r *ConversationRepository) UnreadCount(conversationID, userID uuid.UUID) (int64, error) {
	lastReadAt, err := r.LastReadAt(conversationID, userID)
	if err != nil {
		return 0, err
	}

	var count int64
	err = r.db.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id <> ? AND created_at > ?", conversationID, userID, lastReadAt).
		Count(&count).Error
	return count, err
}
