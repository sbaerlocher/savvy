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

// GetColor returns the merchant color or a default red
func (g *GiftCard) GetColor() string {
	if g.Merchant != nil && g.Merchant.Color != "" {
		return g.Merchant.Color
	}
	return "#DC2626"
}

// IsExpired checks if the gift card has expired
func (g *GiftCard) IsExpired() bool {
	if g.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*g.ExpiresAt)
}

// IsEmpty checks if the gift card has no balance left
func (g *GiftCard) IsEmpty() bool {
	return g.CurrentBalance <= 0
}

// GetComputedStatus returns the computed status based on balance and expiry
// Returns: "redeemed" (balance 0), "expired" (date passed + balance > 0), "active"
func (g *GiftCard) GetComputedStatus() string {
	if g.IsEmpty() {
		return "redeemed"
	}
	if g.IsExpired() {
		return "expired"
	}
	return "active"
}

// IsUsable checks if the gift card can still be used
func (g *GiftCard) IsUsable() bool {
	return !g.IsExpired() && !g.IsEmpty()
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
