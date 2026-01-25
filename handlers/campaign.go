package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"genfity-wa-support/database"
	"genfity-wa-support/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateCampaign creates a new campaign template
func CreateCampaign(c *gin.Context) {
	var req models.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.CampaignResponse{
			Code:    400,
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.CampaignResponse{
			Code:    401,
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	// Validate based on campaign type
	if req.Type == models.CampaignTypeText && req.MessageBody == "" {
		c.JSON(http.StatusBadRequest, models.CampaignResponse{
			Code:    400,
			Success: false,
			Message: "Message body is required for text campaigns",
		})
		return
	}

	if req.Type == models.CampaignTypeImage && req.ImageURL == "" && req.ImageBase64 == "" {
		c.JSON(http.StatusBadRequest, models.CampaignResponse{
			Code:    400,
			Success: false,
			Message: "Image URL or base64 is required for image campaigns",
		})
		return
	}

	// Create campaign
	campaign := models.Campaign{
		UserID:      userID.(string),
		Name:        req.Name,
		Type:        req.Type,
		Status:      models.CampaignStatusActive,
		MessageBody: req.MessageBody,
		ImageURL:    req.ImageURL,
		ImageBase64: req.ImageBase64,
		Caption:     req.Caption,
	}

	if err := database.TransactionalDB.Create(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.CampaignResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("Failed to create campaign: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, models.CampaignResponse{
		Code:    201,
		Success: true,
		Message: "Campaign created successfully",
		Data:    &campaign,
	})
}

// GetCampaigns retrieves all campaigns for a session
func GetCampaigns(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.CampaignListResponse{
			Code:    401,
			Success: false,
		})
		return
	}

	var campaigns []models.Campaign
	if err := database.TransactionalDB.Where("user_id = ?", userID).Find(&campaigns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.CampaignListResponse{
			Code:    500,
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, models.CampaignListResponse{
		Code:    200,
		Success: true,
		Data:    campaigns,
	})
}

// GetCampaign retrieves a specific campaign by ID
func GetCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.CampaignResponse{
			Code:    401,
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	campaignID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.CampaignResponse{
			Code:    400,
			Success: false,
			Message: "Invalid campaign ID",
		})
		return
	}

	var campaign models.Campaign
	if err := database.TransactionalDB.Where("id = ? AND user_id = ?", campaignID, userID).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, models.CampaignResponse{
			Code:    404,
			Success: false,
			Message: "Campaign not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.CampaignResponse{
		Code:    200,
		Success: true,
		Data:    &campaign,
	})
}

// UpdateCampaign updates an existing campaign
func UpdateCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.CampaignResponse{
			Code:    401,
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	campaignID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.CampaignResponse{
			Code:    400,
			Success: false,
			Message: "Invalid campaign ID",
		})
		return
	}

	var req models.UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.CampaignResponse{
			Code:    400,
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	var campaign models.Campaign
	if err := database.TransactionalDB.Where("id = ? AND user_id = ?", campaignID, userID).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, models.CampaignResponse{
			Code:    404,
			Success: false,
			Message: "Campaign not found",
		})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.MessageBody != "" {
		updates["message_body"] = req.MessageBody
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.ImageBase64 != "" {
		updates["image_base64"] = req.ImageBase64
	}
	if req.Caption != "" {
		updates["caption"] = req.Caption
	}

	if err := database.TransactionalDB.Model(&campaign).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.CampaignResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("Failed to update campaign: %v", err),
		})
		return
	}

	// Reload the updated campaign
	database.TransactionalDB.First(&campaign, campaignID)

	c.JSON(http.StatusOK, models.CampaignResponse{
		Code:    200,
		Success: true,
		Message: "Campaign updated successfully",
		Data:    &campaign,
	})
}

