package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"genfity-wa-support/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

var jakartaLoc = mustLoadJakarta()

func mustLoadJakarta() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Failed to load Asia/Jakarta timezone, fallback UTC: %v", err)
		return time.UTC
	}
	return loc
}

// InitDatabase initializes database connection
func InitDatabase() {
	initDatabase()
}

func initDatabase() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	if err := autoMigrateTables(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

func autoMigrateTables() error {
	return DB.AutoMigrate(
		&models.ServiceUser{},
		&models.UserSubscription{},
		&models.WhatsAppSession{},
		&models.SessionMessageStat{},
		&models.SessionContact{},
	)
}

func GetDB() *gorm.DB {
	return DB
}

func StartSubscriptionExpiryCron() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			nowWIB := time.Now().In(jakartaLoc)
			result := DB.Model(&models.UserSubscription{}).
				Where("status = ? AND expires_at <= ?", models.SubscriptionActive, nowWIB).
				Updates(map[string]interface{}{"status": models.SubscriptionExpired})
			if result.Error != nil {
				log.Printf("Subscription expiry cron error: %v", result.Error)
			}
		}
	}()
}
