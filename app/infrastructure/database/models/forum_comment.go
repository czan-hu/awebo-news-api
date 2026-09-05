package models

import (
	"time"

	"github.com/google/uuid"
)

type ForumComment struct {
	ID string `gorm:"primaryKey"`

	TopicID  string    `gorm:"not null;index"`
	ParentID *string   `gorm:"index"`
	AuthorID uuid.UUID `gorm:"type:uuid;not null;index"`
	Author   User      `gorm:"foreignKey:AuthorID"`
	Body     string    `gorm:"type:text;not null"`
	Score    int       `gorm:"not null;default:0"`

	CreatedAt time.Time
}

func (ForumComment) TableName() string { return "forum_comments" }
