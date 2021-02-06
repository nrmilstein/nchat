package models

import (
	"errors"
	"time"

	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

var ErrConversationNotFound = errors.New("Conversation not found.")

type Conversation struct {
	ID        int    `gorm:"primaryKey,not null"`
	Users     []User `gorm:"many2many:conversation_users;"`
	Messages  []Message
	CreatedAt time.Time `gorm:"not null"`
}

func GetConversation(sender *User, recipient *User) (*Conversation, error) {
	db := db.GetDb()

	var senderConversations []Conversation
	err := db.Model(&sender).Preload("Users", "ID = ?", recipient.ID).
		Association("Conversations").Find(&senderConversations)
	if err != nil {
		return nil, utils.NewGormError(err)
	}

	for _, conversation := range senderConversations {
		if len(conversation.Users) == 1 {
			return &conversation, nil
		}
	}

	return nil, ErrConversationNotFound
}
