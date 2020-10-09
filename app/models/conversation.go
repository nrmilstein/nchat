package models

import (
	"time"
)

type Conversation struct {
	ID        int    `gorm:"primaryKey,not null"`
	Users     []User `gorm:"many2many:conversation_users;"`
	Messages  []Message
	CreatedAt time.Time `gorm:"not null"`
}
