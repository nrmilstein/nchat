package models

import (
	"time"
)

type Message struct {
	Id           int
	Conversation *Conversation
	User         *User
	Sent         time.Time
	Body         string
}
