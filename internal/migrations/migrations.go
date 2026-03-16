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
	// Drop existing trigger first (separate statement)
	if err := tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerName, tableName)).Error; err != nil {
		return err
	}

	// Create new trigger (separate statement)
	return tx.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s
			%s %s ON %s
			FOR EACH ROW
			EXECUTE FUNCTION %s()
	`, triggerName, timing, event, tableName, functionName)).Error
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
		addNotifications(),
		addNotificationsSoftDelete(),
		fixShareUniqueConstraintsForSoftDelete(),
		addUserSecurityColumns(),
		addEmailVerificationAndTokens(),
		addUserLanguage(),
		addPushSubscriptions(),
		addExpiryReminderTracking(),
		addUserTOTP(),
		addVoucherCurrency(),
		addUserExpiryRemindersEnabled(),
		addUserShareEmailsEnabled(),
		addPerChannelNotificationPreferences(),
		addUserPasswordChangedAt(),
		addServerSideSessions(),
		fixNotificationPreferenceDefaults(),
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

			// Add foreign key (idempotent)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_favorites_user'
					) THEN
						ALTER TABLE user_favorites
						ADD CONSTRAINT fk_user_favorites_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
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

			// Add foreign key constraint for user_id (idempotent)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_audit_logs_user'
					) THEN
						ALTER TABLE audit_logs
						ADD CONSTRAINT fk_audit_logs_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
					END IF;
				END $$;
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

			// Add column comments (each must be a separate statement)
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.action IS 'Type of action: delete, hard_delete, restore'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.resource_type IS 'Type of resource: cards, vouchers, gift_cards, etc.'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.resource_data IS 'JSON snapshot of the deleted resource'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.ip_address IS 'IP address of the user who performed the action'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN audit_logs.user_agent IS 'Browser user agent string'").Error; err != nil {
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

// addNotifications creates the notifications table for in-app notifications
// Migration 000015 - 2026-02-06
func addNotifications() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602060015_add_notifications",
		Migrate: func(tx *gorm.DB) error {
			// Define Notification struct for migration
			type Notification struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_notifications_user_id"`
				Type         string     `gorm:"type:varchar(50);not null;index:idx_notifications_type"`
				ResourceType string     `gorm:"type:varchar(50);not null"`
				ResourceID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_notifications_resource"`
				Metadata     string     `gorm:"type:jsonb;default:'{}'"`
				IsRead       bool       `gorm:"default:false;index:idx_notifications_is_read"`
				ReadAt       *time.Time `gorm:"type:timestamp with time zone"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;index:idx_notifications_created_at"`
				UpdatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			// Create table
			if err := tx.AutoMigrate(&Notification{}); err != nil {
				return err
			}

			// Add foreign key constraint for user_id (idempotent)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_user'
					) THEN
						ALTER TABLE notifications
						ADD CONSTRAINT fk_notifications_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add composite index for unread notifications query (most common query)
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
				ON notifications (user_id, created_at DESC)
				WHERE is_read = FALSE;
			`).Error; err != nil {
				return err
			}

			// Add composite index for resource lookups
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_resource_lookup
				ON notifications (resource_type, resource_id, created_at DESC);
			`).Error; err != nil {
				return err
			}

			// Add table comment
			if err := tx.Exec(`
				COMMENT ON TABLE notifications IS 'In-app notifications for share and transfer events';
			`).Error; err != nil {
				return err
			}

			// Add column comments (each must be a separate statement)
			if err := tx.Exec("COMMENT ON COLUMN notifications.type IS 'Notification type: share_received, transfer_received'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN notifications.resource_type IS 'Type of resource: card, voucher, gift_card'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN notifications.resource_id IS 'UUID of the resource'").Error; err != nil {
				return err
			}
			if err := tx.Exec("COMMENT ON COLUMN notifications.metadata IS 'JSONB metadata: from_user_id, from_user_name, permissions, etc.'").Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS notifications CASCADE").Error
		},
	}
}

