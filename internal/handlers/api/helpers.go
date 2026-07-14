// Package api contains JSON API handlers and DTOs for the SvelteKit frontend.
package api //nolint:revive // "api" is a meaningful package name for API handlers
import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/validation"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// parsePaginationParams extracts optional page and per_page query parameters.
// Returns (page, perPage, isPaginated). If neither param is provided, isPaginated is false
// and the caller should use the non-paginated code path (backward-compatible).
func parsePaginationParams(c *echo.Context) (int, int, bool) {
	pageStr := c.QueryParam("page")
	perPageStr := c.QueryParam("per_page")

	if pageStr == "" && perPageStr == "" {
		return 0, 0, false
	}

	page := 1
	perPage := 25

	if pageStr != "" {
		if v, err := strconv.Atoi(pageStr); err == nil && v > 0 {
			page = v
		}
	}
	if perPageStr != "" {
		if v, err := strconv.Atoi(perPageStr); err == nil && v > 0 {
			perPage = v
		}
	}

	// Cap per_page to prevent abuse
	if perPage > 100 {
		perPage = 100
	}

	return page, perPage, true
}

// getFavoriteIDsForResources retrieves favorite IDs for a list of resources
// This eliminates code duplication in List handlers for cards, vouchers, and gift cards
func getFavoriteIDsForResources(
	ctx context.Context,
	favoriteService services.FavoriteServiceInterface,
	userID uuid.UUID,
	resourceType string,
	resourceIDs []uuid.UUID,
) map[string]bool {
	favoriteIDs := make(map[string]bool)
	for _, resourceID := range resourceIDs {
		isFav, _ := favoriteService.IsFavorite(ctx, userID, resourceType, resourceID)
		if isFav {
			favoriteIDs[resourceID.String()] = true
		}
	}
	return favoriteIDs
}

// resolveMerchant resolves a merchant ID and name from request data
// Either uses existing merchant_id or creates a new merchant from new_merchant_name
// Returns merchantID, merchantName, and error (which is already an echo.HTTPError for direct return)
func resolveMerchant(
	c *echo.Context,
	merchantService services.MerchantServiceInterface,
	merchantID *string,
	newMerchantName *string,
) (*uuid.UUID, string, error) {
	if merchantID != nil {
		// Use existing merchant
		mid, err := uuid.Parse(*merchantID)
		if err != nil {
			return nil, "", c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_merchant_id",
				Message: "Invalid merchant ID",
			})
		}

		merchant, err := merchantService.GetMerchantByID(c.Request().Context(), mid)
		if err != nil {
			slog.ErrorContext(c.Request().Context(), "failed to get merchant", "error", err)
			return nil, "", c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_merchant",
				Message: "Merchant not found",
			})
		}

		return &mid, merchant.Name, nil
	} else if newMerchantName != nil && *newMerchantName != "" {
		// Create new merchant
		merchant := &models.Merchant{Name: *newMerchantName}
		if err := merchantService.CreateMerchant(c.Request().Context(), merchant); err != nil {
			slog.ErrorContext(c.Request().Context(), "failed to create merchant", "name", logsafe.String(*newMerchantName), "error", err)
			return nil, "", c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to create merchant",
			})
		}

		return &merchant.ID, merchant.Name, nil
	}

	// Neither merchant_id nor new_merchant_name provided
	return nil, "", c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "missing_merchant",
		Message: "Either merchant_id or new_merchant_name is required",
	})
}

// resolveMerchantUpdate resolves a merchant update from a request field value.
// An empty string removes the merchant association. A valid UUID looks up the merchant.
// Returns (merchantID, merchantName, error). If error is non-nil, the response has been sent.
func resolveMerchantUpdate(
	c *echo.Context,
	merchantService services.MerchantServiceInterface,
	merchantIDStr string,
) (*uuid.UUID, string, error) {
	if merchantIDStr == "" {
		return nil, "", nil
	}

	mid, err := uuid.Parse(merchantIDStr)
	if err != nil {
		return nil, "", c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_merchant_id",
			Message: "Invalid merchant ID format",
		})
	}

	merchant, err := merchantService.GetMerchantByID(c.Request().Context(), mid)
	if err != nil {
		return nil, "", c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_merchant",
			Message: "Merchant not found",
		})
	}

	return &mid, merchant.Name, nil
}

// handleResourceDelete provides generic delete handler logic for resources
// This eliminates code duplication in Delete handlers for cards, vouchers, and gift cards
func handleResourceDelete(
	c *echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	deleteResource func(context.Context, uuid.UUID) error,
) error {
	user := c.Get("current_user").(*models.User)

	resourceID, err := parseResourceID(c, resourceType)
	if err != nil {
		return err
	}

	// Check authorization
	perms, err := checkAccess(c.Request().Context(), user.ID, resourceID)
	if err != nil || !perms.CanDelete {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to delete this " + resourceType,
		})
	}

	// Add audit context (user ID, IP address, user agent) for audit logging
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	if err := deleteResource(ctx, resourceID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to delete " + resourceType,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": capitalizeFirst(resourceType) + " deleted"})
}

