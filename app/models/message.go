package models

import (
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

func CreateMessage(sender *User, recipient *User, body string) (*Message, error) {
	db := db.GetDb()

	var senderConversations []Conversation
	err := db.Model(&sender).Preload("Users", "ID = ?", recipient.ID).
		Association("Conversations").Find(&senderConversations)
	if err != nil {
		return nil, newGormError(err)
	}

	newMessage := Message{
		UserID: sender.ID,
		Body:   body,
	}

	if len(senderConversations) == 0 {
		newConversation := &Conversation{
			Users: []User{
				*sender,
				*recipient,
			},
			Messages: []Message{
				newMessage,
			},
		}
		result := db.Create(&newConversation)
		if result.Error != nil {
			return nil, newGormError(result.Error)
		}
	} else {
		conversation := senderConversations[0]
		err := db.Model(&conversation).Association("Messages").Append(&newMessage)
		if err != nil {
			return nil, newGormError(err)
		}
	}
	return &newMessage, nil
}
