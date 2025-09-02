package main

import (
	"log"
	"os"

	"genfity-chat-ai/database"
	"genfity-chat-ai/handlers"

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

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, token")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// WhatsApp Gateway routes - All WA API requests go through this gateway with /wa prefix
	// Admin routes bypass subscription checks, other routes validate subscription
	wa := router.Group("/wa")
	{
		// Admin endpoints (bypass all validation)
		wa.Any("/admin", handlers.WhatsAppGateway)       // Handle exact /wa/admin
		wa.Any("/admin/*path", handlers.WhatsAppGateway) // Handle /wa/admin/...

		// Session endpoints (validate subscription + session limits)
		wa.Any("/session", handlers.WhatsAppGateway)       // Handle exact /wa/session
		wa.Any("/session/*path", handlers.WhatsAppGateway) // Handle /wa/session/...

		// Webhook endpoints (validate subscription)
		wa.Any("/webhook", handlers.WhatsAppGateway)       // Handle exact /wa/webhook
		wa.Any("/webhook/*path", handlers.WhatsAppGateway) // Handle /wa/webhook/...

		// Chat endpoints (validate subscription + message tracking)
		wa.Any("/chat", handlers.WhatsAppGateway)       // Handle exact /wa/chat
		wa.Any("/chat/*path", handlers.WhatsAppGateway) // Handle /wa/chat/...

		// User endpoints (validate subscription)
		wa.Any("/user", handlers.WhatsAppGateway)       // Handle exact /wa/user
		wa.Any("/user/*path", handlers.WhatsAppGateway) // Handle /wa/user/...

		// Group endpoints (validate subscription)
		wa.Any("/group", handlers.WhatsAppGateway)       // Handle exact /wa/group
		wa.Any("/group/*path", handlers.WhatsAppGateway) // Handle /wa/group/...

		// Newsletter endpoints (validate subscription)
		wa.Any("/newsletter", handlers.WhatsAppGateway)       // Handle exact /wa/newsletter
		wa.Any("/newsletter/*path", handlers.WhatsAppGateway) // Handle /wa/newsletter/...
	}

	// Original webhook routes for receiving events from WA server (separate from gateway)
	webhooks := router.Group("/webhook")
	{
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
	log.Printf("Gateway mode: %s", os.Getenv("GATEWAY_MODE"))
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