// addNotificationsSoftDelete adds soft delete support to notifications table
// Migration 000016 - 2026-02-06
func addNotificationsSoftDelete() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602060016_add_notifications_soft_delete",
		Migrate: func(tx *gorm.DB) error {
			// Add deleted_at column with index
			if err := tx.Exec(`
				ALTER TABLE notifications
				ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
			`).Error; err != nil {
				return err
			}

			// Add index on deleted_at for soft delete queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at
				ON notifications (deleted_at);
			`).Error; err != nil {
				return err
			}

			// Add column comment
			if err := tx.Exec(`
				COMMENT ON COLUMN notifications.deleted_at IS 'Soft delete timestamp - NULL means not deleted';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop index and column
			if err := tx.Exec("DROP INDEX IF EXISTS idx_notifications_deleted_at").Error; err != nil {
				return err
			}
			return tx.Exec("ALTER TABLE notifications DROP COLUMN IF EXISTS deleted_at").Error
		},
	}
}

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

// addEmailVerificationAndTokens adds email verification columns to users and creates email_tokens table
// Migration 000019 - 2026-02-19
func addEmailVerificationAndTokens() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602190019_add_email_verification_and_tokens",
		Migrate: func(tx *gorm.DB) error {
			// Add email_verified column with default false
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT false;
			`).Error; err != nil {
				return err
			}

			// Add email_verified_at column (nullable timestamp)
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP WITH TIME ZONE;
			`).Error; err != nil {
				return err
			}

			// Mark all existing users as verified (they registered before verification was required)
			if err := tx.Exec(`
				UPDATE users SET email_verified = true, email_verified_at = CURRENT_TIMESTAMP
				WHERE email_verified = false;
			`).Error; err != nil {
				return err
			}

			// Create email_tokens table
			type EmailToken struct {
				ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_email_tokens_user_type"`
				TokenHash string     `gorm:"type:varchar(64);not null;uniqueIndex:idx_email_tokens_token_hash"`
				TokenType string     `gorm:"type:varchar(50);not null;index:idx_email_tokens_user_type"`
				ExpiresAt time.Time  `gorm:"type:timestamp with time zone;not null"`
				UsedAt    *time.Time `gorm:"type:timestamp with time zone"`
				CreatedAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			if err := tx.AutoMigrate(&EmailToken{}); err != nil {
				return err
			}

			// Add expires_at index for cleanup queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_email_tokens_expires_at
				ON email_tokens (expires_at);
			`).Error; err != nil {
				return err
			}

			// Add foreign key constraint for user_id
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_email_tokens_user'
					) THEN
						ALTER TABLE email_tokens
						ADD CONSTRAINT fk_email_tokens_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add column comments
			if err := tx.Exec(`COMMENT ON COLUMN users.email_verified IS 'Whether the user email address has been verified'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON TABLE email_tokens IS 'Tokens for email verification and password reset'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON COLUMN email_tokens.token_hash IS 'SHA-256 hash of the token (plain token is sent to user)'`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`COMMENT ON COLUMN email_tokens.token_type IS 'Token type: email_verification or password_reset'`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop email_tokens table
			if err := tx.Exec("DROP TABLE IF EXISTS email_tokens CASCADE").Error; err != nil {
				return err
			}

			// Drop columns from users
			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at").Error; err != nil {
				return err
			}
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS email_verified").Error
		},
	}
}

// addUserLanguage adds the language column to users table for localized emails
// Migration 000020 - 2026-02-19
func addUserLanguage() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602190020_add_user_language",
		Migrate: func(tx *gorm.DB) error {
			// Add language column with default 'de'
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS language VARCHAR(5) NOT NULL DEFAULT 'de';
			`).Error; err != nil {
				return err
			}

			// Add column comment
			return tx.Exec(`
				COMMENT ON COLUMN users.language IS 'User preferred language for emails and UI (de, en, fr)';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS language").Error
		},
	}
}

// addUserSecurityColumns adds failed_login_attempts and locked_until columns
// for rate limiting and account lockout functionality
// Migration 000018 - 2026-02-09
func addUserSecurityColumns() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602090018_add_user_security_columns",
		Migrate: func(tx *gorm.DB) error {
			// Add failed_login_attempts column with default 0
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER NOT NULL DEFAULT 0;
			`).Error; err != nil {
				return err
			}

			// Add locked_until column (nullable timestamp)
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;
			`).Error; err != nil {
				return err
			}

			// Add index on locked_until for efficient lockout queries
			if err := tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_users_locked_until
				ON users (locked_until);
			`).Error; err != nil {
				return err
			}

			// Add column comments
			if err := tx.Exec(`
				COMMENT ON COLUMN users.failed_login_attempts IS 'Number of consecutive failed login attempts (resets on successful login)';
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
				COMMENT ON COLUMN users.locked_until IS 'Account locked until this timestamp (NULL = not locked)';
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// Drop index
			if err := tx.Exec("DROP INDEX IF EXISTS idx_users_locked_until").Error; err != nil {
				return err
			}

			// Drop columns
			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS locked_until").Error; err != nil {
				return err
			}

			if err := tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts").Error; err != nil {
				return err
			}

			return nil
		},
	}
}

// addPushSubscriptions creates the push_subscriptions table for Web Push notifications
// Migration 000021 - 2026-02-20
func addPushSubscriptions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602200021_add_push_subscriptions",
		Migrate: func(tx *gorm.DB) error {
			type PushSubscription struct {
				ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_push_subscriptions_user_id"`
				Endpoint  string    `gorm:"type:text;not null;uniqueIndex"`
				P256dhKey string    `gorm:"type:text;not null"`
				AuthKey   string    `gorm:"type:text;not null"` // #nosec G117 -- struct field name, not a hardcoded secret
				UserAgent string    `gorm:"type:text"`
				CreatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				UpdatedAt time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			if err := tx.AutoMigrate(&PushSubscription{}); err != nil {
				return err
			}

			// Add foreign key constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_push_subscriptions_user'
					) THEN
						ALTER TABLE push_subscriptions
						ADD CONSTRAINT fk_push_subscriptions_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add table comment
			return tx.Exec(`COMMENT ON TABLE push_subscriptions IS 'Web Push API subscriptions for real-time browser notifications'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS push_subscriptions CASCADE").Error
		},
	}
}

// addExpiryReminderTracking creates the expiry_reminders_sent table for tracking sent reminders
// Migration 000022 - 2026-02-20
func addExpiryReminderTracking() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602200022_add_expiry_reminder_tracking",
		Migrate: func(tx *gorm.DB) error {
			type ExpiryReminderSent struct {
				ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_expiry_reminders_user"`
				ResourceType string    `gorm:"type:varchar(50);not null"`
				ResourceID   uuid.UUID `gorm:"type:uuid;not null"`
				DaysBefore   int       `gorm:"not null"`
				SentAt       time.Time `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
			}

			if err := tx.AutoMigrate(&ExpiryReminderSent{}); err != nil {
				return err
			}

			// Add unique constraint to prevent duplicate reminders
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'expiry_reminders_sent_unique'
					) THEN
						ALTER TABLE expiry_reminder_sents
						ADD CONSTRAINT expiry_reminders_sent_unique
						UNIQUE (user_id, resource_type, resource_id, days_before);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add foreign key constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_expiry_reminders_user'
					) THEN
						ALTER TABLE expiry_reminder_sents
						ADD CONSTRAINT fk_expiry_reminders_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add table comment
			return tx.Exec(`COMMENT ON TABLE expiry_reminder_sents IS 'Tracks which expiry reminders have been sent to prevent duplicates'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS expiry_reminder_sents CASCADE").Error
		},
	}
}

