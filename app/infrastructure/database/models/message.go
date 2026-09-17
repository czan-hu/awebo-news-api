package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	ConversationID uuid.UUID `gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Sender         User      `gorm:"foreignKey:SenderID"`
	Body           string    `gorm:"type:text;not null"`

	CreatedAt time.Time
}

func (Message) TableName() string { return "messages" }
