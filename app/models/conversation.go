package models

import (
	"errors"
	"net/http"
	"time"

	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
	"gorm.io/gorm"
)

type Conversation struct {
	ID        int    `gorm:"primaryKey,not null"`
	Users     []User `gorm:"many2many:conversation_users;"`
	Messages  []Message
	CreatedAt time.Time `gorm:"not null"`
}

var ErrConversationPartnerNotFound = errors.New(
	"User not part of conversation or conversation not found.")

var ErrGorm = errors.New("Gorm error")

func GetConversationPartner(user *User, conversationId int) (*User, error) {
	db := db.GetDb()

	var conversation Conversation
	response := db.Preload("Users", "ID <> ?", user.ID).
		Take(&conversation, Conversation{ID: conversationId})

	if response.Error != nil {
		return nil, ErrGorm
	}

	if len(conversation.Users) != 1 {
		return nil, ErrConversationPartnerNotFound
	}
	return &conversation.Users[0], nil
}

func CreateConversation(
	user *User,
	recipientEmail string,
	messageBody string,
) (*Conversation, int, error) {
	db := db.GetDb()
	if recipientEmail == user.Email {
		return nil, http.StatusConflict, utils.AppError{"Cannot create conversation with self.", 1, nil}
	}

	var recipientUser User
	readRecipientUserResult := db.Take(&recipientUser, &User{Email: recipientEmail})

	if errors.Is(readRecipientUserResult.Error, gorm.ErrRecordNotFound) {
		return nil, http.StatusUnprocessableEntity, utils.AppError{"Recipient user does not exist", 5, nil}
	} else if readRecipientUserResult.Error != nil {
		return nil, http.StatusInternalServerError, utils.ErrInternalServer
	}

	var myConversations []Conversation
	var otherUsers []User
	err := db.Model(user).Association("Conversations").Find(&myConversations)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, http.StatusInternalServerError, utils.ErrInternalServer
	}

	err = db.Model(&myConversations).
		Where(&User{ID: recipientUser.ID}).Association("Users").Find(&otherUsers)
	if len(otherUsers) > 0 {
		return nil, http.StatusConflict, utils.AppError{"Conversation already exists.", 6, nil}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, http.StatusInternalServerError, utils.ErrInternalServer
	}

	newMessage := Message{
		UserID: user.ID,
		Body:   messageBody,
	}
	newConversation := Conversation{
		Messages: []Message{
			newMessage,
		},
		Users: []User{
			*user,
			recipientUser,
		},
	}
	createConversationResult := db.Create(&newConversation)
	if createConversationResult.Error != nil {
		return nil, http.StatusInternalServerError, utils.ErrInternalServer
	}

	return &newConversation, http.StatusCreated, nil
}
