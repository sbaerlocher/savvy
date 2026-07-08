// Package api contains JSON API handlers for vouchers.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// VouchersHandler handles voucher API endpoints.
type VouchersHandler struct {
	voucherService  services.VoucherServiceInterface
	authzService    services.AuthzServiceInterface
	merchantService services.MerchantServiceInterface
	userService     services.UserServiceInterface
	favoriteService services.FavoriteServiceInterface
	shareService    services.ShareServiceInterface
	transferService services.TransferServiceInterface
}

// NewVouchersHandler creates a new vouchers API handler.
func NewVouchersHandler(
	voucherService services.VoucherServiceInterface,
	authzService services.AuthzServiceInterface,
	merchantService services.MerchantServiceInterface,
	userService services.UserServiceInterface,
	favoriteService services.FavoriteServiceInterface,
	shareService services.ShareServiceInterface,
	transferService services.TransferServiceInterface,
) *VouchersHandler {
	return &VouchersHandler{
		voucherService:  voucherService,
		authzService:    authzService,
		merchantService: merchantService,
		userService:     userService,
		favoriteService: favoriteService,
		shareService:    shareService,
		transferService: transferService,
	}
}

// List returns all vouchers (owned + shared)
// GET /api/v1/vouchers
// Optional query params: ?page=1&per_page=25 for pagination
func (h *VouchersHandler) List(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	page, perPage, isPaginated := parsePaginationParams(c)

	if isPaginated {
		result, err := h.voucherService.GetUserVouchersPaginated(c.Request().Context(), user.ID, page, perPage)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to load vouchers",
			})
		}

		voucherIDs := make([]uuid.UUID, len(result.Items))
		for i, voucher := range result.Items {
			voucherIDs[i] = voucher.ID
		}
		favoriteIDs := getFavoriteIDsForResources(c.Request().Context(), h.favoriteService, user.ID, "voucher", voucherIDs)
		shareCounts, _ := h.shareService.GetVoucherShareCounts(c.Request().Context(), voucherIDs)

		dtos := ToVoucherDTOs(result.Items, favoriteIDs)
		for i, voucher := range result.Items {
			dtos[i].SharedWithCount = int(shareCounts[voucher.ID])
		}

		return c.JSON(http.StatusOK, VoucherListResponse{
			Vouchers: dtos,
			Pagination: &PaginationMeta{
				Total:      result.Total,
				Page:       result.Page,
				PerPage:    result.PerPage,
				TotalPages: result.TotalPages,
			},
		})
	}

	// Non-paginated: return all (backward-compatible)
	vouchers, err := h.voucherService.GetUserVouchers(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load vouchers",
		})
	}

	voucherIDs := make([]uuid.UUID, len(vouchers))
	for i, voucher := range vouchers {
		voucherIDs[i] = voucher.ID
	}
	favoriteIDs := getFavoriteIDsForResources(c.Request().Context(), h.favoriteService, user.ID, "voucher", voucherIDs)
	shareCounts, _ := h.shareService.GetVoucherShareCounts(c.Request().Context(), voucherIDs)

	dtos := ToVoucherDTOs(vouchers, favoriteIDs)
	for i, voucher := range vouchers {
		dtos[i].SharedWithCount = int(shareCounts[voucher.ID])
	}

	return c.JSON(http.StatusOK, VoucherListResponse{
		Vouchers: dtos,
	})
}

