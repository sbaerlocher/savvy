package api

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuplicateGuard_ActiveTakesPrecedence(t *testing.T) {
	id := uuid.New()
	deletedCalled := false
	response := (duplicateGuard{
		ctx: context.Background(),
		active: func() (*duplicateInfo, error) {
			return &duplicateInfo{ID: id, MerchantName: "Merchant", Number: "123"}, nil
		},
		deleted: func() (*duplicateInfo, error) {
			deletedCalled = true
			return &duplicateInfo{ID: uuid.New(), Number: "123"}, nil
		},
		resource: "card", activeMessage: "Active card", deletedMessage: "Deleted card",
	}).check()

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

func TestDuplicateGuard_DeletedDuplicate(t *testing.T) {
	id := uuid.New()
	response := (duplicateGuard{
		ctx:    context.Background(),
		active: func() (*duplicateInfo, error) { return nil, nil },
		deleted: func() (*duplicateInfo, error) {
			return &duplicateInfo{ID: id, MerchantName: "Merchant", Number: "123"}, nil
		},
		resource: "voucher", activeMessage: "Active voucher", deletedMessage: "Deleted voucher",
	}).check()

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

func TestDuplicateGuard_LookupErrorsAreIgnored(t *testing.T) {
	deletedCalled := false
	response := (duplicateGuard{
		ctx:    context.Background(),
		active: func() (*duplicateInfo, error) { return nil, errors.New("active lookup failed") },
		deleted: func() (*duplicateInfo, error) {
			deletedCalled = true
			return nil, errors.New("deleted lookup failed")
		},
		resource: "gift card", activeMessage: "Active gift card", deletedMessage: "Deleted gift card",
	}).check()

	assert.Nil(t, response)
	assert.True(t, deletedCalled)
}

func TestDuplicateGuard_DuplicateWinsOverLookupError(t *testing.T) {
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

			response := (duplicateGuard{
				ctx: context.Background(), active: active, deleted: deleted,
				resource: "card", activeMessage: "Active card", deletedMessage: "Deleted card",
			}).check()
			require.NotNil(t, response)
			require.NotNil(t, response.Duplicate)
			assert.Equal(t, id.String(), response.Duplicate.ExistingID)
			assert.Equal(t, tc.deleted, response.Duplicate.Deleted)
		})
	}
}

func TestDuplicateGuard_RaceDoesNotCallDeletedLookup(t *testing.T) {
	id := uuid.New()
	deletedCalled := false
	response := (duplicateGuard{
		ctx: context.Background(),
		active: func() (*duplicateInfo, error) {
			return &duplicateInfo{ID: id, Number: "123"}, nil
		},
		deleted: func() (*duplicateInfo, error) {
			deletedCalled = true
			return nil, nil
		},
		resource: "card", activeMessage: "Active card", deletedMessage: "Deleted card",
	}).race()

	require.NotNil(t, response)
	require.NotNil(t, response.Duplicate)
	assert.Equal(t, id.String(), response.Duplicate.ExistingID)
	assert.False(t, response.Duplicate.Deleted)
	assert.False(t, deletedCalled)
}

func TestDuplicateGuard_RaceDoesNotLog(t *testing.T) {
	previousLogger := slog.Default()
	var logs bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	for _, tc := range []struct {
		name string
		info *duplicateInfo
		err  error
	}{
		{name: "lookup error", err: errors.New("lookup failed")},
		{name: "duplicate found", info: &duplicateInfo{ID: uuid.New(), Number: "123"}},
		{name: "duplicate and error", info: &duplicateInfo{ID: uuid.New(), Number: "123"}, err: errors.New("lookup failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs.Reset()
			response := (duplicateGuard{
				ctx:      context.Background(),
				active:   func() (*duplicateInfo, error) { return tc.info, tc.err },
				resource: "card", activeMessage: "Active card",
			}).race()
			if tc.info == nil {
				assert.Nil(t, response)
			} else {
				require.NotNil(t, response)
				assert.Equal(t, tc.info.ID.String(), response.Duplicate.ExistingID)
			}
			assert.Empty(t, logs.String())
		})
	}
}
