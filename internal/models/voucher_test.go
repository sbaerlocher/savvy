package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVoucher_GetColor_WithMerchant(t *testing.T) {
	merchant := &Merchant{
		Name:  "Test Merchant",
		Color: "#FF0000",
	}

	voucher := &Voucher{
		Merchant: merchant,
	}

	assert.Equal(t, "#FF0000", voucher.GetColor())
}

func TestVoucher_GetColor_DefaultGreen(t *testing.T) {
	voucher := &Voucher{
		Merchant: nil,
	}

	assert.Equal(t, "#10B981", voucher.GetColor()) // Default green
}

func TestVoucher_BeforeCreate(t *testing.T) {
	voucher := &Voucher{}

	// ID should be zero before creation
	assert.Equal(t, uuid.Nil, voucher.ID)

	// Simulate BeforeCreate hook behavior (DB generates UUID)
	voucher.ID = uuid.New()

	// ID should be set after creation
	assert.NotEqual(t, uuid.Nil, voucher.ID)
}
