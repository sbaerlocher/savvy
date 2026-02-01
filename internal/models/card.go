// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Card represents a savvy card in the system
type Card struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	User         *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	MerchantID   *uuid.UUID     `gorm:"type:uuid;index" json:"merchant_id"`
	Merchant     *Merchant      `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	MerchantName string         `gorm:"default:''" json:"merchant_name"` // Retailer as fallback for free text
	Program      string         `gorm:"not null" json:"program"`         // Savvy program name (e.g. Cumulus, Supercard)
	CardNumber   string         `gorm:"uniqueIndex;not null" json:"card_number"`
	BarcodeType  string         `gorm:"default:CODE128" json:"barcode_type"`
	Status       string         `gorm:"default:active" json:"status"`
	Notes        string         `gorm:"type:text" json:"notes"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GetColor returns the color from the merchant if available, otherwise returns default
func (c *Card) GetColor() string {
	if c.Merchant != nil && c.Merchant.Color != "" {
		return c.Merchant.Color
	}
	return "#0066CC"
}

// CardShare represents a shared card with permissions
type CardShare struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CardID         uuid.UUID      `gorm:"type:uuid;index;not null" json:"card_id"`
	Card           *Card          `gorm:"foreignKey:CardID" json:"card,omitempty"`
	SharedWithID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"shared_with_id"`
	SharedWithUser *User          `gorm:"foreignKey:SharedWithID" json:"shared_with_user,omitempty"`
	CanEdit        bool           `gorm:"default:false" json:"can_edit"`
	CanDelete      bool           `gorm:"default:false" json:"can_delete"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
