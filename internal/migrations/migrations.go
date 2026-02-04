// Package migrations defines all database migrations using Gormigrate.
// This provides Laravel-like migration experience with up/down functions.
package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Helper functions to reduce code duplication

// createTrigger creates a database trigger
func createTrigger(tx *gorm.DB, triggerName, tableName, timing, event, functionName string) error {
	return tx.Exec(fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %s ON %s;
		CREATE TRIGGER %s
			%s %s ON %s
			FOR EACH ROW
			EXECUTE FUNCTION %s();
	`, triggerName, tableName, triggerName, timing, event, tableName, functionName)).Error
}

// dropTrigger drops a database trigger
func dropTrigger(tx *gorm.DB, triggerName, tableName string) error {
	return tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerName, tableName)).Error
}

// createFunction creates a database function
func createFunction(tx *gorm.DB, functionSQL string) error {
	return tx.Exec(functionSQL).Error
}

// dropFunction drops a database function
func dropFunction(tx *gorm.DB, functionName string) error {
	return tx.Exec(fmt.Sprintf("DROP FUNCTION IF EXISTS %s()", functionName)).Error
}

// createIndex creates a database index
func createIndex(tx *gorm.DB, indexSQL string) error {
	return tx.Exec(indexSQL).Error
}

// dropIndex drops a database index
func dropIndex(tx *gorm.DB, indexName string) error {
	return tx.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)).Error
}

// addComment adds a comment to a database object
func addComment(tx *gorm.DB, commentSQL string) error {
	return tx.Exec(commentSQL).Error
}

// GetMigrations returns all migrations in chronological order
func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		initSchema(),
		addGiftCardBalanceConstraint(),
		normalizeEmails(),
		addUserFavorites(),
		addCaseInsensitiveEmailIndex(),
		addGiftCardBalanceCache(),
		addUniqueConstraintsForRaceConditions(),
		replaceGlobalUniqueWithComposite(),
		addAuditLog(),
		addAuthProvider(),
		autoSetGiftCardCurrentBalance(),
		removeVoucherUsedCount(),
		removeUnusedColorColumns(),
		fixGiftCardBalanceExcludeSoftDeletes(),
	}
}

// initSchema creates the initial database schema
// Equivalent to: migrations/000001_init_schema.up.sql
func initSchema() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601230001_init_schema",
		Migrate: func(tx *gorm.DB) error {
			// Enable UUID extension
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
				return err
			}

			// Define temporary structs for migration (matches GORM models)
			type User struct {
				ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				Email        string    `gorm:"type:text;not null;uniqueIndex"`
				PasswordHash string    `gorm:"type:text;not null"`
				FirstName    string    `gorm:"type:text;not null"`
				LastName     string    `gorm:"type:text;not null"`
				Role         string    `gorm:"type:text;default:'user'"`
				CreatedAt    time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt    time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			type Merchant struct {
				ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				Name      string     `gorm:"type:text;not null;uniqueIndex"`
				LogoURL   string     `gorm:"type:text"`
				Website   string     `gorm:"type:text"`
				Color     string     `gorm:"type:text;default:'#0066CC'"`
				CreatedAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			type Card struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       *uuid.UUID `gorm:"type:uuid;index:idx_cards_user_id"`
				MerchantID   *uuid.UUID `gorm:"type:uuid;index:idx_cards_merchant_id"`
				MerchantName string     `gorm:"type:text;default:''"`
				Program      string     `gorm:"type:text;not null"`
				CardNumber   string     `gorm:"type:text;not null;uniqueIndex:idx_cards_card_number"`
				BarcodeType  string     `gorm:"type:text;default:'CODE128'"`
				Status       string     `gorm:"type:text;default:'active'"`
				Notes        string     `gorm:"type:text"`
				Color        string     `gorm:"type:text;default:'#0066CC'"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt    *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			type CardShare struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				CardID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_card_shares_card_id"`
				SharedWithID uuid.UUID  `gorm:"type:uuid;not null;index:idx_card_shares_shared_with_id"`
				CanEdit      bool       `gorm:"default:false"`
				CanDelete    bool       `gorm:"default:false"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt    *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			type Voucher struct {
				ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID            *uuid.UUID `gorm:"type:uuid;index:idx_vouchers_user_id"`
				MerchantID        *uuid.UUID `gorm:"type:uuid;index:idx_vouchers_merchant_id"`
				MerchantName      string     `gorm:"type:text"`
				Code              string     `gorm:"type:text;not null;uniqueIndex:idx_vouchers_code"`
				Type              string     `gorm:"type:text;not null"`
				Value             float64    `gorm:"type:numeric;not null"`
				Description       string     `gorm:"type:text"`
				MinPurchaseAmount float64    `gorm:"type:numeric;default:0"`
				ValidFrom         time.Time  `gorm:"type:timestamp with time zone;not null"`
				ValidUntil        time.Time  `gorm:"type:timestamp with time zone;not null"`
				UsageLimitType    string     `gorm:"type:text;default:'single_use'"`
				BarcodeType       string     `gorm:"type:text;default:'CODE128'"`
				Color             string     `gorm:"type:text;default:'#10B981'"`
				CreatedAt         time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt         time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt         *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			type VoucherShare struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				VoucherID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_voucher_shares_voucher_id"`
				SharedWithID uuid.UUID  `gorm:"type:uuid;not null;index:idx_voucher_shares_shared_with_id"`
				CanEdit      bool       `gorm:"default:false"`
				CanDelete    bool       `gorm:"default:false"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt    *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			type GiftCard struct {
				ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID         *uuid.UUID `gorm:"type:uuid;index:idx_gift_cards_user_id"`
				MerchantID     *uuid.UUID `gorm:"type:uuid;index:idx_gift_cards_merchant_id"`
				MerchantName   string     `gorm:"type:text"`
				CardNumber     string     `gorm:"type:text;not null;uniqueIndex:idx_gift_cards_card_number"`
				InitialBalance float64    `gorm:"type:numeric;not null"`
				Currency       string     `gorm:"type:text;default:'CHF'"`
				PIN            string     `gorm:"type:text"`
				ExpiresAt      *time.Time `gorm:"type:timestamp with time zone"`
				Status         string     `gorm:"type:text;default:'active'"`
				BarcodeType    string     `gorm:"type:text;default:'CODE128'"`
				Color          string     `gorm:"type:text;default:'#10B981'"`
				Notes          string     `gorm:"type:text"`
				CreatedAt      time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt      time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt      *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			type GiftCardTransaction struct {
				ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				GiftCardID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_gift_card_transactions_gift_card_id"`
				Amount          float64    `gorm:"type:numeric;not null"`
				Description     string     `gorm:"type:text"`
				TransactionDate time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`
				CreatedAt       time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt       time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt       *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			type GiftCardShare struct {
				ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				GiftCardID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_gift_card_shares_gift_card_id"`
				SharedWithID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_gift_card_shares_shared_with_id"`
				CanEdit             bool       `gorm:"default:false"`
				CanDelete           bool       `gorm:"default:false"`
				CanEditTransactions bool       `gorm:"default:false"`
				CreatedAt           time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt           time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt           *time.Time `gorm:"type:timestamp with time zone;index"`
			}

			// Create all tables
			tables := []interface{}{
				&User{},
				&Merchant{},
				&Card{},
				&CardShare{},
				&Voucher{},
				&VoucherShare{},
				&GiftCard{},
				&GiftCardTransaction{},
				&GiftCardShare{},
			}

			for _, table := range tables {
				if err := tx.AutoMigrate(table); err != nil {
					return err
				}
			}

			// Add unique constraints that AutoMigrate might miss
			// Use DO blocks to check existence before adding constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'card_shares_unique'
					) THEN
						ALTER TABLE card_shares ADD CONSTRAINT card_shares_unique UNIQUE (card_id, shared_with_id);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'voucher_shares_unique'
					) THEN
						ALTER TABLE voucher_shares ADD CONSTRAINT voucher_shares_unique UNIQUE (voucher_id, shared_with_id);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'gift_card_shares_unique'
					) THEN
						ALTER TABLE gift_card_shares ADD CONSTRAINT gift_card_shares_unique UNIQUE (gift_card_id, shared_with_id);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add foreign key constraints
			if err := tx.Exec(`
				ALTER TABLE cards
				ADD CONSTRAINT fk_cards_user FOREIGN KEY (user_id) REFERENCES users(id),
				ADD CONSTRAINT fk_cards_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE card_shares
				ADD CONSTRAINT fk_card_shares_card FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE,
				ADD CONSTRAINT fk_card_shares_user FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE vouchers
				ADD CONSTRAINT fk_vouchers_user FOREIGN KEY (user_id) REFERENCES users(id),
				ADD CONSTRAINT fk_vouchers_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE voucher_shares
				ADD CONSTRAINT fk_voucher_shares_voucher FOREIGN KEY (voucher_id) REFERENCES vouchers(id) ON DELETE CASCADE,
				ADD CONSTRAINT fk_voucher_shares_user FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE gift_cards
				ADD CONSTRAINT fk_gift_cards_user FOREIGN KEY (user_id) REFERENCES users(id),
				ADD CONSTRAINT fk_gift_cards_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE gift_card_transactions
				ADD CONSTRAINT fk_gift_card_transactions_gift_card
				FOREIGN KEY (gift_card_id) REFERENCES gift_cards(id) ON DELETE CASCADE
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				ALTER TABLE gift_card_shares
				ADD CONSTRAINT fk_gift_card_shares_gift_card FOREIGN KEY (gift_card_id) REFERENCES gift_cards(id) ON DELETE CASCADE,
				ADD CONSTRAINT fk_gift_card_shares_user FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop all tables in reverse order (respecting foreign keys)
			tables := []string{
				"gift_card_shares",
				"gift_card_transactions",
				"gift_cards",
				"voucher_shares",
				"vouchers",
				"card_shares",
				"cards",
				"merchants",
				"users",
			}

			for _, table := range tables {
				if err := tx.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
					return err
				}
			}

			return tx.Exec(`DROP EXTENSION IF EXISTS "pgcrypto" CASCADE`).Error
		},
	}
}

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

// normalizeEmails creates trigger to automatically lowercase emails
// Equivalent to: migrations/000004_normalize_emails.up.sql
func normalizeEmails() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601230004_normalize_emails",
		Migrate: func(tx *gorm.DB) error {
			// Normalize all existing emails to lowercase
			if err := tx.Exec("UPDATE users SET email = LOWER(email)").Error; err != nil {
				return err
			}

			// Create trigger function to automatically lowercase emails on insert/update
			if err := createFunction(tx, `
				CREATE OR REPLACE FUNCTION enforce_lowercase_email()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.email = LOWER(TRIM(NEW.email));
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
			`); err != nil {
				return err
			}

			// Create trigger on users table
			if err := createTrigger(tx, "trigger_lowercase_email", "users",
				"BEFORE", "INSERT OR UPDATE", "enforce_lowercase_email"); err != nil {
				return err
			}

			// Add comment for documentation
			return addComment(tx, `
				COMMENT ON FUNCTION enforce_lowercase_email() IS 'Automatically converts email addresses to lowercase to ensure case-insensitive uniqueness';
			`)
		},
		Rollback: func(tx *gorm.DB) error {
			if err := dropTrigger(tx, "trigger_lowercase_email", "users"); err != nil {
				return err
			}
			return dropFunction(tx, "enforce_lowercase_email")
		},
	}
}

// addUserFavorites creates the user_favorites table
// Equivalent to: migrations/000005_add_user_favorites.up.sql
func addUserFavorites() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601250005_add_user_favorites",
		Migrate: func(tx *gorm.DB) error {
			type UserFavorite struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_favorites"`
				ResourceType string     `gorm:"type:varchar(50);not null;index:idx_user_favorites"`
				ResourceID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_favorites"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				DeletedAt    *time.Time `gorm:"type:timestamp with time zone;index:idx_user_favorites_deleted_at"`
			}

			// Create table
			if err := tx.AutoMigrate(&UserFavorite{}); err != nil {
				return err
			}

			// Add unique constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'user_favorites_unique'
					) THEN
						ALTER TABLE user_favorites ADD CONSTRAINT user_favorites_unique UNIQUE (user_id, resource_type, resource_id);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add foreign key
			if err := tx.Exec(`
				ALTER TABLE user_favorites
				ADD CONSTRAINT fk_user_favorites_user
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			`).Error; err != nil {
				return err
			}

			// Add table comment
			if err := tx.Exec(`
				COMMENT ON TABLE user_favorites IS 'User-specific favorites for cards, vouchers, and gift cards'
			`).Error; err != nil {
				return err
			}

			// Add column comments
			if err := tx.Exec(`
				COMMENT ON COLUMN user_favorites.resource_type IS 'Type of resource: card, voucher, or gift_card'
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				COMMENT ON COLUMN user_favorites.resource_id IS 'UUID of the favorited resource'
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS user_favorites CASCADE").Error
		},
	}
}

// addCaseInsensitiveEmailIndex replaces the case-sensitive email index with a case-insensitive one
// Equivalent to: migrations/000006_case_insensitive_email_index.up.sql
func addCaseInsensitiveEmailIndex() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601250006_case_insensitive_email_index",
		Migrate: func(tx *gorm.DB) error {
			// Drop the existing case-sensitive unique index
			if err := tx.Exec(`
				DROP INDEX IF EXISTS idx_users_email;
			`).Error; err != nil {
				return err
			}

			// Create case-insensitive unique index using LOWER()
			if err := tx.Exec(`
				CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));
			`).Error; err != nil {
				return err
			}

			// Add comment to explain the index
			if err := tx.Exec(`
				COMMENT ON INDEX idx_users_email_lower IS 'Case-insensitive unique index on email to prevent duplicate emails with different cases (e.g., Test@Email.com and test@email.com)';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop case-insensitive index
			if err := tx.Exec("DROP INDEX IF EXISTS idx_users_email_lower").Error; err != nil {
				return err
			}

			// Recreate original case-sensitive index
			if err := tx.Exec(`
				CREATE UNIQUE INDEX idx_users_email ON users (email);
			`).Error; err != nil {
				return err
			}

			return nil
		},
	}
}

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

