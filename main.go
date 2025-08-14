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

	// Routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", handlers.HealthCheck)

		// Data retrieval routes
		api.GET("/messages", handlers.GetMessages)
		api.GET("/events", handlers.GetWebhookEvents)
		api.GET("/users", handlers.GetUsers)
		api.GET("/users/:user_token/stats", handlers.GetUserStats)

		// WhatsApp session management
		api.GET("/sessions", handlers.GetSessions)
		api.GET("/sessions/:user_token/qr", handlers.GetSessionQR)
		api.GET("/sessions/sync", handlers.SyncSessionStatus) // Trigger sync with WA server

		// Message status tracking
		api.GET("/message-statuses", handlers.GetMessageStatuses)

		// Chat presence (typing) tracking
		api.GET("/chat-presences", handlers.GetChatPresences)
	}

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
