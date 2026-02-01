// Package merchants contains HTTP request handlers for merchant operations.
package merchants

import (
	"net/http"
	"savvy/internal/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Create creates a new merchant
func (h *Handler) Create(c echo.Context) error {
	color := c.FormValue("color")
	if color == "" {
		color = "#0066CC"
	}

	name := c.FormValue("name")

	// Check for duplicate name BEFORE Create
	_, err := h.merchantService.GetMerchantByName(c.Request().Context(), name)
	if err == nil {
		// Merchant with this name already exists
		c.Logger().Warnf("Duplicate merchant name attempt: %s", name)
		return c.Redirect(http.StatusSeeOther, "/merchants/new?error=name_exists")
	} else if err != gorm.ErrRecordNotFound {
		// Unexpected error
		c.Logger().Errorf("Failed to check merchant name: %v", err)
		return c.Redirect(http.StatusSeeOther, "/merchants/new?error=database_error")
	}

	merchant := models.Merchant{
		Name:    name,
		LogoURL: c.FormValue("logo_url"),
		Website: c.FormValue("website"),
		Color:   color,
	}

	if err := h.merchantService.CreateMerchant(c.Request().Context(), &merchant); err != nil {
		c.Logger().Errorf("Failed to create merchant: %v", err)
		return c.Redirect(http.StatusSeeOther, "/merchants/new?error=database_error")
	}

	return c.Redirect(http.StatusSeeOther, "/merchants")
}
