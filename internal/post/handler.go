package post

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sefazor/comfyn/internal/models"
	"github.com/sefazor/comfyn/internal/notification"
	"github.com/sefazor/comfyn/pkg/database"
	"gorm.io/gorm"
)

type CreatePostInput struct {
	ImageURL    string    `json:"imageUrl" binding:"required"`
	Description string    `json:"description"`
	Products    []Product `json:"products" binding:"required,dive,required"`
	CategoryIDs []uint    `json:"categoryIds" binding:"required,min=1"`
	Hashtags    []string  `json:"hashtags"`
}

type Product struct {
	Name        string  `json:"name" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Link        string  `json:"link"`
	Description string  `json:"description"`
	CategoryIDs []uint  `json:"categoryIds" binding:"required,min=1"`
}

type CreateCommentInput struct {
	Content string `json:"content" binding:"required"`
}

func CreatePostHandler(c *gin.Context) {
	var input CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	tx := database.DB.Begin()

	// Kategorileri kontrol et
	var categories []models.Category
	if err := tx.Find(&categories, input.CategoryIDs).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category IDs"})
		return
	}

	// Hashtag'leri işle
	var hashtags []models.Hashtag
	for _, tagName := range input.Hashtags {
		normalizedTag := models.NormalizeHashtag(tagName)
		if normalizedTag == "" {
			continue
		}

		var hashtag models.Hashtag
		err := tx.Where("name = ?", normalizedTag).FirstOrCreate(&hashtag, models.Hashtag{
			Name: normalizedTag,
		}).Error
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process hashtags"})
			return
		}
		hashtags = append(hashtags, hashtag)
	}

	// Ürünleri oluştur
	var products []models.Product
	for _, p := range input.Products {
		var productCategories []models.Category
		if err := tx.Find(&productCategories, p.CategoryIDs).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product category IDs"})
			return
		}

		product := models.Product{
			Name:        p.Name,
			Price:       p.Price,
			Link:        p.Link,
			Description: p.Description,
			Categories:  productCategories,
		}
		products = append(products, product)
	}

	post := models.Post{
		UserID:      currentUser.ID,
		ImageURL:    input.ImageURL,
		Description: input.Description,
		Categories:  categories,
		Products:    products,
		Hashtags:    hashtags,
	}

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	// Post'u tüm ilişkileriyle birlikte yükle
	if err := tx.Preload("User").
		Preload("Categories").
		Preload("Products").
		Preload("Products.Categories").
		Preload("Hashtags").
		First(&post, post.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load post data"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post created successfully",
		"post":    post.Response(),
	})
}

func ListPostsHandler(c *gin.Context) {
	var posts []models.Post

	if err := database.DB.Preload("User").
		Preload("Products").
		Preload("Categories").
		Preload("Hashtags").
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	response := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		response[i] = post.Response()
	}

	c.JSON(http.StatusOK, gin.H{"posts": response})
}

func GetPostHandler(c *gin.Context) {
	postID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var post models.Post
	if err := database.DB.Preload("User").
		Preload("Products").
		Preload("Products.Categories").
		Preload("Categories").
		Preload("Hashtags").
		Preload("Likes").
		Preload("Comments").
		Preload("Comments.User").
		First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var isLiked bool
	for _, like := range post.Likes {
		if like.UserID == currentUser.ID {
			isLiked = true
			break
		}
	}

	response := post.Response()
	response["isLiked"] = isLiked
	response["comments"] = post.Comments

	c.JSON(http.StatusOK, gin.H{"post": response})
}

func DeletePostHandler(c *gin.Context) {
	postID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var post models.Post
	if err := database.DB.Preload("Products").First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	if post.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own posts"})
		return
	}

	tx := database.DB.Begin()

	if err := tx.Model(&post).Association("Products").Clear(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post products"})
		return
	}

	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Post deleted successfully",
	})
}

func LikePostHandler(c *gin.Context) {
	postID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	tx := database.DB.Begin()

	var post models.Post
	if err := tx.First(&post, postID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var existingLike models.Like
	err := tx.Where("post_id = ? AND user_id = ?", post.ID, currentUser.ID).First(&existingLike).Error

	if err == nil {
		if err := tx.Delete(&existingLike).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike post"})
			return
		}

		// İlgili bildirimi sil
		tx.Where("actor_id = ? AND post_id = ? AND type = ?",
			currentUser.ID, post.ID, models.NotificationPostLike).
			Delete(&models.Notification{})

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"message": "Post unliked successfully",
			"liked":   false,
		})
		return
	}

	like := models.Like{
		PostID: post.ID,
		UserID: currentUser.ID,
	}

	if err := tx.Create(&like).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like post"})
		return
	}

	if post.UserID != currentUser.ID {
		if err := notification.CreateNotification(
			post.UserID,
			currentUser.ID,
			models.NotificationPostLike,
			&post.ID,
			nil,
		); err != nil {
			log.Printf("Failed to create notification: %v", err)
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Post liked successfully",
		"liked":   true,
	})
}

func CreateCommentHandler(c *gin.Context) {
	postID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var input CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var post models.Post
	if err := tx.First(&post, postID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	comment := models.Comment{
		PostID:  post.ID,
		UserID:  currentUser.ID,
		Content: input.Content,
	}

	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	if err := tx.Preload("User").First(&comment, comment.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load comment"})
		return
	}

	if post.UserID != currentUser.ID {
		if err := notification.CreateNotification(
			post.UserID,
			currentUser.ID,
			models.NotificationComment,
			&post.ID,
			&comment.ID,
		); err != nil {
			log.Printf("Failed to create notification: %v", err)
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Comment added successfully",
		"comment": comment,
	})
}

func IncrementViewHandler(c *gin.Context) {
	postID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	tx := database.DB.Begin()

	var post models.Post
	if err := tx.First(&post, postID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var existingView models.PostView
	isNewView := tx.Where("post_id = ? AND user_id = ?", post.ID, currentUser.ID).
		First(&existingView).Error == gorm.ErrRecordNotFound

	if isNewView {
		view := models.PostView{
			PostID: post.ID,
			UserID: currentUser.ID,
		}
		if err := tx.Create(&view).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record view"})
			return
		}

		if err := tx.Model(&post).Update("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to increment view count"})
			return
		}

		if err := tx.Model(&models.User{}).Where("id = ?", post.UserID).
			Update("total_views", gorm.Expr("total_views + ?", 1)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user total views"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":   "View count updated",
		"isNewView": isNewView,
		"viewCount": post.ViewCount + 1,
	})
}

func GetPersonalFeedHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Takip edilen kullanıcıların postlarını getir
	query := database.DB.Model(&models.Post{}).
		Preload("User").
		Preload("Products").
		Preload("Categories").
		Preload("Hashtags").
		Preload("Likes").
		Joins("JOIN user_followers uf ON posts.user_id = uf.following_id").
		Where("uf.follower_id = ?", currentUser.ID).
		Order("posts.created_at DESC")

	// Toplam post sayısını al
	var total int64
	query.Count(&total)

	// Postları getir
	var posts []models.Post
	if err := query.Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed posts"})
		return
	}

	response := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		postResponse := post.Response()

		// Like durumunu kontrol et
		var isLiked bool
		for _, like := range post.Likes {
			if like.UserID == currentUser.ID {
				isLiked = true
				break
			}
		}
		postResponse["isLiked"] = isLiked

		response[i] = postResponse
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": response,
		"pagination": gin.H{
			"current": page,
			"limit":   limit,
			"total":   total,
			"pages":   (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func GetSuggestedPostsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Takip edilmeyen kullanıcıların popüler postlarını getir
	query := database.DB.Model(&models.Post{}).
		Preload("User").
		Preload("Products").
		Preload("Categories").
		Preload("Hashtags").
		Preload("Likes").
		// Takip edilmeyen kullanıcıların postlarını seç
		Where("posts.user_id NOT IN (?)",
			database.DB.Table("user_followers").
				Select("following_id").
				Where("follower_id = ?", currentUser.ID)).
		// Kendi postlarını hariç tut
		Where("posts.user_id != ?", currentUser.ID).
		// Son 7 günün popüler postları
		Where("posts.created_at >= ?", time.Now().AddDate(0, 0, -7)).
		// Popülerliğe göre sırala
		Order("(posts.view_count * 0.6 + (SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) * 0.4) DESC")

	// Toplam post sayısını al
	var total int64
	query.Count(&total)

	// Postları getir
	var posts []models.Post
	if err := query.Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggested posts"})
		return
	}

	response := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		postResponse := post.Response()

		// Like durumunu kontrol et
		var isLiked bool
		for _, like := range post.Likes {
			if like.UserID == currentUser.ID {
				isLiked = true
				break
			}
		}
		postResponse["isLiked"] = isLiked

		response[i] = postResponse
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": response,
		"pagination": gin.H{
			"current": page,
			"limit":   limit,
			"total":   total,
			"pages":   (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func SearchPostsByHashtagHandler(c *gin.Context) {
	tag := c.Param("tag")
	normalizedTag := models.NormalizeHashtag(tag)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	query := database.DB.Model(&models.Post{}).
		Preload("User").
		Preload("Products").
		Preload("Categories").
		Preload("Hashtags").
		Preload("Likes").
		Joins("JOIN post_hashtags ph ON ph.post_id = posts.id").
		Joins("JOIN hashtags h ON h.id = ph.hashtag_id").
		Where("h.name = ?", normalizedTag).
		Order("posts.created_at DESC")

	// Toplam post sayısını al
	var total int64
	query.Count(&total)

	// Postları getir
	var posts []models.Post
	if err := query.Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	response := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		postResponse := post.Response()

		// Like durumunu kontrol et
		var isLiked bool
		for _, like := range post.Likes {
			if like.UserID == currentUser.ID {
				isLiked = true
				break
			}
		}
		postResponse["isLiked"] = isLiked

		response[i] = postResponse
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":   response,
		"hashtag": normalizedTag,
		"pagination": gin.H{
			"current": page,
			"limit":   limit,
			"total":   total,
			"pages":   (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func GetTrendingHashtagsHandler(c *gin.Context) {
	type HashtagCount struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var trendingHashtags []HashtagCount

	if err := database.DB.Raw(`
        SELECT h.name, COUNT(DISTINCT ph.post_id) as count
        FROM hashtags h
        JOIN post_hashtags ph ON ph.hashtag_id = h.id
        JOIN posts p ON p.id = ph.post_id
        WHERE p.created_at >= NOW() - INTERVAL '7 days'
        GROUP BY h.name
        ORDER BY count DESC
        LIMIT 10
    `).Scan(&trendingHashtags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trending hashtags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trendingHashtags": trendingHashtags})
}

func UpdatePostHandler(c *gin.Context) {
	postID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Post'u bul
	var post models.Post
	if err := database.DB.Preload("Products").
		Preload("Categories").
		Preload("Hashtags").
		First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Post sahibi kontrolü
	if post.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own posts"})
		return
	}

	var input CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	// Description güncelle
	if input.Description != "" {
		post.Description = input.Description
	}

	// ImageURL güncelle
	if input.ImageURL != "" {
		post.ImageURL = input.ImageURL
	}

	// Kategorileri güncelle
	if len(input.CategoryIDs) > 0 {
		var categories []models.Category
		if err := tx.Find(&categories, input.CategoryIDs).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category IDs"})
			return
		}

		if err := tx.Model(&post).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update categories"})
			return
		}
	}

	// Ürünleri güncelle
	if len(input.Products) > 0 {
		if len(input.Products) > models.MaxProductsPerPost {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        "maximum 8 products can be added to a post",
				"currentCount": len(input.Products),
				"maxAllowed":   models.MaxProductsPerPost,
			})
			return
		}

		// Mevcut ürünleri temizle
		if err := tx.Model(&post).Association("Products").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear products"})
			return
		}

		// Yeni ürünleri ekle
		var newProducts []models.Product
		for _, p := range input.Products {
			product := models.Product{
				Name:        p.Name,
				Price:       p.Price,
				Link:        p.Link,
				Description: p.Description,
			}
			newProducts = append(newProducts, product)
		}

		if err := tx.Create(&newProducts).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create products"})
			return
		}

		if err := tx.Model(&post).Association("Products").Replace(newProducts); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update products"})
			return
		}
	}

	// Hashtag'leri güncelle
	if len(input.Hashtags) > 0 {
		// Mevcut hashtag'leri temizle
		if err := tx.Model(&post).Association("Hashtags").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear hashtags"})
			return
		}

		// Yeni hashtag'leri ekle
		for _, tagName := range input.Hashtags {
			normalizedTag := models.NormalizeHashtag(tagName)
			if normalizedTag == "" {
				continue
			}

			var hashtag models.Hashtag
			err := tx.Where("name = ?", normalizedTag).FirstOrCreate(&hashtag, models.Hashtag{
				Name: normalizedTag,
			}).Error
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process hashtags"})
				return
			}

			if err := tx.Model(&post).Association("Hashtags").Append(&hashtag); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add hashtag"})
				return
			}
		}
	}

	// Post'u kaydet
	if err := tx.Save(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	// İlişkili verileri yükle
	if err := tx.Preload("User").
		Preload("Products").
		Preload("Categories").
		Preload("Hashtags").
		First(&post, post.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated post"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    post.Response(),
	})
}
