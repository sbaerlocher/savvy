package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

	db.Exec("DELETE FROM cards WHERE id = ?", card.ID)
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
	defer db.Exec("DELETE FROM cards WHERE id = ?", card.ID)

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
		defer db.Exec("DELETE FROM cards WHERE id = ?", cards[i].ID)
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
	defer db.Exec("DELETE FROM cards WHERE id = ?", card.ID)

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
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "DELETE_ME",
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
	defer db.Exec("DELETE FROM cards WHERE id = ?", card.ID)

	newCount, err := repo.Count(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, initialCount+1, newCount)
}
