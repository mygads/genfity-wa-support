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
		wa.Any("/admin/*path", handlers.WhatsAppGateway)

		// Session endpoints (validate subscription + session limits)
		wa.Any("/session/*path", handlers.WhatsAppGateway)

		// Webhook endpoints (validate subscription)
		wa.Any("/webhook/*path", handlers.WhatsAppGateway)

		// Chat endpoints (validate subscription + message tracking)
		wa.Any("/chat/*path", handlers.WhatsAppGateway)

		// User endpoints (validate subscription)
		wa.Any("/user/*path", handlers.WhatsAppGateway)

		// Group endpoints (validate subscription)
		wa.Any("/group/*path", handlers.WhatsAppGateway)

		// Newsletter endpoints (validate subscription)
		wa.Any("/newsletter/*path", handlers.WhatsAppGateway)
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
