package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID              uint           `json:"id" gorm:"primary_key"`
	UserID          string         `json:"user_id" gorm:"type:char(26);not null;index"`
	DurationSeconds int            `json:"duration_seconds" gorm:"not null"`
	SessionType     string         `json:"session_type" gorm:"not null"`
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
