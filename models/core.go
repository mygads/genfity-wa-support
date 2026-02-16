package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionExpired  SubscriptionStatus = "expired"
	SubscriptionInactive SubscriptionStatus = "inactive"
)

type ServiceUser struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(64)"`
	SourceService   string    `json:"source_service" gorm:"type:varchar(64);index;not null"`
	CustomerAPIKey  string    `json:"customer_api_key" gorm:"type:varchar(128);uniqueIndex;not null"`
	CreatedBy       string    `json:"created_by" gorm:"type:varchar(128)"`
	Status          string    `json:"status" gorm:"type:varchar(32);default:'active';index"`
	ProviderMapping JSONB     `json:"provider_mapping" gorm:"type:jsonb"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ServiceUser) TableName() string {
	return "wa_service_users"
}

type UserSubscription struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	UserID      string             `json:"user_id" gorm:"type:varchar(64);index;not null"`
	Provider    string             `json:"provider" gorm:"type:varchar(32);default:'genfity-wa';index"`
	MaxSessions int                `json:"max_sessions" gorm:"default:1"`
	MaxMessages int                `json:"max_messages" gorm:"default:0"`
	ExpiresAt   time.Time          `json:"expires_at" gorm:"index;not null"`
	Status      SubscriptionStatus `json:"status" gorm:"type:varchar(16);default:'active';index"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func (UserSubscription) TableName() string {
	return "wa_user_subscriptions"
}

type WhatsAppSession struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	UserID          string     `json:"user_id" gorm:"type:varchar(64);index;not null"`
	Provider        string     `json:"provider" gorm:"type:varchar(32);default:'genfity-wa';index"`
	SessionID       string     `json:"session_id" gorm:"type:varchar(128);index;not null"`
	SessionName     string     `json:"session_name" gorm:"type:varchar(255)"`
	SessionToken    string     `json:"session_token" gorm:"type:text;uniqueIndex"`
	WebhookURL      string     `json:"webhook_url" gorm:"type:text"`
	Connected       bool       `json:"connected" gorm:"default:false"`
	LoggedIn        bool       `json:"logged_in" gorm:"default:false"`
	JID             string     `json:"jid" gorm:"type:varchar(255)"`
	Status          string     `json:"status" gorm:"type:varchar(64);default:'inactive';index"`
	LastSyncedAt    *time.Time `json:"last_synced_at"`
	LastActivityAt  *time.Time `json:"last_activity_at"`
	LastMessageSent int64      `json:"last_message_sent" gorm:"default:0"`
	LastMessageFail int64      `json:"last_message_fail" gorm:"default:0"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (WhatsAppSession) TableName() string {
	return "wa_sessions"
}

type SessionMessageStat struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"type:varchar(64);index;not null"`
	SessionID     string    `json:"session_id" gorm:"type:varchar(128);index;not null"`
	MessageType   string    `json:"message_type" gorm:"type:varchar(64);index"`
	TotalSent     int64     `json:"total_sent" gorm:"default:0"`
	TotalFailed   int64     `json:"total_failed" gorm:"default:0"`
	LastSuccessAt time.Time `json:"last_success_at"`
	LastFailedAt  time.Time `json:"last_failed_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (SessionMessageStat) TableName() string {
	return "wa_session_message_stats"
}

type SessionContact struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"type:varchar(64);index;not null"`
	SessionID    string    `json:"session_id" gorm:"type:varchar(128);index;not null"`
	JID          string    `json:"jid" gorm:"type:varchar(255);index;not null"`
	Name         string    `json:"name" gorm:"type:varchar(255)"`
	Phone        string    `json:"phone" gorm:"type:varchar(64)"`
	Raw          JSONB     `json:"raw" gorm:"type:jsonb"`
	LastSyncedAt time.Time `json:"last_synced_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SessionContact) TableName() string {
	return "wa_session_contacts"
}

type GatewayResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
