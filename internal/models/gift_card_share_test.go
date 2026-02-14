package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGiftCardShare_Creation(t *testing.T) {
	giftCardID := uuid.New()
	sharedWithID := uuid.New()

	share := &GiftCardShare{
		GiftCardID:          giftCardID,
		SharedWithID:        sharedWithID,
		CanEdit:             true,
		CanDelete:           false,
		CanEditTransactions: true,
	}

	assert.Equal(t, giftCardID, share.GiftCardID)
	assert.Equal(t, sharedWithID, share.SharedWithID)
	assert.True(t, share.CanEdit)
	assert.False(t, share.CanDelete)
	assert.True(t, share.CanEditTransactions)
}

func TestGiftCardShare_WithAssociations(t *testing.T) {
	giftCard := &GiftCard{
		ID:             uuid.New(),
		CardNumber:     "1234567890",
		InitialBalance: 100.0,
	}

	user := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	share := &GiftCardShare{
		GiftCardID:          giftCard.ID,
		GiftCard:            giftCard,
		SharedWithID:        user.ID,
		SharedWithUser:      user,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
	}

	assert.Equal(t, giftCard.ID, share.GiftCardID)
	assert.Equal(t, user.ID, share.SharedWithID)
	assert.NotNil(t, share.GiftCard)
	assert.NotNil(t, share.SharedWithUser)
	assert.Equal(t, "1234567890", share.GiftCard.CardNumber)
	assert.Equal(t, "test@example.com", share.SharedWithUser.Email)
}

func TestGiftCardShare_DefaultPermissions(t *testing.T) {
	share := &GiftCardShare{
		GiftCardID:   uuid.New(),
		SharedWithID: uuid.New(),
	}

	// Default values should be false (as per GORM tags)
	assert.False(t, share.CanEdit)
	assert.False(t, share.CanDelete)
	assert.False(t, share.CanEditTransactions)
}

func TestGiftCardShare_AllPermissionsGranted(t *testing.T) {
	share := &GiftCardShare{
		GiftCardID:          uuid.New(),
		SharedWithID:        uuid.New(),
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
	}

	assert.True(t, share.CanEdit)
	assert.True(t, share.CanDelete)
	assert.True(t, share.CanEditTransactions)
}

func TestGiftCardShare_PartialPermissions(t *testing.T) {
	share := &GiftCardShare{
		GiftCardID:          uuid.New(),
		SharedWithID:        uuid.New(),
		CanEdit:             true,
		CanDelete:           false,
		CanEditTransactions: true,
	}

	assert.True(t, share.CanEdit)
	assert.False(t, share.CanDelete)
	assert.True(t, share.CanEditTransactions)
}

func TestGiftCardShare_TransactionsOnlyPermission(t *testing.T) {
	share := &GiftCardShare{
		GiftCardID:          uuid.New(),
		SharedWithID:        uuid.New(),
		CanEdit:             false,
		CanDelete:           false,
		CanEditTransactions: true,
	}

	assert.False(t, share.CanEdit)
	assert.False(t, share.CanDelete)
	assert.True(t, share.CanEditTransactions)
}
