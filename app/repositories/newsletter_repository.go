package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type NewsletterRepository struct {
	db *gorm.DB
}

func NewNewsletterRepository(db *gorm.DB) *NewsletterRepository {
	return &NewsletterRepository{db: db}
}

func (r *NewsletterRepository) IsSubscribed(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.NewsletterSubscriber{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *NewsletterRepository) Subscribe(email string) error {
	return r.db.Create(&models.NewsletterSubscriber{ID: uuid.New(), Email: email}).Error
}
