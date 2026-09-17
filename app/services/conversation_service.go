package services

import (
	"github.com/google/uuid"

	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/database/models"
	"awebo/app/infrastructure/realtime"
	"awebo/app/repositories"
)

type ConversationService struct {
	repo     *repositories.ConversationRepository
	userRepo *repositories.UserRepository
	hub      *realtime.Hub
}

func NewConversationService(repo *repositories.ConversationRepository, userRepo *repositories.UserRepository, hub *realtime.Hub) *ConversationService {
	return &ConversationService{repo: repo, userRepo: userRepo, hub: hub}
}

// StartOrGet resolves otherUsername to a user and returns the (possibly
// newly created) 1:1 conversation with them.
func (s *ConversationService) StartOrGet(userID uuid.UUID, otherUsername string) (*entities.Conversation, *exceptions.AppError) {
	other, err := s.userRepo.FindByUsername(otherUsername)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if other == nil {
		return nil, exceptions.NotFound("Пользователь не найден.")
	}
	if other.ID == userID {
		return nil, exceptions.BadRequest("Нельзя написать самому себе.")
	}

	conversation, err := s.repo.FindOrCreateDirect(userID, other.ID)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return s.toConversationFor(userID, *conversation)
}

func (s *ConversationService) ListMine(userID uuid.UUID, page, limit int) (*entities.ConversationList, *exceptions.AppError) {
	page, limit = normalizePage(page, limit)

	conversations, total, err := s.repo.ListForUser(userID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	data := make([]entities.Conversation, 0, len(conversations))
	for _, c := range conversations {
		entity, appErr := s.toConversationFor(userID, c)
		if appErr != nil {
			return nil, appErr
		}
		data = append(data, *entity)
	}

	return &entities.ConversationList{Data: data, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

// Send persists a message and, if the recipient has an open WebSocket
// connection, pushes it to them immediately. An offline recipient sees it
// on their next fetch — no special-casing needed.
func (s *ConversationService) Send(userID, conversationID uuid.UUID, body string) (*entities.Message, *exceptions.AppError) {
	if err := s.mustBeParticipant(conversationID, userID); err != nil {
		return nil, err
	}

	message, err := s.repo.AddMessage(conversationID, userID, body)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	entity := toMessage(*message)

	otherID, err := s.repo.OtherParticipant(conversationID, userID)
	if err == nil && otherID != uuid.Nil {
		s.hub.SendToUser(otherID, realtime.Envelope{Type: "message.new", Payload: entity})
	}

	return entity, nil
}

func (s *ConversationService) ListMessages(userID, conversationID uuid.UUID, page, limit int) (*entities.MessageList, *exceptions.AppError) {
	page, limit = normalizePage(page, limit)

	if err := s.mustBeParticipant(conversationID, userID); err != nil {
		return nil, err
	}

	messages, total, err := s.repo.ListMessages(conversationID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	data := make([]entities.Message, 0, len(messages))
	for _, m := range messages {
		data = append(data, *toMessage(m))
	}

	return &entities.MessageList{Data: data, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *ConversationService) MarkRead(userID, conversationID uuid.UUID) *exceptions.AppError {
	if err := s.mustBeParticipant(conversationID, userID); err != nil {
		return err
	}
	if err := s.repo.MarkRead(conversationID, userID); err != nil {
		return exceptions.Internal("")
	}
	return nil
}

func (s *ConversationService) mustBeParticipant(conversationID, userID uuid.UUID) *exceptions.AppError {
	conversation, err := s.repo.FindByID(conversationID)
	if err != nil {
		return exceptions.Internal("")
	}
	if conversation == nil {
		return exceptions.NotFound("")
	}
	ok, err := s.repo.IsParticipant(conversationID, userID)
	if err != nil {
		return exceptions.Internal("")
	}
	if !ok {
		return exceptions.Forbidden("")
	}
	return nil
}

func (s *ConversationService) toConversationFor(viewerID uuid.UUID, c models.Conversation) (*entities.Conversation, *exceptions.AppError) {
	otherID, err := s.repo.OtherParticipant(c.ID, viewerID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	other, err := s.userRepo.FindByID(otherID)
	if err != nil || other == nil {
		return nil, exceptions.Internal("")
	}

	unread, err := s.repo.UnreadCount(c.ID, viewerID)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	messages, _, err := s.repo.ListMessages(c.ID, 1, 1)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	var lastMessage *entities.Message
	if len(messages) > 0 {
		lastMessage = toMessage(messages[0])
	}

	return &entities.Conversation{
		ID:          c.ID.String(),
		Other:       entities.PublicUser{Name: other.Name, Username: usernameOf(*other), Avatar: other.AvatarPath},
		LastMessage: lastMessage,
		UnreadCount: unread,
		CreatedAt:   c.CreatedAt,
	}, nil
}

func toMessage(m models.Message) *entities.Message {
	return &entities.Message{
		ID:             m.ID.String(),
		ConversationID: m.ConversationID.String(),
		Sender:         entities.PublicUser{Name: m.Sender.Name, Username: usernameOf(m.Sender), Avatar: m.Sender.AvatarPath},
		Body:           m.Body,
		CreatedAt:      m.CreatedAt,
	}
}
