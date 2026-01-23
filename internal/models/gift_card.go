// Package models defines the database models for the savvy system.
package models

import (
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GiftCard represents a prepaid gift card in the system
type GiftCard struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	User           *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	MerchantID     *uuid.UUID     `gorm:"type:uuid;index" json:"merchant_id"`
	Merchant       *Merchant      `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	MerchantName   string         `gorm:"default:''" json:"merchant_name"` // Retailer as fallback
	CardNumber     string         `gorm:"uniqueIndex;not null" json:"card_number"`
	InitialBalance float64        `gorm:"not null" json:"initial_balance"`
	CurrentBalance float64        `gorm:"not null" json:"current_balance"` // Cached balance (auto-updated by trigger)
	Currency       string         `gorm:"default:CHF" json:"currency"`
	PIN            string         `json:"pin"`
	ExpiresAt      *time.Time     `json:"expires_at"`
	Status         string         `gorm:"default:active" json:"status"`
	BarcodeType    string         `gorm:"default:CODE128" json:"barcode_type"`
	Notes          string         `gorm:"type:text" json:"notes"`
	Color          string         `gorm:"default:#DC2626" json:"color"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Transactions []GiftCardTransaction `gorm:"foreignKey:GiftCardID" json:"transactions,omitempty"`
}

// GetCurrentBalance returns the cached current balance
// Note: Balance is automatically maintained by database trigger on gift_card_transactions
// This method is kept for backward compatibility but now just returns the cached value
func (g *GiftCard) GetCurrentBalance() float64 {
	// Round to 2 decimal places to avoid floating point precision issues
	return math.Round(g.CurrentBalance*100) / 100
}

// CurrentBalance method removed - use direct field access or GetCurrentBalance() method
// The balance is now cached in the CurrentBalance field and auto-updated by DB trigger

// GetColor returns the merchant color or a default red
func (g *GiftCard) GetColor() string {
	if g.Merchant != nil && g.Merchant.Color != "" {
		return g.Merchant.Color
	}
	// Fallback to gift card's own color or default red
	if g.Color != "" {
		return g.Color
	}
	return "#DC2626" // Default red
}

// GiftCardTransaction represents a transaction (purchase or reload) on a gift card
type GiftCardTransaction struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GiftCardID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"gift_card_id"`
	GiftCard        *GiftCard      `gorm:"foreignKey:GiftCardID" json:"gift_card,omitempty"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Description     string         `gorm:"type:text" json:"description"`
	TransactionDate time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"transaction_date"`
	CreatedByUserID *uuid.UUID     `gorm:"type:uuid;index" json:"created_by_user_id"`
	CreatedByUser   *User          `gorm:"foreignKey:CreatedByUserID" json:"created_by_user,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GiftCardShare represents a shared gift card with granular permissions
type GiftCardShare struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GiftCardID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"gift_card_id"`
	GiftCard            *GiftCard      `gorm:"foreignKey:GiftCardID" json:"gift_card,omitempty"`
	SharedWithID        uuid.UUID      `gorm:"type:uuid;index;not null" json:"shared_with_id"`
	SharedWithUser      *User          `gorm:"foreignKey:SharedWithID" json:"shared_with_user,omitempty"`
	CanEdit             bool           `gorm:"default:false" json:"can_edit"`
	CanDelete           bool           `gorm:"default:false" json:"can_delete"`
	CanEditTransactions bool           `gorm:"default:false" json:"can_edit_transactions"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