// Show returns a single voucher with permissions
// GET /api/v1/vouchers/:id
func (h *VouchersHandler) Show(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	voucherID, err := parseResourceID(c, "voucher")
	if err != nil {
		return err
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have access to this voucher",
		})
	}

	voucher, err := h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Voucher not found",
		})
	}

	// Check if favorited
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "voucher", voucherID)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "failed to check favorite status", "resource_type", "voucher", "resource_id", voucherID, "error", err)
	}

	voucherDTO := ToVoucherDTO(voucher, isFavorite)
	voucherDTO.Permissions = &PermissionDTO{
		CanView:   perms.CanView,
		CanEdit:   perms.CanEdit,
		CanDelete: perms.CanDelete,
		IsOwner:   perms.IsOwner,
	}

	// Get shares if owner
	var shares []ShareDTO
	if perms.IsOwner {
		voucherShares, err := h.shareService.GetVoucherShares(c.Request().Context(), voucherID)
		if err != nil {
			slog.WarnContext(c.Request().Context(), "failed to load shares", "resource_type", "voucher", "resource_id", voucherID, "error", err)
		}
		shares = ToVoucherShareDTOs(voucherShares)
	}

	return c.JSON(http.StatusOK, VoucherDetailResponse{
		Voucher:     voucherDTO,
		Permissions: *voucherDTO.Permissions,
		Shares:      shares,
	})
}

