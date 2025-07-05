package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetUserProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/user/profile", GetUserProfile)

	t.Run("return user profile when user exists in context", func(t *testing.T) {
		testUser := testutils.CreateTestUser("test_clerk_id")

		req := httptest.NewRequest("GET", "/user/profile", nil)
		w := httptest.NewRecorder()

		// Create context with user
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		GetUserProfile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testUser.ID)
		assert.Contains(t, w.Body.String(), testUser.Email)
		assert.Contains(t, w.Body.String(), "John") // FirstName
		assert.Contains(t, w.Body.String(), "Doe")  // LastName
	})

	t.Run("return unauthorized when user not in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/profile", nil)
		w := httptest.NewRecorder()

		// Create context without user
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		GetUserProfile(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("return unauthorized when user has wrong type in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/profile", nil)
		w := httptest.NewRecorder()

		// Create context with wrong user type
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", "invalid_user_type")

		GetUserProfile(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("return user profile with nil first and last names", func(t *testing.T) {
		testUser := testutils.CreateTestUser("test_clerk_id")
		testUser.FirstName = nil
		testUser.LastName = nil

		req := httptest.NewRequest("GET", "/user/profile", nil)
		w := httptest.NewRecorder()

		// Create context with user
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		GetUserProfile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testUser.ID)
		assert.Contains(t, w.Body.String(), testUser.Email)
		assert.Contains(t, w.Body.String(), "null") // FirstName should be null
		assert.Contains(t, w.Body.String(), "null") // LastName should be null
	})

	t.Run("return user profile with empty email", func(t *testing.T) {
		testUser := testutils.CreateTestUser("test_clerk_id")
		testUser.Email = ""

		req := httptest.NewRequest("GET", "/user/profile", nil)
		w := httptest.NewRecorder()

		// Create context with user
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", *testUser)

		GetUserProfile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testUser.ID)
		assert.Contains(t, w.Body.String(), `"email":""`)
	})
}
