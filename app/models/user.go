package models

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"neal-chat/db"
)

var ErrUserNotFound = errors.New("No user found.")

type User struct {
	Id      int
	Email   string
	Name    string
	Created time.Time
}

func GetUserFromKey(key string) (*User, error) {
	db := db.GetDb()
	user := new(User)
	err := db.QueryRow(
		"SELECT users.id, users.email, users.name, users.created "+
			"FROM auth_keys JOIN users ON auth_keys.user_id = users.id "+
			"WHERE auth_key = $1",
		key,
	).Scan(&user.Id, &user.Email, &user.Name, &user.Created)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserFromRequest(c *gin.Context) (*User, error) {
	return GetUserFromKey(c.GetHeader("X-API-Key"))
}

func HashPassword(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(str)))
}
