// Package models defines the database models for the savvy system.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// NotificationTypeShareReceived is sent when a resource is shared with the user
	NotificationTypeShareReceived NotificationType = "share_received"
	// NotificationTypeTransferReceived is sent when resource ownership is transferred to the user
	NotificationTypeTransferReceived NotificationType = "transfer_received"
)

// NotificationMetadata represents the JSONB metadata stored with a notification
type NotificationMetadata map[string]interface{}

// Value implements the driver.Valuer interface for JSONB serialization
func (m NotificationMetadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB deserialization
func (m *NotificationMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(NotificationMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		*m = make(NotificationMetadata)
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// Notification represents an in-app notification for users
type Notification struct {
	ID           uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID            `gorm:"type:uuid;not null;index" json:"user_id"`
	Type         NotificationType     `gorm:"type:varchar(50);not null" json:"type"`
	ResourceType string               `gorm:"type:varchar(50);not null" json:"resource_type"` // "card", "voucher", "gift_card"
	ResourceID   uuid.UUID            `gorm:"type:uuid;not null" json:"resource_id"`
	Metadata     NotificationMetadata `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	IsRead       bool                 `gorm:"default:false" json:"is_read"`
	ReadAt       *time.Time           `gorm:"type:timestamp with time zone" json:"read_at,omitempty"`
	CreatedAt    time.Time            `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time            `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt       `gorm:"index" json:"deleted_at,omitempty"`

	// Associations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate ensures a UUID is generated before creating a notification
func (n *Notification) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// MarkAsRead marks the notification as read and sets the read timestamp
func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
}

// GetFromUserID returns the user ID who triggered the notification
func (n *Notification) GetFromUserID() string {
	if id, ok := n.Metadata["from_user_id"].(string); ok {
		return id
	}
	return ""
}

// GetFromUserName returns the name of the user who triggered the notification
func (n *Notification) GetFromUserName() string {
	if name, ok := n.Metadata["from_user_name"].(string); ok {
		return name
	}
	return "Unknown User"
}

// GetPermissions returns the permissions metadata for share notifications
func (n *Notification) GetPermissions() map[string]interface{} {
	if perms, ok := n.Metadata["permissions"].(map[string]interface{}); ok {
		return perms
	}
	return nil
}

// IsShareNotification returns true if this is a share notification
func (n *Notification) IsShareNotification() bool {
	return n.Type == NotificationTypeShareReceived
}

// IsTransferNotification returns true if this is a transfer notification
func (n *Notification) IsTransferNotification() bool {
	return n.Type == NotificationTypeTransferReceived
}