// DeleteCampaign deletes a campaign
func DeleteCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.CampaignResponse{
			Code:    401,
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	campaignID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.CampaignResponse{
			Code:    400,
			Success: false,
			Message: "Invalid campaign ID",
		})
		return
	}

	var campaign models.Campaign
	if err := database.TransactionalDB.Where("id = ? AND user_id = ?", campaignID, userID).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, models.CampaignResponse{
			Code:    404,
			Success: false,
			Message: "Campaign not found",
		})
		return
	}

	if err := database.TransactionalDB.Delete(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.CampaignResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("Failed to delete campaign: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.CampaignResponse{
		Code:    200,
		Success: true,
		Message: "Campaign deleted successfully",
	})
}

// CreateBulkCampaign creates a bulk campaign execution
func CreateBulkCampaign(c *gin.Context) {
	var req models.CreateBulkCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BulkCampaignResponse{
			Code:    400,
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.BulkCampaignResponse{
			Code:    401,
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	// Get campaign details
	var campaign models.Campaign
	if err := database.TransactionalDB.Where("id = ? AND user_id = ?", req.CampaignID, userID).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, models.BulkCampaignResponse{
			Code:    404,
			Success: false,
			Message: "Campaign not found",
		})
		return
	}

	// Parse scheduling with timezone
	scheduledAt, timezone, err := parseSendSyncWithTimezone(req.SendSync, req.Timezone)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BulkCampaignResponse{
			Code:    400,
			Success: false,
			Message: fmt.Sprintf("Invalid send_sync or timezone: %v", err),
		})
		return
	}

	// Determine status based on scheduling
	var status models.BulkCampaignStatus
	if scheduledAt == nil {
		status = models.BulkCampaignStatusPending // Immediate execution
	} else {
		status = models.BulkCampaignStatusScheduled // Scheduled for later
	}

	// Create bulk campaign with copied campaign data
	bulkCampaign := models.BulkCampaign{
		UserID:      userID.(string),
		CampaignID:  &req.CampaignID,
		Name:        req.Name,
		Type:        campaign.Type,
		MessageBody: campaign.MessageBody,
		ImageURL:    campaign.ImageURL,
		ImageBase64: campaign.ImageBase64,
		Caption:     campaign.Caption,
		Status:      status,
		TotalCount:  len(req.Phone),
		ScheduledAt: scheduledAt,
		Timezone:    timezone,
	}

	// Start transaction
	tx := database.TransactionalDB.Begin()

	// Create bulk campaign
	if err := tx.Create(&bulkCampaign).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.BulkCampaignResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("Failed to create bulk campaign: %v", err),
		})
		return
	}

	// Create bulk campaign items
	for _, phone := range req.Phone {
		item := models.BulkCampaignItem{
			BulkCampaignID: bulkCampaign.ID,
			Phone:          phone,
			Status:         models.BulkCampaignItemStatusPending,
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.BulkCampaignResponse{
				Code:    500,
				Success: false,
				Message: fmt.Sprintf("Failed to create bulk campaign item: %v", err),
			})
			return
		}
	}

	// Commit transaction
	tx.Commit()

	// If immediate execution, process now
	if status == models.BulkCampaignStatusPending {
		go processBulkCampaign(bulkCampaign.ID)
	}

	responseData := struct {
		BulkCampaignID  uint       `json:"bulk_campaign_id"`
		TotalRecipients int        `json:"total_recipients"`
		Status          string     `json:"status"`
		ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
		Timezone        string     `json:"timezone,omitempty"`
	}{
		BulkCampaignID:  bulkCampaign.ID,
		TotalRecipients: len(req.Phone),
		Status:          string(status),
		ScheduledAt:     scheduledAt,
		Timezone:        timezone,
	}

	c.JSON(http.StatusCreated, models.BulkCampaignResponse{
		Code:    201,
		Success: true,
		Message: "Bulk campaign created successfully",
		Data:    &responseData,
	})
}

