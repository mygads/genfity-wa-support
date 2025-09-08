package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"genfity-wa-support/database"
	"genfity-wa-support/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BulkCreateText creates a bulk text message campaign
func BulkCreateText(c *gin.Context) {
	// Get token from header
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.BulkMessageResponse{
			Code:    401,
			Success: false,
			Message: "Token is required",
		})
		return
	}

	// Validate token exists in transactional database
	var session models.WhatsappSession
	if err := database.GetTransactionalDB().Where("token = ?", token).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.BulkMessageResponse{
			Code:    401,
			Success: false,
			Message: "Invalid token",
		})
		return
	}

	// Parse request body
	var req models.BulkTextMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BulkMessageResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Parse schedule time
	scheduledAt, err := parseSendSync(req.SendSync)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BulkMessageResponse{
			Code:    400,
			Success: false,
			Message: "Invalid SendSync format: " + err.Error(),
		})
		return
	}

	// Convert phone numbers to JSONB
	phoneNumbers := make([]string, len(req.Phone))
	copy(phoneNumbers, req.Phone)
	phoneJSON, _ := json.Marshal(phoneNumbers)
	var phoneJSONB models.JSONB
	json.Unmarshal(phoneJSON, &phoneJSONB)

	// Determine status
	status := models.BulkMessageStatusPending
	if scheduledAt != nil {
		status = models.BulkMessageStatusScheduled
	}

	// Create bulk message record
	bulkMessage := models.BulkMessage{
		SessionID:       session.ID,
		MessageType:     models.BulkMessageTypeText,
		PhoneNumbers:    phoneJSONB,
		Body:            req.Body,
		SendSync:        req.SendSync,
		ScheduledAt:     scheduledAt,
		Status:          status,
		TotalRecipients: len(req.Phone),
	}

	db := database.GetTransactionalDB()
	if err := db.Create(&bulkMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.BulkMessageResponse{
			Code:    500,
			Success: false,
			Message: "Failed to create bulk message: " + err.Error(),
		})
		return
	}

	// Create individual message items
	if err := createBulkMessageItems(db, bulkMessage.ID, req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, models.BulkMessageResponse{
			Code:    500,
			Success: false,
			Message: "Failed to create message items: " + err.Error(),
		})
		return
	}

	// If immediate send, trigger processing
	if scheduledAt == nil {
		go processBulkMessage(bulkMessage.ID)
	}

	c.JSON(http.StatusOK, models.BulkMessageResponse{
		Code:    200,
		Success: true,
		Message: "Bulk text message created successfully",
		Data: &struct {
			BulkID          uint       `json:"bulk_id"`
			TotalRecipients int        `json:"total_recipients"`
			Status          string     `json:"status"`
			ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
		}{
			BulkID:          bulkMessage.ID,
			TotalRecipients: len(req.Phone),
			Status:          string(status),
			ScheduledAt:     scheduledAt,
		},
	})
}

// BulkCreateImage creates a bulk image message campaign
func BulkCreateImage(c *gin.Context) {
	// Get token from header
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.BulkMessageResponse{
			Code:    401,
			Success: false,
			Message: "Token is required",
		})
		return
	}

	// Validate token exists in transactional database
	var session models.WhatsappSession
	if err := database.GetTransactionalDB().Where("token = ?", token).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.BulkMessageResponse{
			Code:    401,
			Success: false,
			Message: "Invalid token",
		})
		return
	}

	// Parse request body
	var req models.BulkImageMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BulkMessageResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Parse schedule time
	scheduledAt, err := parseSendSync(req.SendSync)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BulkMessageResponse{
			Code:    400,
			Success: false,
			Message: "Invalid SendSync format: " + err.Error(),
		})
		return
	}

	// Convert phone numbers to JSONB
	phoneNumbers := make([]string, len(req.Phone))
	copy(phoneNumbers, req.Phone)
	phoneJSON, _ := json.Marshal(phoneNumbers)
	var phoneJSONB models.JSONB
	json.Unmarshal(phoneJSON, &phoneJSONB)

	// Determine status
	status := models.BulkMessageStatusPending
	if scheduledAt != nil {
		status = models.BulkMessageStatusScheduled
	}

	// Create bulk message record
	bulkMessage := models.BulkMessage{
		SessionID:       session.ID,
		MessageType:     models.BulkMessageTypeImage,
		PhoneNumbers:    phoneJSONB,
		Image:           req.Image,
		Caption:         req.Caption,
		SendSync:        req.SendSync,
		ScheduledAt:     scheduledAt,
		Status:          status,
		TotalRecipients: len(req.Phone),
	}

	db := database.GetTransactionalDB()
	if err := db.Create(&bulkMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.BulkMessageResponse{
			Code:    500,
			Success: false,
			Message: "Failed to create bulk message: " + err.Error(),
		})
		return
	}

	// Create individual message items
	if err := createBulkMessageItems(db, bulkMessage.ID, req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, models.BulkMessageResponse{
			Code:    500,
			Success: false,
			Message: "Failed to create message items: " + err.Error(),
		})
		return
	}

	// If immediate send, trigger processing
	if scheduledAt == nil {
		go processBulkMessage(bulkMessage.ID)
	}

	c.JSON(http.StatusOK, models.BulkMessageResponse{
		Code:    200,
		Success: true,
		Message: "Bulk image message created successfully",
		Data: &struct {
			BulkID          uint       `json:"bulk_id"`
			TotalRecipients int        `json:"total_recipients"`
			Status          string     `json:"status"`
			ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
		}{
			BulkID:          bulkMessage.ID,
			TotalRecipients: len(req.Phone),
			Status:          string(status),
			ScheduledAt:     scheduledAt,
		},
	})
}

