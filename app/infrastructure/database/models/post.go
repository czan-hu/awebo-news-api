package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	AuthorID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Author    User      `gorm:"foreignKey:AuthorID"`
	Body      string    `gorm:"type:text;not null"`
	ImagePath string    `gorm:"column:image_path;not null;default:''"`

	LikeCount    int64 `gorm:"not null;default:0"`
	CommentCount int64 `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Post) TableName() string { return "posts" }
