package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
	"github.com/mindful-minutes/mindful-minutes-api/internal/testutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestGetDashboard(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	// Set the global database connection for services to use
	database.DB = db

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	t.Run("return dashboard data when user is authenticated", func(t *testing.T) {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")

		// Create test user
		user := &models.User{
			ID:           "01JAXXXXXXXXXXXXXXXXXXX1",
			ClerkUserID:  "user_test123",
			Email:        "test@example.com",
			FirstName:    lo.ToPtr("Test"),
			LastName:     lo.ToPtr("User"),
		}
		err := db.Create(user).Error
		assert.NoError(t, err)

		// Create test sessions
		sessions := []models.Session{
			{
				UserID:          user.ID,
				DurationSeconds: 600,
				SessionType:     "mindfulness",
				Notes:          "Morning session",
				CreatedAt:      time.Date(2025, 7, 5, 8, 0, 0, 0, time.UTC),
			},
			{
				UserID:          user.ID,
				DurationSeconds: 900,
				SessionType:     "breathing",
				Notes:          "Evening session",
				CreatedAt:      time.Date(2025, 7, 4, 20, 0, 0, 0, time.UTC),
			},
		}
		for _, session := range sessions {
			err := db.Create(&session).Error
			assert.NoError(t, err)
		}

		// Setup request
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/dashboard", nil)

		// Mock authentication
		c.Set("user", *user)

		// Call handler
		GetDashboard(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response contains expected structure
		body := w.Body.String()
		assert.Contains(t, body, "user")
		assert.Contains(t, body, "streaks")
		assert.Contains(t, body, "weekly_progress")
		assert.Contains(t, body, "yearly_progress")
		assert.Contains(t, body, "recent_sessions")
		assert.Contains(t, body, "test@example.com")
	})

	t.Run("return dashboard data with custom year parameter", func(t *testing.T) {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")

		// Create test user
		user := &models.User{
			ID:           "01JAXXXXXXXXXXXXXXXXXXX2",
			ClerkUserID:  "user_test456",
			Email:        "test2@example.com",
			FirstName:    lo.ToPtr("Test2"),
			LastName:     lo.ToPtr("User2"),
		}
		err := db.Create(user).Error
		assert.NoError(t, err)

		// Setup request with year parameter
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/dashboard?year=2024", nil)

		// Mock authentication
		c.Set("user", *user)

		// Call handler
		GetDashboard(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response structure
		body := w.Body.String()
		assert.Contains(t, body, "yearly_progress")
		assert.Contains(t, body, "test2@example.com")
	})

	t.Run("return dashboard data with custom sessions limit", func(t *testing.T) {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")

		// Create test user
		user := &models.User{
			ID:           "01JAXXXXXXXXXXXXXXXXXXX3",
			ClerkUserID:  "user_test789",
			Email:        "test3@example.com",
			FirstName:    lo.ToPtr("Test3"),
			LastName:     lo.ToPtr("User3"),
		}
		err := db.Create(user).Error
		assert.NoError(t, err)

		// Create multiple test sessions
		for i := 0; i < 10; i++ {
			session := models.Session{
				UserID:          user.ID,
				DurationSeconds: 300 + i*60,
				SessionType:     "mindfulness",
				Notes:          "Session " + string(rune(i+'1')),
				CreatedAt:      time.Now().AddDate(0, 0, -i),
			}
			err := db.Create(&session).Error
			assert.NoError(t, err)
		}

		// Setup request with sessions limit
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/dashboard?sessions=3", nil)

		// Mock authentication
		c.Set("user", *user)

		// Call handler
		GetDashboard(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response contains limited sessions
		body := w.Body.String()
		assert.Contains(t, body, "recent_sessions")
		assert.Contains(t, body, "test3@example.com")
	})

	t.Run("return unauthorized when user is not authenticated", func(t *testing.T) {
		// Setup request without authentication
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/dashboard", nil)

		// Call handler without setting user in context
		GetDashboard(c)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("handle invalid year parameter gracefully", func(t *testing.T) {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")

		// Create test user
		user := &models.User{
			ID:           "01JAXXXXXXXXXXXXXXXXXXX4",
			ClerkUserID:  "user_test000",
			Email:        "test4@example.com",
			FirstName:    lo.ToPtr("Test4"),
			LastName:     lo.ToPtr("User4"),
		}
		err := db.Create(user).Error
		assert.NoError(t, err)

		// Setup request with invalid year
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/dashboard?year=invalid", nil)

		// Mock authentication
		c.Set("user", *user)

		// Call handler
		GetDashboard(c)

		// Should still return OK with default year
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "yearly_progress")
	})

	t.Run("handle invalid sessions parameter gracefully", func(t *testing.T) {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")

		// Create test user
		user := &models.User{
			ID:           "01JAXXXXXXXXXXXXXXXXXXX5",
			ClerkUserID:  "user_test111",
			Email:        "test5@example.com",
			FirstName:    lo.ToPtr("Test5"),
			LastName:     lo.ToPtr("User5"),
		}
		err := db.Create(user).Error
		assert.NoError(t, err)

		// Setup request with invalid sessions limit
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/dashboard?sessions=invalid", nil)

		// Mock authentication
		c.Set("user", *user)

		// Call handler
		GetDashboard(c)

		// Should still return OK with default limit
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "recent_sessions")
	})
}