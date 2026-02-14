package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGiftCardTransaction_Creation(t *testing.T) {
	giftCardID := uuid.New()
	createdByUserID := uuid.New()

	transaction := &GiftCardTransaction{
		GiftCardID:      giftCardID,
		Amount:          -25.50,
		Description:     "Coffee shop purchase",
		CreatedByUserID: &createdByUserID,
	}

	assert.Equal(t, giftCardID, transaction.GiftCardID)
	assert.Equal(t, -25.50, transaction.Amount)
	assert.Equal(t, "Coffee shop purchase", transaction.Description)
	assert.NotNil(t, transaction.CreatedByUserID)
	assert.Equal(t, createdByUserID, *transaction.CreatedByUserID)
}

func TestGiftCardTransaction_PositiveAmount(t *testing.T) {
	transaction := &GiftCardTransaction{
		GiftCardID:  uuid.New(),
		Amount:      50.00,
		Description: "Reload",
	}

	assert.Equal(t, 50.00, transaction.Amount)
}

func TestGiftCardTransaction_NegativeAmount(t *testing.T) {
	transaction := &GiftCardTransaction{
		GiftCardID:  uuid.New(),
		Amount:      -30.00,
		Description: "Purchase",
	}

	assert.Equal(t, -30.00, transaction.Amount)
}

func TestGiftCardTransaction_WithAssociations(t *testing.T) {
	giftCard := &GiftCard{
		ID:             uuid.New(),
		CardNumber:     "1234567890",
		InitialBalance: 100.0,
	}

	user := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	transaction := &GiftCardTransaction{
		GiftCardID:      giftCard.ID,
		GiftCard:        giftCard,
		Amount:          -15.00,
		Description:     "Grocery purchase",
		CreatedByUserID: &user.ID,
		CreatedByUser:   user,
	}

	assert.Equal(t, giftCard.ID, transaction.GiftCardID)
	assert.Equal(t, user.ID, *transaction.CreatedByUserID)
	assert.NotNil(t, transaction.GiftCard)
	assert.NotNil(t, transaction.CreatedByUser)
	assert.Equal(t, "1234567890", transaction.GiftCard.CardNumber)
	assert.Equal(t, "test@example.com", transaction.CreatedByUser.Email)
}

func TestGiftCardTransaction_WithTransactionDate(t *testing.T) {
	transactionDate := time.Now()

	transaction := &GiftCardTransaction{
		GiftCardID:      uuid.New(),
		Amount:          -10.00,
		Description:     "Test purchase",
		TransactionDate: transactionDate,
	}

	assert.Equal(t, transactionDate, transaction.TransactionDate)
}

func TestGiftCardTransaction_NoUserID(t *testing.T) {
	transaction := &GiftCardTransaction{
		GiftCardID:      uuid.New(),
		Amount:          -5.00,
		Description:     "Anonymous transaction",
		CreatedByUserID: nil,
	}

	assert.Nil(t, transaction.CreatedByUserID)
}

func TestGiftCardTransaction_EmptyDescription(t *testing.T) {
	transaction := &GiftCardTransaction{
		GiftCardID:  uuid.New(),
		Amount:      -20.00,
		Description: "",
	}

	assert.Equal(t, "", transaction.Description)
}

func TestGiftCardTransaction_ZeroAmount(t *testing.T) {
	transaction := &GiftCardTransaction{
		GiftCardID:  uuid.New(),
		Amount:      0.00,
		Description: "Zero amount transaction",
	}

	assert.Equal(t, 0.00, transaction.Amount)
}
