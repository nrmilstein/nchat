package models

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nrmilstein/nchat/db"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("No user found.")

type User struct {
	ID            int            `gorm:"primaryKey,not null"`
	Email         string         `gorm:"not null"`
	Password      string         `gorm:"not null"`
	Name          string         `gorm:"not null"`
	Conversations []Conversation `gorm:"many2many:conversation_users;"`
	Messages      []Message
	CreatedAt     time.Time `gorm:"not null"`
}

func GetUserFromKey(key string) (*User, error) {
	if key == "" {
		return nil, ErrUserNotFound
	}
	db := db.GetDb()
	var session Session
	readSession := db.Joins("User").Take(&session, &Session{Key: key}) // TODO: exclude password
	if errors.Is(readSession.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	} else if readSession.Error != nil {
		return nil, readSession.Error
	}
	return &session.User, nil
}

func GetUserFromRequest(c *gin.Context) (*User, error) {
	return GetUserFromKey(c.GetHeader("X-API-Key"))
}

func HashPassword(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(str)))
}
