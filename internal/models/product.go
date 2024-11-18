package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"not null"`
	Price       float64 `gorm:"not null"`
	Link        string
	TrackingURL string // Yeni alan
	Description string
	Categories  []Category `gorm:"many2many:product_categories;"`
	Posts       []Post     `gorm:"many2many:post_products;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// Tracking URL oluşturma metodu
func (p *Product) GenerateTrackingURL(userID uint, postID uint) {
	trackingID := fmt.Sprintf("cmf_%d_%d_%d", userID, postID, p.ID)
	p.TrackingURL = fmt.Sprintf("http://localhost:8080/go/%s", trackingID)
}
