package testutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/mindful-minutes/mindful-minutes-api/internal/models"
)

func CreateTestUser(clerkUserID string) *models.User {
	return &models.User{
		ID:          ulid.Make().String(),
		ClerkUserID: clerkUserID,
		Email:       "test@example.com",
		FirstName:   lo.ToPtr("John"),
		LastName:    lo.ToPtr("Doe"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func CreateTestSession(userID string) *models.Session {
	return &models.Session{
		UserID:          userID,
		DurationSeconds: 600,
		SessionType:     "mindfulness",
		Notes:           "Test session",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func GenerateValidClerkSignature(payload, timestamp, secret string) string {
	signedPayload := timestamp + "." + payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signedPayload))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("v1,%s", signature)
}