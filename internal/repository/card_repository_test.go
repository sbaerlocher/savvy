package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"savvy/internal/models"
)

func TestCardRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "TEST123456",
		MerchantName: "Test Card Merchant",
	}

	err := repo.Create(ctx, card)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, card.ID)

}

func TestCardRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "TEST789",
		MerchantName: "Test Merchant",
	}
	db.Create(card)

	found, err := repo.GetByID(ctx, card.ID)
	assert.NoError(t, err)
	assert.Equal(t, card.ID, found.ID)
	assert.Equal(t, "TEST789", found.CardNumber)
}

func TestCardRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	cards := []models.Card{
		{UserID: &userID, CardNumber: "TEST1", MerchantName: "M1"},
		{UserID: &userID, CardNumber: "TEST2", MerchantName: "M2"},
	}
	for i := range cards {
		db.Create(&cards[i])
	}

	found, err := repo.GetByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestCardRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "ORIGINAL",
		MerchantName: "Original",
	}
	db.Create(card)

	card.CardNumber = "UPDATED"
	err := repo.Update(ctx, card)
	assert.NoError(t, err)

	var found models.Card
	db.First(&found, "id = ?", card.ID)
	assert.Equal(t, "UPDATED", found.CardNumber)
}

func TestCardRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	uniqueCardNumber := "DELETE_ME_" + uuid.New().String()[:8]
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   uniqueCardNumber,
		MerchantName: "Test",
	}
	db.Create(card)

	err := repo.Delete(ctx, card.ID)
	assert.NoError(t, err)
}

func TestCardRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)
	initialCount, _ := repo.Count(ctx, userID)

	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "COUNT_TEST",
		MerchantName: "Test",
	}
	db.Create(card)

	newCount, err := repo.Count(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, initialCount+1, newCount)
}

func TestCardRepository_GetSharedWithUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	// Create owner and shared user
	ownerID := createTestUser(t, db)
	sharedUserID := createTestUser(t, db)

	// Create cards owned by owner
	cards := []models.Card{
		{UserID: &ownerID, CardNumber: "SHARED_CARD_1", MerchantName: "Shared M1"},
		{UserID: &ownerID, CardNumber: "SHARED_CARD_2", MerchantName: "Shared M2"},
		{UserID: &ownerID, CardNumber: "NOT_SHARED", MerchantName: "Not Shared"},
	}
	for i := range cards {
		db.Create(&cards[i])
	}

	// Share first two cards with sharedUserID
	shares := []models.CardShare{
		{CardID: cards[0].ID, SharedWithID: sharedUserID, CanEdit: true},
		{CardID: cards[1].ID, SharedWithID: sharedUserID, CanEdit: false},
	}
	for i := range shares {
		db.Create(&shares[i])
	}

	// Get shared cards
	found, err := repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)

	// Verify the right cards are returned
	cardNumbers := make(map[string]bool)
	for _, card := range found {
		cardNumbers[card.CardNumber] = true
	}
	assert.True(t, cardNumbers["SHARED_CARD_1"])
	assert.True(t, cardNumbers["SHARED_CARD_2"])
	assert.False(t, cardNumbers["NOT_SHARED"])
}

func TestCardRepository_GetSharedWithUser_ExcludesDeleted(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	ownerID := createTestUser(t, db)
	sharedUserID := createTestUser(t, db)

	// Create a card
	card := &models.Card{
		UserID:       &ownerID,
		CardNumber:   "DELETED_SHARE_CARD",
		MerchantName: "Test",
	}
	db.Create(card)

	// Create a share
	share := &models.CardShare{
		CardID:       card.ID,
		SharedWithID: sharedUserID,
		CanEdit:      true,
	}
	db.Create(share)

	// Get shared cards (should include it)
	found, err := repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 1)

	// Soft delete the share
	db.Delete(share)

	// Get shared cards again (should exclude it now)
	found, err = repo.GetSharedWithUser(ctx, sharedUserID)
	assert.NoError(t, err)

	// Verify the deleted share is not included
	for _, c := range found {
		assert.NotEqual(t, "DELETED_SHARE_CARD", c.CardNumber)
	}
}

func TestCardRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestCardRepository_GetByID_WithPreloads(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Create a merchant
	merchant := &models.Merchant{
		Name:  "Test Merchant Preload",
		Color: "#FF0000",
	}
	db.Create(merchant)

	// Create a card with merchant
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "PRELOAD_TEST",
		MerchantName: "Test",
		MerchantID:   &merchant.ID,
	}
	err := db.Create(card).Error
	require.NoError(t, err, "Failed to create test card")

	// Get with preloads
	found, err := repo.GetByID(ctx, card.ID, "Merchant", "User")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.NotNil(t, found.Merchant)
	assert.Equal(t, merchant.ID, found.Merchant.ID)
	assert.NotNil(t, found.User)
	assert.Equal(t, userID, found.User.ID)
}

func TestCardRepository_FindDeletedByCardNumber(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	card := &models.Card{UserID: &userID, Program: "P", CardNumber: "DEL-1"}
	require.NoError(t, repo.Create(ctx, card))
	require.NoError(t, repo.Delete(ctx, card.ID)) // soft-delete

	// Active lookup does not see it
	active, err := repo.FindByCardNumber(ctx, "DEL-1", userID)
	require.NoError(t, err)
	require.Nil(t, active)

	// Deleted lookup finds it
	found, err := repo.FindDeletedByCardNumber(ctx, "DEL-1", userID)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, card.ID, found.ID)

	// Restore brings it back
	require.NoError(t, repo.RestoreByID(ctx, card.ID, userID))
	active2, err := repo.FindByCardNumber(ctx, "DEL-1", userID)
	require.NoError(t, err)
	require.NotNil(t, active2)
}

// TestCardRepository_ReinsertAfterSoftDelete is the core regression for this
// feature: the partial unique index must exclude soft-deleted rows so a user can
// create a new card whose number matches one they previously soft-deleted.
func TestCardRepository_ReinsertAfterSoftDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	first := &models.Card{UserID: &userID, Program: "P", CardNumber: "REUSE-1"}
	require.NoError(t, repo.Create(ctx, first))
	require.NoError(t, repo.Delete(ctx, first.ID)) // soft-delete

	// Re-insert with the same number must succeed (index excludes the deleted row).
	second := &models.Card{UserID: &userID, Program: "P", CardNumber: "REUSE-1"}
	require.NoError(t, repo.Create(ctx, second))
	require.NotEqual(t, first.ID, second.ID)
}

// TestCardRepository_SameNumberDifferentUsers verifies the composite index is
// per-user: two different users may hold the same card number simultaneously.
func TestCardRepository_SameNumberDifferentUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCardRepository(db)
	ctx := context.Background()
	userA := createTestUser(t, db)
	userB := createTestUser(t, db)

	cardA := &models.Card{UserID: &userA, Program: "P", CardNumber: "SHARED-1"}
	require.NoError(t, repo.Create(ctx, cardA))

	cardB := &models.Card{UserID: &userB, Program: "P", CardNumber: "SHARED-1"}
	require.NoError(t, repo.Create(ctx, cardB))
}
