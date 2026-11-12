package api

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDuplicateConflict_ActiveTakesPrecedence(t *testing.T) {
	id := uuid.New()
	deletedCalled := false
	response := checkDuplicateConflict(
		context.Background(),
		func() (*duplicateInfo, error) {
			return &duplicateInfo{ID: id, MerchantName: "Merchant", Number: "123"}, nil
		},
		func() (*duplicateInfo, error) {
			deletedCalled = true
			return &duplicateInfo{ID: uuid.New(), Number: "123"}, nil
		},
		"card", "Active card", "Deleted card",
	)

	require.NotNil(t, response)
	assert.Equal(t, "duplicate_barcode", response.Error)
	assert.Equal(t, "Active card", response.Message)
	require.NotNil(t, response.Duplicate)
	assert.Equal(t, id.String(), response.Duplicate.ExistingID)
	assert.Equal(t, "Merchant", response.Duplicate.MerchantName)
	assert.Equal(t, "123", response.Duplicate.ResourceNumber)
	assert.True(t, response.Duplicate.HasDuplicate)
	assert.False(t, response.Duplicate.Deleted)
	assert.False(t, deletedCalled)
}

func TestCheckDuplicateConflict_DeletedDuplicate(t *testing.T) {
	id := uuid.New()
	response := checkDuplicateConflict(
		context.Background(),
		func() (*duplicateInfo, error) { return nil, nil },
		func() (*duplicateInfo, error) {
			return &duplicateInfo{ID: id, MerchantName: "Merchant", Number: "123"}, nil
		},
		"voucher", "Active voucher", "Deleted voucher",
	)

	require.NotNil(t, response)
	assert.Equal(t, "duplicate_barcode", response.Error)
	assert.Equal(t, "Deleted voucher", response.Message)
	require.NotNil(t, response.Duplicate)
	assert.Equal(t, id.String(), response.Duplicate.ExistingID)
	assert.Equal(t, "Merchant", response.Duplicate.MerchantName)
	assert.Equal(t, "123", response.Duplicate.ResourceNumber)
	assert.True(t, response.Duplicate.HasDuplicate)
	assert.True(t, response.Duplicate.Deleted)
}

func TestCheckDuplicateConflict_LookupErrorsAreIgnored(t *testing.T) {
	deletedCalled := false
	response := checkDuplicateConflict(
		context.Background(),
		func() (*duplicateInfo, error) { return nil, errors.New("active lookup failed") },
		func() (*duplicateInfo, error) {
			deletedCalled = true
			return nil, errors.New("deleted lookup failed")
		},
		"gift card", "Active gift card", "Deleted gift card",
	)

	assert.Nil(t, response)
	assert.True(t, deletedCalled)
}

func TestCheckDuplicateConflict_DuplicateWinsOverLookupError(t *testing.T) {
	for _, tc := range []struct {
		name    string
		deleted bool
	}{
		{name: "active"},
		{name: "deleted", deleted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.New()
			active := duplicateLookup(func() (*duplicateInfo, error) {
				if tc.deleted {
					return nil, nil
				}
				return &duplicateInfo{ID: id, Number: "123"}, errors.New("active lookup failed")
			})
			deleted := duplicateLookup(func() (*duplicateInfo, error) {
				return &duplicateInfo{ID: id, Number: "123"}, errors.New("deleted lookup failed")
			})

			response := checkDuplicateConflict(context.Background(), active, deleted, "card", "Active card", "Deleted card")
			require.NotNil(t, response)
			require.NotNil(t, response.Duplicate)
			assert.Equal(t, id.String(), response.Duplicate.ExistingID)
			assert.Equal(t, tc.deleted, response.Duplicate.Deleted)
		})
	}
}

func TestCheckDuplicateConflict_ActiveOnlyForRace(t *testing.T) {
	id := uuid.New()
	response := checkDuplicateConflict(
		context.Background(),
		func() (*duplicateInfo, error) {
			return &duplicateInfo{ID: id, Number: "123"}, nil
		},
		nil,
		"card", "Active card", "Deleted card",
	)

	require.NotNil(t, response)
	require.NotNil(t, response.Duplicate)
	assert.Equal(t, id.String(), response.Duplicate.ExistingID)
	assert.False(t, response.Duplicate.Deleted)
}
