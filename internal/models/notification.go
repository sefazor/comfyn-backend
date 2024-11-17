package models

import (
	"time"

	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationNewFollower NotificationType = "new_follower"
	NotificationPostLike    NotificationType = "post_like"
	NotificationComment     NotificationType = "comment"
)

type Notification struct {
	ID        uint             `gorm:"primaryKey"`
	UserID    uint             `gorm:"not null"`
	ActorID   uint             `gorm:"not null"`
	Type      NotificationType `gorm:"not null"`
	PostID    *uint            `gorm:"default:null"`
	CommentID *uint            `gorm:"default:null"`
	IsRead    bool             `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	User    User     `gorm:"foreignkey:UserID"`
	Actor   User     `gorm:"foreignkey:ActorID"`
	Post    *Post    `gorm:"foreignkey:PostID"`
	Comment *Comment `gorm:"foreignkey:CommentID"`
}

type NotificationPreference struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint `gorm:"not null;uniqueIndex"`
	NewFollower bool `gorm:"default:true"`
	PostLike    bool `gorm:"default:true"`
	Comment     bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	User User `gorm:"foreignkey:UserID"`
}

func (n *Notification) Response() map[string]interface{} {
	resp := map[string]interface{}{
		"id":        n.ID,
		"type":      n.Type,
		"isRead":    n.IsRead,
		"createdAt": n.CreatedAt,
		"actor":     n.Actor.SafeResponse(),
	}

	switch n.Type {
	case NotificationPostLike:
		if n.Post != nil {
			resp["post"] = n.Post.Response()
		}
	case NotificationComment:
		if n.Post != nil {
			resp["post"] = n.Post.Response()
		}
		if n.Comment != nil {
			resp["comment"] = n.Comment
		}
	}

	return resp
}
