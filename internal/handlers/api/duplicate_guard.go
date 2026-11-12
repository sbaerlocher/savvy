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

// checkDuplicateConflict checks owned active and deleted records in that order.
// Lookup errors are logged and ignored, leaving the database constraint as a fallback.
func checkDuplicateConflict(ctx context.Context, active, deleted duplicateLookup, resource, activeMessage, deletedMessage string) *DuplicateErrorResponse {
	duplicate, err := active()
	if err != nil {
		slog.ErrorContext(ctx, "failed to check duplicate", "error", err)
	}
	if duplicate != nil {
		slog.InfoContext(ctx, "duplicate "+resource+" blocked", "existing_id", duplicate.ID)
		return duplicateConflictResponse(duplicate, activeMessage, false)
	}

	if deleted != nil {
		duplicate, err := deleted()
		if err != nil {
			slog.ErrorContext(ctx, "failed to check deleted duplicate", "error", err)
		}
		if duplicate != nil {
			return duplicateConflictResponse(duplicate, deletedMessage, true)
		}
	}

	return nil
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
