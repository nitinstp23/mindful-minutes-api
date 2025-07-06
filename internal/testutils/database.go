package testutils

import (
	"log"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mindful-minutes/mindful-minutes-api/internal/config"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	// Load config to get test database URL
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Use test database URL, defaulting to config's database URL with test suffix
	testDBURL := cfg.Database.URL
	if testDBURL == "postgres://mindful_user:mindful_pass@localhost:5432/mindful_minutes?sslmode=disable" {
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

	err = sqlDB.Close()
	if err != nil {
		log.Printf("Failed to close SQL DB: %v", err)
		return
	}
}

func TruncateTable(db *gorm.DB, table string) {
	db.Exec("TRUNCATE TABLE " + table + " CASCADE")
}
