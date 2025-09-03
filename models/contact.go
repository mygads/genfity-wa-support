package models

import (
	"time"

	"gorm.io/gorm"
)

// WhatsAppContact represents a contact from WhatsApp
type WhatsAppContact struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    string         `json:"user_id" gorm:"column:user_id;not null;index"` // User ID from JWT context
	Phone     string         `json:"phone" gorm:"column:phone;not null;index"`
	Name      string         `json:"name" gorm:"column:name"`
	FullName  string         `json:"full_name" gorm:"column:full_name"`
	PushName  string         `json:"push_name" gorm:"column:push_name"`
	Short     string         `json:"short" gorm:"column:short"`
	Notify    string         `json:"notify" gorm:"column:notify"`
	Business  bool           `json:"business" gorm:"column:business;default:false"`
	Verified  bool           `json:"verified" gorm:"column:verified;default:false"`
	Source    string         `json:"source" gorm:"column:source;default:'sync'"` // 'sync' or 'manual'
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}

// TableName specifies the table name for WhatsAppContact
func (WhatsAppContact) TableName() string {
	return "WhatsAppContact"
}

// ContactSyncRequest represents request for contact sync
type ContactSyncRequest struct {
	SessionToken string `json:"session_token"`
}

// ContactSyncResponse represents response for contact sync
type ContactSyncResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *struct {
		SyncedCount int               `json:"synced_count"`
		Contacts    []WhatsAppContact `json:"contacts"`
	} `json:"data,omitempty"`
}

// ContactListResponse represents response for contact list
type ContactListResponse struct {
	Code    int             `json:"code"`
	Success bool            `json:"success"`
	Data    []ContactSimple `json:"data"`
}

// ContactSimple represents simplified contact data for listing
type ContactSimple struct {
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
	Source   string `json:"source"`
}

// AddContactsRequest represents request for adding contacts manually
type AddContactsRequest struct {
	Contacts []ContactData `json:"contacts" binding:"required,min=1"`
}

// ContactData represents contact data for manual addition
type ContactData struct {
	Phone    string `json:"phone" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
}

// AddContactsResponse represents response for adding contacts
type AddContactsResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *struct {
		AddedCount   int               `json:"added_count"`
		UpdatedCount int               `json:"updated_count"`
		Contacts     []WhatsAppContact `json:"contacts"`
	} `json:"data,omitempty"`
}

// ExternalContactData represents individual contact data from external API
type ExternalContactData struct {
	BusinessName string `json:"BusinessName"`
	FirstName    string `json:"FirstName"`
	Found        bool   `json:"Found"`
	FullName     string `json:"FullName"`
	PushName     string `json:"PushName"`
}

// ExternalContactSyncResponse represents the response from external WhatsApp API
type ExternalContactSyncResponse struct {
	Code    int                            `json:"code"`
	Data    map[string]ExternalContactData `json:"data"`
	Success bool                           `json:"success"`
}

// BulkContactListResponse represents the response for bulk contact list
type BulkContactListResponse struct {
	Code    int                   `json:"code"`
	Data    []ContactListResponse `json:"data"`
	Success bool                  `json:"success"`
}

// Tambahan untuk kompatibilitas dengan handler contact yang ada
type ContactListResponseOld struct {
	Telp     string `json:"telp"`
	FullName string `json:"fullname"`
}
