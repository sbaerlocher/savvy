package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// replaceGlobalUniqueWithComposite drops the old global UNIQUE indexes created by GORM
// and relies on the composite (user_id, card_number/code) indexes instead.
// This allows multiple users to have the same card number/voucher code.
// Migration 000009 - 2026-01-25
func replaceGlobalUniqueWithComposite() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601250009_replace_global_unique_with_composite",
		Migrate: func(tx *gorm.DB) error {
			// Drop old global UNIQUE indexes (created by GORM AutoMigrate)
			if err := dropIndex(tx, "idx_cards_card_number"); err != nil {
				return err
			}
			if err := dropIndex(tx, "idx_vouchers_code"); err != nil {
				return err
			}
			return dropIndex(tx, "idx_gift_cards_card_number")
		},
		Rollback: func(tx *gorm.DB) error {
			// Recreate global UNIQUE indexes
			if err := createIndex(tx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_card_number ON cards (card_number)"); err != nil {
				return err
			}
			if err := createIndex(tx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_vouchers_code ON vouchers (code)"); err != nil {
				return err
			}
			return createIndex(tx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_card_number ON gift_cards (card_number)")
		},
	}
}
