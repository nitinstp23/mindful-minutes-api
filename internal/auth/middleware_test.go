package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jarcoal/httpmock"
	"github.com/mindful-minutes/mindful-minutes-api/internal/config"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := testutils.SetupTestDB(t)
	database.DB = db
	defer testutils.CleanupTestDB(t, db)

	// Create test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			ClerkSecretKey: "test_secret_key",
			ClerkVerifyURL: "https://api.clerk.com/v1/verify_token",
		},
	}

	router := gin.New()
	router.Use(AuthMiddleware(cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Helper function to clean database before each test
	cleanDB := func() {
		testutils.TruncateTable(db, "users")
		testutils.TruncateTable(db, "sessions")
	}

	// Setup httpmock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("return unauthorized when authorization header is missing", func(t *testing.T) {
		cleanDB()

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Missing authorization header")
	})

	t.Run("return unauthorized when authorization header format is invalid", func(t *testing.T) {
		cleanDB()

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("return unauthorized when bearer token is missing", func(t *testing.T) {
		cleanDB()

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("return unauthorized when token verification fails", func(t *testing.T) {
		cleanDB()
		httpmock.Reset()

		// Mock failed Clerk API response
		httpmock.RegisterResponder("GET", cfg.Auth.ClerkVerifyURL,
			httpmock.NewStringResponder(401, "Unauthorized"))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("return unauthorized when user not found in database", func(t *testing.T) {
		cleanDB()
		httpmock.Reset()

		// Mock successful Clerk API response
		httpmock.RegisterResponder("GET", cfg.Auth.ClerkVerifyURL,
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"sub": "user_12345",
			}))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid_token_but_user_not_in_db")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("return unauthorized when secret key is missing", func(t *testing.T) {
		cleanDB()

		// Create config with empty secret key
		emptyCfg := &config.Config{
			Auth: config.AuthConfig{
				ClerkSecretKey: "",
				ClerkVerifyURL: "https://api.clerk.com/v1/verify_token",
			},
		}

		emptyRouter := gin.New()
		emptyRouter.Use(AuthMiddleware(emptyCfg))
		emptyRouter.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid_token")
		w := httptest.NewRecorder()

		emptyRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("return success when token is valid and user exists", func(t *testing.T) {
		cleanDB()
		httpmock.Reset()

		// Create test user in database
		testUser := testutils.CreateTestUser("user_12345")
		db.Create(testUser)

		// Mock successful Clerk API response
		httpmock.RegisterResponder("GET", cfg.Auth.ClerkVerifyURL,
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"sub": "user_12345",
			}))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid_token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}

func TestGetCurrentUser(t *testing.T) {
	t.Run("return user when user exists in context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		testUser := testutils.CreateTestUser("test_clerk_id")
		c.Set("user", *testUser)

		user := GetCurrentUser(c)

		assert.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.ClerkUserID, user.ClerkUserID)
	})

	t.Run("return nil when user not in context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		user := GetCurrentUser(c)

		assert.Nil(t, user)
	})

	t.Run("return nil when user has wrong type in context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		c.Set("user", "invalid_user_type")

		user := GetCurrentUser(c)

		assert.Nil(t, user)
	})
}

func TestGetCurrentUserID(t *testing.T) {
	t.Run("return user ID when user ID exists in context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedID := "test_user_id_123"
		c.Set("user_id", expectedID)

		userID := GetCurrentUserID(c)

		assert.Equal(t, expectedID, userID)
	})

	t.Run("return empty string when user ID not in context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userID := GetCurrentUserID(c)

		assert.Equal(t, "", userID)
	})

	t.Run("return empty string when user ID has wrong type in context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		c.Set("user_id", 123) // Wrong type

		userID := GetCurrentUserID(c)

		assert.Equal(t, "", userID)
	})
}

func TestVerifyClerkToken(t *testing.T) {
	// Setup httpmock for this test suite
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("return error when secret key is missing", func(t *testing.T) {
		cfg := &config.Config{
			Auth: config.AuthConfig{
				ClerkSecretKey:   "",
				ClerkVerifyURL:   "https://api.clerk.com/v1/verify_token",
			},
		}

		_, err := verifyClerkToken("test_token", cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clerk secret key not configured")
	})

	t.Run("return error when token verification fails", func(t *testing.T) {
		httpmock.Reset()
		cfg := &config.Config{
			Auth: config.AuthConfig{
				ClerkSecretKey: "test_secret_key",
				ClerkVerifyURL: "https://api.clerk.com/v1/verify_token",
			},
		}

		// Mock failed Clerk API response
		httpmock.RegisterResponder("GET", cfg.Auth.ClerkVerifyURL,
			httpmock.NewStringResponder(401, "Unauthorized"))

		_, err := verifyClerkToken("invalid_token", cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token verification failed")
	})

	t.Run("return user ID when token is empty but API responds successfully", func(t *testing.T) {
		httpmock.Reset()
		cfg := &config.Config{
			Auth: config.AuthConfig{
				ClerkSecretKey: "test_secret_key",
				ClerkVerifyURL: "https://api.clerk.com/v1/verify_token",
			},
		}

		// Mock successful Clerk API response
		httpmock.RegisterResponder("GET", cfg.Auth.ClerkVerifyURL,
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"sub": "user_12345",
			}))

		userID, err := verifyClerkToken("", cfg)

		// Empty token still makes the request and can succeed if API allows it
		assert.NoError(t, err)
		assert.Equal(t, "user_12345", userID)
	})

	t.Run("return user ID when token is valid", func(t *testing.T) {
		httpmock.Reset()
		cfg := &config.Config{
			Auth: config.AuthConfig{
				ClerkSecretKey: "test_secret_key",
				ClerkVerifyURL: "https://api.clerk.com/v1/verify_token",
			},
		}

		// Mock successful Clerk API response
		httpmock.RegisterResponder("GET", cfg.Auth.ClerkVerifyURL,
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"sub": "user_12345",
			}))

		userID, err := verifyClerkToken("valid_token", cfg)

		assert.NoError(t, err)
		assert.Equal(t, "user_12345", userID)
	})

	t.Run("return error when response JSON is invalid", func(t *testing.T) {
		httpmock.Reset()
		cfg := &config.Config{
			Auth: config.AuthConfig{
				ClerkSecretKey: "test_secret_key",
				ClerkVerifyURL: "https://api.clerk.com/v1/verify_token",
			},
		}

		// Mock response with invalid JSON
		httpmock.RegisterResponder("GET", cfg.Auth.ClerkVerifyURL,
			httpmock.NewStringResponder(200, "invalid json"))

		_, err := verifyClerkToken("valid_token", cfg)

		assert.Error(t, err)
	})
}
