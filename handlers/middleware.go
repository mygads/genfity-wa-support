package handlers

import (
	"net/http"

	"genfity-chat-ai/database"
	"genfity-chat-ai/models"

	"github.com/gin-gonic/gin"
)

// SessionMiddleware validates token and sets session_id in context
func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		token := c.GetHeader("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"success": false,
				"message": "Token is required",
			})
			c.Abort()
			return
		}

		// Validate token exists in transactional database
		var session models.WhatsappSession
		if err := database.GetTransactionalDB().Where("token = ?", token).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"success": false,
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// Set session_id in context for use in handlers
		c.Set("session_id", session.SessionID)
		c.Next()
	}
}
