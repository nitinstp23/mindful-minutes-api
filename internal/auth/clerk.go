package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/mindful-minutes/mindful-minutes-api/internal/database"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
)

type ClerkWebhookEvent struct {
	Data   ClerkUser `json:"data"`
	Object string    `json:"object"`
	Type   string    `json:"type"`
}

type ClerkUser struct {
	ID                string                 `json:"id"`
	EmailAddresses    []ClerkEmailAddress    `json:"email_addresses"`
	FirstName         *string                `json:"first_name"`
	LastName          *string                `json:"last_name"`
	CreatedAt         int64                  `json:"created_at"`
	UpdatedAt         int64                  `json:"updated_at"`
	ExternalAccounts  []ClerkExternalAccount `json:"external_accounts"`
}

type ClerkEmailAddress struct {
	EmailAddress string `json:"email_address"`
	Primary      bool   `json:"primary"`
}

type ClerkExternalAccount struct {
	Provider string `json:"provider"`
}

func VerifyClerkWebhook(c *gin.Context) {
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Clerk secret key not configured"})
		return
	}

	// Get the signature from headers
	signature := c.GetHeader("svix-signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature header"})
		return
	}

	// Get the timestamp from headers
	timestamp := c.GetHeader("svix-timestamp")
	if timestamp == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing timestamp header"})
		return
	}

	// Get the body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify the signature
	if !verifySignature(body, signature, timestamp, secretKey) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse the webhook event
	var event ClerkWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Handle different event types
	switch event.Type {
	case "user.created":
		handleUserCreated(c, event.Data)
	case "user.updated":
		handleUserUpdated(c, event.Data)
	case "user.deleted":
		handleUserDeleted(c, event.Data)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event type not handled"})
	}
}

func verifySignature(payload []byte, signature, timestamp, secret string) bool {
	// Create the signed payload
	signedPayload := timestamp + "." + string(payload)

	// Create HMAC
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Extract signature from header (format: "v1,signature1 v1,signature2")
	signatures := strings.Split(signature, " ")
	for _, sig := range signatures {
		if strings.HasPrefix(sig, "v1,") {
			providedSignature := strings.TrimPrefix(sig, "v1,")
			if hmac.Equal([]byte(expectedSignature), []byte(providedSignature)) {
				return true
			}
		}
	}

	return false
}

func handleUserCreated(c *gin.Context, clerkUser ClerkUser) {
	// Get primary email
	var email string
	for _, emailAddr := range clerkUser.EmailAddresses {
		if emailAddr.Primary {
			email = emailAddr.EmailAddress
			break
		}
	}

	if email == "" && len(clerkUser.EmailAddresses) > 0 {
		email = clerkUser.EmailAddresses[0].EmailAddress
	}

	// Generate ULID
	id := ulid.Make().String()

	// Create user
	user := models.User{
		ID:          id,
		ClerkUserID: clerkUser.ID,
		Email:       email,
		FirstName:   clerkUser.FirstName,
		LastName:    clerkUser.LastName,
	}

	// Save to database
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

func handleUserUpdated(c *gin.Context, clerkUser ClerkUser) {
	// Find existing user
	var user models.User
	if err := database.DB.Where("clerk_user_id = ?", clerkUser.ID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get primary email
	var email string
	for _, emailAddr := range clerkUser.EmailAddresses {
		if emailAddr.Primary {
			email = emailAddr.EmailAddress
			break
		}
	}

	if email == "" && len(clerkUser.EmailAddresses) > 0 {
		email = clerkUser.EmailAddresses[0].EmailAddress
	}

	// Update user
	user.Email = email
	user.FirstName = clerkUser.FirstName
	user.LastName = clerkUser.LastName

	// Save to database
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user_id": user.ID,
	})
}

func handleUserDeleted(c *gin.Context, clerkUser ClerkUser) {
	// Soft delete user
	if err := database.DB.Where("clerk_user_id = ?", clerkUser.ID).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}