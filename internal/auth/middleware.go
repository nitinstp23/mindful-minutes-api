package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
)

type ClerkJWTClaims struct {
	Sub string `json:"sub"`
	Iss string `json:"iss"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Azp string `json:"azp"`
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Verify token with Clerk
		clerkUserID, err := verifyClerkToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := database.DB.Where("clerk_user_id = ?", clerkUserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("clerk_user_id", clerkUserID)

		c.Next()
	}
}

func verifyClerkToken(token string) (string, error) {
	// In a real implementation, you would verify the JWT token against Clerk's JWKS endpoint
	// For now, we'll implement a simple verification mechanism
	
	// Get Clerk secret key
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		return "", fmt.Errorf("clerk secret key not configured")
	}

	// For development/testing, we'll use a simplified token verification
	// In production, you should use proper JWT library like golang-jwt/jwt
	
	// Make HTTP request to Clerk's verification endpoint
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.clerk.com/v1/verify_token", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token verification failed")
	}

	// Parse response to get user ID
	var response struct {
		Sub string `json:"sub"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Sub, nil
}

func GetCurrentUser(c *gin.Context) *models.User {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(models.User); ok {
			return &u
		}
	}
	return nil
}

func GetCurrentUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}