package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                uint   `gorm:"primaryKey"`
	FullName          string `gorm:"size:100;not null"`
	Email             string `gorm:"size:100;not null;unique"`
	Username          string `gorm:"size:50;not null;unique"`
	Password          string `gorm:"not null"`
	ProfileImage      string `gorm:"size:255"`
	Biography         string `gorm:"size:500"`
	InstagramUsername string `gorm:"size:50"`
	FollowerCount     int    `gorm:"default:0"`
	FollowingCount    int    `gorm:"default:0"`
	TotalViews        int    `gorm:"default:0"`
	Followers         []User `gorm:"many2many:user_followers;joinForeignKey:following_id;joinReferences:follower_id"`
	Following         []User `gorm:"many2many:user_followers;joinForeignKey:follower_id;joinReferences:following_id"`
	Posts             []Post `gorm:"foreignKey:UserID"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (user *User) SafeResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":                user.ID,
		"fullName":          user.FullName,
		"email":             user.Email,
		"username":          user.Username,
		"profileImage":      user.ProfileImage,
		"biography":         user.Biography,
		"instagramUsername": user.InstagramUsername,
		"followerCount":     user.FollowerCount,
		"followingCount":    user.FollowingCount,
		"totalViews":        user.TotalViews,
		"createdAt":         user.CreatedAt,
	}
}
