package models

import (
	"time"

	"github.com/google/uuid"
)

type ArticleLike struct {
	ArticleID int       `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time
}

func (ArticleLike) TableName() string { return "article_likes" }

type ArticleSave struct {
	ArticleID int       `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time
}

func (ArticleSave) TableName() string { return "article_saves" }

// ArticleView doubles as the user's reading history: ViewedAt is bumped on
// every repeat visit instead of inserting duplicate rows.
type ArticleView struct {
	ArticleID int       `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid"`
	ViewedAt  time.Time `gorm:"not null"`
}

func (ArticleView) TableName() string { return "article_views" }
