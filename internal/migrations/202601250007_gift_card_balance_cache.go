package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addGiftCardBalanceCache adds a cached current_balance column with trigger-based updates
// Equivalent to: migrations/000007_gift_card_balance_cache.up.sql
func addGiftCardBalanceCache() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601250007_gift_card_balance_cache",
		Migrate: func(tx *gorm.DB) error {
			// Add current_balance column (nullable initially for migration)
			if err := tx.Exec(`
				ALTER TABLE gift_cards ADD COLUMN IF NOT EXISTS current_balance DECIMAL(10,2);
			`).Error; err != nil {
				return err
			}

			// Calculate and populate current_balance for all existing gift cards
			if err := tx.Exec(`
				UPDATE gift_cards
				SET current_balance = initial_balance - (
					SELECT COALESCE(SUM(amount), 0)
					FROM gift_card_transactions
					WHERE gift_card_id = gift_cards.id
				);
			`).Error; err != nil {
				return err
			}

			// Make current_balance NOT NULL now that it's populated
			if err := tx.Exec(`
				ALTER TABLE gift_cards ALTER COLUMN current_balance SET NOT NULL;
			`).Error; err != nil {
				return err
			}

			// Create trigger function to recalculate balance
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
			`); err != nil {
				return err
			}

			// Create trigger on gift_card_transactions
			if err := createTrigger(tx, "trigger_recalculate_gift_card_balance", "gift_card_transactions",
				"AFTER", "INSERT OR UPDATE OR DELETE", "recalculate_gift_card_balance"); err != nil {
				return err
			}

			// Add comment to column
			return addComment(tx, `
				COMMENT ON COLUMN gift_cards.current_balance IS 'Cached balance calculated as initial_balance - SUM(transactions.amount). Auto-updated by trigger on gift_card_transactions.';
			`)
		},
		Rollback: func(tx *gorm.DB) error {
			if err := dropTrigger(tx, "trigger_recalculate_gift_card_balance", "gift_card_transactions"); err != nil {
				return err
			}
			if err := dropFunction(tx, "recalculate_gift_card_balance"); err != nil {
				return err
			}
			return tx.Exec("ALTER TABLE gift_cards DROP COLUMN IF EXISTS current_balance").Error
		},
	}
}
