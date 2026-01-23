// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog tracks all deletion operations for compliance and auditing
type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	User         *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action       string     `gorm:"not null;index" json:"action"`        // "delete", "hard_delete", "restore"
	ResourceType string     `gorm:"not null;index" json:"resource_type"` // "card", "voucher", "gift_card", "transaction", etc.
	ResourceID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"resource_id"`
	ResourceData string     `gorm:"type:jsonb" json:"resource_data"`    // JSON snapshot of deleted object
	IPAddress    string     `gorm:"type:varchar(45)" json:"ip_address"` // IPv4 or IPv6
	UserAgent    string     `gorm:"type:text" json:"user_agent"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`
}

// TableName overrides the table name
func (AuditLog) TableName() string {
	return "audit_logs"
}