// Create creates a new voucher
// POST /api/v1/vouchers
func (h *VouchersHandler) Create(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	var req VoucherCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Code == "" || req.Type == "" || req.ValidFrom == "" || req.ValidUntil == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_fields",
			Message: "Code, type, valid_from, and valid_until are required",
		})
	}

	// Validate string lengths
	if err := validateStringLength(req.Code, "code"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.Description, "description"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.NewMerchantName, "name"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}

	// Validate enum fields
	if err := validateEnum(c, req.Type, validVoucherTypes, "type"); err != nil {
		return err
	}

	// Validate merchant (either MerchantID or MerchantName required)
	if req.MerchantID == nil && (req.NewMerchantName == nil || *req.NewMerchantName == "") {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_merchant",
			Message: "Merchant is required: select an existing merchant or enter a new merchant name",
		})
	}

	// Validate value > 0 (free vouchers carry no value and are exempt)
	if req.Type != "free" && req.Value <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_value",
			Message: "Value must be greater than 0",
		})
	}

	// Parse dates
	validFrom, err := time.Parse(time.RFC3339, req.ValidFrom)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid valid_from date format (use ISO 8601/RFC3339)",
		})
	}

	validUntil, err := time.Parse(time.RFC3339, req.ValidUntil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid valid_until date format (use ISO 8601/RFC3339)",
		})
	}

	// Handle merchant
	merchantID, merchantName, err := resolveMerchant(c, h.merchantService, req.MerchantID, req.NewMerchantName)
	if err != nil {
		return err
	}

	// Validate optional enum fields
	barcodeType := stringOrDefault(req.BarcodeType, "CODE128")
	if err := validateEnum(c, barcodeType, validBarcodeTypes, "barcode_type"); err != nil {
		return err
	}
	usageLimitType := stringOrDefault(req.UsageLimitType, "single_use")
	if err := validateEnum(c, usageLimitType, validUsageLimitTypes, "usage_limit_type"); err != nil {
		return err
	}

	// Check for duplicates — blocks creation to prevent DB unique constraint violation
	duplicate, err := h.voucherService.CheckDuplicate(c.Request().Context(), req.Code, user.ID, nil)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to check duplicate", "error", err)
		// Don't fail the request, just log — worst case the DB constraint catches it
	}
	if duplicate != nil {
		slog.InfoContext(c.Request().Context(), "duplicate voucher blocked", "existing_id", duplicate.ID)
		return c.JSON(http.StatusConflict, DuplicateErrorResponse{
			Error:   "duplicate_barcode",
			Message: "A voucher with this code already exists",
			Duplicate: &DuplicateWarning{
				HasDuplicate:   true,
				MerchantName:   duplicate.MerchantName,
				ResourceNumber: duplicate.Code,
				ExistingID:     duplicate.ID.String(),
			},
		})
	}

	// Soft-deleted twin owned by this user → offer restore instead of a hard failure
	deletedDup, err := h.voucherService.FindDeletedDuplicate(c.Request().Context(), req.Code, user.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to check deleted duplicate", "error", err)
	}
	if deletedDup != nil {
		return c.JSON(http.StatusConflict, DuplicateErrorResponse{
			Error:   "duplicate_barcode",
			Message: "A soft-deleted voucher with this code exists and can be restored",
			Duplicate: &DuplicateWarning{
				HasDuplicate:   true,
				MerchantName:   deletedDup.MerchantName,
				ResourceNumber: deletedDup.Code,
				ExistingID:     deletedDup.ID.String(),
				Deleted:        true,
			},
		})
	}

	// Create voucher
	voucher := &models.Voucher{
		UserID:            &user.ID,
		MerchantID:        merchantID,
		MerchantName:      merchantName,
		Code:              req.Code,
		Type:              req.Type,
		Value:             req.Value,
		Currency:          stringOrDefault(req.Currency, "CHF"),
		Description:       stringOrDefault(req.Description, ""),
		MinPurchaseAmount: floatOrDefault(req.MinPurchaseAmount, 0),
		ValidFrom:         validFrom,
		ValidUntil:        validUntil,
		UsageLimitType:    usageLimitType,
		BarcodeType:       barcodeType,
	}

	if err := h.voucherService.CreateVoucher(c.Request().Context(), voucher); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			// TOCTOU race: re-fetch to return details
			if existing, _ := h.voucherService.CheckDuplicate(c.Request().Context(), req.Code, user.ID, nil); existing != nil {
				return c.JSON(http.StatusConflict, DuplicateErrorResponse{
					Error:   "duplicate_barcode",
					Message: "A voucher with this code already exists",
					Duplicate: &DuplicateWarning{
						HasDuplicate:   true,
						MerchantName:   existing.MerchantName,
						ResourceNumber: existing.Code,
						ExistingID:     existing.ID.String(),
					},
				})
			}
		}
		slog.ErrorContext(c.Request().Context(), "failed to create voucher", "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create voucher",
		})
	}

	// Handle optional sharing on creation
	if req.ShareWithEmail != nil && *req.ShareWithEmail != "" {
		email := strings.ToLower(strings.TrimSpace(*req.ShareWithEmail))
		sharedUser, err := h.userService.GetUserByEmail(c.Request().Context(), email)
		if err == nil && sharedUser.ID != user.ID {
			// Create share (vouchers are always read-only)
			if err := h.shareService.CreateVoucherShare(c.Request().Context(), user.ID, voucher.ID, sharedUser.ID); err != nil {
				slog.WarnContext(c.Request().Context(), "failed to create share on voucher creation", "voucher_id", voucher.ID, "error", err)
			} else {
				slog.InfoContext(c.Request().Context(), "share created on voucher creation", "voucher_id", voucher.ID, "shared_with_id", sharedUser.ID)
			}
		}
	}

	// Reload with merchant relation
	voucher, err = h.voucherService.GetVoucher(c.Request().Context(), voucher.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to reload voucher after creation", "voucher_id", voucher.ID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Voucher created but failed to reload",
		})
	}

	return c.JSON(http.StatusCreated, VoucherDetailResponse{
		Voucher: ToVoucherDTO(voucher, false),
		Permissions: PermissionDTO{
			CanView:   true,
			CanEdit:   true,
			CanDelete: true,
			IsOwner:   true,
		},
	})
}