// handleResourceToggleFavorite provides generic toggle favorite handler logic
// This eliminates code duplication in ToggleFavorite handlers for cards, vouchers, and gift cards
func handleResourceToggleFavorite(
	c *echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	favoriteService services.FavoriteServiceInterface,
) error {
	user := c.Get("current_user").(*models.User)

	resourceID, err := parseResourceID(c, resourceType)
	if err != nil {
		return err
	}

	// Check access
	_, err = checkAccess(c.Request().Context(), user.ID, resourceID)
	if err != nil {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have access to this " + resourceType,
		})
	}

	if err := favoriteService.ToggleFavorite(c.Request().Context(), user.ID, resourceType, resourceID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to toggle favorite",
		})
	}

	isFavorite, err := favoriteService.IsFavorite(c.Request().Context(), user.ID, resourceType, resourceID)
	if err != nil {
		slog.WarnContext(c.Request().Context(), "failed to check favorite status", "resource_type", resourceType, "resource_id", resourceID, "error", err)
	}

	return c.JSON(http.StatusOK, map[string]bool{"is_favorite": isFavorite})
}

// handleResourceDeleteShare provides generic delete share handler logic
// This eliminates code duplication in DeleteShare handlers for cards, vouchers, and gift cards
func handleResourceDeleteShare(
	c *echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	deleteShare func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error,
) error {
	user := c.Get("current_user").(*models.User)

	resourceID, err := parseResourceID(c, resourceType)
	if err != nil {
		return err
	}

	sharedWithID, err := parseUUIDParam(c, "sharedWithID", "invalid_user_id", "Invalid user ID")
	if err != nil {
		return err
	}

	// Check authorization - only owner can unshare
	perms, err := checkAccess(c.Request().Context(), user.ID, resourceID)
	if err != nil || !perms.IsOwner {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only the owner can unshare this " + resourceType,
		})
	}

	// Add audit context (user ID, IP address, user agent) for audit logging
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	// Delete share (callerUserID passed for defense-in-depth ownership check in service layer)
	if err := deleteShare(ctx, user.ID, resourceID, sharedWithID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to delete share",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Share removed successfully"})
}

// handleResourceDeleteAllShares provides generic bulk-revoke handler logic:
// removes every share for a resource. Only the owner may bulk-revoke.
func handleResourceDeleteAllShares(
	c *echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	deleteAllShares func(context.Context, uuid.UUID, uuid.UUID) error,
) error {
	user := c.Get("current_user").(*models.User)

	resourceID, err := parseResourceID(c, resourceType)
	if err != nil {
		return err
	}

	// Check authorization - only owner can unshare
	perms, err := checkAccess(c.Request().Context(), user.ID, resourceID)
	if err != nil || !perms.IsOwner {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only the owner can unshare this " + resourceType,
		})
	}

	// Add audit context (user ID, IP address, user agent) for audit logging
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	// Delete all shares (callerUserID passed for defense-in-depth ownership check in service layer)
	if err := deleteAllShares(ctx, user.ID, resourceID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to delete shares",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "All shares removed successfully"})
}

// handleResourceTransfer provides generic transfer handler logic for resources
// This eliminates code duplication in Transfer handlers for cards, vouchers, and gift cards
func handleResourceTransfer(
	c *echo.Context,
	resourceType string,
	checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error),
	transferOwnership func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error,
	userService services.UserServiceInterface,
) error {
	user := c.Get("current_user").(*models.User)

	resourceID, err := parseResourceID(c, resourceType)
	if err != nil {
		return err
	}

	// Check authorization - only owner can transfer
	perms, err := checkAccess(c.Request().Context(), user.ID, resourceID)
	if err != nil || !perms.IsOwner {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Only the owner can transfer this " + resourceType,
		})
	}

	var req TransferRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.NewOwnerEmail == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_email",
			Message: "New owner email is required",
		})
	}

	// Find new owner by email
	email := strings.ToLower(strings.TrimSpace(req.NewOwnerEmail))
	newOwner, err := userService.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "transfer_failed",
			Message: "Could not transfer to this email address",
		})
	}

	// Prevent self-transfer
	if newOwner.ID == user.ID {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "self_transfer",
			Message: "Cannot transfer to yourself",
		})
	}

	// Add audit context (user ID, IP address, user agent) for audit logging
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	if err := transferOwnership(ctx, resourceID, newOwner.ID, user.ID); err != nil {
		// Log the detailed error server-side
		slog.ErrorContext(c.Request().Context(), "failed to transfer ownership", "resource_type", resourceType, "resource_id", resourceID, "error", err)

		// Return generic error message to client (don't leak internal details)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "transfer_failed",
			Message: "Failed to transfer ownership. Please try again later.",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": capitalizeFirst(resourceType) + " transferred successfully"})
}

