// Package api contains JSON API handlers for the SvelteKit frontend.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"
	"strings"

	"github.com/labstack/echo/v5"
)

// ImportHandler handles data import API endpoints.
type ImportHandler struct {
	importService services.ImportServiceInterface
}

// NewImportHandler creates a new import API handler.
func NewImportHandler(importService services.ImportServiceInterface) *ImportHandler {
	return &ImportHandler{importService: importService}
}

// maxImportSize is the maximum upload size for imports (10MB).
const maxImportSize = 10 << 20

// maxImportItems is the maximum number of items per JSON import (across all types).
const maxImportItems = 500

// ImportJSON imports data from the Savvy export JSON format.
// POST /api/v1/import/json
func (h *ImportHandler) ImportJSON(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	data, err := h.parseJSONBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: err.Error(),
		})
	}

	totalItems := len(data.Cards) + len(data.Vouchers) + len(data.GiftCards)
	if totalItems > maxImportItems {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "too_many_items",
			Message: fmt.Sprintf("Import exceeds maximum of %d items (got %d)", maxImportItems, totalItems),
		})
	}

	result, err := h.importService.ImportJSON(c.Request().Context(), user.ID, data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "import_failed",
			Message: "Failed to import data",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// PreviewJSON previews what would be imported from JSON without executing.
// POST /api/v1/import/json/preview
func (h *ImportHandler) PreviewJSON(c echo.Context) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	data, err := h.parseJSONBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: err.Error(),
		})
	}

	totalItems := len(data.Cards) + len(data.Vouchers) + len(data.GiftCards)
	if totalItems > maxImportItems {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "too_many_items",
			Message: fmt.Sprintf("Import exceeds maximum of %d items (got %d)", maxImportItems, totalItems),
		})
	}

	preview, err := h.importService.PreviewJSON(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, preview)
}

// ImportCardsCSV imports cards from a CSV file.
// POST /api/v1/import/csv/cards
func (h *ImportHandler) ImportCardsCSV(c echo.Context) error {
	return h.handleCSVImport(c, "cards")
}

// ImportVouchersCSV imports vouchers from a CSV file.
// POST /api/v1/import/csv/vouchers
func (h *ImportHandler) ImportVouchersCSV(c echo.Context) error {
	return h.handleCSVImport(c, "vouchers")
}

// ImportGiftCardsCSV imports gift cards from a CSV file.
// POST /api/v1/import/csv/gift-cards
func (h *ImportHandler) ImportGiftCardsCSV(c echo.Context) error {
	return h.handleCSVImport(c, "gift-cards")
}

func (h *ImportHandler) handleCSVImport(c echo.Context, resourceType string) error {
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "Not authenticated",
		})
	}

	// Limit upload size
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxImportSize)

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "No file uploaded",
		})
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "File must be a CSV file",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to read uploaded file",
		})
	}
	defer func() { _ = src.Close() }()

	ctx := c.Request().Context()
	var result *services.ImportResult

	switch resourceType {
	case "cards":
		result, err = h.importService.ImportCardsCSV(ctx, user.ID, src)
	case "vouchers":
		result, err = h.importService.ImportVouchersCSV(ctx, user.ID, src)
	case "gift-cards":
		result, err = h.importService.ImportGiftCardsCSV(ctx, user.ID, src)
	default:
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid resource type",
		})
	}

	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "import_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *ImportHandler) parseJSONBody(c echo.Context) (*services.ExportData, error) {
	// Limit body size
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxImportSize)

	var data services.ExportData
	if err := json.NewDecoder(c.Request().Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
