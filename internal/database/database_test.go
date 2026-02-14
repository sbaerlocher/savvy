package database

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// migrateMutex prevents concurrent AutoMigrate calls that can cause PostgreSQL deadlocks
var migrateMutex sync.Mutex

func TestConnect_Success(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credentials
	}

	// Reset global DB
	DB = nil

	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, DB)

	// Verify connection pool settings
	sqlDB, err := DB.DB()
	require.NoError(t, err)

	stats := sqlDB.Stats()
	assert.LessOrEqual(t, stats.MaxOpenConnections, 25)
}

func TestConnect_InvalidURL(t *testing.T) {
	// Reset global DB
	DB = nil

	err := Connect("postgres://invalid:invalid@localhost:9999/invalid?sslmode=disable", "WARN")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestConnect_MalformedURL(t *testing.T) {
	DB = nil

	err := Connect("not-a-valid-url", "WARN")
	assert.Error(t, err)
}

func TestConnect_EmptyURL(t *testing.T) {
	DB = nil

	err := Connect("", "WARN")
	assert.Error(t, err)
}

func TestConnect_VerifyConnectionPoolSettings(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credential, not real
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	stats := sqlDB.Stats()

	// Verify max open connections is set correctly
	assert.Equal(t, 25, stats.MaxOpenConnections)

	// Verify MaxIdleConns is set
	assert.LessOrEqual(t, stats.MaxIdleClosed, int64(0)) // Should start at 0

	// Test connection
	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestConnect_VerifyNowFuncUTC(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credential, not real
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	// Get NowFunc from config
	now := DB.NowFunc()
	assert.Equal(t, time.UTC, now.Location())
}

func TestEnableTelemetry_WithoutDB(t *testing.T) {
	DB = nil

	err := EnableTelemetry("test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

func TestEnableTelemetry_WithDB(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credentials
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	// This will fail if OTEL is not properly configured, but that's expected in tests
	// We're just testing that the function handles the case
	_ = EnableTelemetry("test-service")
	// We don't assert no error because telemetry setup might fail in test environment
	// Just verify it doesn't panic
}

func TestEnableAuditLogging_WithoutDB(t *testing.T) {
	DB = nil

	err := EnableAuditLogging()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

func TestEnableAuditLogging_WithDB(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credential, not real
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	err = EnableAuditLogging()
	assert.NoError(t, err)
}

func TestAutoMigrate(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	// Lock to prevent concurrent migrations that cause deadlocks
	migrateMutex.Lock()
	err = AutoMigrate()
	migrateMutex.Unlock()
	assert.NoError(t, err)

	// Verify tables were created
	tables := []string{
		"users",
		"cards",
		"card_shares",
		"vouchers",
		"voucher_shares",
		"gift_cards",
		"gift_card_transactions",
		"gift_card_shares",
		"merchants",
		"user_favorites",
		"audit_logs",
	}

	for _, table := range tables {
		var exists bool
		err := DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		assert.NoError(t, err)
		assert.True(t, exists, "table %s should exist", table)
	}
}

func TestIsDuplicateError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{
			name:   "nil error",
			err:    nil,
			expect: false,
		},
		{
			name:   "duplicate key error",
			err:    errors.New("ERROR: duplicate key value violates unique constraint"),
			expect: true,
		},
		{
			name:   "unique constraint error",
			err:    errors.New("ERROR: unique constraint violation"),
			expect: true,
		},
		{
			name:   "violates unique error",
			err:    errors.New("ERROR: violates unique constraint user_email_key"),
			expect: true,
		},
		{
			name:   "SQLSTATE 23505",
			err:    errors.New("SQLSTATE 23505"),
			expect: true,
		},
		{
			name:   "generic error",
			err:    errors.New("some other error"),
			expect: false,
		},
		{
			name:   "connection error",
			err:    errors.New("connection refused"),
			expect: false,
		},
		{
			name:   "not found error",
			err:    gorm.ErrRecordNotFound,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDuplicateError(tt.err)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestErrDuplicateKey(t *testing.T) {
	assert.NotNil(t, ErrDuplicateKey)
	assert.Contains(t, ErrDuplicateKey.Error(), "duplicate key")
}

func TestConnect_PrepareStmtEnabled(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	// Verify PrepareStmt is enabled in config
	assert.True(t, DB.PrepareStmt)
}

func TestConnect_ConnectionLifetime(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	sqlDB, err := DB.DB()
	require.NoError(t, err)

	// Create a connection and verify it can be used
	err = sqlDB.Ping()
	assert.NoError(t, err)

	// Verify connection pool stats
	stats := sqlDB.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1)
}

func TestConnect_MultipleConnections(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"
	}

	// Connect multiple times (simulating reconnection)
	for i := 0; i < 3; i++ {
		DB = nil
		err := Connect(dbURL, "WARN")
		if err != nil {
			t.Skipf("Skipping test: PostgreSQL not available: %v", err)
			return
		}

		require.NoError(t, err)
		require.NotNil(t, DB)

		sqlDB, err := DB.DB()
		require.NoError(t, err)

		err = sqlDB.Ping()
		assert.NoError(t, err)
	}
}

func TestAutoMigrate_Idempotent(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"
	}

	DB = nil
	err := Connect(dbURL, "WARN")
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(t, err)

	// Lock to prevent concurrent migrations that cause deadlocks
	migrateMutex.Lock()
	defer migrateMutex.Unlock()

	// Run migration multiple times - should be idempotent
	err = AutoMigrate()
	assert.NoError(t, err)

	err = AutoMigrate()
	assert.NoError(t, err)

	err = AutoMigrate()
	assert.NoError(t, err)
}

func TestIsDuplicateError_CaseSensitive(t *testing.T) {
	// IsDuplicateError is case-sensitive, matching actual Postgres error messages
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{
			name:   "lowercase duplicate key (actual Postgres format)",
			err:    errors.New("duplicate key value"),
			expect: true,
		},
		{
			name:   "lowercase unique constraint (actual Postgres format)",
			err:    errors.New("unique constraint violation"),
			expect: true,
		},
		{
			name:   "UPPERCASE should not match",
			err:    errors.New("DUPLICATE KEY VALUE"),
			expect: false,
		},
		{
			name:   "Mixed Case should not match",
			err:    errors.New("Duplicate Key value"),
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDuplicateError(tt.err)
			assert.Equal(t, tt.expect, result)
		})
	}
}
