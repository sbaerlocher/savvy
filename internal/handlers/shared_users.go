// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SharedUserResponse represents a user that has been shared with
type SharedUserResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// SharedUsersAutocomplete returns users the current user has previously shared any resource with
// This endpoint is used for autocomplete in share forms
// GET /api/shared-users?q=search_query
func SharedUsersAutocomplete(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	searchQuery := c.QueryParam("q")

	// Query to find all unique users the current user has shared with
	// across cards, vouchers, and gift cards
	var sharedUsers []models.User

	// Build a subquery that unions all share tables
	db := database.DB

	// Get unique user IDs from all share tables
	var userIDs []uuid.UUID

	// Card shares
	var cardSharedUserIDs []uuid.UUID
	db.Table("card_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN cards ON cards.id = card_shares.card_id").
		Where("cards.user_id = ?", user.ID).
		Pluck("shared_with_id", &cardSharedUserIDs)
	userIDs = append(userIDs, cardSharedUserIDs...)

	// Voucher shares
	var voucherSharedUserIDs []uuid.UUID
	db.Table("voucher_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN vouchers ON vouchers.id = voucher_shares.voucher_id").
		Where("vouchers.user_id = ?", user.ID).
		Pluck("shared_with_id", &voucherSharedUserIDs)
	userIDs = append(userIDs, voucherSharedUserIDs...)

	// Gift card shares
	var giftCardSharedUserIDs []uuid.UUID
	db.Table("gift_card_shares").
		Select("DISTINCT shared_with_id").
		Joins("JOIN gift_cards ON gift_cards.id = gift_card_shares.gift_card_id").
		Where("gift_cards.user_id = ?", user.ID).
		Pluck("shared_with_id", &giftCardSharedUserIDs)
	userIDs = append(userIDs, giftCardSharedUserIDs...)

	// Remove duplicates from userIDs
	uniqueUserIDs := make(map[uuid.UUID]bool)
	var filteredUserIDs []uuid.UUID
	for _, id := range userIDs {
		if !uniqueUserIDs[id] {
			uniqueUserIDs[id] = true
			filteredUserIDs = append(filteredUserIDs, id)
		}
	}

	// If no shared users, return empty array
	if len(filteredUserIDs) == 0 {
		return c.JSON(http.StatusOK, []SharedUserResponse{})
	}

	// Fetch users by IDs
	query := db.Where("id IN ?", filteredUserIDs)

	// Apply search filter if provided
	if searchQuery != "" {
		query = query.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			"%"+searchQuery+"%", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	if err := query.Find(&sharedUsers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch shared users",
		})
	}

	// Convert to response format
	response := make([]SharedUserResponse, len(sharedUsers))
	for i, u := range sharedUsers {
		response[i] = SharedUserResponse{
			ID:    u.ID,
			Name:  u.DisplayName(),
			Email: u.Email,
		}
	}

	return c.JSON(http.StatusOK, response)
}