// addUserTOTP creates the user_totps table for 2FA
func addUserTOTP() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602200023_add_user_totp",
		Migrate: func(tx *gorm.DB) error {
			// Create user_totps table
			if err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS user_totps (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					user_id UUID NOT NULL,
					secret TEXT NOT NULL,
					backup_codes TEXT NOT NULL,
					enabled BOOLEAN NOT NULL DEFAULT false,
					verified BOOLEAN NOT NULL DEFAULT false,
					enabled_at TIMESTAMP WITH TIME ZONE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				)
			`).Error; err != nil {
				return err
			}

			// Add unique index on user_id
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_indexes WHERE indexname = 'idx_user_totps_user_id'
					) THEN
						CREATE UNIQUE INDEX idx_user_totps_user_id ON user_totps(user_id);
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			// Add foreign key constraint
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_totps_user'
					) THEN
						ALTER TABLE user_totps
						ADD CONSTRAINT fk_user_totps_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`COMMENT ON TABLE user_totps IS 'Stores TOTP two-factor authentication configuration per user'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS user_totps CASCADE").Error
		},
	}
}

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

// addUserShareEmailsEnabled adds a preference field for opting out of share/transfer email notifications
// Migration 000026 - 2026-02-23
func addUserShareEmailsEnabled() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602230026_add_user_share_emails_enabled",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS share_emails_enabled BOOLEAN NOT NULL DEFAULT true;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN users.share_emails_enabled IS 'Whether the user wants to receive email notifications for shares and transfers';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS share_emails_enabled").Error
		},
	}
}

