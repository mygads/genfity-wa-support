package database

import (
	"fmt"
	"log"
	"os"

	"genfity-wa-support/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // No logging for cleaner output
	})
	if err != nil {
		log.Fatal("Failed to connect to primary database:", err)
	}

	log.Println("Primary database connected successfully")

	// Auto migrate all primary tables (create if not exist)
	if err := autoMigratePrimaryTables(); err != nil {
		log.Fatal("Failed to migrate primary database:", err)
	}
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
	TransactionalDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // No logging for cleaner output
	})
	if err != nil {
		log.Fatal("Failed to connect to transactional database:", err)
	}

	log.Println("Transactional database connected successfully")

	// Auto migrate contact and campaign tables in transactional database
	// err = TransactionalDB.AutoMigrate(
	// 	&models.WhatsAppContact{},
	// 	&models.Campaign{},
	// 	&models.BulkCampaign{},
	// 	&models.BulkCampaignItem{},
	// 	&models.WhatsAppMessageStats{},
	// )
	// if err != nil {
	// 	log.Printf("Warning: Failed to migrate tables in transactional database: %v", err)
	// } else {
	// 	log.Println("Contact and campaign tables migration completed in transactional database")
	// }

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

// autoMigratePrimaryTables checks and migrates only tables that don't exist
func autoMigratePrimaryTables() error {
	tables := []struct {
		name  string
		model interface{}
	}{
		{"gen_event_webhooks", &models.GenEventWebhook{}},
		{"whats_app_messages", &models.WhatsAppMessage{}},
		{"whats_app_read_receipts", &models.WhatsAppReadReceipt{}},
		{"whats_app_presences", &models.WhatsAppPresence{}},
		{"whats_app_chat_presences", &models.WhatsAppChatPresence{}},
		{"whats_app_history_syncs", &models.WhatsAppHistorySync{}},
		{"whats_app_sessions", &models.WhatsAppSession{}},
		{"whats_app_message_statuses", &models.WhatsAppMessageStatus{}},
		{"user_settings", &models.UserSettings{}},
		{"chat_rooms", &models.ChatRoom{}},
		{"chat_messages", &models.ChatMessage{}},
	}

	migratedCount := 0
	skippedCount := 0

	log.Println("Checking primary database tables...")

	for _, table := range tables {
		if !DB.Migrator().HasTable(table.model) {
			log.Printf("Table '%s' not found, creating...", table.name)
			err := DB.AutoMigrate(table.model)
			if err != nil {
				return fmt.Errorf("failed to migrate table %s: %v", table.name, err)
			}
			log.Printf("✓ Created table: %s", table.name)
			migratedCount++
		} else {
			log.Printf("✓ Table '%s' already exists, skipping", table.name)
			skippedCount++
		}
	}

	if migratedCount > 0 {
		log.Printf("Primary database migration completed: %d tables created, %d tables skipped", migratedCount, skippedCount)
	} else {
		log.Printf("All primary database tables already exist (%d tables), no migration needed", skippedCount)
	}
	return nil
}
