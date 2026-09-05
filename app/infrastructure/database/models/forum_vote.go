package models

import "github.com/google/uuid"

type ForumVote struct {
	CommentID string    `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid"`
	Direction int       `gorm:"not null"`
}

func (ForumVote) TableName() string { return "forum_votes" }
