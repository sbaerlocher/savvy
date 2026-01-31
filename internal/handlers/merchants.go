// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"fmt"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MerchantsIndex lists all merchants
func MerchantsIndex(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	var merchants []models.Merchant
	if err := database.DB.Order("name ASC").Find(&merchants).Error; err != nil {
		return err
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.MerchantsIndex(c.Request().Context(), csrfToken, merchants, user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}

// MerchantsNew shows the form to create a new merchant
func MerchantsNew(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return templates.MerchantsNew(c.Request().Context(), csrfToken, user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}

// MerchantsCreate creates a new merchant
func MerchantsCreate(c echo.Context) error {
	color := c.FormValue("color")
	if color == "" {
		color = "#0066CC"
	}

	name := c.FormValue("name")

	// Check for duplicate name BEFORE Create
	var existingMerchant models.Merchant
	if err := database.DB.Where("name = ?", name).First(&existingMerchant).Error; err == nil {
		// Merchant with this name already exists
		c.Logger().Warnf("Duplicate merchant name attempt: %s", name)
		return c.Redirect(http.StatusSeeOther, "/merchants/new?error=name_exists")
	}

	merchant := models.Merchant{
		Name:    name,
		LogoURL: c.FormValue("logo_url"),
		Website: c.FormValue("website"),
		Color:   color,
	}

	if err := database.DB.Create(&merchant).Error; err != nil {
		c.Logger().Errorf("Failed to create merchant: %v", err)
		return c.Redirect(http.StatusSeeOther, "/merchants/new?error=database_error")
	}

	return c.Redirect(http.StatusSeeOther, "/merchants")
}

// MerchantsEdit shows the form to edit a merchant
func MerchantsEdit(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	var merchant models.Merchant
	if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.MerchantsEdit(c.Request().Context(), csrfToken, merchant, user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}

// MerchantsUpdate updates a merchant
func MerchantsUpdate(c echo.Context) error {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	var merchant models.Merchant
	if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	merchant.Name = c.FormValue("name")
	merchant.LogoURL = c.FormValue("logo_url")
	merchant.Website = c.FormValue("website")
	merchant.Color = c.FormValue("color")

	if merchant.Color == "" {
		merchant.Color = "#0066CC"
	}

	if err := database.DB.Save(&merchant).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants/"+merchant.ID.String()+"/edit")
	}

	return c.Redirect(http.StatusSeeOther, "/merchants")
}

// MerchantsDelete deletes a merchant
func MerchantsDelete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	var merchant models.Merchant
	if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	// Add user context for audit logging
	ctx := audit.AddUserIDToContext(c.Request().Context(), user.ID)
	if err := database.DB.WithContext(ctx).Delete(&merchant).Error; err != nil {
		return c.Redirect(http.StatusSeeOther, "/merchants")
	}

	// Always use HX-Redirect header for consistent behavior
	c.Response().Header().Set("HX-Redirect", "/merchants")
	return c.Redirect(http.StatusSeeOther, "/merchants")
}

// MerchantsSearch returns merchants as HTML options for autocomplete
func MerchantsSearch(c echo.Context) error {
	query := c.QueryParam("q")

	// Sanitize query to prevent LIKE-pattern injection
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, "%", "\\%")
	query = strings.ReplaceAll(query, "_", "\\_")

	// Limit query length to prevent DoS
	if len(query) > 100 {
		return c.String(http.StatusBadRequest, "Query too long")
	}

	var merchants []models.Merchant
	if query != "" {
		database.DB.Where("name ILIKE ?", "%"+query+"%").Order("name ASC").Limit(10).Find(&merchants)
	} else {
		database.DB.Order("name ASC").Limit(20).Find(&merchants)
	}

	// Return HTML options for datalist
	html := ""
	for _, m := range merchants {
		html += fmt.Sprintf(`<option value="%s" data-id="%s">`, m.Name, m.ID.String())
	}

	return c.HTML(http.StatusOK, html)
}
