package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) IsUsernameTakenByOther(username string, excludeUserID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).
		Where("username = ? AND id <> ?", username, excludeUserID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) UpdateFields(id uuid.UUID, fields map[string]any) (*models.User, error) {
	if len(fields) > 0 {
		if err := r.db.Model(&models.User{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return nil, err
		}
	}
	return r.FindByID(id)
}

func (r *UserRepository) ReplaceLinks(userID uuid.UUID, links []models.UserLink) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserLink{}).Error; err != nil {
			return err
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Create(&links).Error
	})
}

func (r *UserRepository) ListLinks(userID uuid.UUID) ([]models.UserLink, error) {
	var links []models.UserLink
	if err := r.db.Where("user_id = ?", userID).Order("position asc").Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

type UserStatsRow struct {
	Articles int64
	Liked    int64
	Saved    int64
	History  int64
}

func (r *UserRepository) Stats(userID uuid.UUID) (UserStatsRow, error) {
	var stats UserStatsRow

	if err := r.db.Model(&models.Article{}).Where("author_id = ?", userID).Count(&stats.Articles).Error; err != nil {
		return stats, err
	}
	if err := r.db.Model(&models.ArticleLike{}).Where("user_id = ?", userID).Count(&stats.Liked).Error; err != nil {
		return stats, err
	}
	if err := r.db.Model(&models.ArticleSave{}).Where("user_id = ?", userID).Count(&stats.Saved).Error; err != nil {
		return stats, err
	}
	if err := r.db.Model(&models.ArticleView{}).Where("user_id = ?", userID).Count(&stats.History).Error; err != nil {
		return stats, err
	}

	return stats, nil
}
