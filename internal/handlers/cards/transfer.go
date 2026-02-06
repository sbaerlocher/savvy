// Package cards handles card-related HTTP requests.
package cards

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"savvy/internal/audit"
	"savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/templates"
)

// TransferInline renders the transfer form (inline on show page).
func (h *Handler) TransferInline(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.T(c.Request().Context(), "error.invalid_id"))
	}

	// Authorization: only owner can see transfer form
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusForbidden, i18n.T(c.Request().Context(), "error.only_owner_can_transfer"))
	}

	// Render inline transfer form template
	csrfToken := ""
	if token := c.Get("csrf"); token != nil {
		csrfToken = token.(string)
	}
	return templates.CardTransferInlineForm(c.Request().Context(), csrfToken, cardID.String()).
		Render(c.Request().Context(), c.Response().Writer)
}

// CancelTransfer hides the transfer form (HTMX swap with empty string).
func (h *Handler) CancelTransfer(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

// Transfer executes the ownership transfer.
func (h *Handler) Transfer(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/cards")
	}

	// Parse new owner email
	newOwnerEmail := strings.ToLower(strings.TrimSpace(c.FormValue("new_owner_email")))
	if newOwnerEmail == "" {
		// Return error to form
		errMsg := i18n.T(c.Request().Context(), "error.email_required")
		return c.String(http.StatusBadRequest, fmt.Sprintf(`<div class="text-red-600 text-sm">%s</div>`, errMsg))
	}

	// Get new owner by email
	newOwner, err := h.userService.GetUserByEmail(c.Request().Context(), newOwnerEmail)
	if err != nil {
		errMsg := i18n.T(c.Request().Context(), "error.user_not_found")
		return c.String(http.StatusBadRequest, fmt.Sprintf(`<div class="text-red-600 text-sm">%s</div>`, errMsg))
	}

	// Authorization: verify current user is owner
	perms, err := h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardID)
	if err != nil || !perms.IsOwner {
		return c.String(http.StatusForbidden, i18n.T(c.Request().Context(), "error.only_owner_can_transfer"))
	}

	// Execute transfer via TransferService
	if err := h.transferService.TransferCardOwnership(c.Request().Context(), cardID, newOwner.ID, user.ID); err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf(`<div class="text-red-600 text-sm">%s</div>`, err.Error()))
	}

	// Audit log the transfer
	if h.db != nil {
		auditData := map[string]string{
			"action":          "transfer",
			"old_owner_id":    user.ID.String(),
			"old_owner_email": user.Email,
			"new_owner_id":    newOwner.ID.String(),
			"new_owner_email": newOwner.Email,
		}
		if err := audit.LogUpdateFromContext(c, h.db, "cards", cardID, auditData); err != nil {
			c.Logger().Errorf("Failed to log transfer: %v", err)
		}
	}

	// HTMX: Redirect to cards list (user no longer owns this card)
	c.Response().Header().Set("HX-Redirect", "/cards")
	return c.NoContent(http.StatusOK)
}
