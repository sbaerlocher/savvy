package services

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"savvy/internal/models"
)

// setupTestDB creates a test database connection.
// Requires PostgreSQL via DATABASE_URL env var (Docker/CI).
// Skips tests locally if DATABASE_URL is not set.
func setupTestDB(t *testing.T) *gorm.DB {
	// Check if DATABASE_URL is set (Docker/CI environment)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping test: DATABASE_URL not set. Run tests in Docker with PostgreSQL.")
		return nil
	}

	// Use PostgreSQL from environment (production-like testing)
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL test database: %v", err)
	}

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
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Clean up tables before each test
	db.Exec("TRUNCATE users, merchants, cards, card_shares, vouchers, voucher_shares, gift_cards, gift_card_shares CASCADE")

	return db
}

func TestAuthzService_CheckCardAccess_Owner(t *testing.T) {
	db := setupTestDB(t)
	service := NewAuthzService(db)

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
	service := NewAuthzService(db)

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
	assert.True(t, perms.CanEdit)   // Granted by share
	assert.False(t, perms.CanDelete) // Not granted
}

func TestAuthzService_CheckCardAccess_NoAccess(t *testing.T) {
	db := setupTestDB(t)
	service := NewAuthzService(db)

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
	service := NewAuthzService(db)

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
	service := NewAuthzService(db)

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
