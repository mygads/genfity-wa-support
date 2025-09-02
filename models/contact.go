package models

import (
	"time"

	"gorm.io/gorm"
)

// WhatsAppContact represents a contact from WhatsApp sync
type WhatsAppContact struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	SessionID    string         `json:"session_id" gorm:"column:session_id;index;not null"`   // Links to WhatsApp session
	ContactJID   string         `json:"contact_jid" gorm:"column:contact_jid;index;not null"` // WhatsApp ID like "6285xxx@s.whatsapp.net"
	BusinessName string         `json:"business_name" gorm:"column:business_name"`            // Business name if available
	FirstName    string         `json:"first_name" gorm:"column:first_name"`                  // First name
	FullName     string         `json:"full_name" gorm:"column:full_name"`                    // Full name
	PushName     string         `json:"push_name" gorm:"column:push_name"`                    // Push name from WhatsApp
	Found        bool           `json:"found" gorm:"column:found;default:false"`              // Whether contact was found
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}

// TableName specifies the table name for WhatsAppContact
func (WhatsAppContact) TableName() string {
	return "WhatsAppContact"
}

// ContactSyncRequest represents the request to sync contacts
type ContactSyncRequest struct {
	Token string `json:"token" header:"token"`
}

// ContactData represents individual contact data from external API
type ContactData struct {
	BusinessName string `json:"BusinessName"`
	FirstName    string `json:"FirstName"`
	Found        bool   `json:"Found"`
	FullName     string `json:"FullName"`
	PushName     string `json:"PushName"`
}

// ContactSyncResponse represents the response from external WhatsApp API
type ContactSyncResponse struct {
	Code    int                    `json:"code"`
	Data    map[string]ContactData `json:"data"`
	Success bool                   `json:"success"`
}

// ContactListResponse represents a simplified contact for the list endpoint
type ContactListResponse struct {
	Telp     string `json:"telp"`
	FullName string `json:"fullname"`
}

// BulkContactListResponse represents the response for bulk contact list
type BulkContactListResponse struct {
	Code    int                   `json:"code"`
	Data    []ContactListResponse `json:"data"`
	Success bool                  `json:"success"`
}
