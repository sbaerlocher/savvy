// Package models defines GORM database models.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a server-side session stored in PostgreSQL.
type Session struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID `gorm:"type:uuid;index:idx_sessions_user_id" json:"user_id"`
	TokenHash    string     `gorm:"type:text;not null;uniqueIndex:idx_sessions_token_hash" json:"-"`
	Data         []byte     `gorm:"type:bytea;not null" json:"-"`
	IPAddress    string     `gorm:"type:text;not null;default:''" json:"ip_address"`
	UserAgent    string     `gorm:"type:text;not null;default:''" json:"user_agent"`
	CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastActiveAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"last_active_at"`
	ExpiresAt    time.Time  `gorm:"type:timestamp with time zone;not null;index:idx_sessions_expires_at" json:"expires_at"`
}
