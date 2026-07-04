// Package models defines the database models for the savvy system.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Voucher represents a discount voucher in the system
type Voucher struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            *uuid.UUID     `gorm:"type:uuid;index;uniqueIndex:idx_vouchers_user_code,where:user_id IS NOT NULL AND deleted_at IS NULL,composite:user_code,priority:1" json:"user_id"`
	User              *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	MerchantID        *uuid.UUID     `gorm:"type:uuid;index" json:"merchant_id"`
	Merchant          *Merchant      `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	MerchantName      string         `json:"merchant_name"` // Fallback for free text
	Code              string         `gorm:"uniqueIndex:idx_vouchers_user_code,where:user_id IS NOT NULL AND deleted_at IS NULL,composite:user_code,priority:2;not null" json:"code"`
	Type              string         `gorm:"not null" json:"type"` // percentage, fixed_amount, points_multiplier, bonus_points
	Value             float64        `gorm:"not null" json:"value"`
	Currency          string         `gorm:"default:CHF" json:"currency"` // Currency for fixed_amount vouchers (CHF, EUR, USD, GBP)
	Description       string         `gorm:"type:text" json:"description"`
	MinPurchaseAmount float64        `gorm:"default:0" json:"min_purchase_amount"`
	ValidFrom         time.Time      `gorm:"not null" json:"valid_from"`
	ValidUntil        time.Time      `gorm:"not null" json:"valid_until"`
	UsageLimitType    string         `gorm:"default:single_use" json:"usage_limit_type"` // single_use, one_per_customer, multiple_use_with_card, multiple_use_without_card
	BarcodeType       string         `gorm:"default:CODE128" json:"barcode_type"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GetColor returns the merchant color or a default green
func (v *Voucher) GetColor() string {
	if v.Merchant != nil && v.Merchant.Color != "" {
		return v.Merchant.Color
	}
	return "#10B981"
}

// VoucherShare represents a shared voucher (read-only)
type VoucherShare struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VoucherID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"voucher_id"`
	Voucher        *Voucher       `gorm:"foreignKey:VoucherID" json:"voucher,omitempty"`
	SharedWithID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"shared_with_id"`
	SharedWithUser *User          `gorm:"foreignKey:SharedWithID" json:"shared_with_user,omitempty"`
	CanEdit        bool           `gorm:"default:false" json:"can_edit"`
	CanDelete      bool           `gorm:"default:false" json:"can_delete"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
