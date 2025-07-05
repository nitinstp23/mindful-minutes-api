package testutils

import (
	"log"
	"os"
	"testing"

	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	// Use test database URL or in-memory SQLite for testing
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://mindful_user:mindful_pass@localhost:5432/mindful_minutes_test?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Session{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func CleanupTestDB(t *testing.T, db *gorm.DB) {
	// Clean up test data
	db.Exec("DELETE FROM sessions")
	db.Exec("DELETE FROM users")

	// Close the database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get SQL DB: %v", err)
		return
	}
	sqlDB.Close()
}

func TruncateTable(db *gorm.DB, table string) {
	db.Exec("TRUNCATE TABLE " + table + " CASCADE")
}
