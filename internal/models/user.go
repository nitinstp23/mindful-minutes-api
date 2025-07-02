package models

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type User struct {
	ID          ulid.ULID      `json:"id" gorm:"type:char(26);primary_key"`
	ClerkUserID string         `json:"clerk_user_id" gorm:"unique;not null"`
	Email       string         `json:"email" gorm:"not null"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Sessions []Session `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID.Compare(ulid.ULID{}) == 0 {
		u.ID = ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	}
	return nil
}