// AddContacts adds contacts manually to the database
func AddContacts(c *gin.Context) {
	var req struct {
		Contacts []struct {
			Phone    string `json:"phone" binding:"required"`
			FullName string `json:"full_name" binding:"required"`
		} `json:"contacts" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"message": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "User ID not found",
		})
		return
	}

	var addedContacts []models.WhatsAppContact
	var addedCount, updatedCount int

	// Process each contact
	for _, contactData := range req.Contacts {
		// Check if contact already exists
		var existingContact models.WhatsAppContact
		err := database.TransactionalDB.Where("user_id = ? AND phone = ?", userID, contactData.Phone).First(&existingContact).Error

		if err != nil {
			// Contact doesn't exist, create new
			newContact := models.WhatsAppContact{
				UserID:   userID.(string),
				Phone:    contactData.Phone,
				FullName: contactData.FullName,
				Source:   "manual",
			}

			if err := database.TransactionalDB.Create(&newContact).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"success": false,
					"message": fmt.Sprintf("Failed to create contact %s: %v", contactData.Phone, err),
				})
				return
			}

			addedContacts = append(addedContacts, newContact)
			addedCount++
		} else {
			// Contact exists, update full name
			existingContact.FullName = contactData.FullName
			if err := database.TransactionalDB.Save(&existingContact).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"success": false,
					"message": fmt.Sprintf("Failed to update contact %s: %v", contactData.Phone, err),
				})
				return
			}

			addedContacts = append(addedContacts, existingContact)
			updatedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"message": fmt.Sprintf("Processed %d contacts (%d added, %d updated)", len(req.Contacts), addedCount, updatedCount),
		"data": gin.H{
			"added_count":   addedCount,
			"updated_count": updatedCount,
			"contacts":      addedContacts,
		},
	})
}

// GetBulkCampaigns retrieves all bulk campaigns for a session
func GetBulkCampaigns(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.BulkCampaignListResponse{
			Code:    401,
			Success: false,
		})
		return
	}

	var bulkCampaigns []models.BulkCampaign
	if err := database.TransactionalDB.Where("user_id = ?", userID).Preload("Campaign").Find(&bulkCampaigns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.BulkCampaignListResponse{
			Code:    500,
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, models.BulkCampaignListResponse{
		Code:    200,
		Success: true,
		Data:    bulkCampaigns,
	})
}

// GetBulkCampaign retrieves a specific bulk campaign by ID
func GetBulkCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.BulkCampaignDetailResponse{
			Code:    401,
			Success: false,
		})
		return
	}

	bulkCampaignID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BulkCampaignDetailResponse{
			Code:    400,
			Success: false,
		})
		return
	}

	var bulkCampaign models.BulkCampaign
	if err := database.TransactionalDB.Where("id = ? AND user_id = ?", bulkCampaignID, userID).
		Preload("Campaign").Preload("Items").First(&bulkCampaign).Error; err != nil {
		c.JSON(http.StatusNotFound, models.BulkCampaignDetailResponse{
			Code:    404,
			Success: false,
		})
		return
	}

	responseData := struct {
		BulkCampaign *models.BulkCampaign      `json:"bulk_campaign"`
		Items        []models.BulkCampaignItem `json:"items"`
	}{
		BulkCampaign: &bulkCampaign,
		Items:        bulkCampaign.Items,
	}

	c.JSON(http.StatusOK, models.BulkCampaignDetailResponse{
		Code:    200,
		Success: true,
		Data:    &responseData,
	})
}

// DeleteBulkCampaign deletes a bulk campaign
func DeleteBulkCampaign(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "User ID not found",
		})
		return
	}

	bulkCampaignID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"message": "Invalid bulk campaign ID",
		})
		return
	}

	var bulkCampaign models.BulkCampaign
	if err := database.TransactionalDB.Where("id = ? AND user_id = ?", bulkCampaignID, userID).First(&bulkCampaign).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"success": false,
			"message": "Bulk campaign not found",
		})
		return
	}

	// Prevent deletion of campaigns that are currently processing
	if bulkCampaign.Status == models.BulkCampaignStatusProcessing {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"message": "Cannot delete campaign that is currently processing",
		})
		return
	}

	// Delete the bulk campaign (cascade will delete items automatically)
	if err := database.TransactionalDB.Delete(&bulkCampaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": fmt.Sprintf("Failed to delete bulk campaign: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"message": "Bulk campaign deleted successfully",
	})
}

// Helper function to process bulk campaign (placeholder)
func processBulkCampaign(bulkCampaignID uint) {
	log.Printf("[BULK_CAMPAIGN] Starting processing for campaign ID: %d", bulkCampaignID)

	// Get bulk campaign details
	var bulkCampaign models.BulkCampaign
	if err := database.TransactionalDB.Preload("Items").First(&bulkCampaign, bulkCampaignID).Error; err != nil {
		log.Printf("[BULK_CAMPAIGN] Error fetching campaign %d: %v", bulkCampaignID, err)
		return
	}

	// Update status to processing
	if err := database.TransactionalDB.Model(&bulkCampaign).Update("status", models.BulkCampaignStatusProcessing).Error; err != nil {
		log.Printf("[BULK_CAMPAIGN] Error updating status to processing for campaign %d: %v", bulkCampaignID, err)
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"processed_at": &now,
	}
	if err := database.TransactionalDB.Model(&bulkCampaign).Updates(updates).Error; err != nil {
		log.Printf("[BULK_CAMPAIGN] Error updating processed_at for campaign %d: %v", bulkCampaignID, err)
		return
	}

	// Get WhatsApp session for this user
	var whatsappSession models.WhatsappSession
	if err := database.TransactionalDB.Where("\"userId\" = ? AND \"connected\" = ?", bulkCampaign.UserID, true).
		Order("\"updatedAt\" DESC").First(&whatsappSession).Error; err != nil {
		log.Printf("[BULK_CAMPAIGN] No active WhatsApp session found for user %s: %v", bulkCampaign.UserID, err)
		markCampaignFailed(bulkCampaignID, "No active WhatsApp session found")
		return
	}

	// Get WhatsApp server URL
	whatsappServerURL := os.Getenv("WA_SERVER_URL")
	if whatsappServerURL == "" {
		whatsappServerURL = "http://wa-api:8080"
	}

	// Process each item
	sentCount := 0
	failedCount := 0

	for i, item := range bulkCampaign.Items {
		// Add delay between requests to avoid overwhelming the server (except for first request)
		if i > 0 {
			time.Sleep(2 * time.Second)
		}

		// Send message to WhatsApp server with retry
		success, messageID, errorMsg := sendWhatsAppMessageWithRetry(whatsappServerURL, whatsappSession.Token, item.Phone, bulkCampaign)

		// Update item status
		itemUpdates := map[string]interface{}{
			"status": models.BulkCampaignItemStatusSent,
		}

		if success {
			itemUpdates["message_id"] = messageID
			itemUpdates["sent_at"] = &now
			sentCount++
			// Track message stats for successful send
			trackCampaignMessageStats(*whatsappSession.UserID, whatsappSession.Token, true, string(bulkCampaign.Type))
		} else {
			itemUpdates["status"] = models.BulkCampaignItemStatusFailed
			itemUpdates["error_message"] = errorMsg
			failedCount++
			log.Printf("[BULK_CAMPAIGN] Failed to send message to %s: %s", item.Phone, errorMsg)
			// Track message stats for failed send
			trackCampaignMessageStats(*whatsappSession.UserID, whatsappSession.Token, false, string(bulkCampaign.Type))
		}

		if err := database.TransactionalDB.Model(&item).Updates(itemUpdates).Error; err != nil {
			log.Printf("[BULK_CAMPAIGN] Error updating item %d: %v", item.ID, err)
		}
	}

	// Update final campaign status
	finalStatus := models.BulkCampaignStatusCompleted
	if sentCount == 0 && failedCount > 0 {
		finalStatus = models.BulkCampaignStatusFailed
	}

	completedAt := time.Now()
	finalUpdates := map[string]interface{}{
		"status":       finalStatus,
		"sent_count":   sentCount,
		"failed_count": failedCount,
		"completed_at": &completedAt,
	}

	if err := database.TransactionalDB.Model(&bulkCampaign).Updates(finalUpdates).Error; err != nil {
		log.Printf("[BULK_CAMPAIGN] Error updating final status for campaign %d: %v", bulkCampaignID, err)
		return
	}

	// Always log the final result
	log.Printf("[BULK_CAMPAIGN] Campaign %d completed: %d berhasil dikirim, %d gagal",
		bulkCampaignID, sentCount, failedCount)
}

// sendWhatsAppMessage sends a message to WhatsApp server
func sendWhatsAppMessage(serverURL, sessionToken, phone string, campaign models.BulkCampaign) (bool, string, string) {
	// Prepare message payload using gateway format
	var payload map[string]interface{}

	// Add message content based on type using WhatsApp server format
	switch campaign.Type {
	case "text":
		// For text messages, use Body field as expected by WhatsApp server
		payload = map[string]interface{}{
			"Phone": phone,
			"Body":  campaign.MessageBody,
		}
	case "image":
		// Validate image URL
		if campaign.ImageURL == "" {
			return false, "", "Image URL is required for image campaigns"
		}

		// Check if ImageURL is already a base64 data URI
		var imageData string
		if strings.HasPrefix(campaign.ImageURL, "data:image/") {
			// Already in base64 format
			imageData = campaign.ImageURL
		} else {
			// It's a URL, download and encode to base64

			// Check if it's an SVG file (not supported by WhatsApp)
			if strings.HasSuffix(strings.ToLower(campaign.ImageURL), ".svg") {
				return false, "", "SVG images are not supported by WhatsApp. Please use PNG, JPEG, GIF, or WebP format."
			}

			// Parse URL to validate it
			_, err := url.Parse(campaign.ImageURL)
			if err != nil {
				return false, "", fmt.Sprintf("Invalid image URL: %v", err)
			}

			convertedData, err := downloadAndEncodeImageForCampaign(campaign.ImageURL)
			if err != nil {
				return false, "", fmt.Sprintf("Failed to download and encode image: %v", err)
			}
			imageData = convertedData
		}

		// For image messages, use gateway format
		payload = map[string]interface{}{
			"Phone": phone,
			"Image": imageData,
		}
		if campaign.Caption != "" {
			payload["Caption"] = campaign.Caption
		}
	default:
		return false, "", fmt.Sprintf("Unsupported message type: %s", campaign.Type)
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, "", fmt.Sprintf("Failed to marshal payload: %v", err)
	}

	// Create request with correct endpoint based on message type
	var url string
	switch campaign.Type {
	case "text":
		url = fmt.Sprintf("%s/chat/send/text", serverURL)
	case "image":
		url = fmt.Sprintf("%s/chat/send/image", serverURL)
	default:
		return false, "", fmt.Sprintf("Unsupported message type: %s", campaign.Type)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, "", fmt.Sprintf("Failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", sessionToken)

	// Send request with increased timeout for bulk operations
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Check for specific timeout errors
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return false, "", "Request timeout - WhatsApp server took too long to respond"
		}
		if strings.Contains(err.Error(), "connection") {
			return false, "", "Network connection error - please check connectivity"
		}
		return false, "", fmt.Sprintf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Sprintf("Failed to read response: %v", err)
	}

	// Check if response is successful first
	if resp.StatusCode != 200 {
		return false, "", fmt.Sprintf("WhatsApp server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Try to parse as JSON
	var response struct {
		Code    int    `json:"code"`
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Id        string `json:"Id"` // WhatsApp uses "Id" not "message_id"
			Details   string `json:"Details"`
			Timestamp int64  `json:"Timestamp"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		// If JSON parsing fails, check if it's HTML error page
		bodyStr := string(body)
		if strings.Contains(bodyStr, "<html>") || strings.Contains(bodyStr, "<!DOCTYPE") {
			return false, "", "WhatsApp server returned HTML page instead of JSON (possibly server error or wrong endpoint)"
		}
		return false, "", fmt.Sprintf("Failed to parse JSON response: %v (response: %s)", err, bodyStr)
	}

	if !response.Success || resp.StatusCode != 200 {
		return false, "", fmt.Sprintf("WhatsApp server error: %s", response.Message)
	}

	return true, response.Data.Id, ""
}

