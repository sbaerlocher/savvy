// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DefaultMerchantColor is used when a merchant has no color of its own.
const DefaultMerchantColor = "#3B82F6"

// Merchant represents a retailer or brand in the system
type Merchant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"uniqueIndex:idx_merchants_name,where:deleted_at IS NULL;not null" json:"name"`
	LogoURL   string         `gorm:"type:text" json:"logo_url"`
	Website   string         `gorm:"type:text" json:"website"`
	Color     string         `gorm:"default:#0066CC" json:"color"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
