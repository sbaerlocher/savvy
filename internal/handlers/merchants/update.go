// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Update updates a merchant
func (h *Handler) Update(c echo.Context) error {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	merchant, err := h.merchantService.GetMerchantByID(c.Request().Context(), merchantID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	merchant.Name = c.FormValue("name")
	merchant.LogoURL = c.FormValue("logo_url")
	merchant.Website = c.FormValue("website")
	merchant.Color = c.FormValue("color")

	if merchant.Color == "" {
		merchant.Color = "#0066CC"
	}

	if err := h.merchantService.UpdateMerchant(c.Request().Context(), merchant); err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants/"+merchant.ID.String()+"/edit")
	}

	return c.Redirect(http.StatusSeeOther, "/merchants")
}
