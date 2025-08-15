package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONB type for PostgreSQL JSONB columns
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

// GenEventWebhook represents any webhook event received
type GenEventWebhook struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	EventType   string     `json:"event_type" gorm:"not null;index"`
	Source      string     `json:"source" gorm:"not null;index;default:'wa'"` // wa, telegram, etc
	UserToken   string     `json:"user_token" gorm:"index"`                   // Links to specific WA user
	EventData   JSONB      `json:"event_data" gorm:"type:jsonb"`
	RawData     string     `json:"raw_data" gorm:"type:text"`
	Processed   bool       `json:"processed" gorm:"default:false;index"`
	ReceivedAt  time.Time  `json:"received_at" gorm:"autoCreateTime"`
	ProcessedAt *time.Time `json:"processed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// WhatsAppMessage represents processed WhatsApp messages
type WhatsAppMessage struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	MessageID         string    `json:"message_id" gorm:"uniqueIndex;not null"`
	From              string    `json:"from" gorm:"not null;index"`
	To                string    `json:"to" gorm:"not null;index"`
	FromMe            bool      `json:"from_me" gorm:"default:false"`
	PushName          string    `json:"push_name"`
	MessageType       string    `json:"message_type" gorm:"not null;index"`
	Body              string    `json:"body" gorm:"type:text"`
	Caption           string    `json:"caption" gorm:"type:text"`
	MediaData         JSONB     `json:"media_data" gorm:"type:jsonb"`
	LocationData      JSONB     `json:"location_data" gorm:"type:jsonb"`
	ContactData       JSONB     `json:"contact_data" gorm:"type:jsonb"`
	QuotedMessage     JSONB     `json:"quoted_message" gorm:"type:jsonb"`
	MentionedJid      JSONB     `json:"mentioned_jid" gorm:"type:jsonb"`
	GroupJid          string    `json:"group_jid"`
	Participant       string    `json:"participant"`
	Broadcast         bool      `json:"broadcast" gorm:"default:false"`
	Forwarded         bool      `json:"forwarded" gorm:"default:false"`
	EphemeralDuration int       `json:"ephemeral_duration" gorm:"default:0"`
	MessageTimestamp  time.Time `json:"message_timestamp" gorm:"not null;index"`
	WebhookReceived   time.Time `json:"webhook_received" gorm:"autoCreateTime"`
	Status            string    `json:"status" gorm:"default:'received'"`
	Processed         bool      `json:"processed" gorm:"default:false"`
	UserToken         string    `json:"user_token" gorm:"index"` // Links to specific WA user
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// WhatsAppReadReceipt represents read receipt events
type WhatsAppReadReceipt struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	MessageIds     JSONB     `json:"message_ids" gorm:"type:jsonb"`
	From           string    `json:"from" gorm:"not null;index"`
	To             string    `json:"to" gorm:"not null;index"`
	ReceiptType    string    `json:"receipt_type" gorm:"not null"` // read, delivered, played
	EventTimestamp time.Time `json:"event_timestamp" gorm:"not null"`
	ReceivedAt     time.Time `json:"received_at" gorm:"autoCreateTime"`
	UserToken      string    `json:"user_token" gorm:"index"` // Links to specific WA user
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WhatsAppPresence represents presence events (online/offline)
type WhatsAppPresence struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	From       string     `json:"from" gorm:"not null;index"`
	Presence   string     `json:"presence" gorm:"not null"` // available, unavailable
	LastSeen   *time.Time `json:"last_seen"`
	ReceivedAt time.Time  `json:"received_at" gorm:"autoCreateTime"`
	UserToken  string     `json:"user_token" gorm:"index"` // Links to specific WA user
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// WhatsAppChatPresence represents chat presence events (typing/recording)
type WhatsAppChatPresence struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	From        string     `json:"from" gorm:"not null;index"`
	ChatJid     string     `json:"chat_jid" gorm:"not null;index"`
	State       string     `json:"state" gorm:"not null"`             // composing, paused
	Media       string     `json:"media"`                             // text, audio
	AutoStopped bool       `json:"auto_stopped" gorm:"default:false"` // If auto-stopped after 10 seconds
	ExpiresAt   *time.Time `json:"expires_at"`                        // When composing state should auto-expire
	ReceivedAt  time.Time  `json:"received_at" gorm:"autoCreateTime"`
	UserToken   string     `json:"user_token" gorm:"index"` // Links to specific WA user
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// WhatsAppHistorySync represents history sync events
type WhatsAppHistorySync struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	SyncType      string    `json:"sync_type" gorm:"not null"`
	Conversations JSONB     `json:"conversations" gorm:"type:jsonb"`
	ReceivedAt    time.Time `json:"received_at" gorm:"autoCreateTime"`
	UserToken     string    `json:"user_token" gorm:"index"` // Links to specific WA user
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// WhatsAppSession represents session state and QR codes
type WhatsAppSession struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	UserToken      string     `json:"user_token" gorm:"uniqueIndex;not null"` // Unique per session
	SessionName    string     `json:"session_name"`                           // Name from WA server
	SessionID      string     `json:"session_id"`                             // ID from WA server
	JID            string     `json:"jid"`                                    // WhatsApp JID when connected
	SessionState   string     `json:"session_state" gorm:"not null"`          // connecting, qr_waiting, connected, disconnected
	Connected      bool       `json:"connected" gorm:"default:false"`         // From WA server status
	LoggedIn       bool       `json:"logged_in" gorm:"default:false"`         // From WA server status
	QRCode         string     `json:"qr_code" gorm:"type:text"`               // Base64 QR code data
	QRExpiredAt    *time.Time `json:"qr_expired_at"`                          // When QR expires
	ConnectedAt    *time.Time `json:"connected_at"`                           // When session connected
	DisconnectedAt *time.Time `json:"disconnected_at"`                        // When session disconnected
	LastSyncAt     *time.Time `json:"last_sync_at"`                           // Last sync with WA server
	Webhook        string     `json:"webhook"`                                // Webhook URL set on server
	LastActivityAt time.Time  `json:"last_activity_at" gorm:"autoUpdateTime"` // Last activity
	ReceivedAt     time.Time  `json:"received_at" gorm:"autoCreateTime"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// WhatsAppMessageStatus represents message delivery and read status
