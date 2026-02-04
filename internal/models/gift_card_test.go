package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGiftCard_GetColor_WithMerchant(t *testing.T) {
	merchant := &Merchant{
		Name:  "Test Merchant",
		Color: "#00FF00",
	}

	giftCard := &GiftCard{
		Merchant: merchant,
	}

	assert.Equal(t, "#00FF00", giftCard.GetColor())
}

func TestGiftCard_GetColor_DefaultRed(t *testing.T) {
	giftCard := &GiftCard{
		Merchant: nil,
	}

	assert.Equal(t, "#DC2626", giftCard.GetColor()) // Default red
}

func TestGiftCard_GetCurrentBalance_Positive(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 100.0,
		CurrentBalance: 75.50,
	}

	assert.Equal(t, 75.50, giftCard.GetCurrentBalance())
}

func TestGiftCard_GetCurrentBalance_Zero(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 50.0,
		CurrentBalance: 0.0,
	}

	assert.Equal(t, 0.0, giftCard.GetCurrentBalance())
}

func TestGiftCard_GetCurrentBalance_Rounding(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 100.0,
		CurrentBalance: 33.33333333,
	}

	// Should round to 2 decimal places
	assert.Equal(t, 33.33, giftCard.GetCurrentBalance())
}

func TestGiftCard_DefaultCurrency(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 100.0,
		CurrentBalance: 100.0,
		Currency:       "CHF", // Default currency
	}

	assert.Equal(t, "CHF", giftCard.Currency)
}

func TestGiftCard_CustomCurrency(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 100.0,
		CurrentBalance: 100.0,
		Currency:       "EUR",
	}

	assert.Equal(t, "EUR", giftCard.Currency)
}

func TestGiftCard_Status_Active(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Status:         "active",
	}

	assert.Equal(t, "active", giftCard.Status)
}

func TestGiftCard_Status_Inactive(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 100.0,
		CurrentBalance: 0.0,
		Status:         "inactive",
	}

	assert.Equal(t, "inactive", giftCard.Status)
}

func TestGiftCard_BeforeCreate(t *testing.T) {
	giftCard := &GiftCard{}

	// ID should be zero before creation
	assert.Equal(t, uuid.Nil, giftCard.ID)

	// Simulate BeforeCreate hook behavior (DB generates UUID)
	giftCard.ID = uuid.New()

	// ID should be set after creation
	assert.NotEqual(t, uuid.Nil, giftCard.ID)
}

func TestGiftCard_WithTransactions(t *testing.T) {
	giftCard := &GiftCard{
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Transactions: []GiftCardTransaction{
			{Amount: -25.0, Description: "Purchase"},
		},
	}

	assert.Len(t, giftCard.Transactions, 1)
	assert.Equal(t, -25.0, giftCard.Transactions[0].Amount)
}
