package models

import (
	"time"

	"github.com/google/uuid"
)

type Friendship struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	RequesterID uuid.UUID `gorm:"type:uuid;not null;index"`
	Requester   User      `gorm:"foreignKey:RequesterID"`

	AddresseeID uuid.UUID `gorm:"type:uuid;not null;index"`
	Addressee   User      `gorm:"foreignKey:AddresseeID"`

	Status string `gorm:"not null;default:'pending'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Friendship) TableName() string { return "friendships" }
