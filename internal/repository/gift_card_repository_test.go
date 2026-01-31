package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"savvy/internal/models"
)

func TestGiftCardRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "GIFT-123",
		MerchantName:   "Test Merchant",
		InitialBalance: 100.0,
		CurrentBalance: 100.0,
		Currency:       "CHF",
	}

	err := repo.Create(ctx, giftCard)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, giftCard.ID)

	db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCard.ID)
}

func TestGiftCardRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "GIFT-GET",
		MerchantName:   "Test",
		InitialBalance: 50.0,
		CurrentBalance: 50.0,
		Currency:       "CHF",
	}
	db.Create(giftCard)
	defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCard.ID)

	found, err := repo.GetByID(ctx, giftCard.ID)
	assert.NoError(t, err)
	assert.Equal(t, giftCard.ID, found.ID)
	assert.Equal(t, "GIFT-GET", found.CardNumber)
	assert.Equal(t, 50.0, found.CurrentBalance)
}

func TestGiftCardRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	giftCards := []models.GiftCard{
		{UserID: &userID, CardNumber: "GC1", MerchantName: "M1", InitialBalance: 100, CurrentBalance: 100, Currency: "CHF"},
		{UserID: &userID, CardNumber: "GC2", MerchantName: "M2", InitialBalance: 50, CurrentBalance: 50, Currency: "CHF"},
	}
	for i := range giftCards {
		db.Create(&giftCards[i])
		defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCards[i].ID)
	}

	found, err := repo.GetByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestGiftCardRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "UPDATE-TEST",
		MerchantName:   "Original",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard)
	defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCard.ID)

	giftCard.CurrentBalance = 75.5
	err := repo.Update(ctx, giftCard)
	assert.NoError(t, err)

	var found models.GiftCard
	db.First(&found, "id = ?", giftCard.ID)
	assert.Equal(t, 75.5, found.CurrentBalance)
}

func TestGiftCardRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "DELETE-ME",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard)

	err := repo.Delete(ctx, giftCard.ID)
	assert.NoError(t, err)
}

func TestGiftCardRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	initialCount, _ := repo.Count(ctx, userID)

	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "COUNT-TEST",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard)
	defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCard.ID)

	newCount, err := repo.Count(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, initialCount+1, newCount)
}

func TestGiftCardRepository_GetTotalBalance(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create multiple gift cards
	giftCards := []models.GiftCard{
		{UserID: &userID, CardNumber: "BAL1", MerchantName: "M1", InitialBalance: 100, CurrentBalance: 75, Currency: "CHF"},
		{UserID: &userID, CardNumber: "BAL2", MerchantName: "M2", InitialBalance: 50, CurrentBalance: 25, Currency: "CHF"},
	}
	for i := range giftCards {
		db.Create(&giftCards[i])
		defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCards[i].ID)
	}

	totalBalance, err := repo.GetTotalBalance(ctx, userID)
	assert.NoError(t, err)
	// GetTotalBalance sums all gift cards for this user
	assert.GreaterOrEqual(t, totalBalance, 100.0) // At least our 2 test cards (75+25=100)
}
