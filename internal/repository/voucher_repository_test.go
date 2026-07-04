package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"savvy/internal/models"
)

func TestVoucherRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "TEST-VOUCHER-123",
		MerchantName:   "Test Merchant",
		ValidFrom:      time.Now().Add(-24 * time.Hour),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}

	err := repo.Create(ctx, voucher)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, voucher.ID)

}

func TestVoucherRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "GET-TEST",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	found, err := repo.GetByID(ctx, voucher.ID)
	assert.NoError(t, err)
	assert.Equal(t, voucher.ID, found.ID)
	assert.Equal(t, "GET-TEST", found.Code)
}

func TestVoucherRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	validFrom := time.Now()
	validUntil := time.Now().Add(24 * time.Hour)

	vouchers := []models.Voucher{
		{UserID: &userID, Code: "V1", MerchantName: "M1", ValidFrom: validFrom, ValidUntil: validUntil, UsageLimitType: "unlimited"},
		{UserID: &userID, Code: "V2", MerchantName: "M2", ValidFrom: validFrom, ValidUntil: validUntil, UsageLimitType: "unlimited"},
	}
	for i := range vouchers {
		db.Create(&vouchers[i])
	}

	found, err := repo.GetByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestVoucherRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "ORIGINAL-CODE",
		MerchantName:   "Original",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	voucher.Code = "UPDATED-CODE"
	err := repo.Update(ctx, voucher)
	assert.NoError(t, err)

	var found models.Voucher
	db.First(&found, "id = ?", voucher.ID)
	assert.Equal(t, "UPDATED-CODE", found.Code)
}

func TestVoucherRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	uniqueCode := "DELETE-ME-" + uuid.New().String()[:8]
	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           uniqueCode,
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	err := repo.Delete(ctx, voucher.ID)
	assert.NoError(t, err)
}

func TestVoucherRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	initialCount, _ := repo.Count(ctx, userID)

	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "COUNT-TEST",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	newCount, err := repo.Count(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, initialCount+1, newCount)
}

func TestVoucherRepository_GetSharedWithUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	// Create owner and shared user
	ownerID := createTestUser(t, db)
	sharedUserID := createTestUser(t, db)

	validFrom := time.Now()
	validUntil := time.Now().Add(24 * time.Hour)

	// Create vouchers owned by owner
	vouchers := []models.Voucher{
		{UserID: &ownerID, Code: "SHARED_V1", MerchantName: "Shared M1", ValidFrom: validFrom, ValidUntil: validUntil, UsageLimitType: "unlimited"},
		{UserID: &ownerID, Code: "SHARED_V2", MerchantName: "Shared M2", ValidFrom: validFrom, ValidUntil: validUntil, UsageLimitType: "unlimited"},
		{UserID: &ownerID, Code: "NOT_SHARED_V", MerchantName: "Not Shared", ValidFrom: validFrom, ValidUntil: validUntil, UsageLimitType: "unlimited"},
	}
	for i := range vouchers {
		db.Create(&vouchers[i])
	}

	// Share first two vouchers with sharedUserID
	shares := []models.VoucherShare{
		{VoucherID: vouchers[0].ID, SharedWithID: sharedUserID},
		{VoucherID: vouchers[1].ID, SharedWithID: sharedUserID},
	}
	for i := range shares {
		db.Create(&shares[i])
	}

	// Get shared vouchers
	found, err := repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)

	// Verify the right vouchers are returned
	voucherCodes := make(map[string]bool)
	for _, voucher := range found {
		voucherCodes[voucher.Code] = true
	}
	assert.True(t, voucherCodes["SHARED_V1"])
	assert.True(t, voucherCodes["SHARED_V2"])
	assert.False(t, voucherCodes["NOT_SHARED_V"])
}

func TestVoucherRepository_GetSharedWithUser_ExcludesDeleted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	ownerID := createTestUser(t, db)
	sharedUserID := createTestUser(t, db)

	validFrom := time.Now()
	validUntil := time.Now().Add(24 * time.Hour)

	// Create a voucher
	voucher := &models.Voucher{
		UserID:         &ownerID,
		Code:           "DELETED_SHARE_V",
		MerchantName:   "Test",
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	// Create a share
	share := &models.VoucherShare{
		VoucherID:    voucher.ID,
		SharedWithID: sharedUserID,
	}
	db.Create(share)

	// Get shared vouchers (should include it)
	found, err := repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 1)

	// Soft delete the share
	db.Delete(share)

	// Get shared vouchers again (should exclude it now)
	found, err = repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)

	// Verify the deleted share is not included
	for _, v := range found {
		assert.NotEqual(t, "DELETED_SHARE_V", v.Code)
	}
}

func TestVoucherRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestVoucherRepository_GetByID_WithPreloads(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create a merchant
	merchant := &models.Merchant{
		Name:  "Test Voucher Merchant",
		Color: "#00FF00",
	}
	db.Create(merchant)

	validFrom := time.Now()
	validUntil := time.Now().Add(24 * time.Hour)

	// Create a voucher with merchant
	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "PRELOAD_VOUCHER",
		MerchantName:   "Test",
		MerchantID:     &merchant.ID,
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	// Get with preloads
	found, err := repo.GetByID(ctx, voucher.ID, "Merchant", "User")
	assert.NoError(t, err)
	assert.NotNil(t, found.Merchant)
	assert.Equal(t, merchant.ID, found.Merchant.ID)
	assert.NotNil(t, found.User)
	assert.Equal(t, userID, found.User.ID)
}

func TestVoucherRepository_FindDeletedByCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVoucherRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	voucher := &models.Voucher{
		UserID:    &userID,
		Code:      "DEL-V-1",
		ValidFrom: time.Now().Add(-24 * time.Hour),
		ValidUntil: time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, repo.Create(ctx, voucher))
	require.NoError(t, repo.Delete(ctx, voucher.ID)) // soft-delete

	// Active lookup does not see it
	active, err := repo.FindByVoucherCode(ctx, "DEL-V-1", userID)
	require.NoError(t, err)
	require.Nil(t, active)

	// Deleted lookup finds it
	deleted, err := repo.FindDeletedByCode(ctx, "DEL-V-1", userID)
	require.NoError(t, err)
	require.NotNil(t, deleted)
	require.Equal(t, voucher.ID, deleted.ID)

	// Restore brings it back
	require.NoError(t, repo.RestoreByID(ctx, voucher.ID, userID))
	active2, err := repo.FindByVoucherCode(ctx, "DEL-V-1", userID)
	require.NoError(t, err)
	require.NotNil(t, active2)
}