// markCampaignFailed marks a campaign as failed
func markCampaignFailed(bulkCampaignID uint, reason string) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.BulkCampaignStatusFailed,
		"completed_at": &now,
	}

	if err := database.TransactionalDB.Model(&models.BulkCampaign{}).Where("id = ?", bulkCampaignID).Updates(updates).Error; err != nil {
		log.Printf("[BULK_CAMPAIGN] Error marking campaign %d as failed: %v", bulkCampaignID, err)
	}
}

// BulkCampaignCronJob processes scheduled bulk campaigns (called by cron every minute)
func BulkCampaignCronJob(c *gin.Context) {
	db := database.GetTransactionalDB()

	// Find scheduled campaigns that are ready to be sent
	var scheduledCampaigns []models.BulkCampaign
	now := time.Now()

	err := db.Where("status = ? AND scheduled_at <= ?", models.BulkCampaignStatusScheduled, now).Find(&scheduledCampaigns).Error
	if err != nil {
		log.Printf("[CRON_JOB] Error fetching scheduled campaigns: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to fetch scheduled campaigns",
		})
		return
	}

	processedCount := 0
	for _, campaign := range scheduledCampaigns {
		go processBulkCampaign(campaign.ID)
		processedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"message": "Bulk campaign cron job completed",
		"data": gin.H{
			"processed_count": processedCount,
			"checked_at":      now,
		},
	})
}

