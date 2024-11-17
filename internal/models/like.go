package models

import (
	"time"
)

type Like struct {
	ID        uint `gorm:"primaryKey"`
	PostID    uint `gorm:"not null"`
	UserID    uint `gorm:"not null"`
	CreatedAt time.Time

	User User `gorm:"foreignkey:UserID"`
	Post Post `gorm:"foreignkey:PostID"`
}
