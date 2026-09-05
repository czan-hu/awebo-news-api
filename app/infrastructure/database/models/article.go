package models

import (
	"time"

	"github.com/google/uuid"

	"awebo/app/entities"
)

type Article struct {
	ID int `gorm:"primaryKey;autoIncrement"`

	Slug         string    `gorm:"uniqueIndex;not null"`
	Title        string    `gorm:"not null"`
	Excerpt      string    `gorm:"type:text;not null"`
	CategorySlug string    `gorm:"column:category_slug;not null;index"`
	AuthorID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Author       User      `gorm:"foreignKey:AuthorID"`

	PublishedAt    time.Time `gorm:"not null;index"`
	ReadingMinutes int       `gorm:"not null;default:1"`
	Cover          string    `gorm:"not null;default:''"`
	Featured       bool      `gorm:"not null;default:false;index"`

	Tags    JSONArray[string]                 `gorm:"type:jsonb;not null;default:'[]'"`
	Content JSONArray[entities.ContentBlock]  `gorm:"type:jsonb;not null;default:'[]'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Article) TableName() string { return "articles" }
