// Package api contains JSON API handlers for merchants.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"log/slog"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/labstack/echo/v5"
)

// MerchantsHandler handles merchant API endpoints.
type MerchantsHandler struct {
	merchantService services.MerchantServiceInterface
}

// NewMerchantsHandler creates a new merchants API handler.
func NewMerchantsHandler(merchantService services.MerchantServiceInterface) *MerchantsHandler {
	return &MerchantsHandler{
		merchantService: merchantService,
	}
}

// List returns all merchants
// GET /api/v1/merchants
func (h *MerchantsHandler) List(c *echo.Context) error {
	merchants, err := h.merchantService.GetAllMerchants(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load merchants",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"merchants": ToMerchantDTOs(merchants),
	})
}

// Search searches merchants by query (autocomplete)
// GET /api/v1/merchants/search?q=...
func (h *MerchantsHandler) Search(c *echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		// Return all merchants if no query
		return h.List(c)
	}

	merchants, err := h.merchantService.SearchMerchants(c.Request().Context(), query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to search merchants",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"merchants": ToMerchantDTOs(merchants),
	})
}

// Show returns a single merchant
// GET /api/v1/merchants/:id
func (h *MerchantsHandler) Show(c *echo.Context) error {
	merchantID, err := parseResourceID(c, "merchant")
	if err != nil {
		return err
	}

	merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Merchant not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"merchant": ToMerchantDTO(merchant),
	})
}

// ==================== Admin Endpoints ====================

// Create creates a new merchant (Admin only)
// POST /api/v1/admin/merchants
func (h *MerchantsHandler) Create(c *echo.Context) error {
	var req MerchantCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_name",
			Message: "Merchant name is required",
		})
	}

	// Set default color if not provided
	color := "#3B82F6"
	if req.Color != nil && *req.Color != "" {
		color = *req.Color
	}

	merchant := &models.Merchant{
		Name:    req.Name,
		Color:   color,
		LogoURL: stringPtrValue(req.LogoURL),
		Website: stringPtrValue(req.Website),
	}

	if err := h.merchantService.CreateMerchant(c.Request().Context(), merchant); err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to create merchant", "name", logsafe.String(req.Name), "error", logsafe.Error(err))

		// Check for duplicate name error
		if err.Error() == "merchant with this name already exists" {
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "duplicate_name",
				Message: "A merchant with this name already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create merchant",
		})
	}

	slog.InfoContext(c.Request().Context(), "merchant created", "name", logsafe.String(merchant.Name), "id", logsafe.UUID(merchant.ID))

	return c.JSON(http.StatusCreated, map[string]any{
		"message":  "Merchant created successfully",
		"merchant": ToMerchantDTO(merchant),
	})
}

// Update updates a merchant (Admin only)
// PATCH /api/v1/admin/merchants/:id
func (h *MerchantsHandler) Update(c *echo.Context) error {
	merchantID, err := parseResourceID(c, "merchant")
	if err != nil {
		return err
	}

	var req MerchantUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Get existing merchant
	merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Merchant not found",
		})
	}

	// Apply updates
	if req.Name != nil && *req.Name != "" {
		merchant.Name = *req.Name
	}
	if req.Color != nil {
		merchant.Color = *req.Color
	}
	if req.LogoURL != nil {
		merchant.LogoURL = *req.LogoURL
	}
	if req.Website != nil {
		merchant.Website = *req.Website
	}

	// Validate
	if merchant.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_name",
			Message: "Merchant name is required",
		})
	}

	if err := h.merchantService.UpdateMerchant(c.Request().Context(), merchant); err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to update merchant", "merchant_id", merchantID, "error", err)

		// Check for duplicate name error
		if err.Error() == "merchant with this name already exists" {
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "duplicate_name",
				Message: "A merchant with this name already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update merchant",
		})
	}

	slog.InfoContext(c.Request().Context(), "merchant updated", "name", logsafe.String(merchant.Name), "id", logsafe.UUID(merchant.ID))

	return c.JSON(http.StatusOK, map[string]any{
		"message":  "Merchant updated successfully",
		"merchant": ToMerchantDTO(merchant),
	})
}

// Delete deletes a merchant (Admin only)
// DELETE /api/v1/admin/merchants/:id
func (h *MerchantsHandler) Delete(c *echo.Context) error {
	merchantID, err := parseResourceID(c, "merchant")
	if err != nil {
		return err
	}

	// Check if merchant exists
	merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Merchant not found",
		})
	}

	// Add audit context (user ID, IP address, user agent) for audit logging
	user := c.Get("current_user").(*models.User)
	ctx := audit.AddAuditContextToContext(c.Request().Context(), user.ID, c.RealIP(), c.Request().UserAgent())

	if err := h.merchantService.DeleteMerchant(ctx, merchantID); err != nil {
		slog.ErrorContext(ctx, "failed to delete merchant", "merchant_id", merchantID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to delete merchant",
		})
	}

	slog.InfoContext(ctx, "merchant deleted", "name", merchant.Name, "merchant_id", merchant.ID)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Merchant deleted successfully",
	})
}

// stringPtrValue returns the value of a string pointer, or empty string if nil
func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
