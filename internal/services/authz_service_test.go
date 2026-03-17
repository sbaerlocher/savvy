package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/repository"
	"savvy/internal/testutil"
)

// setupTestDB returns a transaction-isolated test database.
// Every test gets its own transaction that is rolled back automatically,
// so no TRUNCATE or DELETE cleanup is needed.
func setupTestDB(t *testing.T) *gorm.DB {
	return testutil.NewTestDB(t)
}

// setupDirectTestDB returns a non-transactional test database with schema isolation.
// Use for tests that are incompatible with transactions (goroutines, GORM hooks).
func setupDirectTestDB(t *testing.T) *gorm.DB {
	return testutil.NewTestDBDirect(t)
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
		FirstName:    "Test",
		LastName:     "User",
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
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(owner)

	// Create shared user
	sharedUser := &models.User{
		Email:        "shared@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
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
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(owner)

	// Create unauthorized user
	unauthorized := &models.User{
		Email:        "unauthorized@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
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
		FirstName:    "Test",
		LastName:     "User",
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
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(owner)

	// Create shared user
	sharedUser := &models.User{
		Email:        "shared@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
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
		FirstName:    "Test",
		LastName:     "User",
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
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(owner)

	// Create shared user
	sharedUser := &models.User{
		Email:        "voucher-shared@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
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
