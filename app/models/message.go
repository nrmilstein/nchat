package models

import (
	"net/http"
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

func CreateMessage(
	user *User,
	conversationId int,
	messageBody string,
) (*Message, int, error) {
	db := db.GetDb()

	var conversations []Conversation
	db.Model(&user).Association("Conversations").
		Find(&conversations, Conversation{ID: conversationId})

	if len(conversations) == 0 {
		return nil, http.StatusNotFound, utils.AppError{"Conversation does not exist.", 1, nil}
	}

	newMessage := Message{
		UserID:         user.ID,
		ConversationID: conversationId,
		Body:           messageBody,
	}
	createMessageResult := db.Create(&newMessage)
	if createMessageResult.Error != nil {
		return nil, http.StatusInternalServerError, utils.ErrInternalServer
	}

	return &newMessage, http.StatusCreated, nil
}
