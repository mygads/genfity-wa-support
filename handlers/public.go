package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"genfity-wa-support/database"
	"genfity-wa-support/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createSessionRequest struct {
	SessionName   string `json:"session_name" binding:"required"`
	WebhookURL    string `json:"webhook_url"`
	Events        string `json:"events"`
	ExpirationSec int    `json:"expiration_sec"`
	AutoConnect   bool   `json:"auto_connect"`
	AutoRead      bool   `json:"auto_read_enabled"`
	Typing        bool   `json:"typing_enabled"`
	History       int    `json:"history"`
}

type updateSessionRequest struct {
	SessionName   *string `json:"session_name"`
	WebhookURL    *string `json:"webhook_url"`
	Events        *string `json:"events"`
	ExpirationSec *int    `json:"expiration_sec"`
	AutoRead      *bool   `json:"auto_read_enabled"`
	Typing        *bool   `json:"typing_enabled"`
	History       *int    `json:"history"`
}

func GetCurrentUser(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sub, err := getActiveSubscription(user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "subscription": sub})
}

func ListSessions(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	var sessions []models.WhatsAppSession
	if err := database.GetDB().Where("user_id = ?", user.ID).Order("updated_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list sessions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func CreateSession(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sub, err := getActiveSubscription(user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	var req createSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if req.Events == "" {
		req.Events = "Message,Connected,Disconnected,QR"
	}

	var current int64
	database.GetDB().Model(&models.WhatsAppSession{}).
		Where("user_id = ? AND status IN ?", user.ID, []string{"active", "connected", "qr_waiting"}).
		Count(&current)
	if int(current) >= sub.MaxSessions {
		c.JSON(http.StatusForbidden, gin.H{"message": "session limit exceeded"})
		return
	}

	sessionTokenRaw, _, err := generateAPIKey("wat")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed generating session token"})
		return
	}

	adminPayload := map[string]interface{}{
		"name":       req.SessionName,
		"token":      sessionTokenRaw,
		"webhook":    req.WebhookURL,
		"expiration": req.ExpirationSec,
		"events":     req.Events,
		"history":    req.History,
	}
	status, body, err := proxyAdminToWAServer(http.MethodPost, "/admin/users", adminPayload)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	if status < 200 || status >= 300 {
		c.Data(status, "application/json", body)
		return
	}

	waUserID, waToken, waWebhook := parseWAAdminUserResponse(body)
	if waUserID == "" {
		c.JSON(http.StatusBadGateway, gin.H{"message": "invalid wa response for session provisioning"})
		return
	}
	if waToken == "" {
		waToken = sessionTokenRaw
	}
	if waWebhook == "" {
		waWebhook = req.WebhookURL
	}

	now := time.Now()
	session := models.WhatsAppSession{
		UserID:          user.ID,
		Provider:        sub.Provider,
		SessionID:       waUserID,
		SessionName:     req.SessionName,
		SessionToken:    waToken,
		WebhookURL:      waWebhook,
		AutoReadEnabled: req.AutoRead,
		TypingEnabled:   req.Typing,
		Status:          "created",
		LastSyncedAt:    &now,
	}
	if err := database.GetDB().Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to store session"})
		return
	}

	if req.AutoConnect {
		_, _, _ = proxyWithToken(http.MethodPost, "/session/connect", waToken, map[string]interface{}{"subscribe": strings.Split(req.Events, ",")})
	}

	c.JSON(http.StatusCreated, gin.H{"session": session})
	return
}

func UpdateSession(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sessionID := c.Param("session_id")

	var req updateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var session models.WhatsAppSession
	if err := database.GetDB().Where("user_id = ? AND session_id = ?", user.ID, sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "session not found for user"})
		return
	}

	adminPayload := map[string]interface{}{}
	if req.SessionName != nil {
		adminPayload["name"] = *req.SessionName
		session.SessionName = *req.SessionName
	}
	if req.WebhookURL != nil {
		adminPayload["webhook"] = *req.WebhookURL
		session.WebhookURL = *req.WebhookURL
	}
	if req.Events != nil {
		adminPayload["events"] = *req.Events
	}
	if req.ExpirationSec != nil {
		adminPayload["expiration"] = *req.ExpirationSec
	}
	if req.History != nil {
		adminPayload["history"] = *req.History
	}

	if len(adminPayload) > 0 {
		status, body, err := proxyAdminToWAServer(http.MethodPut, "/admin/users/"+sessionID, adminPayload)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
			return
		}
		if status < 200 || status >= 300 {
			c.Data(status, "application/json", body)
			return
		}
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.AutoRead != nil {
		updates["auto_read_enabled"] = *req.AutoRead
		session.AutoReadEnabled = *req.AutoRead
	}
	if req.Typing != nil {
		updates["typing_enabled"] = *req.Typing
		session.TypingEnabled = *req.Typing
	}
	if req.WebhookURL != nil {
		updates["webhook_url"] = *req.WebhookURL
	}

	if err := database.GetDB().Model(&models.WhatsAppSession{}).Where("id = ?", session.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update local session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": session})
}

