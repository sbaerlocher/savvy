// Package api contains JSON API handlers for gift cards.
package api

//
//nolint:revive // "api" is a meaningful package name for API handlers

import (
	"log/slog"
	"net/http"
	"time"

	"savvy/internal/models"
	"savvy/internal/validation"

	"github.com/labstack/echo/v5"
)

// ==================== Transaction Endpoints ====================

// ListTransactions returns all transactions for a gift card
// GET /api/v1/gift-cards/:id/transactions
func (h *GiftCardsHandler) ListTransactions(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	// Check authorization
	_, err = h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have access to this gift card",
		})
	}

	// Get gift card with transactions
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Gift card not found",
		})
	}

	// Convert transactions to DTOs
	var transactions []GiftCardTransactionDTO
	if giftCard.Transactions != nil {
		transactions = ToTransactionDTOs(giftCard.Transactions)
	} else {
		transactions = []GiftCardTransactionDTO{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"transactions": transactions,
	})
}

// CreateTransaction creates a new transaction (updates balance via DB trigger)
// POST /api/v1/gift-cards/:id/transactions
func (h *GiftCardsHandler) CreateTransaction(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	// Check authorization - needs can_edit_transactions permission
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEditTransactions {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to edit transactions for this gift card",
		})
	}

	var req TransactionCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate amount bounds (negative for spending, positive for refunds)
	if err := validation.ValidateTransactionAmount(req.Amount); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_amount",
			Message: err.Error(),
		})
	}

	// Parse transaction date (defaults to now)
	transactionDate := time.Now()
	if req.TransactionDate != nil {
		txDate, err := time.Parse(time.RFC3339, *req.TransactionDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid transaction_date format (use ISO 8601/RFC3339)",
			})
		}
		transactionDate = txDate
	}

	// Create transaction
	transaction := &models.GiftCardTransaction{
		GiftCardID:      giftCardID,
		Amount:          req.Amount,
		Description:     stringOrDefault(req.Description, ""),
		TransactionDate: transactionDate,
	}

	if err := h.giftCardService.CreateTransaction(c.Request().Context(), transaction); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create transaction",
		})
	}

	// Return created transaction + updated balance
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to reload gift card after transaction", "gift_card_id", giftCardID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Transaction created but failed to reload gift card",
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message":         "Transaction created successfully",
		"transaction":     ToTransactionDTO(transaction),
		"current_balance": giftCard.CurrentBalance,
	})
}

// DeleteTransaction deletes a transaction (balance recalculated via DB trigger)
// DELETE /api/v1/gift-cards/:id/transactions/:transactionID
func (h *GiftCardsHandler) DeleteTransaction(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	giftCardID, err := parseResourceID(c, "gift card")
	if err != nil {
		return err
	}

	transactionID, err := parseUUIDParam(c, "transactionID", "invalid_transaction_id", "Invalid transaction ID")
	if err != nil {
		return err
	}

	// Check authorization - needs can_edit_transactions permission
	perms, err := h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardID)
	if err != nil || !perms.CanEditTransactions {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to edit transactions for this gift card",
		})
	}

	// Verify transaction belongs to this gift card
	transaction, err := h.giftCardService.GetTransaction(c.Request().Context(), transactionID, giftCardID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Transaction not found",
		})
	}

	if transaction.GiftCardID != giftCardID {
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Message: "Transaction does not belong to this gift card",
		})
	}

	if err := h.giftCardService.DeleteTransaction(c.Request().Context(), transactionID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to delete transaction",
		})
	}

	// Return updated balance
	giftCard, err := h.giftCardService.GetGiftCard(c.Request().Context(), giftCardID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to reload gift card after transaction deletion", "gift_card_id", giftCardID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Transaction deleted but failed to reload gift card",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":         "Transaction deleted successfully",
		"current_balance": giftCard.CurrentBalance,
	})
}
