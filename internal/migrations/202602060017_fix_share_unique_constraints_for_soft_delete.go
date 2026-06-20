package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// fixShareUniqueConstraintsForSoftDelete fixes unique constraints to allow re-sharing after soft delete
// Migration 000017 - 2026-02-06
func fixShareUniqueConstraintsForSoftDelete() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602060017_fix_share_unique_constraints_for_soft_delete",
		Migrate: func(tx *gorm.DB) error {
			// Card Shares: Drop constraint and create partial unique index (only for non-deleted)
			if err := tx.Exec("ALTER TABLE card_shares DROP CONSTRAINT IF EXISTS card_shares_unique").Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS card_shares_unique_active
				ON card_shares (card_id, shared_with_id)
				WHERE deleted_at IS NULL
			`).Error; err != nil {
				return err
			}

			// Voucher Shares: Drop constraint and create partial unique index (only for non-deleted)
			if err := tx.Exec("ALTER TABLE voucher_shares DROP CONSTRAINT IF EXISTS voucher_shares_unique").Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS voucher_shares_unique_active
				ON voucher_shares (voucher_id, shared_with_id)
				WHERE deleted_at IS NULL
			`).Error; err != nil {
				return err
			}

			// Gift Card Shares: Drop constraint and create partial unique index (only for non-deleted)
			if err := tx.Exec("ALTER TABLE gift_card_shares DROP CONSTRAINT IF EXISTS gift_card_shares_unique").Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS gift_card_shares_unique_active
				ON gift_card_shares (gift_card_id, shared_with_id)
				WHERE deleted_at IS NULL
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Restore original unique constraints (without WHERE clause)
			if err := tx.Exec("DROP INDEX IF EXISTS card_shares_unique_active").Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				CREATE UNIQUE INDEX card_shares_unique
				ON card_shares (card_id, shared_with_id)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec("DROP INDEX IF EXISTS voucher_shares_unique_active").Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				CREATE UNIQUE INDEX voucher_shares_unique
				ON voucher_shares (voucher_id, shared_with_id)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec("DROP INDEX IF EXISTS gift_card_shares_unique_active").Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				CREATE UNIQUE INDEX gift_card_shares_unique
				ON gift_card_shares (gift_card_id, shared_with_id)
			`).Error; err != nil {
				return err
			}

			return nil
		},
	}
}
