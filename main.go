package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sefazor/comfyn/configs"
	"github.com/sefazor/comfyn/internal/affiliate/link"
	"github.com/sefazor/comfyn/internal/auth"
	"github.com/sefazor/comfyn/internal/notification"
	"github.com/sefazor/comfyn/internal/post"
	"github.com/sefazor/comfyn/internal/user"
	"github.com/sefazor/comfyn/pkg/database"
	"github.com/sefazor/comfyn/pkg/middleware"
)

func main() {
	configs.LoadEnv()
	database.InitDB()

	r := gin.Default()

	// Public routes
	r.POST("/api/auth/register", auth.RegisterHandler)
	r.POST("/api/auth/login", auth.LoginHandler)
	r.GET("/go/:tracking_id", link.RedirectHandler)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// User routes
		protected.GET("/users/me", user.GetProfileHandler)
		protected.GET("/users/:id", user.GetUserProfileHandler)
		protected.PUT("/users/profile", user.UpdateProfileHandler)
		protected.PUT("/users/security", user.UpdateSecurityHandler)
		protected.POST("/users/:id/follow", user.FollowUserHandler)
		protected.GET("/users/search", user.SearchUsersHandler)

		// Post routes
		protected.POST("/posts", post.CreatePostHandler)
		protected.GET("/posts", post.ListPostsHandler)
		protected.GET("/posts/:id", post.GetPostHandler)
		protected.PUT("/posts/:id", post.UpdatePostHandler)
		protected.DELETE("/posts/:id", post.DeletePostHandler)
		protected.POST("/posts/:id/like", post.LikePostHandler)
		protected.POST("/posts/:id/comment", post.CreateCommentHandler)
		protected.POST("/posts/:id/view", post.IncrementViewHandler)

		// Feed routes
		protected.GET("/feed", post.GetPersonalFeedHandler)
		protected.GET("/feed/suggested", post.GetSuggestedPostsHandler)

		// Hashtag routes
		protected.GET("/posts/hashtag/:tag", post.SearchPostsByHashtagHandler)
		protected.GET("/hashtags/trending", post.GetTrendingHashtagsHandler)

		// Notification routes
		protected.GET("/notifications", notification.GetNotificationsHandler)
		protected.PUT("/notifications/:id/read", notification.MarkNotificationReadHandler)
		protected.PUT("/notifications/preferences", notification.UpdateNotificationPreferencesHandler)
		protected.GET("/analytics/links", link.GetLinkAnalyticsHandler)

		protected.GET("/analytics/clicks", post.GetClickStatsHandler)

	}

	log.Printf("Server starting on :8080")
	r.Run(":8080")
}
