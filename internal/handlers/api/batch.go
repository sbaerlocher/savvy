// Package api contains JSON API handlers for batch operations.
package api //nolint:revive // "api" is a meaningful package name for API handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/services"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// maxBatchSize is the maximum number of items allowed in a single batch request.
const maxBatchSize = 50

// sanitizeBatchError returns the error message if it's a known business error,
// or a generic message if it looks like an internal/database error.
func sanitizeBatchError(err error, fallback string) string {
	msg := err.Error()
	// Hide raw database errors (SQLSTATE, pq:, etc.)
	if strings.Contains(msg, "SQLSTATE") || strings.Contains(msg, "pq:") || strings.Contains(msg, "sql:") {
		return fallback
	}
	return msg
}

// BatchHandler handles batch API endpoints for cards, vouchers, and gift cards.
type BatchHandler struct {
	cardService     services.CardServiceInterface
	voucherService  services.VoucherServiceInterface
	giftCardService services.GiftCardServiceInterface
	authzService    services.AuthzServiceInterface
	shareService    services.ShareServiceInterface
	transferService services.TransferServiceInterface
	userService     services.UserServiceInterface
	exportService   services.ExportServiceInterface
}

// NewBatchHandler creates a new batch API handler.
func NewBatchHandler(
	cardService services.CardServiceInterface,
	voucherService services.VoucherServiceInterface,
	giftCardService services.GiftCardServiceInterface,
	authzService services.AuthzServiceInterface,
	shareService services.ShareServiceInterface,
	transferService services.TransferServiceInterface,
	userService services.UserServiceInterface,
	exportService services.ExportServiceInterface,
) *BatchHandler {
	return &BatchHandler{
		cardService:     cardService,
		voucherService:  voucherService,
		giftCardService: giftCardService,
		authzService:    authzService,
		shareService:    shareService,
		transferService: transferService,
		userService:     userService,
		exportService:   exportService,
	}
}

// ==================== Batch Delete ====================

// DeleteCards batch-deletes cards (all-or-nothing: all checks must pass before any deletion).
// POST /api/v1/cards/batch/delete
func (h *BatchHandler) DeleteCards(c echo.Context) error {
	return h.batchDelete(c, "card", h.authzService.CheckCardAccess, func(ctx context.Context, id uuid.UUID) error {
		return h.cardService.DeleteCard(ctx, id)
	})
}

// DeleteVouchers batch-deletes vouchers.
// POST /api/v1/vouchers/batch/delete
func (h *BatchHandler) DeleteVouchers(c echo.Context) error {
	return h.batchDelete(c, "voucher", h.authzService.CheckVoucherAccess, func(ctx context.Context, id uuid.UUID) error {
		return h.voucherService.DeleteVoucher(ctx, id)
	})
}

// DeleteGiftCards batch-deletes gift cards.
// POST /api/v1/gift-cards/batch/delete
func (h *BatchHandler) DeleteGiftCards(c echo.Context) error {
	return h.batchDelete(c, "gift_card", h.authzService.CheckGiftCardAccess, func(ctx context.Context, id uuid.UUID) error {
		return h.giftCardService.DeleteGiftCard(ctx, id)
	})
}

// ==================== Batch Share ====================

// ShareCards batch-shares cards (partial success).
// POST /api/v1/cards/batch/share
func (h *BatchHandler) ShareCards(c echo.Context) error {
	return h.batchShare(c, "card", h.authzService.CheckCardAccess,
		func(ctx context.Context, resourceID, callerID, sharedWithID uuid.UUID, req BatchShareRequest) error {
			canEdit := req.CanEdit != nil && *req.CanEdit
			canDelete := req.CanDelete != nil && *req.CanDelete
			return h.shareService.CreateCardShare(ctx, callerID, resourceID, sharedWithID, canEdit, canDelete)
		})
}

// ShareVouchers batch-shares vouchers (partial success, always read-only).
// POST /api/v1/vouchers/batch/share
func (h *BatchHandler) ShareVouchers(c echo.Context) error {
	return h.batchShare(c, "voucher", h.authzService.CheckVoucherAccess,
		func(ctx context.Context, resourceID, callerID, sharedWithID uuid.UUID, _ BatchShareRequest) error {
			return h.shareService.CreateVoucherShare(ctx, callerID, resourceID, sharedWithID)
		})
}

// ShareGiftCards batch-shares gift cards (partial success).
// POST /api/v1/gift-cards/batch/share
func (h *BatchHandler) ShareGiftCards(c echo.Context) error {
	return h.batchShare(c, "gift_card", h.authzService.CheckGiftCardAccess,
		func(ctx context.Context, resourceID, callerID, sharedWithID uuid.UUID, req BatchShareRequest) error {
			canEdit := req.CanEdit != nil && *req.CanEdit
			canDelete := req.CanDelete != nil && *req.CanDelete
			canEditTx := req.CanEditTransactions != nil && *req.CanEditTransactions
			return h.shareService.CreateGiftCardShare(ctx, callerID, resourceID, sharedWithID, canEdit, canDelete, canEditTx)
		})
}

