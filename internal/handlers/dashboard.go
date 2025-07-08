package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/auth"
	"github.com/mindful-minutes/mindful-minutes-api/internal/services"
)

// GetDashboard returns all dashboard data for the authenticated user
// Query parameters:
// - year: Year for yearly progress (defaults to current year)
// - sessions: Number of recent sessions to return (defaults to 5, max 100)
func GetDashboard(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})

		return
	}

	year := parseYear(c)
	sessionLimit := parseSessionLimit(c)

	dashboardData, err := services.GetDashboardData(user, year, sessionLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard data", "details": err.Error()})

		return
	}

	c.JSON(http.StatusOK, dashboardData)
}

// parseYear parses and validates the year query parameter
func parseYear(c *gin.Context) int {
	yearStr := c.Query("year")
	if yearStr == "" {
		return 0 // Will default to current year in service
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1900 || year > 3000 {
		return 0 // Invalid year, will default to current year
	}

	return year
}

// parseSessionLimit parses and validates the sessions limit query parameter
func parseSessionLimit(c *gin.Context) int {
	limitStr := c.Query("sessions")
	if limitStr == "" {
		return 5 // Default limit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		return 5 // Invalid limit, use default
	}

	return limit
}