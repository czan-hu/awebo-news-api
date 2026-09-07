package models

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Label     string    `gorm:"not null"`
	Href      string    `gorm:"not null"`
	Icon      string    `gorm:"type:text;not null"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time
}

func (Contact) TableName() string { return "contacts" }
