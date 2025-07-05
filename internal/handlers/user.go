package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/auth"
)

func GetUserProfile(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}
