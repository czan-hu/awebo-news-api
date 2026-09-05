package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awebo/app/infrastructure/database/models"
)

// AuthRepository covers only what the email + one-time-code auth flow needs.
// Everything else about a user account lives in UserRepository.
type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpsertVerificationCode creates the user row on first request-code call, or
// reissues a fresh code for an existing one.
func (r *AuthRepository) UpsertVerificationCode(email, code string, expiresAt, requestedAt time.Time) (*models.User, error) {
	user, err := r.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user = &models.User{
			ID:                        uuid.New(),
			Email:                     email,
			VerificationCode:          &code,
			VerificationCodeExpiresAt: &expiresAt,
			CodeRequestedAt:           &requestedAt,
		}
		if err := r.db.Create(user).Error; err != nil {
			return nil, err
		}
		return user, nil
	}

	updates := map[string]any{
		"verification_code":            code,
		"verification_code_expires_at": expiresAt,
		"code_requested_at":            requestedAt,
	}
	if err := r.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.FindByID(user.ID)
}

func (r *AuthRepository) ClearVerificationCode(userID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{
			"verification_code":            nil,
			"verification_code_expires_at": nil,
		}).Error
}

func (r *AuthRepository) IsUsernameTaken(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AuthRepository) CompleteProfile(userID uuid.UUID, name, username string) (*models.User, error) {
	updates := map[string]any{
		"name":              name,
		"username":          username,
		"profile_completed": true,
	}
	if err := r.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.FindByID(userID)
}
