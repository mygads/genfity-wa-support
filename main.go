package main

import (
	"log"
	"os"

	"genfity-wa-support/database"
	"genfity-wa-support/handlers"

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
	database.StartSubscriptionExpiryCron()

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, token, x-api-key, x-internal-api-key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Home page
	router.GET("/", handlers.HomePage)

	// Health check - support both GET and HEAD methods
	router.GET("/health", handlers.HealthCheck)
	router.HEAD("/health", handlers.HealthCheck)

	internal := router.Group("/internal")
	internal.Use(handlers.InternalAPIKeyMiddleware())
	{
		internal.GET("/me", handlers.InternalMe)
		internal.GET("/users", handlers.InternalListUsers)
		internal.POST("/users", handlers.InternalUpsertUser)
		internal.PUT("/users/:user_id", handlers.InternalUpdateUser)
		internal.GET("/users/:user_id/apikey", handlers.InternalGetUserAPIKey)
		internal.POST("/users/:user_id/apikey/rotate", handlers.InternalRotateUserAPIKey)
	}

	public := router.Group("/v1")
	public.Use(handlers.PublicRateLimiter(), handlers.CustomerAPIKeyMiddleware())
	wa := router.Group("/wa")
	wa.Use(handlers.PublicRateLimiter())
	{
		public.GET("/me", handlers.GetCurrentUser)
		public.GET("/sessions", handlers.ListSessions)
		public.POST("/sessions", handlers.CreateSession)
		public.PUT("/sessions/:session_id", handlers.UpdateSession)
		public.DELETE("/sessions/:session_id", handlers.DeleteSession)
		public.GET("/sessions/:session_id/settings", handlers.GetSessionSettings)
		public.PUT("/sessions/:session_id/settings", handlers.UpdateSessionSettings)
		public.GET("/sessions/:session_id/contacts", handlers.ListSessionContacts)
		public.POST("/sessions/:session_id/contacts/sync", handlers.SyncSessionContacts)

		wa.Any("/*path", handlers.WhatsAppGateway)
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
