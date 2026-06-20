package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addUniqueConstraintsForRaceConditions adds unique constraints to prevent race conditions
// on card_number, code, and card_number for cards, vouchers, and gift_cards respectively.
// Prevents TOCTOU (Time-of-check to time-of-use) vulnerabilities.
// Migration 000008 - 2026-01-25
func addUniqueConstraintsForRaceConditions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601250008_unique_constraints_race_conditions",
		Migrate: func(tx *gorm.DB) error {
			// Cards: UNIQUE (user_id, card_number)
			// Multiple users can have same card number, but one user can't have duplicate card numbers
			if err := createIndex(tx, `
				CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_user_card_number
				ON cards (user_id, card_number)
				WHERE user_id IS NOT NULL;
			`); err != nil {
				return err
			}

			if err := addComment(tx, `
				COMMENT ON INDEX idx_cards_user_card_number IS 'Prevents duplicate card numbers per user. Allows different users to have same card number (e.g., family cards).';
			`); err != nil {
				return err
			}

			// Vouchers: UNIQUE (user_id, code)
			// Same logic - different users can have same voucher code
			if err := createIndex(tx, `
				CREATE UNIQUE INDEX IF NOT EXISTS idx_vouchers_user_code
				ON vouchers (user_id, code)
				WHERE user_id IS NOT NULL;
			`); err != nil {
				return err
			}

			if err := addComment(tx, `
				COMMENT ON INDEX idx_vouchers_user_code IS 'Prevents duplicate voucher codes per user. Allows different users to have same voucher code.';
			`); err != nil {
				return err
			}

			// Gift Cards: UNIQUE (user_id, card_number)
			// Same logic - different users can have same gift card number
			if err := createIndex(tx, `
				CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_user_card_number
				ON gift_cards (user_id, card_number)
				WHERE user_id IS NOT NULL;
			`); err != nil {
				return err
			}

			return addComment(tx, `
				COMMENT ON INDEX idx_gift_cards_user_card_number IS 'Prevents duplicate gift card numbers per user. Allows different users to have same card number.';
			`)
		},
		Rollback: func(tx *gorm.DB) error {
			if err := dropIndex(tx, "idx_cards_user_card_number"); err != nil {
				return err
			}
			if err := dropIndex(tx, "idx_vouchers_user_code"); err != nil {
				return err
			}
			return dropIndex(tx, "idx_gift_cards_user_card_number")
		},
	}
}