// BulkMessageList returns list of bulk messages for the user
func BulkMessageList(c *gin.Context) {
	// Get token from header
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.BulkMessageListResponse{
			Code:    401,
			Success: false,
		})
		return
	}

	// Validate token exists in transactional database
	var session models.WhatsappSession
	if err := database.GetTransactionalDB().Where("token = ?", token).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.BulkMessageListResponse{
			Code:    401,
			Success: false,
		})
		return
	}

	// Get bulk messages for this session
	var bulkMessages []models.BulkMessage
	if err := database.GetTransactionalDB().Where("session_id = ?", session.ID).Order("created_at DESC").Find(&bulkMessages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.BulkMessageListResponse{
			Code:    500,
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, models.BulkMessageListResponse{
		Code:    200,
		Success: true,
		Data:    bulkMessages,
	})
}

// BulkMessageDetail returns detailed information about a bulk message
func BulkMessageDetail(c *gin.Context) {
	// Get token from header
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.BulkMessageDetailResponse{
			Code:    401,
			Success: false,
		})
		return
	}

	// Get bulk ID from URL parameter
	bulkIDStr := c.Param("id")
	bulkID, err := strconv.ParseUint(bulkIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BulkMessageDetailResponse{
			Code:    400,
			Success: false,
		})
		return
	}

	// Validate token exists in transactional database
	var session models.WhatsappSession
	if err := database.GetTransactionalDB().Where("token = ?", token).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.BulkMessageDetailResponse{
			Code:    401,
			Success: false,
		})
		return
	}

	// Get bulk message
	var bulkMessage models.BulkMessage
	if err := database.GetTransactionalDB().Where("id = ? AND session_id = ?", bulkID, session.ID).First(&bulkMessage).Error; err != nil {
		c.JSON(http.StatusNotFound, models.BulkMessageDetailResponse{
			Code:    404,
			Success: false,
		})
		return
	}

	// Get message items
	var items []models.BulkMessageItem
	if err := database.GetTransactionalDB().Where("bulk_message_id = ?", bulkID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.BulkMessageDetailResponse{
			Code:    500,
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, models.BulkMessageDetailResponse{
		Code:    200,
		Success: true,
		Data: &struct {
			BulkMessage *models.BulkMessage      `json:"bulk_message"`
			Items       []models.BulkMessageItem `json:"items"`
		}{
			BulkMessage: &bulkMessage,
			Items:       items,
		},
	})
}

// Helper functions

// parseSendSync parses the SendSync field to determine if it's immediate or scheduled
func parseSendSync(sendSync string) (*time.Time, error) {
	sendSync = strings.TrimSpace(strings.ToLower(sendSync))

	if sendSync == "now" || sendSync == "sekarang" {
		return nil, nil // Immediate send
	}

	// Try to parse as datetime in various formats
	formats := []string{
		"2006-01-02 15:04:05",       // 2023-12-25 14:30:00
		"2006-01-02T15:04:05",       // 2023-12-25T14:30:00
		"2006-01-02T15:04:05Z",      // 2023-12-25T14:30:00Z
		"2006-01-02T15:04:05-07:00", // 2023-12-25T14:30:00+07:00
		"2006-01-02 15:04",          // 2023-12-25 14:30
		"02/01/2006 15:04",          // 25/12/2023 14:30
	}

	for _, format := range formats {
		if t, err := time.Parse(format, sendSync); err == nil {
			// Ensure the scheduled time is in the future
			if t.Before(time.Now()) {
				return nil, errors.New("scheduled time must be in the future")
			}
			return &t, nil
		}
	}

	return nil, errors.New("invalid datetime format. Use 'now' for immediate or 'YYYY-MM-DD HH:MM:SS' for scheduled")
}

// createBulkMessageItems creates individual message items for each phone number
func createBulkMessageItems(db *gorm.DB, bulkMessageID uint, phoneNumbers []string) error {
	items := make([]models.BulkMessageItem, len(phoneNumbers))
	for i, phone := range phoneNumbers {
		items[i] = models.BulkMessageItem{
			BulkMessageID: bulkMessageID,
			PhoneNumber:   phone,
			Status:        "pending",
		}
	}

	return db.Create(&items).Error
}

// processBulkMessage processes a bulk message (sends to WhatsApp server)
// This function will be called immediately for "now" messages or by cron for scheduled messages
func processBulkMessage(bulkMessageID uint) {
	// TODO: Implement actual message sending logic
	// This is a placeholder that will be implemented to call the WhatsApp server

	db := database.GetTransactionalDB()

	// Update status to processing
	db.Model(&models.BulkMessage{}).Where("id = ?", bulkMessageID).Updates(map[string]interface{}{
		"status":       models.BulkMessageStatusProcessing,
		"processed_at": time.Now(),
	})

	// TODO: Get bulk message details and send to WhatsApp server
	// TODO: Update individual item statuses based on sending results
	// TODO: Update bulk message final status when complete
}

// BulkMessageCronJob processes scheduled bulk messages (called by cron every minute)
func BulkMessageCronJob(c *gin.Context) {
	db := database.GetTransactionalDB()

	// Find scheduled messages that are ready to be sent
	var scheduledMessages []models.BulkMessage
	now := time.Now()

	err := db.Where("status = ? AND scheduled_at <= ?", models.BulkMessageStatusScheduled, now).Find(&scheduledMessages).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to fetch scheduled messages",
		})
		return
	}

	processedCount := 0
	for _, msg := range scheduledMessages {
		go processBulkMessage(msg.ID)
		processedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"message": "Cron job completed",
		"data": gin.H{
			"processed_count": processedCount,
			"checked_at":      now,
		},
	})
}
