package repositories

import (
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type ContactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) List() ([]models.Contact, error) {
	var contacts []models.Contact
	if err := r.db.Order("sort_order asc, created_at asc").Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}