// ==================== Batch Transfer ====================

// TransferCards batch-transfers card ownership (partial success).
// POST /api/v1/cards/batch/transfer
func (h *BatchHandler) TransferCards(c echo.Context) error {
	return h.batchTransfer(c, "card", h.authzService.CheckCardAccess,
		func(ctx context.Context, resourceID, newOwnerID, currentOwnerID uuid.UUID) error {
			return h.transferService.TransferCardOwnership(ctx, resourceID, newOwnerID, currentOwnerID)
		})
}

// TransferVouchers batch-transfers voucher ownership (partial success).
// POST /api/v1/vouchers/batch/transfer
func (h *BatchHandler) TransferVouchers(c echo.Context) error {
	return h.batchTransfer(c, "voucher", h.authzService.CheckVoucherAccess,
		func(ctx context.Context, resourceID, newOwnerID, currentOwnerID uuid.UUID) error {
			return h.transferService.TransferVoucherOwnership(ctx, resourceID, newOwnerID, currentOwnerID)
		})
}

// TransferGiftCards batch-transfers gift card ownership (partial success).
// POST /api/v1/gift-cards/batch/transfer
func (h *BatchHandler) TransferGiftCards(c echo.Context) error {
	return h.batchTransfer(c, "gift_card", h.authzService.CheckGiftCardAccess,
		func(ctx context.Context, resourceID, newOwnerID, currentOwnerID uuid.UUID) error {
			return h.transferService.TransferGiftCardOwnership(ctx, resourceID, newOwnerID, currentOwnerID)
		})
}

// ==================== Batch Export ====================

// ExportCards batch-exports selected cards as JSON download.
// POST /api/v1/cards/batch/export
func (h *BatchHandler) ExportCards(c echo.Context) error {
	return h.batchExport(c, "cards", h.authzService.CheckCardAccess,
		func(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*services.BatchExportData, error) {
			return h.exportService.ExportCardsByIDs(ctx, userID, ids)
		})
}

// ExportVouchers batch-exports selected vouchers as JSON download.
// POST /api/v1/vouchers/batch/export
func (h *BatchHandler) ExportVouchers(c echo.Context) error {
	return h.batchExport(c, "vouchers", h.authzService.CheckVoucherAccess,
		func(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*services.BatchExportData, error) {
			return h.exportService.ExportVouchersByIDs(ctx, userID, ids)
		})
}

// ExportGiftCards batch-exports selected gift cards as JSON download.
// POST /api/v1/gift-cards/batch/export
func (h *BatchHandler) ExportGiftCards(c echo.Context) error {
	return h.batchExport(c, "gift-cards", h.authzService.CheckGiftCardAccess,
		func(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*services.BatchExportData, error) {
			return h.exportService.ExportGiftCardsByIDs(ctx, userID, ids)
		})
}

// batchExport performs a batch export with access checks.
func (h *BatchHandler) batchExport(
	c echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	exportFn func(context.Context, uuid.UUID, []uuid.UUID) (*services.BatchExportData, error),
) error {
	user := c.Get("current_user").(*models.User)

	var req BatchDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	ids, err := parseBatchIDs(req.IDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Check view access for all items
	for _, id := range ids {
		perms, err := checkAccess(ctx, user.ID, id)
		if err != nil || !perms.CanView {
			return c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to export all selected items",
			})
		}
	}

	data, err := exportFn(ctx, user.ID, ids)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to export data",
		})
	}

	filename := fmt.Sprintf("savvy-export-%s-%s.json", resourceType, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.JSON(http.StatusOK, data)
}

// ==================== Generic Batch Logic ====================

// batchDelete performs a batch delete with all-or-nothing semantics.
// First validates all permissions, then deletes all resources.
func (h *BatchHandler) batchDelete(
	c echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	deleteResource func(context.Context, uuid.UUID) error,
) error {
	user := c.Get("current_user").(*models.User)

	var req BatchDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	ids, err := parseBatchIDs(req.IDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Phase 1: Check all permissions (all-or-nothing)
	for _, id := range ids {
		perms, err := checkAccess(ctx, user.ID, id)
		if err != nil || !perms.CanDelete {
			return c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to delete all selected " + resourceType + "s",
			})
		}
	}

	// Phase 2: Delete all (permissions already verified)
	auditCtx := audit.AddAuditContextToContext(ctx, user.ID, c.RealIP(), c.Request().UserAgent())
	failed := make([]BatchFailedItem, 0)
	successCount := 0

	for _, id := range ids {
		if err := deleteResource(auditCtx, id); err != nil {
			slog.ErrorContext(ctx, "batch delete failed", "resource_type", resourceType, "resource_id", id, "error", err)
			failed = append(failed, BatchFailedItem{
				ID:    id.String(),
				Error: sanitizeBatchError(err, "Failed to delete "+resourceType),
			})
		} else {
			successCount++
		}
	}

	return c.JSON(http.StatusOK, BatchResponse{
		SuccessCount: successCount,
		Failed:       failed,
	})
}

