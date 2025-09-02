package models

import (
	"time"

	"gorm.io/gorm"
)

// BulkMessageType represents the type of bulk message
type BulkMessageType string

const (
	BulkMessageTypeText  BulkMessageType = "text"
	BulkMessageTypeImage BulkMessageType = "image"
)

// BulkMessageStatus represents the status of bulk message
type BulkMessageStatus string

const (
	BulkMessageStatusPending    BulkMessageStatus = "pending"
	BulkMessageStatusScheduled  BulkMessageStatus = "scheduled"
	BulkMessageStatusProcessing BulkMessageStatus = "processing"
	BulkMessageStatusCompleted  BulkMessageStatus = "completed"
	BulkMessageStatusFailed     BulkMessageStatus = "failed"
)

// BulkMessage represents a bulk message campaign
type BulkMessage struct {
	ID              uint              `json:"id" gorm:"primaryKey"`
	SessionID       string            `json:"session_id" gorm:"column:session_id;index;not null"`   // Links to WhatsApp session
	MessageType     BulkMessageType   `json:"message_type" gorm:"column:message_type;not null"`     // text or image
	PhoneNumbers    JSONB             `json:"phone_numbers" gorm:"column:phone_numbers;type:jsonb"` // Array of phone numbers
	Body            string            `json:"body" gorm:"column:body;type:text"`                    // Message body for text
	Image           string            `json:"image" gorm:"column:image;type:text"`                  // Image URL or base64
	Caption         string            `json:"caption" gorm:"column:caption;type:text"`              // Caption for image
	SendSync        string            `json:"send_sync" gorm:"column:send_sync"`                    // "now" or datetime
	ScheduledAt     *time.Time        `json:"scheduled_at" gorm:"column:scheduled_at"`              // When to send (null if immediate)
	Status          BulkMessageStatus `json:"status" gorm:"column:status;default:'pending'"`        // Current status
	TotalRecipients int               `json:"total_recipients" gorm:"column:total_recipients"`      // Total phone numbers
	SentCount       int               `json:"sent_count" gorm:"column:sent_count;default:0"`        // Successfully sent count
	FailedCount     int               `json:"failed_count" gorm:"column:failed_count;default:0"`    // Failed count
	ProcessedAt     *time.Time        `json:"processed_at" gorm:"column:processed_at"`              // When processing started
	CompletedAt     *time.Time        `json:"completed_at" gorm:"column:completed_at"`              // When completed
	ErrorMessage    string            `json:"error_message" gorm:"column:error_message;type:text"`  // Error details if failed
	CreatedAt       time.Time         `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time         `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt    `json:"deleted_at" gorm:"column:deleted_at;index"`
}

// TableName specifies the table name for BulkMessage
func (BulkMessage) TableName() string {
	return "BulkMessage"
}

// BulkMessageItem represents individual message in a bulk campaign
type BulkMessageItem struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	BulkMessageID uint           `json:"bulk_message_id" gorm:"column:bulk_message_id;index;not null"`
	PhoneNumber   string         `json:"phone_number" gorm:"column:phone_number;not null"`
	Status        string         `json:"status" gorm:"column:status;default:'pending'"`       // pending, sent, failed
	MessageID     string         `json:"message_id" gorm:"column:message_id"`                 // Message ID from WA server
	ErrorMessage  string         `json:"error_message" gorm:"column:error_message;type:text"` // Error details if failed
	SentAt        *time.Time     `json:"sent_at" gorm:"column:sent_at"`                       // When sent
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	// Relations
	BulkMessage *BulkMessage `json:"bulk_message,omitempty" gorm:"foreignKey:BulkMessageID"`
}

// TableName specifies the table name for BulkMessageItem
func (BulkMessageItem) TableName() string {
	return "BulkMessageItems"
}

// Request/Response structs for bulk message APIs

// BulkTextMessageRequest represents request for bulk text message
type BulkTextMessageRequest struct {
	Phone    []string `json:"Phone" binding:"required,min=1"`
	Body     string   `json:"Body" binding:"required"`
	SendSync string   `json:"SendSync" binding:"required"`
}

// BulkImageMessageRequest represents request for bulk image message
type BulkImageMessageRequest struct {
	Phone    []string `json:"Phone" binding:"required,min=1"`
	Image    string   `json:"Image" binding:"required"`
	Caption  string   `json:"Caption"`
	SendSync string   `json:"SendSync" binding:"required"`
}

// BulkMessageResponse represents response for bulk message creation
type BulkMessageResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *struct {
		BulkID          uint       `json:"bulk_id"`
		TotalRecipients int        `json:"total_recipients"`
		Status          string     `json:"status"`
		ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	} `json:"data,omitempty"`
}

// BulkMessageListResponse represents response for bulk message list
type BulkMessageListResponse struct {
	Code    int           `json:"code"`
	Success bool          `json:"success"`
	Data    []BulkMessage `json:"data"`
}

// BulkMessageDetailResponse represents response for bulk message detail
type BulkMessageDetailResponse struct {
	Code    int  `json:"code"`
	Success bool `json:"success"`
	Data    *struct {
		BulkMessage *BulkMessage      `json:"bulk_message"`
		Items       []BulkMessageItem `json:"items"`
	} `json:"data,omitempty"`
}
