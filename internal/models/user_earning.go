// internal/models/user_earning.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentPending    PaymentStatus = "pending"
	PaymentProcessing PaymentStatus = "processing"
	PaymentCompleted  PaymentStatus = "completed"
	PaymentFailed     PaymentStatus = "failed"
)

type UserEarning struct {
	ID            uint          `gorm:"primaryKey"`
	UserID        uint          `gorm:"not null"`
	TransactionID uint          `gorm:"not null"`
	Amount        float64       `gorm:"not null"`
	Status        PaymentStatus `gorm:"not null;default:'pending'"`
	PaymentDate   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	User        User                 `gorm:"foreignkey:UserID"`
	Transaction AffiliateTransaction `gorm:"foreignkey:TransactionID"`
}
