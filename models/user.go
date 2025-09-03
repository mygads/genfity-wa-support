package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents a user in the system (view only - table already exists)
type User struct {
	ID                            string     `json:"id" gorm:"primaryKey;column:id"`
	Name                          *string    `json:"name" gorm:"column:name"`
	Email                         *string    `json:"email" gorm:"column:email;unique"`
	Phone                         *string    `json:"phone" gorm:"column:phone;unique"`
	Password                      *string    `json:"password,omitempty" gorm:"column:password"`
	OTP                           *string    `json:"otp,omitempty" gorm:"column:otp"`
	OTPExpires                    *time.Time `json:"otp_expires,omitempty" gorm:"column:otpExpires"`
	OTPVerificationDeadline       *time.Time `json:"otp_verification_deadline,omitempty" gorm:"column:otpVerificationDeadline"`
	EmailVerified                 *time.Time `json:"email_verified" gorm:"column:emailVerified"`
	PhoneVerified                 *time.Time `json:"phone_verified" gorm:"column:phoneVerified"`
	Image                         *string    `json:"image" gorm:"column:image"`
	EmailVerificationToken        *string    `json:"email_verification_token,omitempty" gorm:"column:emailVerificationToken;unique"`
	EmailVerificationTokenExpires *time.Time `json:"email_verification_token_expires,omitempty" gorm:"column:emailVerificationTokenExpires"`
	Role                          string     `json:"role" gorm:"column:role;default:customer"`
	ApiKey                        *string    `json:"api_key,omitempty" gorm:"column:apiKey;unique"`
	UpdatedAt                     time.Time  `json:"updated_at" gorm:"column:updatedAt"`
	EmailOTP                      *string    `json:"email_otp,omitempty" gorm:"column:emailOtp"`
	EmailOTPExpires               *time.Time `json:"email_otp_expires,omitempty" gorm:"column:emailOtpExpires"`
	ResetPasswordOTP              *string    `json:"reset_password_otp,omitempty" gorm:"column:resetPasswordOtp"`
	ResetPasswordOTPExpires       *time.Time `json:"reset_password_otp_expires,omitempty" gorm:"column:resetPasswordOtpExpires"`
	ResetPasswordLastRequestAt    *time.Time `json:"reset_password_last_request_at,omitempty" gorm:"column:resetPasswordLastRequestAt"`
	SSOOTP                        *string    `json:"sso_otp,omitempty" gorm:"column:ssoOtp"`
	SSOOTPExpires                 *time.Time `json:"sso_otp_expires,omitempty" gorm:"column:ssoOtpExpires"`
	SSOLastRequestAt              *time.Time `json:"sso_last_request_at,omitempty" gorm:"column:ssoLastRequestAt"`
	IsActive                      bool       `json:"is_active" gorm:"column:isActive;default:true"`
	CreatedAt                     time.Time  `json:"created_at" gorm:"column:createdAt"`

	// Relationships
	UserSessions []UserSession `json:"user_sessions,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "User"
}

// UserSession represents a user session (view only - table already exists)
type UserSession struct {
	UserID     string    `json:"user_id" gorm:"column:userId;not null;index"`
	Token      string    `json:"token" gorm:"column:token;unique;size:512;not null;index"`
	DeviceInfo *string   `json:"device_info" gorm:"column:deviceInfo;type:text"`
	IPAddress  *string   `json:"ip_address" gorm:"column:ipAddress"`
	UserAgent  *string   `json:"user_agent" gorm:"column:userAgent;type:text"`
	IsActive   bool      `json:"is_active" gorm:"column:isActive;default:true;index"`
	LastUsed   time.Time `json:"last_used" gorm:"column:lastUsed"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"column:expiresAt;index"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:createdAt"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updatedAt"`
	ID         string    `json:"id" gorm:"primaryKey;column:id"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for UserSession
func (UserSession) TableName() string {
	return "UserSession"
}

// JWTClaims represents the JWT payload structure
type JWTClaims struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	SessionID string `json:"sessionId"`
	jwt.RegisteredClaims
}
