package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// setupTestDB creates a test database connection.
// Uses DATABASE_URL env var or falls back to local PostgreSQL from docker-compose.
func setupTestDB(t *testing.T) *gorm.DB {
	// Check if DATABASE_URL is set (Docker/CI environment)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback to local docker-compose PostgreSQL
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable" // #nosec G101 -- test credential, not real
	}

	// Use PostgreSQL from environment (production-like testing)
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return nil
	}

	// Limit connection pool to prevent deadlocks with parallel test packages
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	t.Cleanup(func() { _ = sqlDB.Close() })

	// Auto-migrate models
	err = db.AutoMigrate(
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
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Use targeted DELETEs (not TRUNCATE) to avoid deadlocks with
	// parallel test packages that share the same database.
	db.Exec(`DO $$
	BEGIN
		DELETE FROM user_favorites WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM card_shares WHERE shared_with_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM voucher_shares WHERE shared_with_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM gift_card_shares WHERE shared_with_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM gift_card_transactions WHERE gift_card_id IN (SELECT id FROM gift_cards WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com'));
		DELETE FROM audit_logs WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM cards WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM vouchers WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM gift_cards WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
		DELETE FROM merchants WHERE name LIKE 'Test%';
		DELETE FROM users WHERE email LIKE '%@example.com';
	END $$`)

	return db
}

// setupAuthzService creates an AuthzService with real repositories backed by the test DB.
func setupAuthzService(db *gorm.DB) AuthzServiceInterface {
	return NewAuthzService(
		repository.NewCardRepository(db),
		repository.NewVoucherRepository(db),
		repository.NewGiftCardRepository(db),
		repository.NewCardShareRepository(db),
		repository.NewVoucherShareRepository(db),
		repository.NewGiftCardShareRepository(db),
	)
}

func TestAuthzService_CheckCardAccess_Owner(t *testing.T) {
	db := setupTestDB(t)
	service := setupAuthzService(db)

	// Create test user with explicit ID for SQLite
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "owner@example.com",
		PasswordHash: "hashed",
	}
	db.Create(user)

	// Create test card owned by user
	card := &models.Card{
		UserID:       &user.ID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}
	db.Create(card)

	// Test: Owner should have full access
	perms, err := service.CheckCardAccess(context.Background(), user.ID, card.ID)

	assert.NoError(t, err)
	assert.NotNil(t, perms)
	assert.True(t, perms.IsOwner)
	assert.True(t, perms.CanView)
	assert.True(t, perms.CanEdit)
	assert.True(t, perms.CanDelete)
}

func TestAuthzService_CheckCardAccess_SharedUser(t *testing.T) {
	db := setupTestDB(t)
	service := setupAuthzService(db)

	// Create owner
	owner := &models.User{
		Email:        "owner@example.com",
		PasswordHash: "hashed",
	}
	db.Create(owner)

	// Create shared user
	sharedUser := &models.User{
		Email:        "shared@example.com",
		PasswordHash: "hashed",
	}
	db.Create(sharedUser)

	// Create card owned by owner
	card := &models.Card{
		UserID:       &owner.ID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}
	db.Create(card)

	// Create share with edit permission only
	share := &models.CardShare{
		CardID:       card.ID,
		SharedWithID: sharedUser.ID,
		CanEdit:      true,
		CanDelete:    false,
	}
	db.Create(share)

	// Test: Shared user should have limited access
	perms, err := service.CheckCardAccess(context.Background(), sharedUser.ID, card.ID)

	assert.NoError(t, err)
	assert.NotNil(t, perms)
	assert.False(t, perms.IsOwner)
	assert.True(t, perms.CanView)
	assert.True(t, perms.CanEdit)    // Granted by share
	assert.False(t, perms.CanDelete) // Not granted
}

