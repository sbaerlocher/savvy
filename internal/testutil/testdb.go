// Package testutil provides shared test helpers for database integration tests.
//
// Each test gets its own transaction that is automatically rolled back after
// the test completes. This provides complete isolation between tests —
// even across packages running in parallel — without any cleanup logic.
package testutil

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"savvy/internal/models"
)

var (
	sharedDB *gorm.DB
	once     sync.Once
	initErr  error
)

// allModels lists every model that needs to exist in the test database.
var allModels = []interface{}{
	&models.User{},
	&models.Merchant{},
	&models.Card{},
	&models.CardShare{},
	&models.Voucher{},
	&models.VoucherShare{},
	&models.GiftCard{},
	&models.GiftCardShare{},
	&models.GiftCardTransaction{},
	&models.UserFavorite{},
	&models.AuditLog{},
	&models.Notification{},
	&models.Session{},
	&models.UserTOTP{},
	&models.EmailToken{},
	&models.PushSubscription{},
	&models.ExpiryReminderSent{},
}

// initSharedDB opens ONE connection to the test database and runs migrations.
// Called once per process via sync.Once.
func initSharedDB() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credential
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		initErr = fmt.Errorf("PostgreSQL not available: %w", err)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		initErr = fmt.Errorf("failed to get underlying DB: %w", err)
		return
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	if err := db.AutoMigrate(allModels...); err != nil {
		initErr = fmt.Errorf("failed to migrate: %w", err)
		return
	}

	// One-time cleanup of stale test data from previous runs.
	// With transaction isolation, each test rolls back its own data,
	// but old data from non-transactional runs may linger.
	db.Exec(`DO $$
	BEGIN
		DELETE FROM user_favorites WHERE user_id IN (SELECT id FROM users WHERE email NOT LIKE '%@production.%');
		DELETE FROM card_shares WHERE card_id IN (SELECT id FROM cards WHERE card_number LIKE 'TEST%' OR card_number LIKE 'PRELOAD%' OR card_number LIKE 'DASH%' OR card_number LIKE 'SHARED%' OR card_number LIKE 'DELETE%' OR card_number LIKE 'GIFT%' OR card_number LIKE 'GC%' OR card_number LIKE 'ORIGINAL' OR card_number LIKE 'UPDATED' OR card_number LIKE 'COUNT%' OR card_number LIKE 'NOT%' OR card_number LIKE '12345%' OR card_number LIKE 'BAL%');
		DELETE FROM voucher_shares WHERE voucher_id IN (SELECT id FROM vouchers WHERE code LIKE 'TEST%' OR code LIKE 'DASH%' OR code LIKE 'GET%' OR code LIKE 'V1' OR code LIKE 'V2' OR code LIKE 'ORIGINAL%' OR code LIKE 'UPDATED%' OR code LIKE 'DELETE%' OR code LIKE 'COUNT%' OR code LIKE 'SHARED%' OR code LIKE 'NOT%' OR code LIKE 'PRELOAD%');
		DELETE FROM gift_card_shares WHERE gift_card_id IN (SELECT id FROM gift_cards WHERE card_number LIKE 'TEST%' OR card_number LIKE 'GIFT%' OR card_number LIKE 'GC%' OR card_number LIKE 'DASH%' OR card_number LIKE 'SHARED%' OR card_number LIKE 'NOT%' OR card_number LIKE 'DELETE%' OR card_number LIKE 'UPDATE%' OR card_number LIKE 'COUNT%' OR card_number LIKE 'BAL%' OR card_number LIKE 'PRELOAD%');
		DELETE FROM gift_card_transactions WHERE gift_card_id IN (SELECT id FROM gift_cards WHERE card_number LIKE 'TEST%' OR card_number LIKE 'GIFT%' OR card_number LIKE 'GC%' OR card_number LIKE 'DASH%');
		DELETE FROM notifications WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'test-%@example.com' OR email LIKE '%@test.com' OR email LIKE 'dashboard%' OR email LIKE 'admin@example.com' OR email LIKE 'user@example.com' OR email LIKE 'user1@example.com' OR email LIKE 'user2@example.com');
		DELETE FROM audit_logs WHERE resource_type = 'test' OR user_id IN (SELECT id FROM users WHERE email LIKE 'test-%@example.com' OR email LIKE '%@test.com' OR email LIKE '%@example.com');
		DELETE FROM cards WHERE card_number LIKE 'TEST%' OR card_number LIKE 'PRELOAD%' OR card_number LIKE 'DASH%' OR card_number LIKE 'SHARED%' OR card_number LIKE 'DELETE%' OR card_number LIKE 'ORIGINAL' OR card_number LIKE 'UPDATED' OR card_number LIKE 'COUNT%' OR card_number LIKE 'NOT%' OR card_number LIKE '12345%';
		DELETE FROM vouchers WHERE code LIKE 'TEST%' OR code LIKE 'DASH%' OR code LIKE 'GET%';
		DELETE FROM gift_cards WHERE card_number LIKE 'TEST%' OR card_number LIKE 'GIFT%' OR card_number LIKE 'GC%' OR card_number LIKE 'DASH%';
		DELETE FROM merchants WHERE name LIKE 'Test%';
		DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'test-%@example.com' OR email LIKE '%@test.com' OR email LIKE '%@example.com');
		DELETE FROM users WHERE email LIKE 'test-%@example.com' OR email LIKE '%@test.com' OR email LIKE '%@example.com';
	END $$`)

	sharedDB = db
}

