package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/config"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
	"github.com/mindful-minutes/mindful-minutes-api/internal/testutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestVerifyClerkWebhook(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := testutils.SetupTestDB(t)
	database.DB = db
	defer testutils.CleanupTestDB(t, db)

	// Create test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			ClerkSecretKey: "test_secret_key",
		},
	}

	router := gin.New()
	router.POST("/webhooks/clerk", VerifyClerkWebhook(cfg))

	// Helper function to clean database before each test
	cleanDB := func() {
		testutils.TruncateTable(db, "users")
		testutils.TruncateTable(db, "sessions")
	}

	t.Run("return internal server error when secret key is missing", func(t *testing.T) {
		cleanDB()

		// Create config with empty secret key
		emptyCfg := &config.Config{
			Auth: config.AuthConfig{
				ClerkSecretKey: "",
			},
		}

		emptyRouter := gin.New()
		emptyRouter.POST("/webhooks/clerk", VerifyClerkWebhook(emptyCfg))

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer([]byte("{}")))
		w := httptest.NewRecorder()

		emptyRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Clerk secret key not configured")
	})

	t.Run("return bad request when signature header is missing", func(t *testing.T) {
		cleanDB()

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer([]byte("{}")))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing signature header")
	})

	t.Run("return bad request when timestamp header is missing", func(t *testing.T) {
		cleanDB()

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("svix-signature", "v1,test_signature")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing timestamp header")
	})

	t.Run("return unauthorized when signature is invalid", func(t *testing.T) {
		cleanDB()

		payload := `{"type": "user.created", "data": {"id": "test_user"}}`
		timestamp := "1234567890"

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer([]byte(payload)))
		req.Header.Set("svix-signature", "v1,invalid_signature")
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid signature")
	})

	t.Run("return bad request when JSON payload is invalid", func(t *testing.T) {
		cleanDB()

		payload := `{invalid json}`
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(payload, timestamp, "test_secret_key")

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer([]byte(payload)))
		req.Header.Set("svix-signature", signature)
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid JSON payload")
	})

	t.Run("successfully create user when user.created event is received", func(t *testing.T) {
		cleanDB()

		event := ClerkWebhookEvent{
			Type: "user.created",
			Data: ClerkUser{
				ID: "test_user_123",
				EmailAddresses: []ClerkEmailAddress{
					{EmailAddress: "test@example.com", Primary: true},
				},
				FirstName: lo.ToPtr("John"),
				LastName:  lo.ToPtr("Doe"),
				CreatedAt: time.Now().Unix() * 1000,
				UpdatedAt: time.Now().Unix() * 1000,
			},
		}

		payload, _ := json.Marshal(event)
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(string(payload), timestamp, "test_secret_key")

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer(payload))
		req.Header.Set("svix-signature", signature)
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User created successfully")

		// Verify user was created in database
		var user models.User
		err := db.Where("clerk_user_id = ?", "test_user_123").First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "John", *user.FirstName)
		assert.Equal(t, "Doe", *user.LastName)
	})

	t.Run("successfully create user with empty email when no email addresses provided", func(t *testing.T) {
		cleanDB()

		event := ClerkWebhookEvent{
			Type: "user.created",
			Data: ClerkUser{
				ID:             "test_user_123",
				EmailAddresses: []ClerkEmailAddress{},
				FirstName:      lo.ToPtr("John"),
				LastName:       lo.ToPtr("Doe"),
				CreatedAt:      time.Now().Unix() * 1000,
				UpdatedAt:      time.Now().Unix() * 1000,
			},
		}

		payload, _ := json.Marshal(event)
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(string(payload), timestamp, "test_secret_key")

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer(payload))
		req.Header.Set("svix-signature", signature)
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify user was created with empty email
		var user models.User
		err := db.Where("clerk_user_id = ?", "test_user_123").First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, "", user.Email)
	})

	t.Run("successfully update existing user when user.updated event is received", func(t *testing.T) {
		cleanDB()

		// Create existing user
		existingUser := testutils.CreateTestUser("test_user_123")
		db.Create(existingUser)

		event := ClerkWebhookEvent{
			Type: "user.updated",
			Data: ClerkUser{
				ID: "test_user_123",
				EmailAddresses: []ClerkEmailAddress{
					{EmailAddress: "updated@example.com", Primary: true},
				},
				FirstName: lo.ToPtr("Jane"),
				LastName:  lo.ToPtr("Smith"),
				CreatedAt: time.Now().Unix() * 1000,
				UpdatedAt: time.Now().Unix() * 1000,
			},
		}

		payload, _ := json.Marshal(event)
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(string(payload), timestamp, "test_secret_key")

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer(payload))
		req.Header.Set("svix-signature", signature)
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User updated successfully")

		// Verify user was updated in database
		var user models.User
		err := db.Where("clerk_user_id = ?", "test_user_123").First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, "updated@example.com", user.Email)
		assert.Equal(t, "Jane", *user.FirstName)
		assert.Equal(t, "Smith", *user.LastName)
	})

	t.Run("return not found when updating non-existent user", func(t *testing.T) {
		cleanDB()

		event := ClerkWebhookEvent{
			Type: "user.updated",
			Data: ClerkUser{
				ID: "nonexistent_user",
				EmailAddresses: []ClerkEmailAddress{
					{EmailAddress: "test@example.com", Primary: true},
				},
				FirstName: lo.ToPtr("John"),
				LastName:  lo.ToPtr("Doe"),
				CreatedAt: time.Now().Unix() * 1000,
				UpdatedAt: time.Now().Unix() * 1000,
			},
		}

		payload, _ := json.Marshal(event)
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(string(payload), timestamp, "test_secret_key")

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer(payload))
		req.Header.Set("svix-signature", signature)
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("successfully soft delete user when user.deleted event is received", func(t *testing.T) {
		cleanDB()

		// Create existing user
		existingUser := testutils.CreateTestUser("test_user_123")
		db.Create(existingUser)

		event := ClerkWebhookEvent{
			Type: "user.deleted",
			Data: ClerkUser{
				ID: "test_user_123",
			},
		}

		payload, _ := json.Marshal(event)
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(string(payload), timestamp, "test_secret_key")

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer(payload))
		req.Header.Set("svix-signature", signature)
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User deleted successfully")

		// Verify user was soft deleted
		var user models.User
		err := db.Unscoped().Where("clerk_user_id = ?", "test_user_123").First(&user).Error
		assert.NoError(t, err)
		assert.NotNil(t, user.DeletedAt)
	})

	t.Run("return ok with message when unhandled event type is received", func(t *testing.T) {
		cleanDB()

		event := ClerkWebhookEvent{
			Type: "user.unknown",
			Data: ClerkUser{ID: "test_user_123"},
		}

		payload, _ := json.Marshal(event)
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(string(payload), timestamp, "test_secret_key")

		req := httptest.NewRequest("POST", "/webhooks/clerk", bytes.NewBuffer(payload))
		req.Header.Set("svix-signature", signature)
		req.Header.Set("svix-timestamp", timestamp)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Event type not handled")
	})
}

