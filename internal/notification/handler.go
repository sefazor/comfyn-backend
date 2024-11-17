package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sefazor/comfyn/internal/models"
	"github.com/sefazor/comfyn/pkg/database"
)

// Bildirimleri listeleme
func GetNotificationsHandler(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var notifications []models.Notification
	if err := database.DB.Where("user_id = ?", currentUser.ID).
		Preload("Actor").
		Preload("Post").
		Preload("Comment").
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	// Response'ları hazırla
	response := make([]map[string]interface{}, len(notifications))
	for i, notification := range notifications {
		response[i] = notification.Response()
	}

	c.JSON(http.StatusOK, gin.H{"notifications": response})
}

// Bildirimi okundu olarak işaretle
func MarkNotificationReadHandler(c *gin.Context) {
	notificationID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var notification models.Notification
	if err := database.DB.Where("id = ? AND user_id = ?", notificationID, currentUser.ID).
		First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if err := database.DB.Model(&notification).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// Bildirim tercihlerini güncelleme input'u
type UpdatePreferencesInput struct {
	NewFollower bool `json:"newFollower"`
	PostLike    bool `json:"postLike"`
	Comment     bool `json:"comment"`
}

// Bildirim tercihlerini güncelleme
func UpdateNotificationPreferencesHandler(c *gin.Context) {
	var input UpdatePreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var pref models.NotificationPreference
	// Tercihler yoksa oluştur, varsa güncelle
	if err := database.DB.FirstOrCreate(&pref, models.NotificationPreference{UserID: currentUser.ID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get preferences"})
		return
	}

	// Tercihleri güncelle
	updates := map[string]interface{}{
		"new_follower": input.NewFollower,
		"post_like":    input.PostLike,
		"comment":      input.Comment,
	}

	if err := database.DB.Model(&pref).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Notification preferences updated",
		"preferences": pref,
	})
}

// Tüm bildirimleri okundu olarak işaretle
func MarkAllNotificationsReadHandler(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	if err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", currentUser.ID, false).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// Okunmamış bildirim sayısını getir
func GetUnreadNotificationCountHandler(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var count int64
	if err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", currentUser.ID, false).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread notification count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unreadCount": count})
}
