package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/constants"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/handlers"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
	"github.com/mindful-minutes/mindful-minutes-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutils.SetupTestDB(t)
	database.DB = db
	defer testutils.CleanupTestDB(t, db)

	router := gin.New()
	router.POST("/sessions", handlers.CreateSession)

	// Helper function to clean database before each test
	cleanDB := func() {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")
	}

	t.Run("successfully create session when valid data provided", func(t *testing.T) {
		cleanDB()

		testUser := testutils.CreateTestUser("test_clerk_id")
		db.Create(testUser)

		requestBody := map[string]interface{}{
			"duration_seconds": 600,
			"session_type":     constants.SessionTypeMindfulness,
			"notes":            "Morning meditation",
		}

		jsonData, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		handlers.CreateSession(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Session created successfully")
		assert.Contains(t, w.Body.String(), strconv.Itoa(requestBody["duration_seconds"].(int)))
		assert.Contains(t, w.Body.String(), requestBody["session_type"].(string))
	})

	t.Run("return unauthorized when user not in context", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"duration_seconds": 600,
			"session_type":     constants.SessionTypeMindfulness,
		}

		jsonData, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handlers.CreateSession(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("return bad request when invalid session type provided", func(t *testing.T) {
		cleanDB()

		testUser := testutils.CreateTestUser("test_clerk_id")
		db.Create(testUser)

		requestBody := map[string]interface{}{
			"duration_seconds": 600,
			"session_type":     "invalid_type",
			"notes":            "Test session",
		}

		jsonData, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		handlers.CreateSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid session type")
	})

	t.Run("return bad request when duration is missing", func(t *testing.T) {
		cleanDB()

		testUser := testutils.CreateTestUser("test_clerk_id")
		db.Create(testUser)

		requestBody := map[string]interface{}{
			"session_type": constants.SessionTypeMindfulness,
		}

		jsonData, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		handlers.CreateSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request data")
	})

	t.Run("return bad request when duration is zero or negative", func(t *testing.T) {
		cleanDB()

		testUser := testutils.CreateTestUser("test_clerk_id")
		db.Create(testUser)

		requestBody := map[string]interface{}{
			"duration_seconds": 0,
			"session_type":     constants.SessionTypeMindfulness,
		}

		jsonData, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		handlers.CreateSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request data")
	})
}

func TestGetSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutils.SetupTestDB(t)
	database.DB = db
	defer testutils.CleanupTestDB(t, db)

	router := gin.New()
	router.GET("/sessions", handlers.GetSessions)

	// Helper function to clean database and create test data
	setupTestData := func() (*models.User, []models.Session) {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")

		testUser := testutils.CreateTestUser("test_clerk_id")
		db.Create(testUser)

		// Create test sessions
		sessions := []models.Session{
			{UserID: testUser.ID, DurationSeconds: 600, SessionType: constants.SessionTypeMindfulness, Notes: "Session 1"},
			{UserID: testUser.ID, DurationSeconds: 900, SessionType: constants.SessionTypeBreathing, Notes: "Session 2"},
			{UserID: testUser.ID, DurationSeconds: 300, SessionType: constants.SessionTypeMetta, Notes: "Session 3"},
		}

		for i := range sessions {
			db.Create(&sessions[i])
		}

		return testUser, sessions
	}

	t.Run("successfully retrieve sessions for user", func(t *testing.T) {
		testUser, _ := setupTestData()

		req := httptest.NewRequest("GET", "/sessions", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		handlers.GetSessions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.GetSessionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Sessions, 3)
		assert.False(t, response.HasMore)
		assert.Nil(t, response.NextID)
	})

	t.Run("return unauthorized when user not in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sessions", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handlers.GetSessions(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("handle pagination correctly", func(t *testing.T) {
		testUser, _ := setupTestData()

		// Request with limit of 2
		req := httptest.NewRequest("GET", "/sessions?limit=2", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		handlers.GetSessions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.GetSessionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Sessions, 2)
		assert.True(t, response.HasMore)
		assert.NotNil(t, response.NextID)
	})
}

func TestDeleteSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutils.SetupTestDB(t)
	database.DB = db
	defer testutils.CleanupTestDB(t, db)

	router := gin.New()
	router.DELETE("/sessions/:id", handlers.DeleteSession)

	// Helper function to clean database and create test data
	setupTestData := func() (*models.User, models.Session) {
		testutils.TruncateTable(db, "sessions")
		testutils.TruncateTable(db, "users")

		testUser := testutils.CreateTestUser("test_clerk_id")
		db.Create(testUser)

		session := models.Session{
			UserID:          testUser.ID,
			DurationSeconds: 600,
			SessionType:     constants.SessionTypeMindfulness,
			Notes:           "Test session",
		}
		db.Create(&session)

		return testUser, session
	}

	t.Run("successfully delete session when valid ID provided", func(t *testing.T) {
		testUser, session := setupTestData()

		req := httptest.NewRequest("DELETE", "/sessions/"+strconv.Itoa(int(session.ID)), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(session.ID))}}

		handlers.DeleteSession(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Session deleted successfully")
	})

	t.Run("return unauthorized when user not in context", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/sessions/1", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handlers.DeleteSession(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("return bad request when invalid session ID provided", func(t *testing.T) {
		testUser := testutils.CreateTestUser("test_clerk_id")

		req := httptest.NewRequest("DELETE", "/sessions/invalid", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handlers.DeleteSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid session ID")
	})

	t.Run("return not found when session does not exist or belongs to different user", func(t *testing.T) {
		testUser := testutils.CreateTestUser("test_clerk_id")
		db.Create(testUser)

		req := httptest.NewRequest("DELETE", "/sessions/999", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handlers.DeleteSession(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Session not found")
	})
}