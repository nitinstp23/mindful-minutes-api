package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mindful-minutes/mindful-minutes-api/internal/config"
)

func TestLoad(t *testing.T) {
	t.Run("successfully load config with environment variables", func(t *testing.T) {
		t.Setenv("PORT", "9000")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
		t.Setenv("CLERK_SECRET_KEY", "test_secret")
		t.Setenv("ENVIRONMENT", "production")

		cfg, err := config.Load()

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

		cfg, err := config.Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "DATABASE_URL is required")
	})

	t.Run("return error when port is invalid", func(t *testing.T) {
		t.Setenv("PORT", "invalid_port")

		cfg, err := config.Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "PORT must be a valid number")
	})

	t.Run("return error when clerk secret key missing in production", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		t.Setenv("CLERK_SECRET_KEY", "")

		cfg, err := config.Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "CLERK_SECRET_KEY is required in production")
	})

	t.Run("allow empty clerk secret key in development", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "development")
		t.Setenv("CLERK_SECRET_KEY", "")

		cfg, err := config.Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "", cfg.Auth.ClerkSecretKey)
	})
}

func TestConfigMethods(t *testing.T) {
	t.Run("IsProduction returns true for production environment", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Environment: "production",
			},
		}

		assert.True(t, cfg.IsProduction())
		assert.False(t, cfg.IsDevelopment())
		assert.False(t, cfg.IsTest())
	})

	t.Run("IsDevelopment returns true for development environment", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Environment: "development",
			},
		}

		assert.False(t, cfg.IsProduction())
		assert.True(t, cfg.IsDevelopment())
		assert.False(t, cfg.IsTest())
	})

	t.Run("IsTest returns true for test environment", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Environment: "test",
			},
		}

		assert.False(t, cfg.IsProduction())
		assert.False(t, cfg.IsDevelopment())
		assert.True(t, cfg.IsTest())
	})
}
