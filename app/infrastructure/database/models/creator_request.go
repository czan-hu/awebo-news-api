package models

import (
	"time"

	"github.com/google/uuid"
)

type CreatorRequest struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	User   User      `gorm:"foreignKey:UserID"`

	Status     string     `gorm:"not null;default:'pending'"`
	ReviewedBy *uuid.UUID `gorm:"type:uuid"`
	ReviewedAt *time.Time

	CreatedAt time.Time
}

func (CreatorRequest) TableName() string { return "creator_requests" }
