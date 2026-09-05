package models

import (
	"time"

	"github.com/google/uuid"
)

type ForumTopic struct {
	ID string `gorm:"primaryKey"`

	Title    string    `gorm:"not null"`
	Category string    `gorm:"not null;index"`
	AuthorID uuid.UUID `gorm:"type:uuid;not null;index"`
	Author   User      `gorm:"foreignKey:AuthorID"`
	Body     string    `gorm:"type:text;not null"`

	Pinned       bool  `gorm:"not null;default:false"`
	CommentCount int64 `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ForumTopic) TableName() string { return "forum_topics" }
