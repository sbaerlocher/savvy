// Package giftcards provides HTTP handlers for gift card management operations.
package giftcards

import (
	"fmt"
	"net/http"
	"savvy/internal/audit"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/templates"
	"savvy/internal/validation"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TransactionNew shows the inline form for adding a new transaction (HTMX)
func (h *Handler) TransactionNew(c echo.Context) error {
	giftCardID := c.Param("id")
	csrfToken := c.Get("csrf").(string)
	return templates.TransactionNewForm(c.Request().Context(), csrfToken, giftCardID).Render(c.Request().Context(), c.Response().Writer)
}

// TransactionCancel clears the transaction form (HTMX)
func (h *Handler) TransactionCancel(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

// TransactionCreate creates a new transaction for a gift card (HTMX)
func (h *Handler) TransactionCreate(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEditTransactions {
		return c.NoContent(http.StatusForbidden)
	}

	var giftCard models.GiftCard
	if err := database.DB.Where("id = ?", giftCardID).First(&giftCard).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	// Parse and validate amount
	amountStr := c.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.Logger().Errorf("Amount parse failed: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	// Validate amount is positive (expenses are positive values)
	if amount <= 0 {
		c.Logger().Errorf("Amount must be positive, got: %f", amount)
		csrfToken := c.Get("csrf").(string)
		return templates.TransactionNewFormWithError(c.Request().Context(), csrfToken, giftCardID.String(), "Der Betrag muss positiv sein").Render(c.Request().Context(), c.Response().Writer)
	}

	// Check if sufficient balance exists
	var giftCardWithTransactions models.GiftCard
	database.DB.First(&giftCardWithTransactions, giftCardID)
	currentBalance := giftCardWithTransactions.CurrentBalance

	if amount > currentBalance {
		c.Logger().Warnf("Insufficient funds: amount=%f, balance=%f", amount, currentBalance)
		csrfToken := c.Get("csrf").(string)
		errorMsg := fmt.Sprintf("Nicht genügend Guthaben. Verfügbar: %.2f CHF", currentBalance)
		return templates.TransactionNewFormWithError(c.Request().Context(), csrfToken, giftCardID.String(), errorMsg).Render(c.Request().Context(), c.Response().Writer)
	}

	// Parse transaction date
	transactionDateStr := c.FormValue("transaction_date")
	transactionDate, err := validation.ParseAndValidateDate(transactionDateStr, true) // allow past transactions
	if err != nil {
		c.Logger().Errorf("Transaction date validation failed: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}
	// Set to noon
	transactionDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 12, 0, 0, 0, time.UTC)

	transaction := models.GiftCardTransaction{
		GiftCardID:      giftCard.ID,
		Amount:          amount,
		Description:     c.FormValue("description"),
		TransactionDate: transactionDate,
		CreatedByUserID: &user.ID, // Track who created this transaction
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		// Check if error is from balance constraint trigger
		if strings.Contains(err.Error(), "Insufficient balance") ||
			strings.Contains(err.Error(), "check_gift_card_balance") {
			c.Logger().Warnf("Database balance check failed (race condition prevented): %v", err)
			// The application-level check should have caught this, but database trigger
			// prevented race condition. Redirect with error message.
			c.Response().Header().Set("HX-Redirect", "/gift-cards/"+giftCard.ID.String()+"?error=insufficient_balance")
			return c.NoContent(http.StatusBadRequest)
		}
		// Generic database error
		c.Logger().Errorf("Failed to create transaction: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// Return empty content and trigger page reload via HX-Redirect
	c.Response().Header().Set("HX-Redirect", "/gift-cards/"+giftCard.ID.String())
	return c.NoContent(http.StatusOK)
}

// TransactionDelete deletes a transaction (HTMX)
func (h *Handler) TransactionDelete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	transactionID, err := uuid.Parse(c.Param("transaction_id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// Check authorization
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEditTransactions {
		return c.NoContent(http.StatusForbidden)
	}

	var giftCard models.GiftCard
	if err := database.DB.Where("id = ?", giftCardID).First(&giftCard).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	// Delete transaction
	var transaction models.GiftCardTransaction
	if err := database.DB.Where("id = ? AND gift_card_id = ?", transactionID, giftCardID).First(&transaction).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	// Add user context for audit logging (automatic hook will create audit log)
	ctx := audit.AddUserIDToContext(c.Request().Context(), user.ID)
	if err := database.DB.WithContext(ctx).Delete(&transaction).Error; err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	// Trigger page reload
	c.Response().Header().Set("HX-Redirect", "/gift-cards/"+giftCard.ID.String())
	return c.NoContent(http.StatusOK)
}
