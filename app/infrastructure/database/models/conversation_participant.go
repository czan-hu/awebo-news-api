package models

import (
	"time"

	"github.com/google/uuid"
)

type ConversationParticipant struct {
	ConversationID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	User           User      `gorm:"foreignKey:UserID"`

	LastReadAt time.Time
	JoinedAt   time.Time
}

func (ConversationParticipant) TableName() string { return "conversation_participants" }
