// internal/affiliate/link/handler.go
package link

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sefazor/comfyn/internal/models"
	"github.com/sefazor/comfyn/pkg/database"
)

// Link yönlendirme handler'ı
func RedirectHandler(c *gin.Context) {
	trackingID := c.Param("tracking_id")

	var link models.AffiliateLink
	if err := database.DB.Where("tracking_url LIKE ?", "%"+trackingID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Kullanıcı bilgisini al (varsa)
	var userID *uint
	if user, exists := c.Get("user"); exists {
		currentUser := user.(models.User)
		userID = &currentUser.ID
	}

	// Tıklamayı logla
	go LogClick(
		link.ID,
		userID,
		c.ClientIP(),
		c.Request.UserAgent(),
		c.Request.Referer(),
	)

	// Orijinal URL'e yönlendir
	c.Redirect(http.StatusTemporaryRedirect, link.OriginalURL)
}

// Analytics handler'ı
func GetLinkAnalyticsHandler(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var links []models.AffiliateLink
	if err := database.DB.Where("user_id = ?", currentUser.ID).
		Preload("Post").
		Preload("Product").
		Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}

	response := make([]map[string]interface{}, len(links))
	for i, link := range links {
		response[i] = map[string]interface{}{
			"id":          link.ID,
			"originalURL": link.OriginalURL,
			"trackingURL": link.TrackingURL,
			"clickCount":  link.ClickCount,
			"post":        link.Post.Response(),
			"product":     link.Product,
			"createdAt":   link.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"links": response})
}
