package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
		UsageLimitType: "unlimited",
	}

	err := repo.Create(ctx, voucher)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, voucher.ID)

	db.Exec("DELETE FROM vouchers WHERE id = ?", voucher.ID)
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
	defer db.Exec("DELETE FROM vouchers WHERE id = ?", voucher.ID)

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
		defer db.Exec("DELETE FROM vouchers WHERE id = ?", vouchers[i].ID)
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
		UsageLimitType: "unlimited",
	}
	db.Create(voucher)
	defer db.Exec("DELETE FROM vouchers WHERE id = ?", voucher.ID)

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
	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "DELETE-ME",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "unlimited",
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
		UsageLimitType: "unlimited",
	}
	db.Create(voucher)
	defer db.Exec("DELETE FROM vouchers WHERE id = ?", voucher.ID)

	newCount, err := repo.Count(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, initialCount+1, newCount)
}
