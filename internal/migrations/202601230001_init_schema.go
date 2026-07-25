package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

			// Add foreign key constraints.
			// ponytail: each FK is guarded by IF NOT EXISTS so a half-applied
			// init migration (constraints created but migration not recorded)
			// can replay without colliding on SQLSTATE 42710.
			foreignKeys := []struct {
				name, table, definition string
			}{
				{"fk_cards_user", "cards", "FOREIGN KEY (user_id) REFERENCES users(id)"},
				{"fk_cards_merchant", "cards", "FOREIGN KEY (merchant_id) REFERENCES merchants(id)"},
				{"fk_card_shares_card", "card_shares", "FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE"},
				{"fk_card_shares_user", "card_shares", "FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE"},
				{"fk_vouchers_user", "vouchers", "FOREIGN KEY (user_id) REFERENCES users(id)"},
				{"fk_vouchers_merchant", "vouchers", "FOREIGN KEY (merchant_id) REFERENCES merchants(id)"},
				{"fk_voucher_shares_voucher", "voucher_shares", "FOREIGN KEY (voucher_id) REFERENCES vouchers(id) ON DELETE CASCADE"},
				{"fk_voucher_shares_user", "voucher_shares", "FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE"},
				{"fk_gift_cards_user", "gift_cards", "FOREIGN KEY (user_id) REFERENCES users(id)"},
				{"fk_gift_cards_merchant", "gift_cards", "FOREIGN KEY (merchant_id) REFERENCES merchants(id)"},
				{"fk_gift_card_transactions_gift_card", "gift_card_transactions", "FOREIGN KEY (gift_card_id) REFERENCES gift_cards(id) ON DELETE CASCADE"},
				{"fk_gift_card_shares_gift_card", "gift_card_shares", "FOREIGN KEY (gift_card_id) REFERENCES gift_cards(id) ON DELETE CASCADE"},
				{"fk_gift_card_shares_user", "gift_card_shares", "FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE"},
			}
			for _, fk := range foreignKeys {
				if err := validateSQLIdentifiers(fk.name, fk.table); err != nil {
					return fmt.Errorf("addForeignKey %s: %w", fk.name, err)
				}
				if err := tx.Exec(fmt.Sprintf(`
					DO $$
					BEGIN
						IF NOT EXISTS (
							SELECT 1 FROM pg_constraint WHERE conname = '%s'
						) THEN
							ALTER TABLE %s ADD CONSTRAINT %s %s;
						END IF;
					END $$;
				`, fk.name, fk.table, fk.name, fk.definition)).Error; err != nil {
					return err
				}
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
				if err := validateSQLIdentifiers(table); err != nil {
					return fmt.Errorf("dropTable rollback: %w", err)
				}
				if err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
					return err
				}
			}

			return tx.Exec(`DROP EXTENSION IF EXISTS "pgcrypto" CASCADE`).Error
		},
	}
}
