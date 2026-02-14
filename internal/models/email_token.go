// Package models defines GORM database models.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenType constants for email tokens
const (
	TokenTypeEmailVerification       = "email_verification" // #nosec G101 -- constant name, not a credential
	TokenTypePasswordReset           = "password_reset"
	TokenTypeUnsubscribeNotification = "unsubscribe_notifications"
	TokenTypeUnsubscribeReminders    = "unsubscribe_reminders"
)

// EmailToken represents a verification or password reset token.
type EmailToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_email_tokens_user_type" json:"user_id"`
	TokenHash string     `gorm:"type:varchar(64);not null;uniqueIndex:idx_email_tokens_token_hash" json:"-"`
	TokenType string     `gorm:"type:varchar(50);not null;index:idx_email_tokens_user_type" json:"token_type"`
	ExpiresAt time.Time  `gorm:"type:timestamp with time zone;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"type:timestamp with time zone" json:"-"`
	CreatedAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// BeforeCreate ensures a UUID is generated before creating a token
func (t *EmailToken) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// IsExpired returns true if the token has expired.
func (t *EmailToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed returns true if the token has already been used.
func (t *EmailToken) IsUsed() bool {
	return t.UsedAt != nil
}
