package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addGiftCardBalanceConstraint creates trigger to prevent negative gift card balances
// Equivalent to: migrations/000003_gift_card_balance_constraint.up.sql
func addGiftCardBalanceConstraint() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601230003_gift_card_balance_constraint",
		Migrate: func(tx *gorm.DB) error {
			// Create trigger function to validate balance before insert/update
			if err := createFunction(tx, `
				CREATE OR REPLACE FUNCTION check_gift_card_balance()
				RETURNS TRIGGER AS $$
				DECLARE
					current_balance DECIMAL(10,2);
					initial_balance DECIMAL(10,2);
				BEGIN
					-- Get initial balance
					SELECT gc.initial_balance INTO initial_balance
					FROM gift_cards gc
					WHERE gc.id = NEW.gift_card_id;

					-- Calculate current balance (initial - sum of all transactions)
					SELECT initial_balance - COALESCE(SUM(t.amount), 0) INTO current_balance
					FROM gift_card_transactions t
					WHERE t.gift_card_id = NEW.gift_card_id
						AND t.deleted_at IS NULL
						AND t.id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::uuid);

					-- Check if new transaction would result in negative balance
					IF (current_balance - NEW.amount) < 0 THEN
						RAISE EXCEPTION 'Insufficient balance: current=%, transaction=%, would result in=%',
							current_balance, NEW.amount, (current_balance - NEW.amount);
					END IF;

					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
			`); err != nil {
				return err
			}

			// Create trigger to enforce balance check BEFORE insert/update
			if err := createTrigger(tx, "trigger_check_gift_card_balance", "gift_card_transactions",
				"BEFORE", "INSERT OR UPDATE", "check_gift_card_balance"); err != nil {
				return err
			}

			// Create index for performance (speeds up balance calculation)
			return createIndex(tx, `
				CREATE INDEX IF NOT EXISTS idx_gift_card_transactions_gift_card_deleted
				ON gift_card_transactions(gift_card_id, deleted_at);
			`)
		},
		Rollback: func(tx *gorm.DB) error {
			if err := dropTrigger(tx, "trigger_check_gift_card_balance", "gift_card_transactions"); err != nil {
				return err
			}
			if err := dropFunction(tx, "check_gift_card_balance"); err != nil {
				return err
			}
			return dropIndex(tx, "idx_gift_card_transactions_gift_card_deleted")
		},
	}
}
