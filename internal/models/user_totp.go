// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserTOTP stores TOTP two-factor authentication data for a user
type UserTOTP struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	Secret      string     `gorm:"type:text;not null" json:"-"` // Encrypted TOTP secret
	BackupCodes string     `gorm:"type:text;not null" json:"-"` // JSON array of bcrypt-hashed backup codes
	Enabled     bool       `gorm:"not null;default:false" json:"enabled"`
	Verified    bool       `gorm:"not null;default:false" json:"verified"`
	EnabledAt   *time.Time `gorm:"type:timestamp with time zone" json:"enabled_at,omitempty"`
	CreatedAt   time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Associations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate ensures a UUID is generated before creating a TOTP record
func (t *UserTOTP) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
