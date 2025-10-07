package models

import (
	"time"

	"gorm.io/gorm"
)

// CampaignType represents the type of campaign message
type CampaignType string

const (
	CampaignTypeText  CampaignType = "text"
	CampaignTypeImage CampaignType = "image"
)

// CampaignStatus represents the status of a campaign
type CampaignStatus string

const (
	CampaignStatusActive   CampaignStatus = "active"
	CampaignStatusInactive CampaignStatus = "inactive"
	CampaignStatusArchived CampaignStatus = "archived"
)

type Campaign struct {
	ID     uint           `json:"id" gorm:"primaryKey"`
	UserID string         `json:"user_id" gorm:"column:user_id;not null;index"` // Changed from session_id to user_id
	Name   string         `json:"name" gorm:"column:name;not null"`
	Type   CampaignType   `json:"type" gorm:"column:type;not null"`
	Status CampaignStatus `json:"status" gorm:"column:status;default:'active'"`

	// Message content
	MessageBody string `json:"message_body" gorm:"column:message_body;type:text"`
	ImageURL    string `json:"image_url" gorm:"column:image_url;type:text"`
	ImageBase64 string `json:"image_base64" gorm:"column:image_base64;type:text"`
	Caption     string `json:"caption" gorm:"column:caption;type:text"`

	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	// Relationships
	User          *User          `json:"user,omitempty" gorm:"foreignKey:UserID"` // Changed from session
	BulkCampaigns []BulkCampaign `json:"bulk_campaigns,omitempty" gorm:"foreignKey:CampaignID"`
}

// TableName specifies the table name for Campaign
func (Campaign) TableName() string {
	return "WhatsAppCampaigns"
}

// BulkCampaignStatus represents the status of a bulk campaign
type BulkCampaignStatus string

const (
	BulkCampaignStatusPending    BulkCampaignStatus = "pending"
	BulkCampaignStatusScheduled  BulkCampaignStatus = "scheduled"
	BulkCampaignStatusProcessing BulkCampaignStatus = "processing"
	BulkCampaignStatusCompleted  BulkCampaignStatus = "completed"
	BulkCampaignStatusFailed     BulkCampaignStatus = "failed"
)

