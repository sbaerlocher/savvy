package api

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type duplicateInfo struct {
	ID           uuid.UUID
	MerchantName string
	Number       string
}

type duplicateLookup func() (*duplicateInfo, error)

func adaptDuplicateLookup[T any](lookup func() (*T, error), info func(*T) duplicateInfo) duplicateLookup {
	return func() (*duplicateInfo, error) {
		item, err := lookup()
		if item == nil {
			return nil, err
		}
		duplicate := info(item)
		return &duplicate, err
	}
}

type duplicateGuard struct {
	ctx            context.Context
	active         duplicateLookup
	deleted        duplicateLookup
	resource       string
	activeMessage  string
	deletedMessage string
}

// check checks owned active and deleted records in that order. Lookup errors
// are logged and ignored, leaving the database constraint as a fallback.
func (g duplicateGuard) check() *DuplicateErrorResponse {
	duplicate, err := g.active()
	if err != nil {
		slog.ErrorContext(g.ctx, "failed to check duplicate", "error", err)
	}
	if duplicate != nil {
		slog.InfoContext(g.ctx, "duplicate "+g.resource+" blocked", "existing_id", duplicate.ID)
		return duplicateConflictResponse(duplicate, g.activeMessage, false)
	}

	if g.deleted != nil {
		duplicate, err := g.deleted()
		if err != nil {
			slog.ErrorContext(g.ctx, "failed to check deleted duplicate", "error", err)
		}
		if duplicate != nil {
			return duplicateConflictResponse(duplicate, g.deletedMessage, true)
		}
	}

	return nil
}

// race re-fetches an active duplicate after the database rejected an insert.
// The caller already logs a failed insert, so this lookup does not log again.
func (g duplicateGuard) race() *DuplicateErrorResponse {
	duplicate, _ := g.active()
	if duplicate == nil {
		return nil
	}
	return duplicateConflictResponse(duplicate, g.activeMessage, false)
}

func duplicateConflictResponse(duplicate *duplicateInfo, message string, deleted bool) *DuplicateErrorResponse {
	return &DuplicateErrorResponse{
		Error:   "duplicate_barcode",
		Message: message,
		Duplicate: &DuplicateWarning{
			HasDuplicate:   true,
			MerchantName:   duplicate.MerchantName,
			ResourceNumber: duplicate.Number,
			ExistingID:     duplicate.ID.String(),
			Deleted:        deleted,
		},
	}
}