// addAuditLog creates the audit_logs table for tracking all deletion operations
// Migration 000010 - 2026-01-26
func addAuditLog() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601260010_add_audit_log",
		Migrate: func(tx *gorm.DB) error {
			type AuditLog struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       *uuid.UUID `gorm:"type:uuid;index"`
				Action       string     `gorm:"type:varchar(50);not null;index"`
				ResourceType string     `gorm:"type:varchar(50);not null;index"`
				ResourceID   uuid.UUID  `gorm:"type:uuid;not null;index"`
				ResourceData string     `gorm:"type:jsonb"`
				IPAddress    string     `gorm:"type:varchar(45)"`
				UserAgent    string     `gorm:"type:text"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;index"`
			}

			// Create table
			if err := tx.AutoMigrate(&AuditLog{}); err != nil {
				return err
			}

			// Add foreign key constraint for user_id
			if err := tx.Exec(`
				ALTER TABLE audit_logs
				ADD CONSTRAINT fk_audit_logs_user
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
			`).Error; err != nil {
				return err
			}

			// Add composite index for common queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_audit_logs_resource
				ON audit_logs (resource_type, resource_id, created_at DESC);
			`).Error; err != nil {
				return err
			}

			// Add index for user queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created
				ON audit_logs (user_id, created_at DESC)
				WHERE user_id IS NOT NULL;
			`).Error; err != nil {
				return err
			}

			// Add table comment
			if err := tx.Exec(`
				COMMENT ON TABLE audit_logs IS 'Audit trail for all deletion operations in the system for compliance and traceability';
			`).Error; err != nil {
				return err
			}

			// Add column comments
			if err := tx.Exec(`
				COMMENT ON COLUMN audit_logs.action IS 'Type of action: delete, hard_delete, restore';
				COMMENT ON COLUMN audit_logs.resource_type IS 'Type of resource: cards, vouchers, gift_cards, etc.';
				COMMENT ON COLUMN audit_logs.resource_data IS 'JSON snapshot of the deleted resource';
				COMMENT ON COLUMN audit_logs.ip_address IS 'IP address of the user who performed the action';
				COMMENT ON COLUMN audit_logs.user_agent IS 'Browser user agent string';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS audit_logs CASCADE").Error
		},
	}
}

// addAuthProvider adds the auth_provider column to users table to distinguish OAuth from local users
// Migration 000011 - 2026-01-26
func addAuthProvider() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202601260011_add_auth_provider",
		Migrate: func(tx *gorm.DB) error {
			// Add auth_provider column with default 'local'
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(50) NOT NULL DEFAULT 'local';
			`).Error; err != nil {
				return err
			}

			// Add index for auth_provider queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_users_auth_provider
				ON users (auth_provider);
			`).Error; err != nil {
				return err
			}

			// Add column comment
			if err := tx.Exec(`
				COMMENT ON COLUMN users.auth_provider IS 'Authentication provider: "local" for username/password, "oauth" for OAuth/OIDC';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop index
			if err := tx.Exec("DROP INDEX IF EXISTS idx_users_auth_provider").Error; err != nil {
				return err
			}

			// Drop column
			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS auth_provider").Error; err != nil {
				return err
			}

			return nil
		},
	}
}

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
			if err := tx.Exec(`
				DROP TRIGGER IF EXISTS trigger_auto_set_gift_card_current_balance_update ON gift_cards;
				CREATE TRIGGER trigger_auto_set_gift_card_current_balance_update
					BEFORE UPDATE ON gift_cards
					FOR EACH ROW
					WHEN (OLD.initial_balance IS DISTINCT FROM NEW.initial_balance)
					EXECUTE FUNCTION auto_set_gift_card_current_balance();
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

// removeVoucherUsedCount drops the unused used_count column from vouchers table
// Migration 000011 - 2026-02-01
func removeVoucherUsedCount() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602010011_remove_voucher_used_count",
		Migrate: func(tx *gorm.DB) error {
			// Drop used_count column (no longer used after removing redeem functionality)
			return tx.Exec("ALTER TABLE vouchers DROP COLUMN IF EXISTS used_count").Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Re-add used_count column with default value
			return tx.Exec("ALTER TABLE vouchers ADD COLUMN used_count BIGINT DEFAULT 0").Error
		},
	}
}

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

