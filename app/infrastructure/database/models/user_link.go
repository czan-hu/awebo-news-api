package models

import "github.com/google/uuid"

type UserLink struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Label    string    `gorm:"not null"`
	URL      string    `gorm:"not null"`
	Icon     string    `gorm:"not null"`
	Position int       `gorm:"not null;default:0"`
}

func (UserLink) TableName() string { return "user_links" }
