package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null;unique"`
	Slug        string    `gorm:"size:100;not null;unique"`
	Description string    `gorm:"type:text"`
	Products    []Product `gorm:"many2many:product_categories;"`
	Posts       []Post    `gorm:"many2many:post_categories;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
