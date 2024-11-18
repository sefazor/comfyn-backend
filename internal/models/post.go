package models

import (
	"time"

	"gorm.io/gorm"
)

const MaxProductsPerPost = 8

type Post struct {
	ID          uint       `gorm:"primaryKey"`
	UserID      uint       `gorm:"not null"`
	ImageURL    string     `gorm:"not null"`
	Description string     `gorm:"type:text"`
	ViewCount   int        `gorm:"default:0"`
	Products    []Product  `gorm:"many2many:post_products;"`
	Categories  []Category `gorm:"many2many:post_categories;"`
	Hashtags    []Hashtag  `gorm:"many2many:post_hashtags;"`
	Likes       []Like     `gorm:"foreignKey:PostID"`
	Comments    []Comment  `gorm:"foreignKey:PostID"`
	User        User       `gorm:"foreignkey:UserID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (post *Post) Response() map[string]interface{} {
	productsResponse := make([]map[string]interface{}, len(post.Products))
	for i, product := range post.Products {
		productsResponse[i] = map[string]interface{}{
			"id":          product.ID,
			"name":        product.Name,
			"price":       product.Price,
			"link":        product.Link,
			"trackingUrl": product.TrackingURL,
			"description": product.Description,
		}
	}

	return map[string]interface{}{
		"id":           post.ID,
		"user":         post.User.SafeResponse(),
		"imageUrl":     post.ImageURL,
		"description":  post.Description,
		"products":     productsResponse,
		"categories":   post.Categories,
		"hashtags":     post.Hashtags,
		"likeCount":    len(post.Likes),
		"commentCount": len(post.Comments),
		"createdAt":    post.CreatedAt,
	}
}
