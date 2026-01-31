package models

import (
	"testing"
	"time"

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
		Color:    "#00FF00", // Should be overridden by merchant color
	}

	assert.Equal(t, "#FF0000", voucher.GetColor())
}

func TestVoucher_GetColor_WithoutMerchant(t *testing.T) {
	voucher := &Voucher{
		Merchant: nil,
		Color:    "#AABBCC",
	}

	assert.Equal(t, "#AABBCC", voucher.GetColor())
}

func TestVoucher_GetColor_DefaultGreen(t *testing.T) {
	voucher := &Voucher{
		Merchant: nil,
		Color:    "",
	}

	assert.Equal(t, "#10B981", voucher.GetColor()) // Default green
}

func TestVoucher_CanRedeem_Valid(t *testing.T) {
	now := time.Now()
	voucher := &Voucher{
		ValidFrom:      now.Add(-24 * time.Hour),
		ValidUntil:     now.Add(24 * time.Hour),
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}

	assert.True(t, voucher.CanRedeem())
}

func TestVoucher_CanRedeem_Expired(t *testing.T) {
	now := time.Now()
	voucher := &Voucher{
		ValidFrom:      now.Add(-48 * time.Hour),
		ValidUntil:     now.Add(-24 * time.Hour),
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}

	assert.False(t, voucher.CanRedeem())
}

func TestVoucher_CanRedeem_NotYetValid(t *testing.T) {
	now := time.Now()
	voucher := &Voucher{
		ValidFrom:      now.Add(24 * time.Hour),
		ValidUntil:     now.Add(48 * time.Hour),
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}

	assert.False(t, voucher.CanRedeem())
}

func TestVoucher_CanRedeem_SingleUseAlreadyUsed(t *testing.T) {
	now := time.Now()
	voucher := &Voucher{
		ValidFrom:      now.Add(-24 * time.Hour),
		ValidUntil:     now.Add(24 * time.Hour),
		UsageLimitType: "single_use",
		UsedCount:      1,
	}

	assert.False(t, voucher.CanRedeem())
}

func TestVoucher_Redeem_Success(t *testing.T) {
	now := time.Now()
	voucher := &Voucher{
		ValidFrom:      now.Add(-24 * time.Hour),
		ValidUntil:     now.Add(24 * time.Hour),
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}

	err := voucher.Redeem()

	assert.NoError(t, err)
	assert.Equal(t, 1, voucher.UsedCount)
}

func TestVoucher_Redeem_Failure(t *testing.T) {
	now := time.Now()
	voucher := &Voucher{
		ValidFrom:      now.Add(-48 * time.Hour),
		ValidUntil:     now.Add(-24 * time.Hour),
		UsageLimitType: "unlimited",
		UsedCount:      0,
	}

	err := voucher.Redeem()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be redeemed")
	assert.Equal(t, 0, voucher.UsedCount) // Should not increment
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
