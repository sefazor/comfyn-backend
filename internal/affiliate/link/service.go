// internal/affiliate/link/service.go
package link

import (
	"fmt"

	"github.com/sefazor/comfyn/internal/models"
	"github.com/sefazor/comfyn/pkg/database"
	"gorm.io/gorm"
)

func GenerateTrackingURL(userID uint, postID uint, productID uint, originalURL string) (string, error) {
	// Benzersiz bir tracking ID oluştur
	trackingID := fmt.Sprintf("cmf_%d_%d_%d", userID, postID, productID)

	// Tracking URL oluştur
	trackingURL := fmt.Sprintf("https://comfyn.com/go/%s", trackingID)

	// Link kaydını oluştur
	link := models.AffiliateLink{
		UserID:      userID,
		PostID:      postID,
		ProductID:   productID,
		OriginalURL: originalURL,
		TrackingURL: trackingURL,
	}

	if err := database.DB.Create(&link).Error; err != nil {
		return "", err
	}

	return trackingURL, nil
}

func LogClick(linkID uint, userID *uint, ip string, userAgent string, referer string) error {
	// Tıklama logunu kaydet
	clickLog := models.ClickLog{
		AffiliateLinkID: linkID,
		UserID:          userID,
		IP:              ip,
		UserAgent:       userAgent,
		RefererURL:      referer,
	}

	tx := database.DB.Begin()

	if err := tx.Create(&clickLog).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Tıklanma sayısını artır
	if err := tx.Model(&models.AffiliateLink{}).
		Where("id = ?", linkID).
		Update("click_count", gorm.Expr("click_count + ?", 1)).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
