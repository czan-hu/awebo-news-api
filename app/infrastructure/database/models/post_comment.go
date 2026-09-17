package models

import (
	"time"

	"github.com/google/uuid"
)

type PostComment struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	PostID   uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorID uuid.UUID `gorm:"type:uuid;not null;index"`
	Author   User      `gorm:"foreignKey:AuthorID"`
	Body     string    `gorm:"type:text;not null"`

	CreatedAt time.Time
}

func (PostComment) TableName() string { return "post_comments" }
