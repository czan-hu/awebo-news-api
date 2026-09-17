package models

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time
}

func (Conversation) TableName() string { return "conversations" }
