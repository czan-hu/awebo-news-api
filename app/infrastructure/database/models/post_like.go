package models

import (
	"time"

	"github.com/google/uuid"
)

type PostLike struct {
	PostID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time
}

func (PostLike) TableName() string { return "post_likes" }
