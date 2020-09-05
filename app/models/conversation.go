package models

import (
	"time"
)

type Conversation struct {
  Id       int
  Created  time.Time
  Users    []*User
  Messages []*Message
}
