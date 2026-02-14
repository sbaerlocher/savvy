package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCardShare_Creation(t *testing.T) {
	cardID := uuid.New()
	sharedWithID := uuid.New()

	share := &CardShare{
		CardID:       cardID,
		SharedWithID: sharedWithID,
		CanEdit:      true,
		CanDelete:    false,
	}

	assert.Equal(t, cardID, share.CardID)
	assert.Equal(t, sharedWithID, share.SharedWithID)
	assert.True(t, share.CanEdit)
	assert.False(t, share.CanDelete)
}

func TestCardShare_WithAssociations(t *testing.T) {
	card := &Card{
		ID:      uuid.New(),
		Program: "Test Card",
	}

	user := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	share := &CardShare{
		CardID:         card.ID,
		Card:           card,
		SharedWithID:   user.ID,
		SharedWithUser: user,
		CanEdit:        true,
		CanDelete:      true,
	}

	assert.Equal(t, card.ID, share.CardID)
	assert.Equal(t, user.ID, share.SharedWithID)
	assert.NotNil(t, share.Card)
	assert.NotNil(t, share.SharedWithUser)
	assert.Equal(t, "Test Card", share.Card.Program)
	assert.Equal(t, "test@example.com", share.SharedWithUser.Email)
}

func TestCardShare_DefaultPermissions(t *testing.T) {
	share := &CardShare{
		CardID:       uuid.New(),
		SharedWithID: uuid.New(),
	}

	// Default values should be false (as per GORM tags)
	assert.False(t, share.CanEdit)
	assert.False(t, share.CanDelete)
}

func TestCardShare_AllPermissionsGranted(t *testing.T) {
	share := &CardShare{
		CardID:       uuid.New(),
		SharedWithID: uuid.New(),
		CanEdit:      true,
		CanDelete:    true,
	}

	assert.True(t, share.CanEdit)
	assert.True(t, share.CanDelete)
}