// NewTestDB returns a *gorm.DB that runs inside a transaction.
// The transaction is automatically rolled back when the test ends,
// so every test starts with a clean slate — no TRUNCATE or DELETE needed.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    db := testutil.NewTestDB(t)
//	    // use db normally — all changes are rolled back after test
//	}
// NewTestDBDirect returns a non-transactional *gorm.DB with schema-based cleanup.
// Use this ONLY for tests that are incompatible with transaction isolation:
//   - Tests that use goroutines for parallel queries (e.g. dashboard service)
//   - Tests with GORM hooks that use sqlDB.ExecContext (bypasses transaction)
//
// Data is cleaned up via TRUNCATE on test completion.
func NewTestDBDirect(t *testing.T) *gorm.DB {
	t.Helper()

	once.Do(initSharedDB)
	if initErr != nil {
		t.Skipf("Skipping test: %v", initErr)
		return nil
	}

	// Use a unique schema per test to avoid conflicts with other parallel tests
	schemaName := fmt.Sprintf("test_%s", sanitizeName(t.Name()))

	// Create schema, set search_path, migrate, and clean up on exit
	sharedDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	sharedDB.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))

	// Open a new connection with this schema as search_path
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credential
	}
	dbURL += fmt.Sprintf("&search_path=%s,public", schemaName)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open schema DB: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)

	if err := db.AutoMigrate(allModels...); err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
		sharedDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	})

	return db
}

// sanitizeName converts a test name to a valid PostgreSQL schema name.
func sanitizeName(name string) string {
	result := make([]byte, 0, len(name))
	for _, c := range []byte(name) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			result = append(result, c)
		} else if c >= 'A' && c <= 'Z' {
			result = append(result, c+32) // lowercase
		} else {
			result = append(result, '_')
		}
	}
	// Truncate to fit PostgreSQL identifier limit (63 chars)
	if len(result) > 50 {
		result = result[:50]
	}
	return string(result)
}

func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	once.Do(initSharedDB)
	if initErr != nil {
		t.Skipf("Skipping test: %v", initErr)
		return nil
	}

	tx := sharedDB.Begin()
	if tx.Error != nil {
		t.Fatalf("Failed to begin transaction: %v", tx.Error)
	}

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