// addPerChannelNotificationPreferences adds 6 per-channel notification preference columns
// (2 channel toggles + 4 category toggles) and migrates old category preferences.
//
// Channels: push_notifications_enabled, email_notifications_enabled
// Categories per channel: {push,email}_expiry_enabled (expiry + validity start),
//
//	{push,email}_share_enabled (share + transfer)
//
// Data migration: copies expiry_reminders_enabled → email_reminders_enabled,
//
//	share_emails_enabled → email_sharing_enabled
//
// Migration 000027 - 2026-02-28
func addPerChannelNotificationPreferences() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602280027_add_per_channel_notification_preferences",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS push_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS email_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS push_reminders_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS push_sharing_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS email_reminders_enabled BOOLEAN NOT NULL DEFAULT true,
				ADD COLUMN IF NOT EXISTS email_sharing_enabled BOOLEAN NOT NULL DEFAULT true;
			`).Error; err != nil {
				return err
			}

			// Migrate old category preferences to new per-channel email fields
			if err := tx.Exec(`
				UPDATE users
				SET email_reminders_enabled = expiry_reminders_enabled,
				    email_sharing_enabled = share_emails_enabled;
			`).Error; err != nil {
				return err
			}

			// Add column comments
			for _, stmt := range []string{
				"COMMENT ON COLUMN users.push_notifications_enabled IS 'Global toggle for push notifications'",
				"COMMENT ON COLUMN users.email_notifications_enabled IS 'Global toggle for email notifications'",
				"COMMENT ON COLUMN users.push_reminders_enabled IS 'Push notifications for expiry reminders and validity start'",
				"COMMENT ON COLUMN users.push_sharing_enabled IS 'Push notifications for share and transfer events'",
				"COMMENT ON COLUMN users.email_reminders_enabled IS 'Email notifications for expiry reminders and validity start'",
				"COMMENT ON COLUMN users.email_sharing_enabled IS 'Email notifications for share and transfer events'",
			} {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec(`
				ALTER TABLE users
				DROP COLUMN IF EXISTS push_notifications_enabled,
				DROP COLUMN IF EXISTS email_notifications_enabled,
				DROP COLUMN IF EXISTS push_reminders_enabled,
				DROP COLUMN IF EXISTS push_sharing_enabled,
				DROP COLUMN IF EXISTS email_reminders_enabled,
				DROP COLUMN IF EXISTS email_sharing_enabled;
			`).Error
		},
	}
}

// addUserExpiryRemindersEnabled adds a preference field for opting out of expiry reminders
// Migration 000025 - 2026-02-23
func addUserExpiryRemindersEnabled() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602230025_add_user_expiry_reminders_enabled",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS expiry_reminders_enabled BOOLEAN NOT NULL DEFAULT true;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN users.expiry_reminders_enabled IS 'Whether the user wants to receive expiry reminders for vouchers and gift cards';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS expiry_reminders_enabled").Error
		},
	}
}

// addUserPasswordChangedAt adds a timestamp column to track when the password was last changed.
// Used to invalidate sessions created before the password change (M1/M2 security fix).
// Migration 000028 - 2026-02-28
func addUserPasswordChangedAt() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602280028_add_user_password_changed_at",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE users
				ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`
				COMMENT ON COLUMN users.password_changed_at IS 'When the password was last changed; sessions created before this timestamp are invalidated';
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS password_changed_at").Error
		},
	}
}

