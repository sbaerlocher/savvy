package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVoucherShare_Creation(t *testing.T) {
	voucherID := uuid.New()
	sharedWithID := uuid.New()

	share := &VoucherShare{
		VoucherID:    voucherID,
		SharedWithID: sharedWithID,
		CanEdit:      false,
		CanDelete:    false,
	}

	assert.Equal(t, voucherID, share.VoucherID)
	assert.Equal(t, sharedWithID, share.SharedWithID)
	assert.False(t, share.CanEdit)
	assert.False(t, share.CanDelete)
}

func TestVoucherShare_WithAssociations(t *testing.T) {
	voucher := &Voucher{
		ID:   uuid.New(),
		Code: "VOUCHER123",
	}

	user := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	share := &VoucherShare{
		VoucherID:      voucher.ID,
		Voucher:        voucher,
		SharedWithID:   user.ID,
		SharedWithUser: user,
		CanEdit:        false,
		CanDelete:      false,
	}

	assert.Equal(t, voucher.ID, share.VoucherID)
	assert.Equal(t, user.ID, share.SharedWithID)
	assert.NotNil(t, share.Voucher)
	assert.NotNil(t, share.SharedWithUser)
	assert.Equal(t, "VOUCHER123", share.Voucher.Code)
	assert.Equal(t, "test@example.com", share.SharedWithUser.Email)
}

func TestVoucherShare_DefaultPermissions(t *testing.T) {
	share := &VoucherShare{
		VoucherID:    uuid.New(),
		SharedWithID: uuid.New(),
	}

	// Vouchers are read-only, permissions default to false
	assert.False(t, share.CanEdit)
	assert.False(t, share.CanDelete)
}

func TestVoucherShare_ReadOnly(t *testing.T) {
	// Vouchers are always read-only (no edit/delete rights)
	share := &VoucherShare{
		VoucherID:    uuid.New(),
		SharedWithID: uuid.New(),
		CanEdit:      false,
		CanDelete:    false,
	}

	assert.False(t, share.CanEdit)
	assert.False(t, share.CanDelete)
}
