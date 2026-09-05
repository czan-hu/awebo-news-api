package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type CreatorRequestRepository struct {
	db *gorm.DB
}

func NewCreatorRequestRepository(db *gorm.DB) *CreatorRequestRepository {
	return &CreatorRequestRepository{db: db}
}

func (r *CreatorRequestRepository) LatestForUser(userID uuid.UUID) (*models.CreatorRequest, error) {
	var req models.CreatorRequest
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *CreatorRequestRepository) HasPending(userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.CreatorRequest{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CreatorRequestRepository) Create(userID uuid.UUID) (*models.CreatorRequest, error) {
	req := &models.CreatorRequest{
		ID:     uuid.New(),
		UserID: userID,
		Status: "pending",
	}
	if err := r.db.Create(req).Error; err != nil {
		return nil, err
	}
	return req, nil
}

func (r *CreatorRequestRepository) FindByID(id uuid.UUID) (*models.CreatorRequest, error) {
	var req models.CreatorRequest
	if err := r.db.Preload("User").First(&req, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *CreatorRequestRepository) ListPending(page, limit int) ([]models.CreatorRequest, int64, error) {
	query := r.db.Model(&models.CreatorRequest{}).Where("status = ?", "pending")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var requests []models.CreatorRequest
	if err := query.Preload("User").
		Order("created_at asc").
		Offset((page - 1) * limit).Limit(limit).
		Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func (r *CreatorRequestRepository) Review(id, reviewerID uuid.UUID, status string) error {
	now := time.Now()
	return r.db.Model(&models.CreatorRequest{}).Where("id = ?", id).Updates(map[string]any{
		"status":      status,
		"reviewed_by": reviewerID,
		"reviewed_at": now,
	}).Error
}
