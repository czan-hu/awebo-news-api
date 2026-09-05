package models

import (
	"time"

	"github.com/google/uuid"
)

type NewsletterSubscriber struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
}

func (NewsletterSubscriber) TableName() string { return "newsletter_subscribers" }
