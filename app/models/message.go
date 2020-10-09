package models

import (
	"time"
)

type Message struct {
	ID             int       `gorm:"primaryKey,not null"`
	UserID         int       `gorm:"not null"`
	ConversationID int       `gorm:"not null"`
	Body           string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"not null"`
}
