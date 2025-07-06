package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	App      AppConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port    string
	GinMode string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL string
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	ClerkSecretKey string
	ClerkVerifyURL string
}

// AppConfig holds general application configuration
type AppConfig struct {
	Environment string
}

// Load loads configuration from environment variables and returns a config instance
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:    getEnvWithDefault("PORT", "8080"),
			GinMode: getEnvWithDefault("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			URL: getEnvWithDefault("DATABASE_URL", "postgres://mindful_user:mindful_pass@localhost:5432/mindful_minutes?sslmode=disable"),
		},
		Auth: AuthConfig{
			ClerkSecretKey: getEnvWithDefault("CLERK_SECRET_KEY", ""),
			ClerkVerifyURL: getEnvWithDefault("CLERK_VERIFY_URL", "https://api.clerk.com/v1/verify_token"),
		},
		App: AppConfig{
			Environment: getEnvWithDefault("ENVIRONMENT", "development"),
		},
	}

	// Validate required configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsTest returns true if running in test environment
func (c *Config) IsTest() bool {
	return c.App.Environment == "test"
}

// validateConfig validates that required configuration values are set
func validateConfig(config *Config) error {
	if strings.TrimSpace(config.Database.URL) == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	// Only require Clerk secret key in production
	if config.IsProduction() && config.Auth.ClerkSecretKey == "" {
		return fmt.Errorf("CLERK_SECRET_KEY is required in production")
	}

	// Validate port is a valid number
	if config.Server.Port != "" {
		if _, err := strconv.Atoi(config.Server.Port); err != nil {
			return fmt.Errorf("PORT must be a valid number: %w", err)
		}
	}

	return nil
}

// getEnvWithDefault gets an environment variable with a fallback default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