func TestAuthzService_CheckCardAccess_NoAccess(t *testing.T) {
	db := setupTestDB(t)
	service := setupAuthzService(db)

	// Create owner
	owner := &models.User{
		Email:        "owner@example.com",
		PasswordHash: "hashed",
	}
	db.Create(owner)

	// Create unauthorized user
	unauthorized := &models.User{
		Email:        "unauthorized@example.com",
		PasswordHash: "hashed",
	}
	db.Create(unauthorized)

	// Create card owned by owner (no share)
	card := &models.Card{
		UserID:       &owner.ID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}
	db.Create(card)

	// Test: Unauthorized user should be denied
	perms, err := service.CheckCardAccess(context.Background(), unauthorized.ID, card.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Nil(t, perms)
}

func TestAuthzService_CheckCardAccess_NonExistentCard(t *testing.T) {
	db := setupTestDB(t)
	service := setupAuthzService(db)

	// Create test user
	user := &models.User{
		Email:        "user@example.com",
		PasswordHash: "hashed",
	}
	db.Create(user)

	// Test: Non-existent card should return ErrForbidden
	nonExistentID := uuid.New()
	perms, err := service.CheckCardAccess(context.Background(), user.ID, nonExistentID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Nil(t, perms)
}

func TestAuthzService_CheckGiftCardAccess_TransactionPermission(t *testing.T) {
	db := setupTestDB(t)
	service := setupAuthzService(db)

	// Create owner
	owner := &models.User{
		Email:        "owner@example.com",
		PasswordHash: "hashed",
	}
	db.Create(owner)

	// Create shared user
	sharedUser := &models.User{
		Email:        "shared@example.com",
		PasswordHash: "hashed",
	}
	db.Create(sharedUser)

	// Create gift card
	giftCard := &models.GiftCard{
		UserID:         &owner.ID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Merchant",
		InitialBalance: 100.0,
		CurrentBalance: 100.0,
	}
	db.Create(giftCard)

	// Create share with transaction permission
	share := &models.GiftCardShare{
		GiftCardID:          giftCard.ID,
		SharedWithID:        sharedUser.ID,
		CanEdit:             false,
		CanDelete:           false,
		CanEditTransactions: true, // Special gift card permission
	}
	db.Create(share)

	// Test: Shared user should have transaction permission
	perms, err := service.CheckGiftCardAccess(context.Background(), sharedUser.ID, giftCard.ID)

	assert.NoError(t, err)
	assert.NotNil(t, perms)
	assert.False(t, perms.IsOwner)
	assert.True(t, perms.CanView)
	assert.False(t, perms.CanEdit)
	assert.False(t, perms.CanDelete)
	assert.True(t, perms.CanEditTransactions) // Gift card specific permission
}

func TestAuthzService_CheckVoucherAccess_Owner(t *testing.T) {
	db := setupTestDB(t)
	service := setupAuthzService(db)

	// Create owner
	owner := &models.User{
		Email:        "voucher-owner@example.com",
		PasswordHash: "hashed",
	}
	db.Create(owner)

	// Create voucher
	voucher := &models.Voucher{
		UserID:         &owner.ID,
		Code:           "TEST-VOUCHER",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	// Test: Owner should have full access
	perms, err := service.CheckVoucherAccess(context.Background(), owner.ID, voucher.ID)

	assert.NoError(t, err)
	assert.NotNil(t, perms)
	assert.True(t, perms.IsOwner)
	assert.True(t, perms.CanView)
	assert.True(t, perms.CanEdit)
	assert.True(t, perms.CanDelete)
}

func TestAuthzService_CheckVoucherAccess_SharedUser(t *testing.T) {
	db := setupTestDB(t)
	service := setupAuthzService(db)

	// Create owner
	owner := &models.User{
		Email:        "voucher-owner2@example.com",
		PasswordHash: "hashed",
	}
	db.Create(owner)

	// Create shared user
	sharedUser := &models.User{
		Email:        "voucher-shared@example.com",
		PasswordHash: "hashed",
	}
	db.Create(sharedUser)

	// Create voucher
	voucher := &models.Voucher{
		UserID:         &owner.ID,
		Code:           "TEST-VOUCHER-SHARED",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	// Create share (vouchers are always read-only when shared)
	share := &models.VoucherShare{
		VoucherID:    voucher.ID,
		SharedWithID: sharedUser.ID,
	}
	db.Create(share)

	// Test: Shared user should have view-only access
	perms, err := service.CheckVoucherAccess(context.Background(), sharedUser.ID, voucher.ID)

	assert.NoError(t, err)
	assert.NotNil(t, perms)
	assert.False(t, perms.IsOwner)
	assert.True(t, perms.CanView)
	assert.False(t, perms.CanEdit)   // Vouchers are always read-only
	assert.False(t, perms.CanDelete) // Shared users can't delete
}
