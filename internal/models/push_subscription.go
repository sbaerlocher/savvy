// Package models defines GORM database models.
package models

import (
	"time"

	"github.com/google/uuid"
)

// PushSubscription represents a Web Push subscription for a user.
type PushSubscription struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_push_subscriptions_user_id" json:"user_id"`
	Endpoint  string    `gorm:"type:text;not null;uniqueIndex" json:"endpoint"`
	P256dhKey string    `gorm:"type:text;not null" json:"p256dh_key"`
	AuthKey   string    `gorm:"type:text;not null" json:"auth_key"` // #nosec G117 -- struct field name, not a hardcoded secret
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