func DeleteSession(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sessionID := c.Param("session_id")
	if !userOwnsSession(user.ID, sessionID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "session not found for user"})
		return
	}

	status, body, err := proxyAdminToWAServer(http.MethodDelete, "/admin/users/"+sessionID+"/full", nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}

	if status >= 200 && status < 300 {
		_ = database.GetDB().Where("user_id = ? AND session_id = ?", user.ID, sessionID).Delete(&models.WhatsAppSession{}).Error
	}
	c.Data(status, "application/json", body)
}

func GetSessionSettings(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sessionID := c.Param("session_id")

	var session models.WhatsAppSession
	if err := database.GetDB().Where("user_id = ? AND session_id = ?", user.ID, sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":          session.SessionID,
		"auto_read_enabled":   session.AutoReadEnabled,
		"typing_enabled":      session.TypingEnabled,
		"webhook_url":         session.WebhookURL,
		"message_stat_sent":   session.LastMessageSent,
		"message_stat_failed": session.LastMessageFail,
	})
}

func UpdateSessionSettings(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sessionID := c.Param("session_id")
	var session models.WhatsAppSession
	if err := database.GetDB().Where("user_id = ? AND session_id = ?", user.ID, sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "session not found"})
		return
	}

	var req struct {
		AutoReadEnabled *bool   `json:"auto_read_enabled"`
		TypingEnabled   *bool   `json:"typing_enabled"`
		WebhookURL      *string `json:"webhook_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.AutoReadEnabled != nil {
		updates["auto_read_enabled"] = *req.AutoReadEnabled
	}
	if req.TypingEnabled != nil {
		updates["typing_enabled"] = *req.TypingEnabled
	}
	if req.WebhookURL != nil {
		updates["webhook_url"] = *req.WebhookURL
	}

	if err := database.GetDB().Model(&models.WhatsAppSession{}).
		Where("user_id = ? AND session_id = ?", user.ID, sessionID).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update settings"})
		return
	}

	if req.WebhookURL != nil {
		_, _, _ = proxyWithToken(http.MethodPut, "/webhook", session.SessionToken, map[string]interface{}{"webhookURL": *req.WebhookURL})
	}
	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

func ListSessionContacts(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sessionID := c.Param("session_id")

	var session models.WhatsAppSession
	if err := database.GetDB().Where("user_id = ? AND session_id = ?", user.ID, sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "session not found"})
		return
	}

	autoSync := c.DefaultQuery("sync", "true")
	if strings.EqualFold(autoSync, "true") {
		status, body, err := proxyWithToken(http.MethodGet, "/user/contacts", session.SessionToken, nil)
		if err == nil && status >= 200 && status < 300 {
			_, _ = upsertContacts(user.ID, sessionID, body)
		}
	}

	var contacts []models.SessionContact
	if err := database.GetDB().Where("user_id = ? AND session_id = ?", user.ID, sessionID).Order("name asc").Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list contacts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contacts": contacts})
}

func SyncSessionContacts(c *gin.Context) {
	user := c.MustGet("user").(models.ServiceUser)
	sessionID := c.Param("session_id")

	var session models.WhatsAppSession
	if err := database.GetDB().Where("user_id = ? AND session_id = ?", user.ID, sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "session not found"})
		return
	}

	status, body, err := proxyWithToken(http.MethodGet, "/user/contacts", session.SessionToken, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	if status < 200 || status >= 300 {
		c.Data(status, "application/json", body)
		return
	}

	count, err := upsertContacts(user.ID, sessionID, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"synced": count})
}

func WhatsAppGateway(c *gin.Context) {
	token := getTokenFromRequest(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "token required"})
		return
	}

	session, sub, err := validateSessionToken(token)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}
	if sub.Status != models.SubscriptionActive {
		c.JSON(http.StatusForbidden, gin.H{"message": "subscription inactive"})
		return
	}

	targetPath := strings.TrimPrefix(c.Request.URL.Path, "/wa")
	if targetPath == "" {
		targetPath = "/"
	}
	if strings.HasPrefix(targetPath, "/admin") {
		c.JSON(http.StatusForbidden, gin.H{"message": "admin path is not exposed"})
		return
	}

	if sub.MaxMessages > 0 && c.Request.Method == http.MethodPost && strings.HasPrefix(targetPath, "/chat/send") {
		remaining := sub.MaxMessages - int(session.LastMessageSent)
		if remaining <= 0 {
			c.JSON(http.StatusForbidden, gin.H{"message": "message quota exceeded"})
			return
		}
	}

	if session.TypingEnabled && c.Request.Method == http.MethodPost && strings.HasPrefix(targetPath, "/chat/send") {
		_, _, _ = proxyWithToken(http.MethodPost, "/chat/presence", session.SessionToken, map[string]interface{}{"state": "composing"})
	}

	status, body, err := proxyToWAServer(c, targetPath, false)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}

	if c.Request.Method == http.MethodPost && strings.HasPrefix(targetPath, "/chat/send") {
		incSent := int64(0)
		incFail := int64(1)
		if status >= 200 && status < 300 {
			incSent = 1
			incFail = 0
		}
		_ = database.GetDB().Model(&models.WhatsAppSession{}).
			Where("id = ?", session.ID).
			Updates(map[string]interface{}{
				"last_message_sent": gorm.Expr("last_message_sent + ?", incSent),
				"last_message_fail": gorm.Expr("last_message_fail + ?", incFail),
				"last_activity_at":  time.Now(),
			}).Error

		messageType := detectMessageType(targetPath)
		_ = upsertMessageStat(session.UserID, session.SessionID, messageType, incSent, incFail)
	}

	if strings.HasPrefix(targetPath, "/session") && status >= 200 && status < 300 {
		_ = syncSessionFromResponse(session.UserID, body)
	}

	c.Data(status, "application/json", body)
}

func validateSessionToken(token string) (models.WhatsAppSession, models.UserSubscription, error) {
	var session models.WhatsAppSession
	if err := database.GetDB().Where("session_token = ?", token).First(&session).Error; err != nil {
		return session, models.UserSubscription{}, err
	}
	sub, err := getActiveSubscription(session.UserID)
	if err != nil {
		return session, models.UserSubscription{}, err
	}
	return session, sub, nil
}

func getActiveSubscription(userID string) (models.UserSubscription, error) {
	var sub models.UserSubscription
	err := database.GetDB().Where("user_id = ? AND status = ?", userID, models.SubscriptionActive).Order("updated_at desc").First(&sub).Error
	if err != nil {
		return sub, err
	}
	if time.Now().After(sub.ExpiresAt) {
		sub.Status = models.SubscriptionExpired
		_ = database.GetDB().Save(&sub).Error
		return sub, errors.New("subscription expired")
	}
	return sub, nil
}

func userOwnsSession(userID, sessionID string) bool {
	var count int64
	database.GetDB().Model(&models.WhatsAppSession{}).Where("user_id = ? AND session_id = ?", userID, sessionID).Count(&count)
	return count > 0
}

func getTokenFromRequest(c *gin.Context) string {
	token := c.GetHeader("token")
	if token != "" {
		return token
	}
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return authHeader
}

func proxyToWAServer(c *gin.Context, path string, useAdminToken bool) (int, []byte, error) {
	waServerURL := strings.TrimRight(os.Getenv("WA_SERVER_URL"), "/")
	targetURL := waServerURL + path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	for k, values := range c.Request.Header {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}
	if useAdminToken {
		req.Header.Set("Authorization", "Bearer "+os.Getenv("WA_ADMIN_TOKEN"))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusBadGateway, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func proxyJSONToWAServer(method string, path string, payload interface{}) (int, []byte, error) {
	waServerURL := strings.TrimRight(os.Getenv("WA_SERVER_URL"), "/")
	targetURL := waServerURL + path
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("WA_ADMIN_TOKEN"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusBadGateway, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func proxyAdminToWAServer(method string, path string, payload interface{}) (int, []byte, error) {
	waServerURL := strings.TrimRight(os.Getenv("WA_SERVER_URL"), "/")
	targetURL := waServerURL + path

	var reader io.Reader
	if payload != nil {
		body, _ := json.Marshal(payload)
		reader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, targetURL, reader)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", os.Getenv("WA_ADMIN_TOKEN"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusBadGateway, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func proxyWithToken(method string, path string, token string, payload interface{}) (int, []byte, error) {
	waServerURL := strings.TrimRight(os.Getenv("WA_SERVER_URL"), "/")
	targetURL := waServerURL + path

	var reader io.Reader
	if payload != nil {
		body, _ := json.Marshal(payload)
		reader = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(method, targetURL, reader)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusBadGateway, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func parseWAAdminUserResponse(body []byte) (id string, token string, webhook string) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", ""
	}
	data, _ := payload["data"].(map[string]interface{})
	if data == nil {
		data = payload
	}
	id, _ = data["id"].(string)
	token, _ = data["token"].(string)
	webhook, _ = data["webhook"].(string)
	return
}

func detectMessageType(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}

func upsertMessageStat(userID, sessionID, messageType string, sent, failed int64) error {
	now := time.Now()
	db := database.GetDB()
	var stat models.SessionMessageStat
	if err := db.Where("user_id = ? AND session_id = ? AND message_type = ?", userID, sessionID, messageType).First(&stat).Error; err != nil {
		stat = models.SessionMessageStat{
			UserID:      userID,
			SessionID:   sessionID,
			MessageType: messageType,
			TotalSent:   sent,
			TotalFailed: failed,
		}
		if sent > 0 {
			stat.LastSuccessAt = now
		}
		if failed > 0 {
			stat.LastFailedAt = now
		}
		return db.Create(&stat).Error
	}

	updates := map[string]interface{}{
		"total_sent":   gorm.Expr("total_sent + ?", sent),
		"total_failed": gorm.Expr("total_failed + ?", failed),
		"updated_at":   now,
	}
	if sent > 0 {
		updates["last_success_at"] = now
	}
	if failed > 0 {
		updates["last_failed_at"] = now
	}
	return db.Model(&models.SessionMessageStat{}).Where("id = ?", stat.ID).Updates(updates).Error
}

func upsertContacts(userID, sessionID string, raw []byte) (int, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, fmt.Errorf("invalid contacts response")
	}

	now := time.Now()
	db := database.GetDB()
	count := 0

	// Format 1: genfity-wa returns map[jid]contactInfo
	for key, value := range payload {
		entry, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		if upsertContactRow(db, userID, sessionID, key, entry, now) {
			count++
		}
	}

	if count > 0 {
		return count, nil
	}

	// Format 2: array response under data/contacts
	listAny := payload["data"]
	if listAny == nil {
		listAny = payload["contacts"]
	}
	list, ok := listAny.([]interface{})
	if !ok {
		return 0, nil
	}

	for _, item := range list {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		jid, _ := entry["jid"].(string)
		if jid == "" {
			jid, _ = entry["id"].(string)
		}
		if upsertContactRow(db, userID, sessionID, jid, entry, now) {
			count++
		}
	}
	return count, nil
}

func upsertContactRow(db *gorm.DB, userID, sessionID, jid string, entry map[string]interface{}, now time.Time) bool {
	jid = strings.TrimSpace(jid)
	if jid == "" {
		return false
	}

	name, _ := entry["name"].(string)
	if name == "" {
		if pushName, ok := entry["full_name"].(string); ok {
			name = pushName
		}
	}

	phone, _ := entry["phone"].(string)
	if phone == "" {
		phone = normalizePhoneFromJID(jid)
	}

	var contact models.SessionContact
	err := db.Where("user_id = ? AND session_id = ? AND jid = ?", userID, sessionID, jid).First(&contact).Error
	if err != nil {
		contact = models.SessionContact{
			UserID:       userID,
			SessionID:    sessionID,
			JID:          jid,
			Name:         name,
			Phone:        phone,
			Raw:          models.JSONB(entry),
			LastSyncedAt: now,
		}
		return db.Create(&contact).Error == nil
	}

	contact.Name = name
	contact.Phone = phone
	contact.Raw = models.JSONB(entry)
	contact.LastSyncedAt = now
	return db.Save(&contact).Error == nil
}

func normalizePhoneFromJID(jid string) string {
	left := strings.Split(jid, "@")[0]
	left = strings.Split(left, ":")[0]
	return left
}

func syncSessionFromResponse(userID string, body []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	sessionData, ok := payload["data"].(map[string]interface{})
	if !ok {
		sessionData = payload
	}
	sessionID, _ := sessionData["id"].(string)
	if sessionID == "" {
		sessionID, _ = sessionData["sessionId"].(string)
	}
	if sessionID == "" {
		return nil
	}
	name, _ := sessionData["name"].(string)
	token, _ := sessionData["token"].(string)
	jid, _ := sessionData["jid"].(string)
	status, _ := sessionData["status"].(string)
	connected, _ := sessionData["connected"].(bool)
	loggedIn, _ := sessionData["loggedIn"].(bool)
	now := time.Now()

	var session models.WhatsAppSession
	db := database.GetDB()
	if err := db.Where("user_id = ? AND session_id = ?", userID, sessionID).First(&session).Error; err != nil {
		session = models.WhatsAppSession{
			UserID:       userID,
			SessionID:    sessionID,
			SessionName:  name,
			SessionToken: token,
			JID:          jid,
			Status:       status,
			Connected:    connected,
			LoggedIn:     loggedIn,
			LastSyncedAt: &now,
		}
		return db.Create(&session).Error
	}

	session.SessionName = name
	if token != "" {
		session.SessionToken = token
	}
	session.JID = jid
	session.Status = status
	session.Connected = connected
	session.LoggedIn = loggedIn
	session.LastSyncedAt = &now
	return db.Save(&session).Error
}