// Update updates a voucher
// PUT /api/v1/vouchers/:id
func (h *VouchersHandler) Update(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	voucherID, err := parseResourceID(c, "voucher")
	if err != nil {
		return err
	}

	// Check authorization
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.CanEdit {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to edit this voucher",
		})
	}

	var req VoucherUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate string lengths
	if err := validateStringLengthPtr(req.Code, "code"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}
	if err := validateStringLengthPtr(req.Description, "description"); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "field_too_long", Message: err.Error()})
	}

	// Get existing voucher
	voucher, err := h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Voucher not found",
		})
	}

	// Apply updates
	if req.MerchantID != nil {
		mid, name, err := resolveMerchantUpdate(c, h.merchantService, *req.MerchantID)
		if err != nil {
			return err
		}
		voucher.MerchantID = mid
		voucher.MerchantName = name
		// CRITICAL: Clear the Merchant relation to prevent GORM from overwriting MerchantID
		voucher.Merchant = nil
	}
	if req.Code != nil {
		voucher.Code = *req.Code
	}
	if req.Type != nil {
		if err := validateEnum(c, *req.Type, validVoucherTypes, "type"); err != nil {
			return err
		}
		voucher.Type = *req.Type
	}
	if req.Value != nil {
		voucher.Value = *req.Value
	}
	if req.Currency != nil {
		voucher.Currency = *req.Currency
	}
	if req.Description != nil {
		voucher.Description = *req.Description
	}
	if req.MinPurchaseAmount != nil {
		voucher.MinPurchaseAmount = *req.MinPurchaseAmount
	}
	if req.ValidFrom != nil {
		validFrom, err := time.Parse(time.RFC3339, *req.ValidFrom)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid valid_from date format (use ISO 8601/RFC3339)",
			})
		}
		voucher.ValidFrom = validFrom
	}
	if req.ValidUntil != nil {
		validUntil, err := time.Parse(time.RFC3339, *req.ValidUntil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid valid_until date format (use ISO 8601/RFC3339)",
			})
		}
		voucher.ValidUntil = validUntil
	}
	if req.UsageLimitType != nil {
		if err := validateEnum(c, *req.UsageLimitType, validUsageLimitTypes, "usage_limit_type"); err != nil {
			return err
		}
		voucher.UsageLimitType = *req.UsageLimitType
	}
	if req.BarcodeType != nil {
		if err := validateEnum(c, *req.BarcodeType, validBarcodeTypes, "barcode_type"); err != nil {
			return err
		}
		voucher.BarcodeType = *req.BarcodeType
	}

	// Check for duplicates if code changed (warning only)
	duplicateWarning := checkVoucherDuplicate(c, h.voucherService, req.Code, user.ID, &voucherID)

	if err := h.voucherService.UpdateVoucher(c.Request().Context(), voucher); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update voucher",
		})
	}

	// Reload
	voucher, err = h.voucherService.GetVoucher(c.Request().Context(), voucherID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to reload voucher after update", "voucher_id", voucherID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Voucher updated but failed to reload",
		})
	}
	isFavorite, err := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "voucher", voucherID)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "failed to check favorite status", "resource_type", "voucher", "resource_id", voucherID, "error", err)
	}

	// Get shares if owner
	var shares []ShareDTO
	if perms.IsOwner {
		voucherShares, err := h.shareService.GetVoucherShares(c.Request().Context(), voucherID)
		if err != nil {
			slog.WarnContext(c.Request().Context(), "failed to load shares", "resource_type", "voucher", "resource_id", voucherID, "error", err)
		}
		shares = ToVoucherShareDTOs(voucherShares)
	}

	return c.JSON(http.StatusOK, VoucherDetailResponse{
		Voucher: ToVoucherDTO(voucher, isFavorite),
		Permissions: PermissionDTO{
			CanView:   perms.CanView,
			CanEdit:   perms.CanEdit,
			CanDelete: perms.CanDelete,
			IsOwner:   perms.IsOwner,
		},
		Shares:           shares,
		DuplicateWarning: duplicateWarning,
	})
}

// Delete deletes a voucher
// DELETE /api/v1/vouchers/:id
func (h *VouchersHandler) Delete(c *echo.Context) error {
	return handleResourceDelete(c, "voucher", h.authzService.CheckVoucherAccess, h.voucherService.DeleteVoucher)
}

// ToggleFavorite toggles favorite status
// POST /api/v1/vouchers/:id/favorite
func (h *VouchersHandler) ToggleFavorite(c *echo.Context) error {
	return handleResourceToggleFavorite(c, "voucher", h.authzService.CheckVoucherAccess, h.favoriteService)
}

