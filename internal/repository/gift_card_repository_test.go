package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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
	uniqueCardNumber := "DELETE-ME-" + uuid.New().String()[:8]
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     uniqueCardNumber,
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
	}

	totalBalance, err := repo.GetTotalBalance(ctx, userID)
	assert.NoError(t, err)
	// GetTotalBalance sums all gift cards for this user
	assert.GreaterOrEqual(t, totalBalance, 100.0) // At least our 2 test cards (75+25=100)
}

func TestGiftCardRepository_GetSharedWithUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	// Create owner and shared user
	ownerID := createTestUser(t, db)
	sharedUserID := createTestUser(t, db)

	// Create gift cards owned by owner
	giftCards := []models.GiftCard{
		{UserID: &ownerID, CardNumber: "SHARED_GC1", MerchantName: "Shared M1", InitialBalance: 100, CurrentBalance: 100, Currency: "CHF"},
		{UserID: &ownerID, CardNumber: "SHARED_GC2", MerchantName: "Shared M2", InitialBalance: 50, CurrentBalance: 50, Currency: "CHF"},
		{UserID: &ownerID, CardNumber: "NOT_SHARED_GC", MerchantName: "Not Shared", InitialBalance: 75, CurrentBalance: 75, Currency: "CHF"},
	}
	for i := range giftCards {
		db.Create(&giftCards[i])
	}

	// Share first two gift cards with sharedUserID
	shares := []models.GiftCardShare{
		{GiftCardID: giftCards[0].ID, SharedWithID: sharedUserID, CanEdit: true, CanDelete: false, CanEditTransactions: true},
		{GiftCardID: giftCards[1].ID, SharedWithID: sharedUserID, CanEdit: false, CanDelete: false, CanEditTransactions: false},
	}
	for i := range shares {
		db.Create(&shares[i])
	}

	// Get shared gift cards
	found, err := repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)

	// Verify the right gift cards are returned
	cardNumbers := make(map[string]bool)
	for _, gc := range found {
		cardNumbers[gc.CardNumber] = true
	}
	assert.True(t, cardNumbers["SHARED_GC1"])
	assert.True(t, cardNumbers["SHARED_GC2"])
	assert.False(t, cardNumbers["NOT_SHARED_GC"])
}

func TestGiftCardRepository_GetSharedWithUser_ExcludesDeleted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	ownerID := createTestUser(t, db)
	sharedUserID := createTestUser(t, db)

	// Create a gift card
	giftCard := &models.GiftCard{
		UserID:         &ownerID,
		CardNumber:     "DELETED_SHARE_GC",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard)

	// Create a share
	share := &models.GiftCardShare{
		GiftCardID:   giftCard.ID,
		SharedWithID: sharedUserID,
		CanEdit:      true,
	}
	db.Create(share)

	// Get shared gift cards (should include it)
	found, err := repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 1)

	// Soft delete the share
	db.Delete(share)

	// Get shared gift cards again (should exclude it now)
	found, err = repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)

	// Verify the deleted share is not included
	for _, gc := range found {
		assert.NotEqual(t, "DELETED_SHARE_GC", gc.CardNumber)
	}
}

func TestGiftCardRepository_CreateTransaction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create a gift card
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "GC-TRANSACTION",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	err := db.Create(giftCard).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, giftCard.ID)

	// Create a transaction
	transaction := &models.GiftCardTransaction{
		GiftCardID:      giftCard.ID,
		Amount:          -25.50,
		Description:     "Test purchase",
		TransactionDate: time.Now(),
	}

	err = repo.CreateTransaction(ctx, transaction)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, transaction.ID)

	// Verify it was created
	var found models.GiftCardTransaction
	err = db.First(&found, "id = ?", transaction.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, giftCard.ID, found.GiftCardID)
	assert.Equal(t, -25.50, found.Amount)

}

func TestGiftCardRepository_GetTransaction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create a gift card
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "GC-GET-TRANS",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard)

	// Create a transaction
	transaction := &models.GiftCardTransaction{
		GiftCardID:      giftCard.ID,
		Amount:          50.00,
		Description:     "Reload",
		TransactionDate: time.Now(),
	}
	db.Create(transaction)

	// Get the transaction
	found, err := repo.GetTransaction(ctx, transaction.ID, giftCard.ID)
	assert.NoError(t, err)
	assert.Equal(t, transaction.ID, found.ID)
	assert.Equal(t, giftCard.ID, found.GiftCardID)
	assert.Equal(t, 50.00, found.Amount)
}

func TestGiftCardRepository_GetTransaction_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	// Try to get non-existent transaction
	_, err := repo.GetTransaction(ctx, uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGiftCardRepository_GetTransaction_WrongGiftCard(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create two gift cards
	giftCard1 := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "GC-WRONG-1",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard1)

	giftCard2 := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "GC-WRONG-2",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard2)

	// Create a transaction for giftCard1
	transaction := &models.GiftCardTransaction{
		GiftCardID:      giftCard1.ID,
		Amount:          -10.00,
		Description:     "Test",
		TransactionDate: time.Now(),
	}
	db.Create(transaction)

	// Try to get transaction with wrong gift card ID
	_, err := repo.GetTransaction(ctx, transaction.ID, giftCard2.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGiftCardRepository_DeleteTransaction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create a gift card
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "GC-DEL-TRANS",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard)

	// Create a transaction
	transaction := &models.GiftCardTransaction{
		GiftCardID:      giftCard.ID,
		Amount:          -20.00,
		Description:     "To be deleted",
		TransactionDate: time.Now(),
	}
	db.Create(transaction)

	// Delete the transaction
	err := repo.DeleteTransaction(ctx, transaction.ID)
	assert.NoError(t, err)

	// Verify it's soft deleted
	var found models.GiftCardTransaction
	err = db.First(&found, "id = ?", transaction.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Verify it exists with Unscoped
	err = db.Unscoped().First(&found, "id = ?", transaction.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestGiftCardRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGiftCardRepository_GetByID_WithPreloads(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGiftCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create a merchant
	merchant := &models.Merchant{
		Name:  "Test GiftCard Merchant",
		Color: "#0000FF",
	}
	db.Create(merchant)

	// Create a gift card with merchant
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "PRELOAD_GC",
		MerchantName:   "Test",
		MerchantID:     &merchant.ID,
		InitialBalance: 100,
		CurrentBalance: 100,
		Currency:       "CHF",
	}
	db.Create(giftCard)

	// Create a transaction
	transaction := &models.GiftCardTransaction{
		GiftCardID:      giftCard.ID,
		Amount:          -10.00,
		Description:     "Test transaction",
		TransactionDate: time.Now(),
	}
	db.Create(transaction)

	// Get with preloads
	found, err := repo.GetByID(ctx, giftCard.ID, "Merchant", "User", "Transactions")
	assert.NoError(t, err)
	assert.NotNil(t, found.Merchant)
	assert.Equal(t, merchant.ID, found.Merchant.ID)
	assert.NotNil(t, found.User)
	assert.Equal(t, userID, found.User.ID)
	assert.GreaterOrEqual(t, len(found.Transactions), 1)
}
