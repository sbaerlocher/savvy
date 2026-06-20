package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// removeUnusedColorColumns drops the unused color columns from cards, vouchers, and gift_cards tables
// These columns were never used in the frontend (no input fields, no handler processing).
// Color is only retrieved from Merchant.Color via GetColor() methods.
// Migration 000013 - 2026-02-03
func removeUnusedColorColumns() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602030013_remove_unused_color_columns",
		Migrate: func(tx *gorm.DB) error {
			// Drop color column from cards (was in DB but ignored by GORM model)
			if err := tx.Exec("ALTER TABLE cards DROP COLUMN IF EXISTS color").Error; err != nil {
				return err
			}

			// Drop color column from vouchers (had default #10B981 but never editable)
			if err := tx.Exec("ALTER TABLE vouchers DROP COLUMN IF EXISTS color").Error; err != nil {
				return err
			}

			// Drop color column from gift_cards (had default #DC2626 but never editable)
			return tx.Exec("ALTER TABLE gift_cards DROP COLUMN IF EXISTS color").Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Re-add color columns with their original defaults
			if err := tx.Exec("ALTER TABLE cards ADD COLUMN color VARCHAR(7) DEFAULT '#0066CC'").Error; err != nil {
				return err
			}

			if err := tx.Exec("ALTER TABLE vouchers ADD COLUMN color VARCHAR(7) DEFAULT '#10B981'").Error; err != nil {
				return err
			}

			return tx.Exec("ALTER TABLE gift_cards ADD COLUMN color VARCHAR(7) DEFAULT '#DC2626'").Error
		},
	}
}
