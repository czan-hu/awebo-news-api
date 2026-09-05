package repositories

import (
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.Order("slug asc").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) Exists(slug string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Category{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
