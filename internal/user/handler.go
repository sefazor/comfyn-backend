package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sefazor/comfyn/internal/models"
	"github.com/sefazor/comfyn/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UpdateProfileInput struct {
	Biography         string `json:"biography"`
	InstagramUsername string `json:"instagramUsername"`
	Username          string `json:"username"`
	ProfileImage      string `json:"profileImage"`
}

type UpdateSecurityInput struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
	Email           string `json:"email" binding:"omitempty,email"`
}

type SearchUsersQuery struct {
	Query string `form:"q"`
	Page  int    `form:"page,default=1"`
	Limit int    `form:"limit,default=20"`
}

// Kendi profilini görüntüleme
func GetProfileHandler(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	if err := database.DB.First(&currentUser, currentUser.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": currentUser.SafeResponse()})
}

// Başka kullanıcının profilini görüntüleme
func GetUserProfileHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var targetUser models.User
	if err := database.DB.First(&targetUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Takip durumunu kontrol et
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var count int64
	database.DB.Table("user_followers").
		Where("follower_id = ? AND following_id = ?", currentUser.ID, targetUser.ID).
		Count(&count)

	response := targetUser.SafeResponse()
	response["isFollowing"] = count > 0

	c.JSON(http.StatusOK, gin.H{"user": response})
}

// Profil güncelleme
func UpdateProfileHandler(c *gin.Context) {
	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Username kontrolü
	if input.Username != "" && input.Username != currentUser.Username {
		var existingUser models.User
		if err := database.DB.Where("username = ?", input.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
			return
		}
		currentUser.Username = input.Username
	}

	currentUser.Biography = input.Biography
	currentUser.InstagramUsername = input.InstagramUsername
	if input.ProfileImage != "" {
		currentUser.ProfileImage = input.ProfileImage
	}

	if err := database.DB.Save(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": currentUser.SafeResponse()})
}

// Güvenlik ayarlarını güncelleme
func UpdateSecurityHandler(c *gin.Context) {
	var input UpdateSecurityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Mevcut şifreyi kontrol et
	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Email kontrolü
	if input.Email != "" && input.Email != currentUser.Email {
		var existingUser models.User
		if err := database.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
			return
		}
		currentUser.Email = input.Email
	}

	// Yeni şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	currentUser.Password = string(hashedPassword)

	if err := database.DB.Save(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update security settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Security settings updated successfully",
		"user":    currentUser.SafeResponse(),
	})
}

// Follow/Unfollow işlemi
func FollowUserHandler(c *gin.Context) {
	targetUserIDStr := c.Param("id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	if currentUser.ID == uint(targetUserID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot follow yourself"})
		return
	}

	tx := database.DB.Begin()

	var targetUser models.User
	if err := tx.First(&targetUser, targetUserID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var count int64
	tx.Table("user_followers").
		Where("follower_id = ? AND following_id = ?", currentUser.ID, targetUser.ID).
		Count(&count)

	if count > 0 {
		// Unfollow
		if err := tx.Exec("DELETE FROM user_followers WHERE follower_id = ? AND following_id = ?",
			currentUser.ID, targetUser.ID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow user"})
			return
		}

		// Sayaçları güncelle
		if err := tx.Model(&targetUser).Update("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update follower count"})
			return
		}

		if err := tx.Model(&currentUser).Update("following_count", gorm.Expr("following_count - ?", 1)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update following count"})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"message":     "Successfully unfollowed user",
			"isFollowing": false,
		})
		return
	}

	// Follow
	if err := tx.Exec("INSERT INTO user_followers (follower_id, following_id) VALUES (?, ?)",
		currentUser.ID, targetUser.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow user"})
		return
	}

	// Sayaçları güncelle
	if err := tx.Model(&targetUser).Update("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update follower count"})
		return
	}

	if err := tx.Model(&currentUser).Update("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update following count"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":     "Successfully followed user",
		"isFollowing": true,
	})
}

// Kullanıcı arama
func SearchUsersHandler(c *gin.Context) {
	var query SearchUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.Limit > 50 {
		query.Limit = 50
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.DB.Model(&models.User{})

	if query.Query != "" {
		searchQuery := "%" + query.Query + "%"
		db = db.Where("username ILIKE ? OR full_name ILIKE ?", searchQuery, searchQuery)
	}

	db = db.Where("id != ?", currentUser.ID)

	offset := (query.Page - 1) * query.Limit

	var total int64
	db.Count(&total)

	var users []models.User
	if err := db.Select("id, username, full_name, biography, instagram_username, follower_count, following_count, total_views").
		Limit(query.Limit).
		Offset(offset).
		Order("follower_count DESC").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	type UserResponse struct {
		models.User
		IsFollowing bool `json:"isFollowing"`
	}

	response := make([]UserResponse, len(users))
	for i, u := range users {
		var count int64
		database.DB.Table("user_followers").
			Where("follower_id = ? AND following_id = ?", currentUser.ID, u.ID).
			Count(&count)

		response[i] = UserResponse{
			User:        u,
			IsFollowing: count > 0,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"pagination": gin.H{
			"current": query.Page,
			"limit":   query.Limit,
			"total":   total,
			"pages":   (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}
