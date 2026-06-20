package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// fixGiftCardBalanceExcludeSoftDeletes fixes the recalculate_gift_card_balance() function
// to exclude soft-deleted transactions (deleted_at IS NOT NULL) from balance calculation.
// Also recalculates all existing balances to fix any incorrect values.
// Migration 000014 - 2026-02-04
func fixGiftCardBalanceExcludeSoftDeletes() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602040014_fix_gift_card_balance_exclude_soft_deletes",
		Migrate: func(tx *gorm.DB) error {
			// Replace the trigger function with corrected version
			if err := createFunction(tx, `
				CREATE OR REPLACE FUNCTION recalculate_gift_card_balance()
				RETURNS TRIGGER AS $$
				DECLARE
					card_id UUID;
				BEGIN
					-- Determine which gift card was affected
					IF TG_OP = 'DELETE' THEN
						card_id := OLD.gift_card_id;
					ELSE
						card_id := NEW.gift_card_id;
					END IF;

					-- Recalculate and update the balance (exclude soft-deleted transactions)
					UPDATE gift_cards
					SET current_balance = initial_balance - (
						SELECT COALESCE(SUM(amount), 0)
						FROM gift_card_transactions
						WHERE gift_card_id = card_id
						  AND deleted_at IS NULL
					)
					WHERE id = card_id;

					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
			`); err != nil {
				return err
			}

			// Recalculate all existing balances to fix incorrect values
			return tx.Exec(`
				UPDATE gift_cards
				SET current_balance = initial_balance - (
					SELECT COALESCE(SUM(amount), 0)
					FROM gift_card_transactions
					WHERE gift_card_transactions.gift_card_id = gift_cards.id
					  AND gift_card_transactions.deleted_at IS NULL
				);
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Restore old function without soft-delete check
			return createFunction(tx, `
				CREATE OR REPLACE FUNCTION recalculate_gift_card_balance()
				RETURNS TRIGGER AS $$
				DECLARE
					card_id UUID;
				BEGIN
					-- Determine which gift card was affected
					IF TG_OP = 'DELETE' THEN
						card_id := OLD.gift_card_id;
					ELSE
						card_id := NEW.gift_card_id;
					END IF;

					-- Recalculate and update the balance
					UPDATE gift_cards
					SET current_balance = initial_balance - (
						SELECT COALESCE(SUM(amount), 0)
						FROM gift_card_transactions
						WHERE gift_card_id = card_id
					)
					WHERE id = card_id;

					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
			`)
		},
	}
}
