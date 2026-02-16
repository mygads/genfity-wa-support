package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"genfity-wa-support/database"
	"genfity-wa-support/models"

	"github.com/gin-gonic/gin"
)

type upsertUserRequest struct {
	UserID      string    `json:"user_id" binding:"required"`
	Source      string    `json:"source" binding:"required"`
	ExpiresAt   time.Time `json:"expires_at" binding:"required"`
	MaxSessions int       `json:"max_sessions"`
	MaxMessages int       `json:"max_messages"`
	Provider    string    `json:"provider"`
	CreatedBy   string    `json:"created_by"`
}

type internalUserListItem struct {
	UserID        string    `json:"user_id"`
	SourceService string    `json:"source_service"`
	Status        string    `json:"status"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Subscription struct {
		Provider    string                    `json:"provider"`
		Status      models.SubscriptionStatus `json:"status"`
		ExpiresAt   time.Time                 `json:"expires_at"`
		MaxSessions int                       `json:"max_sessions"`
		MaxMessages int                       `json:"max_messages"`
	} `json:"subscription"`

	SessionCount int64 `json:"session_count"`
}

func InternalMe(c *gin.Context) {
	source, scoped := getInternalSourceScope(c)
	mode := "global"
	if scoped {
		mode = "scoped"
	}

	c.JSON(http.StatusOK, gin.H{
		"auth": gin.H{
			"mode":           mode,
			"source_service": source,
		},
	})
}

func InternalListUsers(c *gin.Context) {
	db := database.GetDB()

	requestedSource := strings.TrimSpace(c.Query("source"))
	source, scoped := getInternalSourceScope(c)
	if !scoped {
		source = requestedSource
	}
	provider := strings.TrimSpace(c.DefaultQuery("provider", "genfity-wa"))
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	limit := parsePositiveInt(c.DefaultQuery("limit", "20"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	query := db.Model(&models.ServiceUser{})
	if source != "" {
		query = query.Where("source_service = ?", source)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count users"})
		return
	}

	var users []models.ServiceUser
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list users"})
		return
	}

	items := make([]internalUserListItem, 0, len(users))
	for _, user := range users {
		item := internalUserListItem{
			UserID:        user.ID,
			SourceService: user.SourceService,
			Status:        user.Status,
			CreatedBy:     user.CreatedBy,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}

		var sub models.UserSubscription
		if err := db.Where("user_id = ? AND provider = ?", user.ID, provider).Order("updated_at desc").First(&sub).Error; err == nil {
			item.Subscription.Provider = sub.Provider
			item.Subscription.Status = sub.Status
			item.Subscription.ExpiresAt = sub.ExpiresAt
			item.Subscription.MaxSessions = sub.MaxSessions
			item.Subscription.MaxMessages = sub.MaxMessages
		}

		_ = db.Model(&models.WhatsAppSession{}).Where("user_id = ?", user.ID).Count(&item.SessionCount).Error
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"meta": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"provider": provider,
			"source":   source,
		},
	})
}

func parsePositiveInt(raw string, fallback int) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func InternalUpsertUser(c *gin.Context) {
	var req upsertUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if source, scoped := getInternalSourceScope(c); scoped {
		if req.Source != source {
			c.JSON(http.StatusForbidden, gin.H{"message": "key only allowed for its own source"})
			return
		}
	}

	if req.MaxSessions <= 0 {
		req.MaxSessions = 1
	}
	if req.Provider == "" {
		req.Provider = "genfity-wa"
	}

	db := database.GetDB()
	var user models.ServiceUser
	res := db.Where("id = ?", req.UserID).First(&user)
	plainAPIKey := ""
	if res.Error != nil {
		raw, hashed, err := generateAPIKey("gwa")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate api key"})
			return
		}
		plainAPIKey = raw
		user = models.ServiceUser{
			ID:             req.UserID,
			SourceService:  req.Source,
			CustomerAPIKey: hashed,
			CreatedBy:      req.CreatedBy,
			Status:         "active",
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create user"})
			return
		}
	} else {
		user.SourceService = req.Source
		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user"})
			return
		}
	}

	var sub models.UserSubscription
	if err := db.Where("user_id = ? AND provider = ?", req.UserID, req.Provider).First(&sub).Error; err != nil {
		sub = models.UserSubscription{
			UserID:      req.UserID,
			Provider:    req.Provider,
			ExpiresAt:   req.ExpiresAt,
			MaxSessions: req.MaxSessions,
			MaxMessages: req.MaxMessages,
			Status:      models.SubscriptionActive,
		}
		if err := db.Create(&sub).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create subscription"})
			return
		}
	} else {
		sub.ExpiresAt = req.ExpiresAt
		sub.MaxSessions = req.MaxSessions
		sub.MaxMessages = req.MaxMessages
		sub.Status = models.SubscriptionActive
		if err := db.Save(&sub).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update subscription"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": req.UserID,
		"api_key": plainAPIKey,
		"note":    "api_key hanya tampil saat user baru dibuat",
	})
}

func InternalUpdateUser(c *gin.Context) {
	userID := c.Param("user_id")
	var req upsertUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if source, scoped := getInternalSourceScope(c); scoped {
		if req.Source != source {
			c.JSON(http.StatusForbidden, gin.H{"message": "key only allowed for its own source"})
			return
		}
		if !internalCanAccessUser(userID, source) {
			c.JSON(http.StatusForbidden, gin.H{"message": "user does not belong to this source"})
			return
		}
	}

	if req.Provider == "" {
		req.Provider = "genfity-wa"
	}
	if req.MaxSessions <= 0 {
		req.MaxSessions = 1
	}

	db := database.GetDB()
	if err := db.Model(&models.ServiceUser{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"source_service": req.Source,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user"})
		return
	}

	if err := db.Model(&models.UserSubscription{}).
		Where("user_id = ? AND provider = ?", userID, req.Provider).
		Updates(map[string]interface{}{
			"expires_at":   req.ExpiresAt,
			"max_sessions": req.MaxSessions,
			"max_messages": req.MaxMessages,
			"status":       models.SubscriptionActive,
			"updated_at":   time.Now(),
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func InternalGetUserAPIKey(c *gin.Context) {
	userID := c.Param("user_id")
	if source, scoped := getInternalSourceScope(c); scoped {
		if !internalCanAccessUser(userID, source) {
			c.JSON(http.StatusForbidden, gin.H{"message": "user does not belong to this source"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"note":    "api key tidak dapat dibaca kembali karena disimpan dalam bentuk hash",
	})
}

func InternalRotateUserAPIKey(c *gin.Context) {
	userID := c.Param("user_id")
	if source, scoped := getInternalSourceScope(c); scoped {
		if !internalCanAccessUser(userID, source) {
			c.JSON(http.StatusForbidden, gin.H{"message": "user does not belong to this source"})
			return
		}
	}

	raw, hashed, err := generateAPIKey("gwa")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate key"})
		return
	}
	if err := database.GetDB().Model(&models.ServiceUser{}).Where("id = ?", userID).Update("customer_api_key", hashed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to rotate key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "api_key": raw})
}

func getInternalSourceScope(c *gin.Context) (string, bool) {
	sourceAny, hasSource := c.Get("internal_source")
	scopedAny, hasScoped := c.Get("internal_scoped")
	if !hasSource || !hasScoped {
		return "", false
	}
	source, _ := sourceAny.(string)
	scoped, _ := scopedAny.(bool)
	return source, scoped
}

func internalCanAccessUser(userID, source string) bool {
	var count int64
	database.GetDB().Model(&models.ServiceUser{}).
		Where("id = ? AND source_service = ?", userID, source).
		Count(&count)
	return count > 0
}
