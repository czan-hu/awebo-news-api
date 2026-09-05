package models

import (
	"time"

	"github.com/google/uuid"
)

// User covers both the auth flow (email + one-time code) and the public
// profile. Username/Name stay nullable because a freshly created account
// (right after verify-code) has neither yet — ProfileCompleted tracks that.
type User struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	Role  string `gorm:"not null;default:'user'"`
	Title string `gorm:"not null;default:''"` // display line, e.g. "Автор · Технологии"

	Name     string  `gorm:"not null;default:''"`
	Username *string `gorm:"uniqueIndex"`
	Email    string  `gorm:"uniqueIndex;not null"`

	AvatarPath string `gorm:"column:avatar_path;not null;default:''"`
	Bio        string `gorm:"type:text;not null;default:''"`
	Location   string `gorm:"not null;default:''"`

	VerificationCode          *string
	VerificationCodeExpiresAt *time.Time
	CodeRequestedAt           *time.Time

	ProfileCompleted bool `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string { return "users" }