func TestVerifySignature(t *testing.T) {
	secret := "test_secret_key"

	t.Run("return true when signature is valid", func(t *testing.T) {
		payload := "test payload"
		timestamp := "1234567890"
		signature := testutils.GenerateValidClerkSignature(payload, timestamp, secret)

		result := verifySignature([]byte(payload), signature, timestamp, secret)
		assert.True(t, result)
	})

	t.Run("return false when signature is invalid", func(t *testing.T) {
		payload := "test payload"
		timestamp := "1234567890"
		signature := "v1,invalid_signature"

		result := verifySignature([]byte(payload), signature, timestamp, secret)
		assert.False(t, result)
	})

	t.Run("return true when multiple signatures provided with at least one valid", func(t *testing.T) {
		payload := "test payload"
		timestamp := "1234567890"
		validSig := testutils.GenerateValidClerkSignature(payload, timestamp, secret)
		multiSig := "v1,invalid_signature " + validSig

		result := verifySignature([]byte(payload), multiSig, timestamp, secret)
		assert.True(t, result)
	})

	t.Run("return false when signature format is invalid", func(t *testing.T) {
		payload := "test payload"
		timestamp := "1234567890"

		result := verifySignature([]byte(payload), "invalid_format", timestamp, secret)
		assert.False(t, result)
	})

	t.Run("return false when signature is empty", func(t *testing.T) {
		payload := "test payload"
		timestamp := "1234567890"

		result := verifySignature([]byte(payload), "", timestamp, secret)
		assert.False(t, result)
	})
}
