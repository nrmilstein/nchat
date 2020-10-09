package models

import (
	"time"
)

type Session struct {
	ID        int    `gorm:"primaryKey"`
	Key       string `gorm:"not null"`
	User      User
	UserID    int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	// AccessedAt time.Time
}