// trackCampaignMessageStats tracks message statistics for campaign sends
func trackCampaignMessageStats(userID, whatsappToken string, success bool, messageType string) {
	// Find WhatsApp session by token to get sessionId
	var session models.WhatsappSession
	if err := database.TransactionalDB.Where("token = ?", whatsappToken).First(&session).Error; err != nil {
		return
	}

	// Try to get existing stats record or create new one
	var stats models.WhatsAppMessageStats
	err := database.TransactionalDB.Where("\"userId\" = ? AND \"sessionId\" = ?", userID, session.SessionID).First(&stats).Error

	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// Create new stats record
		stats = models.WhatsAppMessageStats{
			ID:        uuid.New().String(),
			UserID:    userID,
			SessionID: session.SessionID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Initialize counters based on message type
		if success {
			stats.TotalMessagesSent = 1
			updateMessageTypeCounter(&stats, messageType, true)
			stats.LastMessageSentAt = &now
		} else {
			stats.TotalMessagesFailed = 1
			updateMessageTypeCounter(&stats, messageType, false)
			stats.LastMessageFailedAt = &now
		}

		if err := database.TransactionalDB.Create(&stats).Error; err != nil {
			return
		}
	} else if err == nil {
		// Update existing stats record
		if success {
			stats.TotalMessagesSent++
			updateMessageTypeCounter(&stats, messageType, true)
			stats.LastMessageSentAt = &now
		} else {
			stats.TotalMessagesFailed++
			updateMessageTypeCounter(&stats, messageType, false)
			stats.LastMessageFailedAt = &now
		}
		stats.UpdatedAt = now

		if err := database.TransactionalDB.Save(&stats).Error; err != nil {
			return
		}
	}
}

