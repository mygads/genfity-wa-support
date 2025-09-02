package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"genfity-chat-ai/database"
	"genfity-chat-ai/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Global endpoints that don't require token validation
var globalEndpoints = []string{
	"/webhook/events", // WhatsApp calls this endpoint
	"/health",
}

// isGlobalEndpoint checks if the given path is a global endpoint
func isGlobalEndpoint(path string) bool {
	for _, endpoint := range globalEndpoints {
		if path == endpoint || strings.HasPrefix(path, endpoint+"/") {
			return true
		}
	}
	return false
}

// WhatsAppGateway handles all WhatsApp API requests with /wa prefix
func WhatsAppGateway(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	log.Printf("DEBUG: Gateway received request - Method: %s, Path: %s", method, path)

	// Remove /wa prefix to get actual WA server path
	actualPath := strings.TrimPrefix(path, "/wa")
	log.Printf("DEBUG: Actual path after prefix removal: %s", actualPath)

	// Admin routes bypass all validation
	if strings.HasPrefix(actualPath, "/admin") {
		// log.Printf("DEBUG: Admin route detected, bypassing validation")
		proxyToWAServer(c, actualPath)
		return
	}

	// Global endpoints that don't require token validation
	if isGlobalEndpoint(actualPath) {
		log.Printf("DEBUG: Global endpoint detected, bypassing token validation")
		proxyToWAServer(c, actualPath)
		return
	}

	// For non-admin routes, validate token and subscription
	token := getTokenFromRequest(c)
	if token == "" {
		log.Printf("DEBUG: No token provided")
		c.JSON(http.StatusUnauthorized, models.GatewayResponse{
			Status:  http.StatusUnauthorized,
			Message: "Token required",
		})
		return
	}

	log.Printf("DEBUG: Token received: %s", token)

	// Validate token and subscription
	userID, err := validateTokenAndSubscription(token, actualPath)
	if err != nil {
		log.Printf("Validation failed: %v", err)
		c.JSON(http.StatusForbidden, models.GatewayResponse{
			Status:  http.StatusForbidden,
			Message: err.Error(),
		})
		return
	}

	// If this is a session connect request, check session limits
	if actualPath == "/session/connect" && method == "POST" {
		if err := checkSessionLimits(userID); err != nil {
			c.JSON(http.StatusForbidden, models.GatewayResponse{
				Status:  http.StatusForbidden,
				Message: err.Error(),
			})
			return
		}
	}

	// Proxy to WhatsApp server
	statusCode := proxyToWAServer(c, actualPath)

	// Track successful message sends
	if isMessageEndpoint(actualPath) && method == "POST" && statusCode >= 200 && statusCode < 300 {
		go trackMessageStats(userID, token, actualPath, c)
	}
}

// getTokenFromRequest extracts token from Authorization header or token header
func getTokenFromRequest(c *gin.Context) string {
	// Check token header first (as per API documentation)
	token := c.GetHeader("token")
	if token != "" {
		return token
	}

	// Check Authorization header as fallback
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Remove "Bearer " prefix if present
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}

	return ""
}

// validateTokenAndSubscription validates token and checks subscription status
func validateTokenAndSubscription(token, path string) (string, error) {
	// Find session by token in WhatsAppSession table
	var session models.WhatsappSession
	if err := database.TransactionalDB.Where("token = ?", token).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("invalid token")
		}
		return "", fmt.Errorf("database error: %v", err)
	}

	// Check if session has associated user
	if session.UserID == nil {
		return "", fmt.Errorf("session not associated with any user")
	}

	// Get user's active subscription from ServicesWhatsappCustomers
	var subscription models.ServicesWhatsappCustomers
	err := database.TransactionalDB.
		Where("\"customerId\" = ? AND status = ?", *session.UserID, "active").
		First(&subscription).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("no active subscription found")
		}
		return "", fmt.Errorf("subscription check failed: %v", err)
	}

	// Check if subscription is expired and auto-update status
	if time.Now().After(subscription.ExpiredAt) {
		subscription.Status = "expired"
		database.TransactionalDB.Save(&subscription)
		return "", fmt.Errorf("subscription expired on %s", subscription.ExpiredAt.Format("2006-01-02"))
	}

	return *session.UserID, nil
}

