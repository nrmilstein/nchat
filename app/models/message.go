package models

import (
	"errors"
	"time"

	"github.com/nrmilstein/nchat/db"
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

type GormError struct {
	err error
	msg string
}

func (e GormError) Error() string {
	return e.msg
}

func (e GormError) Unwrap() error {
	return e.err
}

func newGormError(e error) GormError {
	return GormError{
		err: e,
		msg: "Gorm error: " + e.Error(),
	}
}

func CreateMessage(sender *User, recipient *User, body string) (*Message, *Conversation, error) {
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
		result := db.Create(&newConversation)
		if result.Error != nil {
			return nil, nil, newGormError(result.Error)
		}
		conversation = newConversation
		newMessage = &newConversation.Messages[0]
	} else {
		err := db.Model(&conversation).Association("Messages").Append(&newMessage)
		if err != nil {
			return nil, nil, newGormError(err)
		}
	}
	return newMessage, conversation, nil
}

func GetConversation(sender *User, recipient *User) (*Conversation, error) {
	db := db.GetDb()

	var senderConversations []Conversation
	err := db.Model(&sender).Preload("Users", "ID = ?", recipient.ID).
		Association("Conversations").Find(&senderConversations)
	if err != nil {
		return nil, newGormError(err)
	}

	for _, conversation := range senderConversations {
		if len(conversation.Users) == 1 {
			return &conversation, nil
		}
	}

	return nil, ErrConversationNotFound
}
