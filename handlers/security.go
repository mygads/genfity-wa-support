package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"genfity-wa-support/database"
	"genfity-wa-support/models"

	"github.com/gin-gonic/gin"
)

type rateWindow struct {
	count     int
	windowEnd time.Time
}

type blockedIP struct {
	until time.Time
}

var (
	rateMutex       sync.Mutex
	rateCounters    = map[string]*rateWindow{}
	blockedIPCaches = map[string]*blockedIP{}
)

func hashAPIKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generateAPIKey(prefix string) (raw string, hashed string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = prefix + "_" + base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashAPIKey(raw), nil
}

func InternalAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		keys := strings.Split(os.Getenv("INTERNAL_API_KEYS"), ",")
		provided := strings.TrimSpace(c.GetHeader("x-internal-api-key"))
		if provided == "" {
			provided = strings.TrimSpace(c.GetHeader("Authorization"))
			provided = strings.TrimPrefix(provided, "Bearer ")
		}
		if provided == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "internal api key required"})
			return
		}

		for _, key := range keys {
			entry := strings.TrimSpace(key)
			if entry == "" {
				continue
			}

			// Scoped format: service-a:actualKey
			if strings.Contains(entry, ":") {
				parts := strings.SplitN(entry, ":", 2)
				source := strings.TrimSpace(parts[0])
				actualKey := strings.TrimSpace(parts[1])
				if actualKey == provided {
					c.Set("internal_source", source)
					c.Set("internal_scoped", true)
					c.Next()
					return
				}
				continue
			}

			// Legacy unscoped key (can access all sources)
			if entry == provided {
				c.Set("internal_source", "")
				c.Set("internal_scoped", false)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid internal api key"})
	}
}

func CustomerAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(c.GetHeader("x-api-key"))
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "x-api-key is required"})
			return
		}

		hashed := hashAPIKey(apiKey)
		var user models.ServiceUser
		if err := database.GetDB().Where("customer_api_key = ?", hashed).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid api key"})
			return
		}

		if user.Status != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "user is not active"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func PublicRateLimiter() gin.HandlerFunc {
	windowSeconds := getEnvInt("PUBLIC_RATE_LIMIT_WINDOW_SECONDS", 60)
	maxPerWindow := getEnvInt("PUBLIC_RATE_LIMIT_MAX_REQUEST", 120)
	spam10s := getEnvInt("PUBLIC_SPAM_MAX_PER_10S", 40)
	blockMinutes := getEnvInt("PUBLIC_SPAM_BLOCK_MINUTES", 10)

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/internal/") {
			c.Next()
			return
		}

		ip := clientIP(c)
		now := time.Now()
		rateMutex.Lock()

		if blocked, ok := blockedIPCaches[ip]; ok && now.Before(blocked.until) {
			rateMutex.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "ip blocked due to spam"})
			return
		}

		globalKey := "global:" + ip
		counter := rateCounters[globalKey]
		if counter == nil || now.After(counter.windowEnd) {
			counter = &rateWindow{count: 0, windowEnd: now.Add(time.Duration(windowSeconds) * time.Second)}
			rateCounters[globalKey] = counter
		}
		counter.count++

		spamKey := "spam:" + ip
		spamCounter := rateCounters[spamKey]
		if spamCounter == nil || now.After(spamCounter.windowEnd) {
			spamCounter = &rateWindow{count: 0, windowEnd: now.Add(10 * time.Second)}
			rateCounters[spamKey] = spamCounter
		}
		spamCounter.count++

		if spamCounter.count > spam10s {
			blockedIPCaches[ip] = &blockedIP{until: now.Add(time.Duration(blockMinutes) * time.Minute)}
			rateMutex.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "spam detected"})
			return
		}

		if counter.count > maxPerWindow {
			rateMutex.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "rate limit exceeded"})
			return
		}

		rateMutex.Unlock()
		c.Next()
	}
}

func clientIP(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	if ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return "unknown"
}

func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
