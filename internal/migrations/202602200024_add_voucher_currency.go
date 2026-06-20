package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addVoucherCurrency adds currency field to vouchers table for multi-currency support
// Migration 000024 - 2026-02-20
func addVoucherCurrency() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602200024_add_voucher_currency",
		Migrate: func(tx *gorm.DB) error {
			// Add currency column with default 'CHF' (Swiss Franc, like gift cards)
			if err := tx.Exec(`
				ALTER TABLE vouchers
				ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'CHF';
			`).Error; err != nil {
				return err
			}

			// Add column comment
			return tx.Exec(`
				COMMENT ON COLUMN vouchers.currency IS 'Currency for fixed_amount vouchers (CHF, EUR, USD, GBP). Applies only when type=fixed_amount.';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE vouchers DROP COLUMN IF EXISTS currency").Error
		},
	}
}