// maxShareRecipients caps recipients per multi-share call, matching maxBatchSize.
const maxShareRecipients = 50

// handleResourceMultiShare provides generic multi-recipient share handler logic for
// cards, vouchers, and gift cards. It resolves each email to a user, calls createShare
// once per recipient (with the permissions already captured in the closure), and returns
// a partial-success response. Unknown emails, self-shares, and already-shared recipients
// become failed[] entries rather than failing the whole request.
func handleResourceMultiShare(
	c *echo.Context,
	resourceType string,
	resourceID uuid.UUID,
	req ShareCreateRequest,
	userService services.UserServiceInterface,
	createShare func(ctx context.Context, sharedWithID uuid.UUID) error,
	loadShares func(ctx context.Context) ([]ShareDTO, error),
) error {
	user := c.Get("current_user").(*models.User)
	ctx := c.Request().Context()

	if len(req.Emails) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_recipients",
			Message: "At least one recipient email is required",
		})
	}
	if len(req.Emails) > maxShareRecipients {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "too_many_recipients",
			Message: fmt.Sprintf("At most %d recipients per share request", maxShareRecipients),
		})
	}

	resp := ShareCreateResponse{Failed: []BatchFailedItem{}}
	seen := make(map[string]bool, len(req.Emails))

	for _, raw := range req.Emails {
		email := strings.ToLower(strings.TrimSpace(raw))
		if email == "" || seen[email] {
			continue
		}
		seen[email] = true

		sharedUser, err := userService.GetUserByEmail(ctx, email)
		if err != nil {
			resp.Failed = append(resp.Failed, BatchFailedItem{ID: email, Error: "user not found"})
			continue
		}
		if sharedUser.ID == user.ID {
			resp.Failed = append(resp.Failed, BatchFailedItem{ID: email, Error: "cannot share with yourself"})
			continue
		}

		if err := createShare(ctx, sharedUser.ID); err != nil {
			if errors.Is(err, services.ErrAlreadyShared) {
				resp.Failed = append(resp.Failed, BatchFailedItem{ID: email, Error: "already shared with this user"})
				continue
			}
			slog.ErrorContext(ctx, "failed to create share", "resource_type", resourceType, "resource_id", resourceID, "error", err)
			resp.Failed = append(resp.Failed, BatchFailedItem{ID: email, Error: "share failed"})
			continue
		}
		resp.SuccessCount++
	}

	shares, err := loadShares(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to load shares", "resource_type", resourceType, "resource_id", resourceID, "error", err)
	}
	resp.Shares = shares

	// If every recipient failed, surface it as a 4xx so callers, logs and
	// monitoring don't read a total failure as success. Partial success stays 201.
	status := http.StatusCreated
	if resp.SuccessCount == 0 && len(resp.Failed) > 0 {
		status = http.StatusUnprocessableEntity
	}
	return c.JSON(status, resp)
}

// parseResourceID extracts the "id" path parameter as a UUID.
// On failure, sends a 400 JSON error response and returns a non-nil error.
// Callers should return the error directly (response is already sent).
func parseResourceID(c *echo.Context, resourceType string) (uuid.UUID, error) {
	return parseUUIDParam(c, "id", "invalid_id", "Invalid "+resourceType+" ID")
}

// parseUUIDParam extracts and parses a UUID from an Echo path parameter.
// On failure, sends a 400 JSON error response and returns a non-nil error.
// Callers should return the error directly (response is already sent).
func parseUUIDParam(c *echo.Context, param string, errorCode string, errorMessage string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   errorCode,
			Message: errorMessage,
		})
		return uuid.Nil, fmt.Errorf("invalid %s: %w", param, err)
	}
	return id, nil
}

// ==================== Enum Validation ====================

// Package-level aliases for the shared validation maps (for convenience within handlers).
var (
	validStatuses        = validation.ValidStatuses
	validBarcodeTypes    = validation.ValidBarcodeTypes
	validVoucherTypes    = validation.ValidVoucherTypes
	validUsageLimitTypes = validation.ValidUsageLimitTypes
	validCurrencies      = validation.ValidCurrencies
)

// validateEnum checks if a value is in the allowed set.
// Returns a 400 error response if invalid, nil if valid.
func validateEnum(c *echo.Context, value string, allowed map[string]bool, fieldName string) error {
	if err := validation.ValidateEnum(value, allowed, fieldName); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_" + fieldName,
			Message: fmt.Sprintf("Invalid %s value: %q", fieldName, value),
		})
	}
	return nil
}

// capitalizeFirst capitalizes the first ASCII letter of a string.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// validateStringLength checks if a string exceeds the maximum allowed length for the given field.
func validateStringLength(value string, fieldName string) error {
	return validation.ValidateStringLength(value, fieldName)
}

// validateStringLengthPtr is the pointer variant for optional fields.
func validateStringLengthPtr(value *string, fieldName string) error {
	return validation.ValidateStringLengthPtr(value, fieldName)
}