// sendWhatsAppMessageWithRetry sends a message with retry mechanism
func sendWhatsAppMessageWithRetry(serverURL, sessionToken, phone string, campaign models.BulkCampaign) (bool, string, string) {
	maxRetries := 3
	var lastError string

	for attempt := 1; attempt <= maxRetries; attempt++ {
		success, messageID, errorMsg := sendWhatsAppMessage(serverURL, sessionToken, phone, campaign)

		if success {
			return true, messageID, ""
		}

		lastError = errorMsg

		// If it's the last attempt, don't wait
		if attempt < maxRetries {
			// Wait before retry, longer for network issues
			if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "connection") {
				time.Sleep(5 * time.Second)
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}

	return false, "", fmt.Sprintf("Failed after %d attempts: %s", maxRetries, lastError)
}

// updateMessageTypeCounter updates the appropriate counter based on message type
func updateMessageTypeCounter(stats *models.WhatsAppMessageStats, messageType string, success bool) {
	switch strings.ToLower(messageType) {
	case "text":
		if success {
			stats.TextMessagesSent++
		} else {
			stats.TextMessagesFailed++
		}
	case "image":
		if success {
			stats.ImageMessagesSent++
		} else {
			stats.ImageMessagesFailed++
		}
	case "document":
		if success {
			stats.DocumentMessagesSent++
		} else {
			stats.DocumentMessagesFailed++
		}
	case "audio":
		if success {
			stats.AudioMessagesSent++
		} else {
			stats.AudioMessagesFailed++
		}
	case "sticker":
		if success {
			stats.StickerMessagesSent++
		} else {
			stats.StickerMessagesFailed++
		}
	case "video":
		if success {
			stats.VideoMessagesSent++
		} else {
			stats.VideoMessagesFailed++
		}
	case "location":
		if success {
			stats.LocationMessagesSent++
		} else {
			stats.LocationMessagesFailed++
		}
	case "contact":
		if success {
			stats.ContactMessagesSent++
		} else {
			stats.ContactMessagesFailed++
		}
	case "template":
		if success {
			stats.TemplateMessagesSent++
		} else {
			stats.TemplateMessagesFailed++
		}
	default:
		// For unknown types, still count in total but log it
		log.Printf("[BULK_CAMPAIGN] Unknown message type for stats: %s", messageType)
	}
}

// downloadAndEncodeImageForCampaign downloads an image from URL and returns base64 encoded data URI
// This function is specific for campaign processing and includes WhatsApp format validation
func downloadAndEncodeImageForCampaign(imageURL string) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the image
	resp, err := client.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image, status: %d", resp.StatusCode)
	}

	// Read the image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %v", err)
	}

	// Get MIME type from content type header or detect from bytes
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = detectMimeTypeForCampaign(imageData)
	}

	// Validate that it's a supported image format for WhatsApp
	if !isSupportedImageFormat(mimeType) {
		return "", fmt.Errorf("unsupported image format: %s. WhatsApp supports PNG, JPEG, GIF, and WebP only", mimeType)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(imageData)

	// Return data URI format
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	return dataURI, nil
}

