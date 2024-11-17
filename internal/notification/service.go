package notification

import (
	"github.com/sefazor/comfyn/internal/models"
	"github.com/sefazor/comfyn/pkg/database"
)

// Bildirim oluşturma servisi
func CreateNotification(userID, actorID uint, notificationType models.NotificationType, postID, commentID *uint) error {
	// Kullanıcının bildirim tercihlerini kontrol et
	var pref models.NotificationPreference
	if err := database.DB.FirstOrCreate(&pref, models.NotificationPreference{UserID: userID}).Error; err != nil {
		return err
	}

	// Tercihlere göre bildirimi kontrol et
	shouldNotify := false
	switch notificationType {
	case models.NotificationNewFollower:
		shouldNotify = pref.NewFollower
	case models.NotificationPostLike:
		shouldNotify = pref.PostLike
	case models.NotificationComment:
		shouldNotify = pref.Comment
	}

	if !shouldNotify {
		return nil
	}

	// Bildirimi oluştur
	notification := models.Notification{
		UserID:    userID,
		ActorID:   actorID,
		Type:      notificationType,
		PostID:    postID,
		CommentID: commentID,
	}

	return database.DB.Create(&notification).Error
}
