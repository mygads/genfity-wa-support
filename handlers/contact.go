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

// BulkContactSync syncs contacts from external WhatsApp server and stores them in database
func BulkContactSync(c *gin.Context) {
	// Get token from header
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "Token is required",
		})
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

	// Add token to request header
	req.Header.Set("token", token)

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
	var syncResponse models.ContactSyncResponse
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

	for contactJID, contactData := range syncResponse.Data {
		// Check if contact already exists using Take() which doesn't log "record not found"
		var existingContact models.WhatsAppContact
		err := db.Where("session_id = ? AND contact_jid = ?", session.ID, contactJID).Take(&existingContact).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Contact doesn't exist, create new one
			newContact := models.WhatsAppContact{
				SessionID:    session.ID,
				ContactJID:   contactJID,
				BusinessName: contactData.BusinessName,
				FirstName:    contactData.FirstName,
				FullName:     contactData.FullName,
				PushName:     contactData.PushName,
				Found:        contactData.Found,
			}

			if err := db.Create(&newContact).Error; err != nil {
				continue // Skip this contact if failed to save
			}
			savedContacts = append(savedContacts, newContact)
		} else if err != nil {
			// Other database error, skip this contact
			continue
		} else {
			// Contact exists, update it
			existingContact.BusinessName = contactData.BusinessName
			existingContact.FirstName = contactData.FirstName
			existingContact.FullName = contactData.FullName
			existingContact.PushName = contactData.PushName
			existingContact.Found = contactData.Found

			if err := db.Save(&existingContact).Error; err != nil {
				continue // Skip this contact if failed to update
			}
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
	// Get token from header
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "Token is required",
		})
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
		return
	}

	// Get contacts for this session from transactional database
	var contacts []models.WhatsAppContact
	if err := database.GetTransactionalDB().Where("session_id = ? AND found = ?", session.ID, true).Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"message": "Failed to fetch contacts",
		})
		return
	}

	// Convert to simplified response format
	var contactList []models.ContactListResponse
	for _, contact := range contacts {
		// Extract phone number from JID
		phoneNumber := extractPhoneFromJID(contact.ContactJID)

		// Determine best name to use (priority: FullName > PushName > FirstName > BusinessName)
		fullName := ""
		if contact.FullName != "" {
			fullName = contact.FullName
		} else if contact.PushName != "" {
			fullName = contact.PushName
		} else if contact.FirstName != "" {
			fullName = contact.FirstName
		} else if contact.BusinessName != "" {
			fullName = contact.BusinessName
		}

		// Only include contacts that have a name
		if fullName != "" && phoneNumber != "" {
			contactList = append(contactList, models.ContactListResponse{
				Telp:     phoneNumber,
				FullName: fullName,
			})
		}
	}

	c.JSON(http.StatusOK, models.BulkContactListResponse{
		Code:    200,
		Success: true,
		Data:    contactList,
	})
}

// extractPhoneFromJID extracts phone number from WhatsApp JID
// Examples: "6285215538030@s.whatsapp.net" -> "6285215538030", "136202103562334@lid" -> "136202103562334"
func extractPhoneFromJID(jid string) string {
	if strings.Contains(jid, "@") {
		parts := strings.Split(jid, "@")
		return parts[0]
	}
	return jid
}
