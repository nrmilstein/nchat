package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
	"gorm.io/gorm"
)

type Session struct {
	ID        int    `gorm:"primaryKey"`
	Key       string `gorm:"not null"`
	User      User
	UserID    int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	// AccessedAt time.Time
}

var ErrInvalidCred = errors.New("Invalid username/password.")

func CreateSession(username string, password string) (*Session, *User, error) {
	db := db.GetDb()

	hashedPassword := HashPassword(password)

	var user User
	readUserResult := db.Take(&user, &User{Username: username, Password: hashedPassword})

	if readUserResult.Error != nil {
		if errors.Is(readUserResult.Error, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCred
		}
		return nil, nil, utils.NewGormError(readUserResult.Error)
	}

	randBytes := make([]byte, 18)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("Error generating session key: %w", err)
	}
	authKey := base64.URLEncoding.EncodeToString(randBytes)

	session := Session{
		Key:    authKey,
		UserID: user.ID,
	}
	createUserResult := db.Create(&session)

	if createUserResult.Error != nil {
		return nil, nil, utils.NewGormError(createUserResult.Error)
	}

	if createUserResult.RowsAffected == 0 {
		return nil, nil, errors.New("Could not create session.")
	}

	return &session, &user, nil
}
