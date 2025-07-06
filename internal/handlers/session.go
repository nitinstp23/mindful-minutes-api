package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/auth"
	"github.com/mindful-minutes/mindful-minutes-api/internal/constants"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
)

type CreateSessionRequest struct {
	DurationSeconds int    `json:"duration_seconds" binding:"required,min=1"`
	SessionType     string `json:"session_type" binding:"required"`
	Notes           string `json:"notes"`
}

type GetSessionsResponse struct {
	Sessions []models.Session `json:"sessions"`
	NextID   *uint            `json:"next_id,omitempty"`
	HasMore  bool             `json:"has_more"`
}

var validSessionTypes = map[string]bool{
	constants.SessionTypeMindfulness: true,
	constants.SessionTypeBreathing:   true,
	constants.SessionTypeMetta:       true,
	constants.SessionTypeBodyScan:    true,
	constants.SessionTypeWalking:     true,
	constants.SessionTypeOther:       true,
}

// isValidSessionType checks if a session type is valid
func isValidSessionType(sessionType string) bool {
	return validSessionTypes[sessionType]
}

// CreateSession creates a new meditation session for the authenticated user
func CreateSession(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})

		return
	}

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})

		return
	}

	// Validate session type
	if !isValidSessionType(req.SessionType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session type"})

		return
	}

	session := models.Session{
		UserID:          user.ID,
		DurationSeconds: req.DurationSeconds,
		SessionType:     req.SessionType,
		Notes:           req.Notes,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session", "details": err.Error()})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Session created successfully",
		"session": session,
	})
}

// GetSessions retrieves user's meditation sessions with cursor-based pagination
func GetSessions(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})

		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	var lastID uint
	if lastIDStr := c.Query("last_id"); lastIDStr != "" {
		if id, err := strconv.ParseUint(lastIDStr, 10, 32); err == nil {
			lastID = uint(id)
		}
	}

	// Build query
	query := database.DB.Where("user_id = ?", user.ID)
	
	if lastID > 0 {
		query = query.Where("id < ?", lastID)
	}

	var sessions []models.Session
	if err := query.Order("id DESC").Limit(limit + 1).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sessions", "details": err.Error()})

		return
	}

	// Check if there are more sessions
	hasMore := len(sessions) > limit
	if hasMore {
		sessions = sessions[:limit]
	}

	var nextID *uint
	if hasMore && len(sessions) > 0 {
		nextID = &sessions[len(sessions)-1].ID
	}

	response := GetSessionsResponse{
		Sessions: sessions,
		NextID:   nextID,
		HasMore:  hasMore,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSession soft deletes a meditation session
func DeleteSession(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})

		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})

		return
	}

	// Check if session exists and belongs to user
	var session models.Session
	if err := database.DB.Where("id = ? AND user_id = ?", uint(sessionID), user.ID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})

		return
	}

	// Soft delete the session
	if err := database.DB.Delete(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session", "details": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session deleted successfully",
	})
}