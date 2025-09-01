package database

import (
	"fmt"
	"log"
	"os"

	"genfity-chat-ai/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	// Auto migrate the schema
	err = DB.AutoMigrate(
		&models.GenEventWebhook{},
		&models.WhatsAppMessage{},
		&models.WhatsAppReadReceipt{},
		&models.WhatsAppPresence{},
		&models.WhatsAppChatPresence{},
		&models.WhatsAppHistorySync{},
		&models.WhatsAppSession{},
		&models.WhatsAppMessageStatus{},
		&models.UserSettings{},
		&models.ChatRoom{},
		&models.ChatMessage{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