// detectMimeTypeForCampaign detects MIME type from image bytes with WhatsApp support validation
func detectMimeTypeForCampaign(data []byte) string {
	if len(data) < 8 {
		return "application/octet-stream"
	}

	// Check PNG signature
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png"
	}

	// Check JPEG signature
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return "image/jpeg"
	}

	// Check GIF signature
	if bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")) {
		return "image/gif"
	}

	// Check WebP signature
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp"
	}

	// Check SVG signature (not supported by WhatsApp)
	if bytes.HasPrefix(data, []byte("<?xml")) || bytes.HasPrefix(data, []byte("<svg")) {
		return "image/svg+xml"
	}

	// Default fallback
	return "application/octet-stream"
}

// isSupportedImageFormat checks if the MIME type is supported by WhatsApp
func isSupportedImageFormat(mimeType string) bool {
	supportedFormats := []string{
		"image/png",
		"image/jpeg",
		"image/jpg",
		"image/gif",
		"image/webp",
	}

	for _, format := range supportedFormats {
		if strings.EqualFold(mimeType, format) {
			return true
		}
	}

	return false
}

// parseSendSyncWithTimezone parses the SendSync field with timezone support
// Returns: scheduledTime (in UTC), timezone string, error
// Supports both UTC offset (+07:00, -05:00) and IANA timezone names (Asia/Jakarta)
func parseSendSyncWithTimezone(sendSync string, userTimezone string) (*time.Time, string, error) {
	sendSync = strings.TrimSpace(strings.ToLower(sendSync))

	// Immediate execution - no timezone needed
	if sendSync == "now" || sendSync == "sekarang" {
		return nil, "", nil
	}

	// For scheduled campaigns, timezone is required
	if userTimezone == "" {
		return nil, "", fmt.Errorf("timezone is required for scheduled campaigns (e.g., '+07:00', '-05:00', or 'Asia/Jakarta')")
	}

	var loc *time.Location
	var err error

	// Try to parse as UTC offset first (e.g., +07:00, -05:00, +05:30)
	if strings.HasPrefix(userTimezone, "+") || strings.HasPrefix(userTimezone, "-") {
		// Parse UTC offset
		loc, err = parseUTCOffset(userTimezone)
		if err != nil {
			return nil, "", fmt.Errorf("invalid UTC offset '%s': %v (use format like '+07:00', '-05:00', '+05:30')", userTimezone, err)
		}
	} else {
		// Try to load as IANA timezone name
		loc, err = time.LoadLocation(userTimezone)
		if err != nil {
			return nil, "", fmt.Errorf("invalid timezone '%s': %v (use UTC offset like '+07:00' or IANA timezone names like 'Asia/Jakarta')", userTimezone, err)
		}
	}

	// Try to parse datetime in various formats (assuming it's in user's timezone)
	formats := []string{
		"2006-01-02 15:04:05", // 2023-12-25 14:30:00
		"2006-01-02 15:04",    // 2023-12-25 14:30
		"2006-01-02T15:04:05", // 2023-12-25T14:30:00
		"02/01/2006 15:04",    // 25/12/2023 14:30
		"02-01-2006 15:04",    // 25-12-2023 14:30
	}

	var parsedTime time.Time
	var parseSuccess bool

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, sendSync, loc); err == nil {
			parsedTime = t
			parseSuccess = true
			break
		}
	}

	if !parseSuccess {
		return nil, "", fmt.Errorf("invalid datetime format. Use formats like: 'YYYY-MM-DD HH:MM:SS', 'YYYY-MM-DD HH:MM', or 'DD/MM/YYYY HH:MM'")
	}

	// Convert to UTC for storage
	utcTime := parsedTime.UTC()

	// Ensure the scheduled time is in the future
	if utcTime.Before(time.Now().UTC()) {
		// Show current time in user's timezone for better error message
		currentTimeInUserTZ := time.Now().In(loc).Format("2006-01-02 15:04:05")
		return nil, "", fmt.Errorf("scheduled time must be in the future (current time in %s: %s)",
			userTimezone, currentTimeInUserTZ)
	}

	return &utcTime, userTimezone, nil
}

