package models

import (
	"time"
)

// WhatsappSession model - sesuai dengan skema Prisma
type WhatsappSession struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(30);column:id"`
	SessionID       string    `json:"sessionId" gorm:"unique;column:sessionId"`
	SessionName     string    `json:"sessionName" gorm:"column:sessionName"`
	Token           string    `json:"token" gorm:"unique;column:token"`
	UserID          *string   `json:"userId" gorm:"column:userId;index"`
	Webhook         *string   `json:"webhook" gorm:"column:webhook"`
	Events          *string   `json:"events" gorm:"column:events"`
	Expiration      int       `json:"expiration" gorm:"default:0;column:expiration"`
	Connected       bool      `json:"connected" gorm:"default:false;column:connected"`
	LoggedIn        bool      `json:"loggedIn" gorm:"default:false;column:loggedIn"`
	JID             *string   `json:"jid" gorm:"column:jid"`
	QRCode          *string   `json:"qrcode" gorm:"type:text;column:qrcode"`
	Status          string    `json:"status" gorm:"default:disconnected;column:status"`
	Message         *string   `json:"message" gorm:"column:message"`
	IsSystemSession bool      `json:"isSystemSession" gorm:"default:false;column:isSystemSession"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime;column:createdAt"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime;column:updatedAt"`
}

// TableName specifies the table name for GORM
func (WhatsappSession) TableName() string {
	return "WhatsAppSession"
}

// WhatsappApiPackage model - sesuai dengan skema Prisma
type WhatsappApiPackage struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(30);column:id"`
	Name        string    `json:"name" gorm:"column:name"`
	Description *string   `json:"description" gorm:"column:description"`
	PriceMonth  int       `json:"priceMonth" gorm:"column:priceMonth"`
	PriceYear   int       `json:"priceYear" gorm:"column:priceYear"`
	MaxSession  int       `json:"maxSession" gorm:"column:maxSession"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime;column:createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime;column:updatedAt"`
}

// TableName specifies the table name for GORM
func (WhatsappApiPackage) TableName() string {
	return "WhatsappApiPackage"
}

// ServicesWhatsappCustomers model - sesuai dengan skema Prisma
type ServicesWhatsappCustomers struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:varchar(30);column:id"`
	CustomerID         string    `json:"customerId" gorm:"index;not null;column:customerId"`
	PackageID          string    `json:"packageId" gorm:"index;not null;column:packageId"`
	Status             string    `json:"status" gorm:"default:active;index;column:status"`
	ActivatedAt        time.Time `json:"activatedAt" gorm:"autoCreateTime;column:activatedAt"`
	ExpiredAt          time.Time `json:"expiredAt" gorm:"index;not null;column:expiredAt"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"autoUpdateTime;column:updatedAt"`
	LastSubscriptionAt time.Time `json:"lastSubscriptionAt" gorm:"autoCreateTime;column:lastSubscriptionAt"`
}

// TableName specifies the table name for GORM
func (ServicesWhatsappCustomers) TableName() string {
	return "ServicesWhatsappCustomers"
}

// WhatsAppMessageStats model - sesuai dengan skema Prisma
type WhatsAppMessageStats struct {
	ID                  string     `json:"id" gorm:"primaryKey;type:varchar(30);column:id"`
	UserID              string     `json:"userId" gorm:"index;not null;column:userId"`
	SessionID           string     `json:"sessionId" gorm:"index;not null;column:sessionId"`
	TotalMessagesSent   int        `json:"totalMessagesSent" gorm:"default:0;column:totalMessagesSent"`
	TotalMessagesFailed int        `json:"totalMessagesFailed" gorm:"default:0;column:totalMessagesFailed"`
	LastMessageSentAt   *time.Time `json:"lastMessageSentAt" gorm:"column:lastMessageSentAt"`
	LastMessageFailedAt *time.Time `json:"lastMessageFailedAt" gorm:"column:lastMessageFailedAt"`
	CreatedAt           time.Time  `json:"createdAt" gorm:"autoCreateTime;column:createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt" gorm:"autoUpdateTime;column:updatedAt"`
}

// TableName specifies the table name for GORM
func (WhatsAppMessageStats) TableName() string {
	return "WhatsAppMessageStats"
}

// Gateway Response Types
type GatewayResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}
