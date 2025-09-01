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
var TransactionalDB *gorm.DB

// InitDatabase initializes both database connections
func InitDatabase() {
	// Initialize primary database (for webhook events)
	initPrimaryDatabase()

	// Initialize transactional database (for subscription/user data)
	initTransactionalDatabase()
}

// initPrimaryDatabase initializes the primary database connection for webhook events
func initPrimaryDatabase() {
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
		log.Fatal("Failed to connect to primary database:", err)
	}

	log.Println("Primary database connected successfully")

	// Auto migrate the schema for webhook events
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
		log.Fatal("Failed to migrate primary database:", err)
	}

	log.Println("Primary database migration completed")
}

// initTransactionalDatabase initializes the transactional database connection
func initTransactionalDatabase() {
	host := os.Getenv("TRANSACTIONAL_DB_HOST")
	port := os.Getenv("TRANSACTIONAL_DB_PORT")
	user := os.Getenv("TRANSACTIONAL_DB_USER")
	password := os.Getenv("TRANSACTIONAL_DB_PASSWORD")
	dbname := os.Getenv("TRANSACTIONAL_DB_NAME")
	sslmode := os.Getenv("TRANSACTIONAL_DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	TransactionalDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to transactional database:", err)
	}

	log.Println("Transactional database connected successfully")

	// Check if required tables exist (read-only gateway)
	checkRequiredTables()
}

// checkRequiredTables verifies that all required tables exist
func checkRequiredTables() {
	requiredTables := []string{
		"WhatsAppSession",           // Table untuk mapping token ke user
		"WhatsappApiPackage",        // Table untuk package configuration
		"ServicesWhatsappCustomers", // Table untuk subscription users
		"WhatsAppMessageStats",      // Table untuk message tracking
	}

	for _, tableName := range requiredTables {
		if !TransactionalDB.Migrator().HasTable(tableName) {
			log.Printf("Warning: Required table '%s' does not exist in database", tableName)
		} else {
			log.Printf("✓ Table '%s' exists", tableName)
		}
	}

	log.Println("Database table check completed")
}

// GetDB returns the primary database instance
func GetDB() *gorm.DB {
	return DB
}

// GetTransactionalDB returns the transactional database instance
func GetTransactionalDB() *gorm.DB {
	return TransactionalDB
}