// batchShare performs a batch share with partial success semantics.
func (h *BatchHandler) batchShare(
	c echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	createShare func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, BatchShareRequest) error,
) error {
	user := c.Get("current_user").(*models.User)

	var req BatchShareRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	ids, err := parseBatchIDs(req.IDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_email",
			Message: "Email is required for sharing",
		})
	}

	// Find target user
	email := strings.ToLower(strings.TrimSpace(req.Email))
	sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "share_failed",
			Message: "Could not share with this email address",
		})
	}

	if sharedUser.ID == user.ID {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "self_share",
			Message: "Cannot share with yourself",
		})
	}

	ctx := c.Request().Context()
	failed := make([]BatchFailedItem, 0)
	successCount := 0

	for _, id := range ids {
		perms, err := checkAccess(ctx, user.ID, id)
		if err != nil || !perms.IsOwner {
			failed = append(failed, BatchFailedItem{
				ID:    id.String(),
				Error: "Not the owner",
			})
			continue
		}

		if err := createShare(ctx, id, user.ID, sharedUser.ID, req); err != nil {
			slog.ErrorContext(ctx, "batch share failed", "resource_type", resourceType, "resource_id", id, "error", err)
			failed = append(failed, BatchFailedItem{
				ID:    id.String(),
				Error: sanitizeBatchError(err, "Failed to share "+resourceType),
			})
		} else {
			successCount++
		}
	}

	return c.JSON(http.StatusOK, BatchResponse{
		SuccessCount: successCount,
		Failed:       failed,
	})
}

// batchTransfer performs a batch transfer with partial success semantics.
func (h *BatchHandler) batchTransfer(
	c echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	transferOwnership func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error,
) error {
	user := c.Get("current_user").(*models.User)

	var req BatchTransferRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	ids, err := parseBatchIDs(req.IDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	if req.NewOwnerEmail == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_email",
			Message: "New owner email is required",
		})
	}

	// Find new owner
	email := strings.ToLower(strings.TrimSpace(req.NewOwnerEmail))
	newOwner, err := h.userService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "transfer_failed",
			Message: "Could not transfer to this email address",
		})
	}

	if newOwner.ID == user.ID {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "self_transfer",
			Message: "Cannot transfer to yourself",
		})
	}

	ctx := c.Request().Context()
	auditCtx := audit.AddAuditContextToContext(ctx, user.ID, c.RealIP(), c.Request().UserAgent())
	failed := make([]BatchFailedItem, 0)
	successCount := 0

	for _, id := range ids {
		perms, err := checkAccess(ctx, user.ID, id)
		if err != nil || !perms.IsOwner {
			failed = append(failed, BatchFailedItem{
				ID:    id.String(),
				Error: "Not the owner",
			})
			continue
		}

		if err := transferOwnership(auditCtx, id, newOwner.ID, user.ID); err != nil {
			slog.ErrorContext(ctx, "batch transfer failed", "resource_type", resourceType, "resource_id", id, "error", err)
			failed = append(failed, BatchFailedItem{
				ID:    id.String(),
				Error: sanitizeBatchError(err, "Failed to transfer "+resourceType),
			})
		} else {
			successCount++
		}
	}

	return c.JSON(http.StatusOK, BatchResponse{
		SuccessCount: successCount,
		Failed:       failed,
	})
}

// parseBatchIDs validates and parses a list of string IDs to UUIDs.
func parseBatchIDs(rawIDs []string) ([]uuid.UUID, error) {
	if len(rawIDs) == 0 {
		return nil, errNoBatchIDs
	}
	if len(rawIDs) > maxBatchSize {
		return nil, errTooManyBatchIDs
	}

	ids := make([]uuid.UUID, 0, len(rawIDs))
	seen := make(map[uuid.UUID]bool, len(rawIDs))

	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, errInvalidBatchID
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	return ids, nil
}

var (
	errNoBatchIDs      = echo.NewHTTPError(http.StatusBadRequest, "No IDs provided")
	errTooManyBatchIDs = echo.NewHTTPError(http.StatusBadRequest, "Too many IDs (max 50)")
	errInvalidBatchID  = echo.NewHTTPError(http.StatusBadRequest, "One or more IDs are invalid")
)