// checkSessionLimits validates session limits for connect requests
func checkSessionLimits(userID string) error {
	// Get user's subscription
	var subscription models.ServicesWhatsappCustomers
	err := database.TransactionalDB.
		Where("\"customerId\" = ? AND status = ?", userID, "active").
		First(&subscription).Error
	if err != nil {
		return fmt.Errorf("no active subscription found")
	}

	// Get package info
	var packageInfo models.WhatsappApiPackage
	err = database.TransactionalDB.Where("id = ?", subscription.PackageID).First(&packageInfo).Error
	if err != nil {
		return fmt.Errorf("package not found")
	}

	// Count current active sessions for this user
	var currentSessions int64
	database.TransactionalDB.Model(&models.WhatsappSession{}).
		Where("\"userId\" = ? AND connected = ?", userID, true).
		Count(&currentSessions)

	// Check if adding new session would exceed limit
	if int(currentSessions) >= packageInfo.MaxSession {
		return fmt.Errorf("session limit exceeded. Maximum allowed: %d, current: %d",
			packageInfo.MaxSession, currentSessions)
	}

	return nil
}

// proxyToWAServer forwards the request to WhatsApp server without modification
func proxyToWAServer(c *gin.Context, targetPath string) int {
	waServerURL := os.Getenv("WA_SERVER_URL")
	if waServerURL == "" {
		c.JSON(http.StatusInternalServerError, models.GatewayResponse{
			Status:  http.StatusInternalServerError,
			Message: "WhatsApp server URL not configured",
		})
		return http.StatusInternalServerError
	}

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create new request to WA server using the stripped path
	targetURL := waServerURL + targetPath
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	log.Printf("DEBUG: Proxying to URL: %s", targetURL)
	log.Printf("DEBUG: Request method: %s", c.Request.Method)
	log.Printf("DEBUG: Request body: %s", string(bodyBytes))

	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.GatewayResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create request",
		})
		return http.StatusInternalServerError
	}

	// Copy all headers exactly as received
	for name, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Execute request to WA server
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.GatewayResponse{
			Status:  http.StatusBadGateway,
			Message: "Failed to reach WhatsApp server",
		})
		return http.StatusBadGateway
	}
	defer resp.Body.Close()

	// Read response from WA server
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.GatewayResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to read response",
		})
		return http.StatusInternalServerError
	}

	// Copy response headers exactly
	for name, values := range resp.Header {
		for _, value := range values {
			c.Header(name, value)
		}
	}

	// Return response with same status code and body as WA server
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)

	return resp.StatusCode
}

// isMessageEndpoint checks if the endpoint is a message sending endpoint
func isMessageEndpoint(path string) bool {
	messageEndpoints := []string{
		"/chat/send/text",
		"/chat/send/image",
		"/chat/send/audio",
		"/chat/send/document",
		"/chat/send/video",
		"/chat/send/sticker",
		"/chat/send/location",
		"/chat/send/contact",
		"/chat/send/template",
		"/chat/send/edit",
	}

	for _, endpoint := range messageEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

// trackMessageStats tracks successful message sends in WhatsAppMessageStats table
func trackMessageStats(userID, token, path string, c *gin.Context) {
	// Find session by token to get sessionId
	var session models.WhatsappSession
	if err := database.TransactionalDB.Where("token = ?", token).First(&session).Error; err != nil {
		log.Printf("Failed to find session for token: %v", err)
		return
	}

	// Try to get existing stats record or create new one
	var stats models.WhatsAppMessageStats
	err := database.TransactionalDB.Where("\"userId\" = ? AND \"sessionId\" = ?", userID, session.SessionID).First(&stats).Error

	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// Create new stats record
		stats = models.WhatsAppMessageStats{
			ID:                  uuid.New().String(),
			UserID:              userID,
			SessionID:           session.SessionID,
			TotalMessagesSent:   1,
			TotalMessagesFailed: 0,
			LastMessageSentAt:   &now,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		if err := database.TransactionalDB.Create(&stats).Error; err != nil {
			log.Printf("Failed to create message stats: %v", err)
		}
	} else if err == nil {
		// Update existing stats record
		stats.TotalMessagesSent++
		stats.LastMessageSentAt = &now
		stats.UpdatedAt = now

		if err := database.TransactionalDB.Save(&stats).Error; err != nil {
			log.Printf("Failed to update message stats: %v", err)
		}
	} else {
		log.Printf("Failed to query message stats: %v", err)
	}
}