// addServerSideSessions creates the sessions table for server-side session management.
// Replaces CookieStore with PostgreSQL-backed sessions enabling session listing and revocation.
func addServerSideSessions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202602280029_add_server_side_sessions",
		Migrate: func(tx *gorm.DB) error {
			type Session struct {
				ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
				UserID       *uuid.UUID `gorm:"type:uuid;index:idx_sessions_user_id"`
				TokenHash    string     `gorm:"type:text;not null;uniqueIndex:idx_sessions_token_hash"`
				Data         []byte     `gorm:"type:bytea;not null"`
				IPAddress    string     `gorm:"type:text;not null;default:''"`
				UserAgent    string     `gorm:"type:text;not null;default:''"`
				CreatedAt    time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				LastActiveAt time.Time  `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
				ExpiresAt    time.Time  `gorm:"type:timestamp with time zone;not null;index:idx_sessions_expires_at"`
			}

			if err := tx.AutoMigrate(&Session{}); err != nil {
				return err
			}

			// Add foreign key constraint (nullable - some sessions may not have a user yet)
			if err := tx.Exec(`
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_constraint WHERE conname = 'fk_sessions_user'
					) THEN
						ALTER TABLE sessions
						ADD CONSTRAINT fk_sessions_user
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
					END IF;
				END $$;
			`).Error; err != nil {
				return err
			}

			return tx.Exec(`COMMENT ON TABLE sessions IS 'Server-side sessions for authentication, enabling session listing and revocation'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS sessions CASCADE").Error
		},
	}
}

// fixNotificationPreferenceDefaults changes notification preference defaults from true to false.
// Email preferences are now only enabled when email is verified.
// Push preferences are now only enabled when a push subscription is registered.
// This migration updates existing users: disable email prefs for unverified users,
// disable push prefs for users without push subscriptions.
func fixNotificationPreferenceDefaults() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202603160030_fix_notification_preference_defaults",
		Migrate: func(tx *gorm.DB) error {
			// Change column defaults from true to false
			if err := tx.Exec(`
				ALTER TABLE users
				ALTER COLUMN push_notifications_enabled SET DEFAULT false,
				ALTER COLUMN email_notifications_enabled SET DEFAULT false,
				ALTER COLUMN push_reminders_enabled SET DEFAULT false,
				ALTER COLUMN push_sharing_enabled SET DEFAULT false,
				ALTER COLUMN email_reminders_enabled SET DEFAULT false,
				ALTER COLUMN email_sharing_enabled SET DEFAULT false;
			`).Error; err != nil {
				return err
			}

			// Disable email preferences for users who have NOT verified their email
			if err := tx.Exec(`
				UPDATE users SET
					email_notifications_enabled = false,
					email_reminders_enabled = false,
					email_sharing_enabled = false
				WHERE email_verified = false;
			`).Error; err != nil {
				return err
			}

			// Disable push preferences for users who have NO push subscriptions
			if err := tx.Exec(`
				UPDATE users SET
					push_notifications_enabled = false,
					push_reminders_enabled = false,
					push_sharing_enabled = false
				WHERE id NOT IN (
					SELECT DISTINCT user_id FROM push_subscriptions
				);
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`COMMENT ON COLUMN users.email_notifications_enabled IS 'Global email channel — auto-enabled on email verification'`).Error; err != nil {
				return err
			}

			return tx.Exec(`COMMENT ON COLUMN users.push_notifications_enabled IS 'Global push channel — auto-enabled on first push subscription'`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Restore original defaults
			if err := tx.Exec(`
				ALTER TABLE users
				ALTER COLUMN push_notifications_enabled SET DEFAULT true,
				ALTER COLUMN email_notifications_enabled SET DEFAULT true,
				ALTER COLUMN push_reminders_enabled SET DEFAULT true,
				ALTER COLUMN push_sharing_enabled SET DEFAULT true,
				ALTER COLUMN email_reminders_enabled SET DEFAULT true,
				ALTER COLUMN email_sharing_enabled SET DEFAULT true;
			`).Error; err != nil {
				return err
			}

			// Re-enable all preferences (original behavior)
			return tx.Exec(`
				UPDATE users SET
					push_notifications_enabled = true,
					email_notifications_enabled = true,
					push_reminders_enabled = true,
					push_sharing_enabled = true,
					email_reminders_enabled = true,
					email_sharing_enabled = true;
			`).Error
		},
	}
}
