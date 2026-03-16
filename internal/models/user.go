// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account in the system
type User struct {
	ID                        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email                     string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash              string     `gorm:"not null" json:"-"`
	FirstName                 string     `gorm:"not null" json:"first_name"`
	LastName                  string     `gorm:"not null" json:"last_name"`
	Role                      string     `gorm:"default:user;not null" json:"role"`
	AuthProvider              string     `gorm:"default:local;not null" json:"auth_provider"`      // "local" or "oauth"
	Language                  string     `gorm:"default:de;not null" json:"language"`              // User's preferred language (de, en, fr)
	PushNotificationsEnabled  bool       `gorm:"default:false" json:"push_notifications_enabled"`  // Global push notifications channel (enabled on first push subscription)
	EmailNotificationsEnabled bool       `gorm:"default:false" json:"email_notifications_enabled"` // Global email notifications channel (enabled on email verification)
	PushRemindersEnabled      bool       `gorm:"default:false" json:"push_reminders_enabled"`      // Push for expiry & validity start
	PushSharingEnabled        bool       `gorm:"default:false" json:"push_sharing_enabled"`        // Push for share & transfer
	EmailRemindersEnabled     bool       `gorm:"default:false" json:"email_reminders_enabled"`     // Email for expiry & validity start
	EmailSharingEnabled       bool       `gorm:"default:false" json:"email_sharing_enabled"`       // Email for share & transfer
	EmailVerified             bool       `gorm:"default:false" json:"email_verified"`              // Whether email is verified
	EmailVerifiedAt           *time.Time `gorm:"" json:"-"`                                        // When email was verified
	PasswordChangedAt         *time.Time `gorm:"" json:"-"`                                        // When password was last changed (for session invalidation)
	FailedLoginAttempts       int        `gorm:"default:0" json:"-"`                               // Count of failed login attempts
	LockedUntil               *time.Time `gorm:"index" json:"-"`                                   // Account lock expiry time
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// BeforeCreate ensures a UUID is generated before creating a user
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// DisplayName returns the full name of the user
func (u *User) DisplayName() string {
	return u.FirstName + " " + u.LastName
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsOAuthUser returns true if the user authenticated via OAuth
func (u *User) IsOAuthUser() bool {
	return u.AuthProvider == "oauth"
}
