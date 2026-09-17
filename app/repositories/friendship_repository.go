package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

type FriendshipRepository struct {
	db *gorm.DB
}

func NewFriendshipRepository(db *gorm.DB) *FriendshipRepository {
	return &FriendshipRepository{db: db}
}

// FindPair returns the non-declined row between the two users, in either
// direction, or nil if none exists. A declined row is ignored so a new
// request can be sent after a decline — the unique index only guards
// pending/accepted rows, so at most one non-declined row can exist per pair.
func (r *FriendshipRepository) FindPair(userA, userB uuid.UUID) (*models.Friendship, error) {
	var f models.Friendship
	err := r.db.Where("status <> ?", "declined").
		Where("(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
			userA, userB, userB, userA).
		First(&f).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (r *FriendshipRepository) Create(requesterID, addresseeID uuid.UUID) (*models.Friendship, error) {
	f := &models.Friendship{
		ID:          uuid.New(),
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      "pending",
	}
	if err := r.db.Create(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FriendshipRepository) FindByID(id uuid.UUID) (*models.Friendship, error) {
	var f models.Friendship
	if err := r.db.Preload("Requester").Preload("Addressee").First(&f, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (r *FriendshipRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Friendship{}).Where("id = ?", id).Update("status", status).Error
}

func (r *FriendshipRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Friendship{}, "id = ?", id).Error
}

func (r *FriendshipRepository) ListFriends(userID uuid.UUID, page, limit int) ([]models.Friendship, int64, error) {
	query := r.db.Model(&models.Friendship{}).
		Where("status = ? AND (requester_id = ? OR addressee_id = ?)", "accepted", userID, userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var friendships []models.Friendship
	if err := query.Preload("Requester").Preload("Addressee").
		Order("updated_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&friendships).Error; err != nil {
		return nil, 0, err
	}
	return friendships, total, nil
}

func (r *FriendshipRepository) ListIncoming(userID uuid.UUID, page, limit int) ([]models.Friendship, int64, error) {
	query := r.db.Model(&models.Friendship{}).
		Where("status = ? AND addressee_id = ?", "pending", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var friendships []models.Friendship
	if err := query.Preload("Requester").
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&friendships).Error; err != nil {
		return nil, 0, err
	}
	return friendships, total, nil
}

func (r *FriendshipRepository) ListOutgoing(userID uuid.UUID, page, limit int) ([]models.Friendship, int64, error) {
	query := r.db.Model(&models.Friendship{}).
		Where("status = ? AND requester_id = ?", "pending", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var friendships []models.Friendship
	if err := query.Preload("Addressee").
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&friendships).Error; err != nil {
		return nil, 0, err
	}
	return friendships, total, nil
}

// FriendIDs returns the ids of every user accepted-friends with userID —
// whichever side of each accepted row isn't userID itself.
func (r *FriendshipRepository) FriendIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	var asRequester []uuid.UUID
	if err := r.db.Model(&models.Friendship{}).
		Where("status = ? AND requester_id = ?", "accepted", userID).
		Pluck("addressee_id", &asRequester).Error; err != nil {
		return nil, err
	}

	var asAddressee []uuid.UUID
	if err := r.db.Model(&models.Friendship{}).
		Where("status = ? AND addressee_id = ?", "accepted", userID).
		Pluck("requester_id", &asAddressee).Error; err != nil {
		return nil, err
	}

	return append(asRequester, asAddressee...), nil
}