type WhatsAppMessageStatus struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	MessageID      string    `json:"message_id" gorm:"not null;index"`
	From           string    `json:"from" gorm:"not null;index"`
	To             string    `json:"to" gorm:"not null;index"`
	Status         string    `json:"status" gorm:"not null"` // delivered, read
	EventTimestamp time.Time `json:"event_timestamp" gorm:"not null"`
	ReceivedAt     time.Time `json:"received_at" gorm:"autoCreateTime"`
	UserToken      string    `json:"user_token" gorm:"index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WhatsAppServerUser represents user data from WA server admin API
type WhatsAppServerUser struct {
	Connected   bool   `json:"connected"`
	Events      string `json:"events"`
	Expiration  int    `json:"expiration"`
	ID          string `json:"id"`
	JID         string `json:"jid"`
	LoggedIn    bool   `json:"loggedIn"`
	Name        string `json:"name"`
	ProxyConfig struct {
		Enabled  bool   `json:"enabled"`
		ProxyURL string `json:"proxy_url"`
	} `json:"proxy_config"`
	ProxyURL string `json:"proxy_url"`
	QRCode   string `json:"qrcode"`
	S3Config struct {
		AccessKey     string `json:"access_key"`
		Bucket        string `json:"bucket"`
		Enabled       bool   `json:"enabled"`
		Endpoint      string `json:"endpoint"`
		MediaDelivery string `json:"media_delivery"`
		PathStyle     bool   `json:"path_style"`
		PublicURL     string `json:"public_url"`
		Region        string `json:"region"`
		RetentionDays int    `json:"retention_days"`
	} `json:"s3_config"`
	Token   string `json:"token"`
	Webhook string `json:"webhook"`
}

// WhatsAppServerResponse represents response from WA server admin API
type WhatsAppServerResponse struct {
	Code    int                  `json:"code"`
	Data    []WhatsAppServerUser `json:"data"`
	Success bool                 `json:"success"`
}

// UserSettings represents user account settings
type UserSettings struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserToken       string    `json:"user_token" gorm:"uniqueIndex;not null"`
	ChatLogEnabled  bool      `json:"chat_log_enabled" gorm:"default:false"`
	AutoReadEnabled bool      `json:"auto_read_enabled" gorm:"default:false"`
	WebhookURL      string    `json:"webhook_url"`
	DisplayName     string    `json:"display_name"`
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ChatRoom represents a chat conversation
type ChatRoom struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ChatID       string    `json:"chat_id" gorm:"uniqueIndex;not null"` // combination of user_token + contact_jid
	UserToken    string    `json:"user_token" gorm:"index;not null"`
	ContactJID   string    `json:"contact_jid" gorm:"index;not null"`
	ContactName  string    `json:"contact_name"`
	ChatType     string    `json:"chat_type" gorm:"default:'individual'"` // individual, group
	IsGroup      bool      `json:"is_group" gorm:"default:false"`
	GroupName    string    `json:"group_name"`
	LastMessage  string    `json:"last_message" gorm:"type:text"`
	LastSender   string    `json:"last_sender"` // 'user' or 'contact'
	LastActivity time.Time `json:"last_activity" gorm:"autoUpdateTime"`
	UnreadCount  int       `json:"unread_count" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ChatMessage represents individual messages in chat rooms with status tracking
type ChatMessage struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	MessageID        string     `json:"message_id" gorm:"uniqueIndex;not null"`
	ChatRoomID       uint       `json:"chat_room_id" gorm:"index;not null"`
	ChatID           string     `json:"chat_id" gorm:"index;not null"`
	UserToken        string     `json:"user_token" gorm:"index;not null"`
	SenderJID        string     `json:"sender_jid" gorm:"index;not null"`
	SenderType       string     `json:"sender_type" gorm:"not null"` // 'user' or 'contact'
	MessageType      string     `json:"message_type" gorm:"not null"`
	Content          string     `json:"content" gorm:"type:text"`
	Caption          string     `json:"caption" gorm:"type:text"`
	MediaData        JSONB      `json:"media_data" gorm:"type:jsonb"`
	QuotedMessageID  string     `json:"quoted_message_id"`
	Status           string     `json:"status" gorm:"default:'sent'"` // sent, delivered, read
	MessageTimestamp time.Time  `json:"message_timestamp" gorm:"not null;index"`
	DeliveredAt      *time.Time `json:"delivered_at"`
	ReadAt           *time.Time `json:"read_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Relations
	ChatRoom *ChatRoom `json:"chat_room,omitempty" gorm:"foreignKey:ChatRoomID"`
}