// BulkCampaign represents a bulk campaign execution
type BulkCampaign struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	UserID     string `json:"user_id" gorm:"column:user_id;not null;index"` // Changed from session_id to user_id
	CampaignID *uint  `json:"campaign_id" gorm:"column:campaign_id;index"`  // Nullable reference
	Name       string `json:"name" gorm:"column:name;not null"`

	// Store message data copy (so it persists even if campaign is deleted)
	Type        CampaignType `json:"type" gorm:"column:type;not null"`
	MessageBody string       `json:"message_body" gorm:"column:message_body;type:text"`
	ImageURL    string       `json:"image_url" gorm:"column:image_url;type:text"`
	ImageBase64 string       `json:"image_base64" gorm:"column:image_base64;type:text"`
	Caption     string       `json:"caption" gorm:"column:caption;type:text"`

	Status      BulkCampaignStatus `json:"status" gorm:"column:status;default:'pending'"`
	TotalCount  int                `json:"total_count" gorm:"column:total_count;default:0"`
	SentCount   int                `json:"sent_count" gorm:"column:sent_count;default:0"`
	FailedCount int                `json:"failed_count" gorm:"column:failed_count;default:0"`

	// Scheduling
	ScheduledAt *time.Time `json:"scheduled_at" gorm:"column:scheduled_at"`
	Timezone    string     `json:"timezone" gorm:"column:timezone"` // User's timezone for scheduled campaigns
	ProcessedAt *time.Time `json:"processed_at" gorm:"column:processed_at"`
	CompletedAt *time.Time `json:"completed_at" gorm:"column:completed_at"`

	ErrorMessage string `json:"error_message" gorm:"column:error_message;type:text"`

	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	// Relationships
	User     *User              `json:"user,omitempty" gorm:"foreignKey:UserID"` // Changed from session
	Campaign *Campaign          `json:"campaign,omitempty" gorm:"foreignKey:CampaignID"`
	Items    []BulkCampaignItem `json:"items,omitempty" gorm:"foreignKey:BulkCampaignID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for BulkCampaign
func (BulkCampaign) TableName() string {
	return "WhatsAppBulkCampaigns"
}

// BulkCampaignItem represents individual message item in a bulk campaign
type BulkCampaignItem struct {
	ID             uint                   `json:"id" gorm:"primaryKey"`
	BulkCampaignID uint                   `json:"bulk_campaign_id" gorm:"column:bulk_campaign_id;not null;index"`
	Phone          string                 `json:"phone" gorm:"column:phone;not null"`
	Status         BulkCampaignItemStatus `json:"status" gorm:"column:status;default:'pending'"`
	MessageID      string                 `json:"message_id" gorm:"column:message_id"`
	ErrorMessage   string                 `json:"error_message" gorm:"column:error_message;type:text"`
	SentAt         *time.Time             `json:"sent_at" gorm:"column:sent_at"`
	CreatedAt      time.Time              `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time              `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt         `json:"deleted_at" gorm:"column:deleted_at;index"`

	// Relations
	BulkCampaign *BulkCampaign `json:"bulk_campaign,omitempty" gorm:"foreignKey:BulkCampaignID"`
}

// TableName specifies the table name for BulkCampaignItem
func (BulkCampaignItem) TableName() string {
	return "WhatsAppBulkCampaignItems"
}

// BulkCampaignItemStatus represents the status of individual message item
type BulkCampaignItemStatus string

const (
	BulkCampaignItemStatusPending BulkCampaignItemStatus = "pending"
	BulkCampaignItemStatusSent    BulkCampaignItemStatus = "sent"
	BulkCampaignItemStatusFailed  BulkCampaignItemStatus = "failed"
)

// Request/Response structs for campaign APIs

// CreateCampaignRequest represents request for creating a campaign
type CreateCampaignRequest struct {
	Name        string       `json:"name" binding:"required"`
	Type        CampaignType `json:"type" binding:"required,oneof=text image"`
	MessageBody string       `json:"message_body"`
	ImageURL    string       `json:"image_url"`
	ImageBase64 string       `json:"image_base64"`
	Caption     string       `json:"caption"`
}

// UpdateCampaignRequest represents request for updating a campaign
type UpdateCampaignRequest struct {
	Name        string         `json:"name"`
	Status      CampaignStatus `json:"status"`
	MessageBody string         `json:"message_body"`
	ImageURL    string         `json:"image_url"`
	ImageBase64 string         `json:"image_base64"`
	Caption     string         `json:"caption"`
}

// CreateBulkCampaignRequest represents request for creating a bulk campaign
type CreateBulkCampaignRequest struct {
	CampaignID uint     `json:"campaign_id" binding:"required"`
	Name       string   `json:"name" binding:"required"`
	Phone      []string `json:"phone" binding:"required,min=1"`
	SendSync   string   `json:"send_sync" binding:"required"`
	Timezone   string   `json:"timezone"` // Required for scheduled campaigns (e.g., "Asia/Jakarta", "America/New_York")
}

// CampaignResponse represents response for campaign operations
type CampaignResponse struct {
	Code    int       `json:"code"`
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    *Campaign `json:"data,omitempty"`
}

// CampaignListResponse represents response for campaign list
type CampaignListResponse struct {
	Code    int        `json:"code"`
	Success bool       `json:"success"`
	Data    []Campaign `json:"data"`
}

// BulkCampaignResponse represents response for bulk campaign creation
type BulkCampaignResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *struct {
		BulkCampaignID  uint       `json:"bulk_campaign_id"`
		TotalRecipients int        `json:"total_recipients"`
		Status          string     `json:"status"`
		ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
		Timezone        string     `json:"timezone,omitempty"`
	} `json:"data,omitempty"`
}

// BulkCampaignListResponse represents response for bulk campaign list
type BulkCampaignListResponse struct {
	Code    int            `json:"code"`
	Success bool           `json:"success"`
	Data    []BulkCampaign `json:"data"`
}

// BulkCampaignDetailResponse represents response for bulk campaign detail
type BulkCampaignDetailResponse struct {
	Code    int  `json:"code"`
	Success bool `json:"success"`
	Data    *struct {
		BulkCampaign *BulkCampaign      `json:"bulk_campaign"`
		Items        []BulkCampaignItem `json:"items"`
	} `json:"data,omitempty"`
}
