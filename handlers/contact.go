package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"genfity-chat-ai/database"
	"genfity-chat-ai/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// extractPhoneNumber extracts clean phone number from WhatsApp format
func extractPhoneNumber(phoneNumber string) string {
	// Remove @s.whatsapp.net suffix
	if strings.HasSuffix(phoneNumber, "@s.whatsapp.net") {
		return strings.TrimSuffix(phoneNumber, "@s.whatsapp.net")
	}

	// Skip @lid format (channel/business accounts) as they're not regular phone numbers
	if strings.HasSuffix(phoneNumber, "@lid") {
		return ""
	}

	// Return as is if no suffix found
	return phoneNumber
}

// BulkContactSync syncs contacts from external WhatsApp server and stores them in database
func BulkContactSync(c *gin.Context) {
	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "User ID not found",
		})
		return
	}

	// Get WhatsApp session token from header
	whatsappToken := c.GetHeader("token")
	if whatsappToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "WhatsApp session token is required",
		})
		return
	}

	// Get base URL for external WhatsApp server
	baseURL := os.Getenv("WHATSAPP_SERVER_URL")
	if baseURL == "" {
		baseURL = "https://wa.genfity.com"
	}

	// Make request to external WhatsApp server
	url := fmt.Sprintf("%s/user/contacts", baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to create request",
		})
		return
	}

	// Add WhatsApp session token to request header
	req.Header.Set("token", whatsappToken)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to connect to WhatsApp server",
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to read response",
		})
		return
	}

	// Parse response from external API
	var syncResponse models.ExternalContactSyncResponse
	if err := json.Unmarshal(body, &syncResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to parse response",
		})
		return
	}

	// Check if external API request was successful
	if !syncResponse.Success || syncResponse.Code != 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    syncResponse.Code,
			"success": false,
			"message": "Failed to sync contacts from WhatsApp server",
		})
		return
	}

	// Store contacts in transactional database
	db := database.GetTransactionalDB()
	var savedContacts []models.WhatsAppContact

	fmt.Printf("Processing %d contacts from sync response\n", len(syncResponse.Data))

	for phoneNumber, contactData := range syncResponse.Data {
		// Extract clean phone number (remove @s.whatsapp.net or @lid)
		cleanPhone := extractPhoneNumber(phoneNumber)
		if cleanPhone == "" {
			fmt.Printf("Skipping invalid phone: %s\n", phoneNumber)
			continue
		}

		// fmt.Printf("Processing contact: %s -> %s (%s)\n", phoneNumber, cleanPhone, contactData.FullName)

		// Check if contact already exists for this user and phone
		var existingContact models.WhatsAppContact
		err := db.Where("user_id = ? AND phone = ?", userID, cleanPhone).Take(&existingContact).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Contact doesn't exist, create new one
			newContact := models.WhatsAppContact{
				UserID:   userID.(string),
				Phone:    cleanPhone,
				Name:     contactData.FirstName,
				FullName: contactData.FullName,
				PushName: contactData.PushName,
				Business: contactData.BusinessName != "",
				Source:   "sync",
			}

			if err := db.Create(&newContact).Error; err != nil {
				fmt.Printf("Failed to create contact %s: %v\n", cleanPhone, err)
				continue // Skip this contact if failed to save
			}
			// fmt.Printf("Created new contact: %s\n", cleanPhone)
			savedContacts = append(savedContacts, newContact)
		} else if err != nil {
			// Other database error, skip this contact
			fmt.Printf("Database error for contact %s: %v\n", cleanPhone, err)
			continue
		} else {
			// Contact exists, replace with new data (sesuai requirement)
			existingContact.Name = contactData.FirstName
			existingContact.FullName = contactData.FullName
			existingContact.PushName = contactData.PushName
			existingContact.Business = contactData.BusinessName != ""
			existingContact.Source = "sync" // Update source to sync

			if err := db.Save(&existingContact).Error; err != nil {
				fmt.Printf("Failed to update contact %s: %v\n", cleanPhone, err)
				continue // Skip this contact if failed to update
			}
			// fmt.Printf("Replaced existing contact: %s\n", cleanPhone)
			savedContacts = append(savedContacts, existingContact)
		}
	} // Return original response from external API along with storage status
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"message": fmt.Sprintf("Successfully synced %d contacts", len(savedContacts)),
		"data":    syncResponse.Data,
		"stored":  len(savedContacts),
	})
}

// BulkContactList returns simplified contact list for the authenticated user
func BulkContactList(c *gin.Context) {
	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "User ID not found",
		})
		return
	}

	// Get contacts for this user from transactional database
	var contacts []models.WhatsAppContact
	if err := database.GetTransactionalDB().Where("user_id = ?", userID).Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to fetch contacts",
		})
		return
	}

	// Convert to simplified response format
	var contactList []models.ContactSimple
	for _, contact := range contacts {
		// Determine best name to use (priority: FullName > PushName > Name)
		fullName := contact.FullName
		if fullName == "" && contact.PushName != "" {
			fullName = contact.PushName
		}
		if fullName == "" && contact.Name != "" {
			fullName = contact.Name
		}

		// Only include contacts that have a phone number
		if contact.Phone != "" {
			contactList = append(contactList, models.ContactSimple{
				Phone:    contact.Phone,
				FullName: fullName,
				Source:   contact.Source,
			})
		}
	}

	c.JSON(http.StatusOK, models.ContactListResponse{
		Code:    200,
		Success: true,
		Data:    contactList,
	})
}

// BulkDeleteContacts deletes multiple contacts by phone numbers for the authenticated user
func BulkDeleteContacts(c *gin.Context) {
	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "User ID not found",
		})
		return
	}

	// Parse request body
	var req struct {
		Phone []string `json:"phone" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"message": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Get database connection
	db := database.GetTransactionalDB()

	// Count contacts before deletion for reporting
	var totalContacts int64
	db.Model(&models.WhatsAppContact{}).Where("user_id = ? AND phone IN ?", userID, req.Phone).Count(&totalContacts)

	if totalContacts == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"success": false,
			"message": "No contacts found for the provided phone numbers",
		})
		return
	}

	// Delete contacts
	result := db.Where("user_id = ? AND phone IN ?", userID, req.Phone).Delete(&models.WhatsAppContact{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": fmt.Sprintf("Failed to delete contacts: %v", result.Error),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"message": fmt.Sprintf("Successfully deleted %d contacts", result.RowsAffected),
		"data": gin.H{
			"deleted_count":   result.RowsAffected,
			"requested_count": len(req.Phone),
		},
	})
}
