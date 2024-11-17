package models

import (
	"time"
)

type ClickLog struct {
	ID              uint   `gorm:"primaryKey"`
	AffiliateLinkID uint   `gorm:"not null"`
	UserID          *uint  // Tıklayan kullanıcı (eğer giriş yapmışsa)
	IP              string `gorm:"not null"`
	UserAgent       string `gorm:"not null"` // Browser/Device bilgisi
	RefererURL      string // Hangi sayfadan geldiği
	CreatedAt       time.Time

	AffiliateLink AffiliateLink `gorm:"foreignkey:AffiliateLinkID"`
	User          *User         `gorm:"foreignkey:UserID"` // Opsiyonel ilişki
}
