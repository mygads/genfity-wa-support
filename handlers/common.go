package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func HomePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "genfity-wa-support",
		"status":  "running",
		"mode":    "subscription-gateway",
	})
}
