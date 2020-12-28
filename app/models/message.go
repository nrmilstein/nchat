package models

import (
	"errors"
	"time"

	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

type Message struct {
	ID             int       `gorm:"primaryKey,not null"`
	UserID         int       `gorm:"not null"`
	ConversationID int       `gorm:"not null"`
	Body           string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"not null"`
}

var ErrConversationNotFound = errors.New("Conversation not found.")
var ErrTooManyConversations = errors.New("Too many conversations found between given users.")
var ErrSameUser = errors.New("Cannot send message to self.")

func CreateMessage(sender *User, recipient *User, body string) (*Message, *Conversation, error) {
	if sender.ID == recipient.ID {
		return nil, nil, ErrSameUser
	}

	db := db.GetDb()

	conversation, err := GetConversation(sender, recipient)
	if !errors.Is(err, ErrConversationNotFound) && err != nil {
		return nil, nil, err
	}

	newMessage := &Message{
		UserID: sender.ID,
		Body:   body,
	}

	if errors.Is(err, ErrConversationNotFound) {
		newConversation := &Conversation{
			Users: []User{
				*sender,
				*recipient,
			},
			Messages: []Message{
				*newMessage,
			},
		}

		result := db.Omit("Users.*").Create(newConversation)
		if result.Error != nil {
			return nil, nil, utils.NewGormError(result.Error)
		}

		conversation = newConversation
		newMessage = &newConversation.Messages[0]
	} else {
		err := db.Model(conversation).Association("Messages").Append(newMessage)
		if err != nil {
			return nil, nil, utils.NewGormError(err)
		}
	}
	return newMessage, conversation, nil
}
