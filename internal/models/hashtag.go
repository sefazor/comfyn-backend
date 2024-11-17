package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Hashtag struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:50;not null;unique"`
	Posts     []Post `gorm:"many2many:post_hashtags;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func NormalizeHashtag(tag string) string {
	// # işaretini kaldır ve küçük harfe çevir
	tag = strings.ToLower(strings.TrimPrefix(tag, "#"))
	// Boşlukları kaldır
	return strings.ReplaceAll(tag, " ", "")
}
