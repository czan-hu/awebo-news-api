package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"awebo/app/infrastructure/database/models"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) FindByID(id uuid.UUID) (*models.Post, error) {
	var post models.Post
	if err := r.db.Preload("Author").First(&post, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

// ListByAuthors powers the feed: posts by the viewer plus (once Phase C
// lands) their friends, newest first.
func (r *PostRepository) ListByAuthors(authorIDs []uuid.UUID, page, limit int) ([]models.Post, int64, error) {
	query := r.db.Model(&models.Post{}).Where("author_id IN ?", authorIDs)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []models.Post
	if err := query.Preload("Author").
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

// Like is idempotent: liking an already-liked post is a no-op rather than an
// error, and only bumps like_count when a row was actually inserted.
func (r *PostRepository) Like(postID, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&models.PostLike{PostID: postID, UserID: userID})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&models.Post{}).Where("id = ?", postID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

// Unlike mirrors Like: a no-op if there was nothing to remove.
func (r *PostRepository) Unlike(postID, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&models.PostLike{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&models.Post{}).Where("id = ?", postID).
			UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
	})
}

func (r *PostRepository) IsLiked(postID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.PostLike{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// LikedPostIDs batch-resolves the viewer's liked state across a page of
// posts in one query, avoiding an IsLiked call per feed item.
func (r *PostRepository) LikedPostIDs(userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(postIDs) == 0 {
		return map[uuid.UUID]bool{}, nil
	}
	var rows []uuid.UUID
	if err := r.db.Model(&models.PostLike{}).
		Where("user_id = ? AND post_id IN ?", userID, postIDs).
		Pluck("post_id", &rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]bool, len(rows))
	for _, id := range rows {
		result[id] = true
	}
	return result, nil
}

func (r *PostRepository) CreateComment(comment *models.PostComment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&models.Post{}).Where("id = ?", comment.PostID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	})
}

func (r *PostRepository) FindCommentByID(id uuid.UUID) (*models.PostComment, error) {
	var comment models.PostComment
	if err := r.db.Preload("Author").First(&comment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (r *PostRepository) ListComments(postID uuid.UUID, page, limit int) ([]models.PostComment, int64, error) {
	query := r.db.Model(&models.PostComment{}).Where("post_id = ?", postID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []models.PostComment
	if err := query.Preload("Author").
		Order("created_at ASC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}
