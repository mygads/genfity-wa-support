package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"genfity-chat-ai/database"
	"genfity-chat-ai/models"

	"github.com/gin-gonic/gin"
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

	// Parse scheduling
	scheduledAt, err := parseSendSync(req.SendSync)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BulkCampaignResponse{
			Code:    400,
			Success: false,
			Message: fmt.Sprintf("Invalid send_sync format: %v", err),
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
	}{
		BulkCampaignID:  bulkCampaign.ID,
		TotalRecipients: len(req.Phone),
		Status:          string(status),
		ScheduledAt:     scheduledAt,
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

// Helper function to process bulk campaign (placeholder)
func processBulkCampaign(bulkCampaignID uint) {
	// TODO: Implement actual WhatsApp message sending
	// This is a placeholder implementation

	// Update status to processing
	database.TransactionalDB.Model(&models.BulkCampaign{}).Where("id = ?", bulkCampaignID).Update("status", models.BulkCampaignStatusProcessing)

	// Simulate processing (replace with actual WhatsApp API calls)
	time.Sleep(2 * time.Second)

	// Mark as completed
	now := time.Now()
	database.TransactionalDB.Model(&models.BulkCampaign{}).Where("id = ?", bulkCampaignID).Updates(map[string]interface{}{
		"status":       models.BulkCampaignStatusCompleted,
		"processed_at": &now,
		"completed_at": &now,
	})
}
