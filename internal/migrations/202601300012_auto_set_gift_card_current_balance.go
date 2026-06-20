package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// autoSetGiftCardCurrentBalance adds triggers to automatically set current_balance
// on INSERT and UPDATE of gift_cards table
// Migration 000012 - 2026-01-30
func autoSetGiftCardCurrentBalance() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601300012_auto_set_gift_card_current_balance",
		Migrate: func(tx *gorm.DB) error {
			// Create trigger function to auto-set current_balance on INSERT/UPDATE
			if err := createFunction(tx, `
				CREATE OR REPLACE FUNCTION auto_set_gift_card_current_balance()
				RETURNS TRIGGER AS $$
				DECLARE
					transaction_sum DECIMAL(10,2);
				BEGIN
					-- Calculate sum of all transactions for this gift card
					SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
					FROM gift_card_transactions
					WHERE gift_card_id = NEW.id
					  AND deleted_at IS NULL;

					-- Set current_balance based on initial_balance and transactions
					NEW.current_balance := NEW.initial_balance - transaction_sum;

					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
			`); err != nil {
				return err
			}

			// Create BEFORE INSERT trigger
			if err := createTrigger(tx, "trigger_auto_set_gift_card_current_balance_insert", "gift_cards",
				"BEFORE", "INSERT", "auto_set_gift_card_current_balance"); err != nil {
				return err
			}

			// Create BEFORE UPDATE trigger with conditional execution
			// Drop existing trigger first (separate statement)
			if err := tx.Exec("DROP TRIGGER IF EXISTS trigger_auto_set_gift_card_current_balance_update ON gift_cards").Error; err != nil {
				return err
			}

			// Create new trigger (separate statement)
			if err := tx.Exec(`
				CREATE TRIGGER trigger_auto_set_gift_card_current_balance_update
					BEFORE UPDATE ON gift_cards
					FOR EACH ROW
					WHEN (OLD.initial_balance IS DISTINCT FROM NEW.initial_balance)
					EXECUTE FUNCTION auto_set_gift_card_current_balance()
			`).Error; err != nil {
				return err
			}

			// Add comment to function
			return addComment(tx, `
				COMMENT ON FUNCTION auto_set_gift_card_current_balance() IS 'Automatically sets current_balance = initial_balance - SUM(transactions) on INSERT/UPDATE of gift_cards';
			`)
		},
		Rollback: func(tx *gorm.DB) error {
			if err := dropTrigger(tx, "trigger_auto_set_gift_card_current_balance_insert", "gift_cards"); err != nil {
				return err
			}
			if err := dropTrigger(tx, "trigger_auto_set_gift_card_current_balance_update", "gift_cards"); err != nil {
				return err
			}
			return dropFunction(tx, "auto_set_gift_card_current_balance")
		},
	}
}
