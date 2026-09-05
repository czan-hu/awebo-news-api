package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type ForumRepository struct {
	db *gorm.DB
}

func NewForumRepository(db *gorm.DB) *ForumRepository {
	return &ForumRepository{db: db}
}

type ForumTopicFilter struct {
	Category string
	Sort     string // "new" | "top"
}

func (r *ForumRepository) TopicExists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ForumTopic{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ForumRepository) CommentExists(id string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ForumComment{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ForumRepository) ListTopics(filter ForumTopicFilter, page, limit int) ([]models.ForumTopic, int64, error) {
	query := r.db.Model(&models.ForumTopic{})
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "pinned DESC, created_at DESC"
	if filter.Sort == "top" {
		order = "pinned DESC, comment_count DESC, created_at DESC"
	}

	var topics []models.ForumTopic
	if err := query.Preload("Author").
		Order(order).
		Offset((page - 1) * limit).Limit(limit).
		Find(&topics).Error; err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

func (r *ForumRepository) CreateTopic(topic *models.ForumTopic) error {
	return r.db.Create(topic).Error
}

func (r *ForumRepository) FindTopicByID(id string) (*models.ForumTopic, error) {
	var topic models.ForumTopic
	if err := r.db.Preload("Author").First(&topic, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &topic, nil
}

func (r *ForumRepository) IncrementCommentCount(topicID string, delta int64) error {
	return r.db.Model(&models.ForumTopic{}).Where("id = ?", topicID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

func (r *ForumRepository) ListComments(topicID, sort string) ([]models.ForumComment, error) {
	order := "score DESC, created_at ASC"
	switch sort {
	case "new":
		order = "created_at DESC"
	case "old":
		order = "created_at ASC"
	}

	var comments []models.ForumComment
	if err := r.db.Preload("Author").
		Where("topic_id = ?", topicID).
		Order(order).
		Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *ForumRepository) FindCommentByID(id string) (*models.ForumComment, error) {
	var comment models.ForumComment
	if err := r.db.Preload("Author").First(&comment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

// CreateComment inserts the comment and its author's automatic up-vote
// (score starts at 1) inside one transaction, per the API contract.
func (r *ForumRepository) CreateComment(comment *models.ForumComment) error {
	comment.Score = 1
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Create(&models.ForumVote{
			CommentID: comment.ID,
			UserID:    comment.AuthorID,
			Direction: 1,
		}).Error
	})
}

func (r *ForumRepository) MyVotes(userID uuid.UUID, commentIDs []string) (map[string]int, error) {
	if len(commentIDs) == 0 {
		return map[string]int{}, nil
	}
	var votes []models.ForumVote
	if err := r.db.Where("user_id = ? AND comment_id IN ?", userID, commentIDs).Find(&votes).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int, len(votes))
	for _, v := range votes {
		result[v.CommentID] = v.Direction
	}
	return result, nil
}

// SetVote applies direction (-1/1) or removes the vote (0) and returns the
// comment's updated score.
func (r *ForumRepository) SetVote(commentID string, userID uuid.UUID, direction int) (int, error) {
	var newScore int

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.ForumVote
		findErr := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&existing).Error
		hadVote := findErr == nil
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}

		var delta int
		switch {
		case direction == 0 && hadVote:
			delta = -existing.Direction
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
		case direction != 0 && hadVote:
			delta = direction - existing.Direction
			if err := tx.Model(&existing).Where("comment_id = ? AND user_id = ?", commentID, userID).
				Update("direction", direction).Error; err != nil {
				return err
			}
		case direction != 0 && !hadVote:
			delta = direction
			if err := tx.Create(&models.ForumVote{CommentID: commentID, UserID: userID, Direction: direction}).Error; err != nil {
				return err
			}
		default:
			delta = 0
		}

		if delta != 0 {
			if err := tx.Model(&models.ForumComment{}).Where("id = ?", commentID).
				UpdateColumn("score", gorm.Expr("score + ?", delta)).Error; err != nil {
				return err
			}
		}

		var comment models.ForumComment
		if err := tx.Select("score").First(&comment, "id = ?", commentID).Error; err != nil {
			return err
		}
		newScore = comment.Score

		return nil
	})

	return newScore, err
}

func (r *ForumRepository) SearchTopics(q string, limit int) ([]models.ForumTopic, error) {
	var topics []models.ForumTopic
	like := "%" + q + "%"
	if err := r.db.Preload("Author").
		Where("title ILIKE ? OR body ILIKE ?", like, like).
		Order("created_at DESC").
		Limit(limit).
		Find(&topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}
