package main

import (
	"log"
	"os"

	"genfity-event-api/database"
	"genfity-event-api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	database.InitDatabase()

	// Setup Gin router
	router := gin.Default()

	// Add middleware for CORS if needed
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Admin routes - requires bearer token authentication
	admin := router.Group("/admin")
	admin.Use(handlers.AdminAuthMiddleware())
	{
		// User management
		admin.GET("/users", handlers.AdminGetUsers)
		admin.GET("/users/:user_token", handlers.AdminGetUser)
		admin.PUT("/users/:user_token/update", handlers.AdminUpdateUser)

		// Session management
		admin.GET("/sessions", handlers.AdminGetSessions)
		admin.GET("/sessions/:user_token", handlers.AdminGetUserSessions)

		// Event monitoring
		admin.GET("/event", handlers.AdminGetEvents)
	}

	// User routes - also requires bearer token authentication
	user := router.Group("/user")
	user.Use(handlers.AdminAuthMiddleware())
	{
		// Chat management
		user.GET("/chat/:user_token", handlers.GetUserChats)
		user.GET("/message/:user_token/:chat_id", handlers.GetChatMessages)
		user.GET("/event/:user_token", handlers.GetUserEvents)
	}

	// Public routes (no authentication required)
	router.GET("/health", handlers.HealthCheck)
	router.GET("/sessions/sync", handlers.SyncSessionStatus)

	// Webhook routes - for receiving events from various platforms
	webhooks := router.Group("/webhook")
	{
		// WhatsApp webhook routes
		wa := webhooks.Group("/wa")
		{
			wa.GET("", handlers.VerifyWebhook)
			wa.POST("", handlers.HandleWhatsAppWebhook)
		}
	}

	// Get port from environment or default to 8070
	port := os.Getenv("PORT")
	if port == "" {
		port = "8070"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
