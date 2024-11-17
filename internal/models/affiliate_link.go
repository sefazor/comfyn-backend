// internal/models/affiliate_link.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type AffiliateLink struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null"`
	PostID      uint   `gorm:"not null"`
	ProductID   uint   `gorm:"not null"`
	OriginalURL string `gorm:"not null"`
	TrackingURL string `gorm:"not null"`
	ClickCount  int    `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	User    User    `gorm:"foreignkey:UserID"`
	Post    Post    `gorm:"foreignkey:PostID"`
	Product Product `gorm:"foreignkey:ProductID"`
}
