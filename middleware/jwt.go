package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"genfity-wa-support/database"
	"genfity-wa-support/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userID string, email string, role string, sessionID string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", fmt.Errorf("JWT_SECRET environment variable not set")
	}

	claims := &models.JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT parses and validates a JWT token
func ValidateJWT(tokenString string) (*models.JWTClaims, error) {
	if len(jwtSecret) == 0 {
		return nil, fmt.Errorf("JWT_SECRET environment variable not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// JWTMiddleware validates JWT tokens and sets user context
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check for Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Simple token validation: match with UserSession table
		var userSession models.UserSession
		err := database.TransactionalDB.Where(`token = ? AND "expiresAt" > ? AND "isActive" = ?`,
			tokenString, time.Now(), true).First(&userSession).Error
		if err != nil {
			// Debug logging
			fmt.Printf("Token lookup failed: %v\n", err)
			fmt.Printf("Looking for token: %s\n", tokenString[:50]+"...")
			fmt.Printf("Current time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Session not found or expired",
			})
			c.Abort()
			return
		}

		// Get user details from user_id
		var user models.User
		err = database.TransactionalDB.Where("id = ?", userSession.UserID).First(&user).Error
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", userSession.UserID)
		c.Set("user", user)
		c.Set("session", userSession)

		c.Next()
	}
}

// CreateUserSession creates a new JWT session for a user
func CreateUserSession(userID string, email string, role string) (string, error) {
	// Generate session ID
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())

	// Generate JWT token
	token, err := GenerateJWT(userID, email, role, sessionID)
	if err != nil {
		return "", err
	}

	// Create session record
	session := models.UserSession{
		ID:        sessionID,
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		LastUsed:  time.Now(),
	}

	err = database.TransactionalDB.Create(&session).Error
	if err != nil {
		return "", err
	}

	return token, nil
}

// RevokeUserSession revokes a JWT session
func RevokeUserSession(token string) error {
	return database.TransactionalDB.Where("token = ?", token).Delete(&models.UserSession{}).Error
}
