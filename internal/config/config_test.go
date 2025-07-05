package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("successfully load config with default values", func(t *testing.T) {
		t.Setenv("PORT", "")
		t.Setenv("GIN_MODE", "")
		t.Setenv("CLERK_SECRET_KEY", "")
		t.Setenv("ENVIRONMENT", "")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "debug", cfg.Server.GinMode)
		assert.Equal(t, "postgres://mindful_user:mindful_pass@localhost:5432/mindful_minutes?sslmode=disable", cfg.Database.URL)
		assert.Equal(t, "", cfg.Auth.ClerkSecretKey)
		assert.Equal(t, "development", cfg.App.Environment)
	})

	t.Run("successfully load config with environment variables", func(t *testing.T) {
		t.Setenv("PORT", "9000")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
		t.Setenv("CLERK_SECRET_KEY", "test_secret")
		t.Setenv("ENVIRONMENT", "production")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "9000", cfg.Server.Port)
		assert.Equal(t, "release", cfg.Server.GinMode)
		assert.Equal(t, "postgres://test:test@localhost:5432/test", cfg.Database.URL)
		assert.Equal(t, "test_secret", cfg.Auth.ClerkSecretKey)
		assert.Equal(t, "production", cfg.App.Environment)
	})

	t.Run("return error when database URL is explicitly empty", func(t *testing.T) {
		t.Setenv("DATABASE_URL", " ") // Set to space which will be trimmed to empty

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "DATABASE_URL is required")
	})

	t.Run("return error when port is invalid", func(t *testing.T) {
		t.Setenv("PORT", "invalid_port")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "PORT must be a valid number")
	})

	t.Run("return error when clerk secret key missing in production", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		t.Setenv("CLERK_SECRET_KEY", "")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "CLERK_SECRET_KEY is required in production")
	})

	t.Run("allow empty clerk secret key in development", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "development")
		t.Setenv("CLERK_SECRET_KEY", "")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "", cfg.Auth.ClerkSecretKey)
	})
}

func TestConfigMethods(t *testing.T) {
	t.Run("IsProduction returns true for production environment", func(t *testing.T) {
		cfg := &Config{
			App: AppConfig{
				Environment: "production",
			},
		}

		assert.True(t, cfg.IsProduction())
		assert.False(t, cfg.IsDevelopment())
		assert.False(t, cfg.IsTest())
	})

	t.Run("IsDevelopment returns true for development environment", func(t *testing.T) {
		cfg := &Config{
			App: AppConfig{
				Environment: "development",
			},
		}

		assert.False(t, cfg.IsProduction())
		assert.True(t, cfg.IsDevelopment())
		assert.False(t, cfg.IsTest())
	})

	t.Run("IsTest returns true for test environment", func(t *testing.T) {
		cfg := &Config{
			App: AppConfig{
				Environment: "test",
			},
		}

		assert.False(t, cfg.IsProduction())
		assert.False(t, cfg.IsDevelopment())
		assert.True(t, cfg.IsTest())
	})
}

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("return environment variable when set", func(t *testing.T) {
		t.Setenv("TEST_VAR", "test_value")

		result := getEnvWithDefault("TEST_VAR", "default_value")

		assert.Equal(t, "test_value", result)
	})

	t.Run("return default value when environment variable not set", func(t *testing.T) {
		t.Setenv("TEST_VAR", "")

		result := getEnvWithDefault("TEST_VAR", "default_value")

		assert.Equal(t, "default_value", result)
	})

	t.Run("return default value when environment variable is empty", func(t *testing.T) {
		t.Setenv("TEST_VAR", "")

		result := getEnvWithDefault("TEST_VAR", "default_value")

		assert.Equal(t, "default_value", result)
	})
}