// parseUTCOffset parses UTC offset string (e.g., "+07:00", "-05:00", "+05:30") and returns a time.Location
func parseUTCOffset(offset string) (*time.Location, error) {
	// Validate format
	if len(offset) < 6 {
		return nil, fmt.Errorf("offset too short, expected format: +HH:MM or -HH:MM")
	}

	// Extract sign
	sign := offset[0]
	if sign != '+' && sign != '-' {
		return nil, fmt.Errorf("offset must start with + or -")
	}

	// Remove sign for parsing
	offsetStr := offset[1:]

	// Parse hours and minutes
	var hours, minutes int
	var err error

	// Support both HH:MM and HHMM formats
	if strings.Contains(offsetStr, ":") {
		parts := strings.Split(offsetStr, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid offset format, expected +HH:MM or -HH:MM")
		}
		hours, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid hours: %v", err)
		}
		minutes, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minutes: %v", err)
		}
	} else if len(offsetStr) == 4 {
		// HHMM format
		hours, err = strconv.Atoi(offsetStr[0:2])
		if err != nil {
			return nil, fmt.Errorf("invalid hours: %v", err)
		}
		minutes, err = strconv.Atoi(offsetStr[2:4])
		if err != nil {
			return nil, fmt.Errorf("invalid minutes: %v", err)
		}
	} else {
		return nil, fmt.Errorf("invalid offset format, expected +HH:MM or +HHMM")
	}

	// Validate ranges
	if hours < 0 || hours > 14 {
		return nil, fmt.Errorf("hours must be between 0 and 14")
	}
	if minutes < 0 || minutes > 59 {
		return nil, fmt.Errorf("minutes must be between 0 and 59")
	}

	// Calculate total offset in seconds
	totalSeconds := (hours * 3600) + (minutes * 60)

	// Apply sign
	if sign == '-' {
		totalSeconds = -totalSeconds
	}

	// Create fixed timezone with the offset
	// Format name as UTC+07:00 or UTC-05:00 for clarity
	zoneName := fmt.Sprintf("UTC%s", offset)
	loc := time.FixedZone(zoneName, totalSeconds)

	return loc, nil
}
