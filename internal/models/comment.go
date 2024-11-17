package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	PostID    uint   `gorm:"not null"`
	UserID    uint   `gorm:"not null"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	User User `gorm:"foreignkey:UserID"`
	Post Post `gorm:"foreignkey:PostID"`
}