// CreateShare creates a new share (read-only for vouchers)
// POST /api/v1/vouchers/:id/share
func (h *VouchersHandler) CreateShare(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	voucherID, err := parseResourceID(c, "voucher")
	if err != nil {
		return err
	}

	// Check authorization - only owner can share
	perms, err := h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherID)
	if err != nil || !perms.IsOwner {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only the owner can share this voucher",
		})
	}

	var req ShareCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Vouchers are always read-only: no permissions honored.
	return handleResourceMultiShare(c, "voucher", voucherID, req, h.userService,
		func(ctx context.Context, sharedWithID uuid.UUID) error {
			return h.shareService.CreateVoucherShare(ctx, user.ID, voucherID, sharedWithID)
		},
		func(ctx context.Context) ([]ShareDTO, error) {
			shares, err := h.shareService.GetVoucherShares(ctx, voucherID)
			return ToVoucherShareDTOs(shares), err
		},
	)
}

// DeleteShare removes a share
// DELETE /api/v1/vouchers/:id/share/:sharedWithID
func (h *VouchersHandler) DeleteShare(c *echo.Context) error {
	return handleResourceDeleteShare(c, "voucher", h.authzService.CheckVoucherAccess, h.shareService.DeleteVoucherShare)
}

// DeleteAllShares removes all shares for a voucher
// DELETE /api/v1/vouchers/:id/shares
func (h *VouchersHandler) DeleteAllShares(c *echo.Context) error {
	return handleResourceDeleteAllShares(c, "voucher", h.authzService.CheckVoucherAccess, h.shareService.DeleteAllVoucherShares)
}

// Transfer transfers ownership
// POST /api/v1/vouchers/:id/transfer
func (h *VouchersHandler) Transfer(c *echo.Context) error {
	return handleResourceTransfer(c, "voucher", h.authzService.CheckVoucherAccess, h.transferService.TransferVoucherOwnership, h.userService)
}

// Restore restores a soft-deleted voucher owned by the current user
// POST /api/v1/vouchers/:id/restore
func (h *VouchersHandler) Restore(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	voucherID, err := parseResourceID(c, "voucher")
	if err != nil {
		return err
	}

	restored, err := h.voucherService.RestoreVoucher(c.Request().Context(), voucherID, user.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to restore voucher", "voucher_id", voucherID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to restore voucher",
		})
	}
	if restored == nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "No restorable voucher found",
		})
	}

	isFavorite, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "voucher", voucherID)
	dto := ToVoucherDTO(restored, isFavorite)
	// A restored voucher is always owned by the caller — return owner permissions.
	perms := PermissionDTO{CanView: true, CanEdit: true, CanDelete: true, IsOwner: true}
	return c.JSON(http.StatusOK, VoucherDetailResponse{Voucher: dto, Permissions: perms})
}

// checkVoucherDuplicate checks for duplicate vouchers by code (warning only, does not block)
func checkVoucherDuplicate(c *echo.Context, svc services.VoucherServiceInterface, code *string, userID uuid.UUID, excludeID *uuid.UUID) *DuplicateWarning {
	if code == nil || *code == "" {
		return nil
	}
	duplicate, err := svc.CheckDuplicate(c.Request().Context(), *code, userID, excludeID)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "failed to check duplicate", "error", err)
		return nil
	}
	if duplicate == nil {
		return nil
	}
	return &DuplicateWarning{
		HasDuplicate:   true,
		MerchantName:   duplicate.MerchantName,
		ResourceNumber: duplicate.Code,
		ExistingID:     duplicate.ID.String(),
	}
}

// floatOrDefault returns the dereferenced float64 or default if nil
func floatOrDefault(f *float64, defaultValue float64) float64 {
	if f == nil {
		return defaultValue
	}
	return *f
}
