// internal/models/affiliate_partner.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type AffiliatePartner struct {
	ID             uint    `gorm:"primaryKey"`
	Name           string  `gorm:"not null"`
	BaseURL        string  `gorm:"not null"`
	CommissionRate float64 `gorm:"not null"` // Yüzde olarak (örn: 5.5)
	WebhookSecret  string  `gorm:"not null"` // Webhook güvenliği için
	ApiKey         string  `gorm:"not null"` // Partner API erişimi için
	IsActive       bool    `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
