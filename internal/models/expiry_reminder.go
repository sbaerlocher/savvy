// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
)

// ExpiryReminderSent tracks which expiry reminders have been sent to prevent duplicates.
type ExpiryReminderSent struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_expiry_reminders_user" json:"user_id"`
	ResourceType string    `gorm:"type:varchar(50);not null" json:"resource_type"`
	ResourceID   uuid.UUID `gorm:"type:uuid;not null" json:"resource_id"`
	DaysBefore   int       `gorm:"not null" json:"days_before"`
	SentAt       time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"sent_at"`
}
