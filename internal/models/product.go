package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"not null"`
	Price       float64 `gorm:"not null"`
	Link        string
	Description string
	Categories  []Category `gorm:"many2many:product_categories;"`
	Posts       []Post     `gorm:"many2many:post_products;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
