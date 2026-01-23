// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/templates"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// FavoritesHandler handles favorite toggling operations
type FavoritesHandler struct {
	authzService services.AuthzServiceInterface
}

// NewFavoritesHandler creates a new favorites handler
func NewFavoritesHandler(authzService services.AuthzServiceInterface) *FavoritesHandler {
	return &FavoritesHandler{
		authzService: authzService,
	}
}

// ToggleCardFavorite toggles the favorite status of a card for the current user
func (h *FavoritesHandler) ToggleCardFavorite(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	cardID := c.Param("id")

	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid card ID"})
	}

	// Check authorization (owner or shared access can favorite)
	_, err = h.authzService.CheckCardAccess(c.Request().Context(), user.ID, cardUUID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	// Toggle favorite
	isFavorite := h.toggleFavorite(user.ID, "card", cardUUID)

	// Return updated button HTML for HTMX swap
	csrfToken := c.Get("csrf").(string)
	return templates.FavoriteButton(c.Request().Context(), cardID, isFavorite, csrfToken).Render(c.Request().Context(), c.Response().Writer)
}

// ToggleVoucherFavorite toggles the favorite status of a voucher for the current user
func (h *FavoritesHandler) ToggleVoucherFavorite(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	voucherID := c.Param("id")

	voucherUUID, err := uuid.Parse(voucherID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid voucher ID"})
	}

	// Check authorization (owner or shared access can favorite)
	_, err = h.authzService.CheckVoucherAccess(c.Request().Context(), user.ID, voucherUUID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	// Toggle favorite
	isFavorite := h.toggleFavorite(user.ID, "voucher", voucherUUID)

	// Return updated button HTML for HTMX swap
	csrfToken := c.Get("csrf").(string)
	return templates.FavoriteButton(c.Request().Context(), voucherID, isFavorite, csrfToken).Render(c.Request().Context(), c.Response().Writer)
}

// ToggleGiftCardFavorite toggles the favorite status of a gift card for the current user
func (h *FavoritesHandler) ToggleGiftCardFavorite(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	giftCardID := c.Param("id")

	giftCardUUID, err := uuid.Parse(giftCardID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid gift card ID"})
	}

	// Check authorization (owner or shared access can favorite)
	_, err = h.authzService.CheckGiftCardAccess(c.Request().Context(), user.ID, giftCardUUID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	// Toggle favorite
	isFavorite := h.toggleFavorite(user.ID, "gift_card", giftCardUUID)

	// Return updated button HTML for HTMX swap
	csrfToken := c.Get("csrf").(string)
	return templates.FavoriteButton(c.Request().Context(), giftCardID, isFavorite, csrfToken).Render(c.Request().Context(), c.Response().Writer)
}

// toggleFavorite is a helper function that handles the favorite toggle logic
// Returns true if the resource is now favorited, false if unfavorited
func (h *FavoritesHandler) toggleFavorite(userID uuid.UUID, resourceType string, resourceID uuid.UUID) bool {
	// Check if favorite already exists (including soft-deleted)
	var favorite models.UserFavorite
	err := database.DB.Unscoped().Where("user_id = ? AND resource_type = ? AND resource_id = ?",
		userID, resourceType, resourceID).First(&favorite).Error

	if err == nil {
		// Favorite exists - check if it's deleted
		if favorite.DeletedAt.Valid {
			// Restore soft-deleted favorite
			database.DB.Unscoped().Model(&favorite).Update("deleted_at", nil)
			return true
		}
		// Active favorite - soft delete it
		database.DB.Delete(&favorite)
		return false
	}

	// Favorite doesn't exist - create it
	favorite = models.UserFavorite{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
	database.DB.Create(&favorite)
	return true
}

// Legacy function names for backward compatibility with routes
// These will be removed once routes are updated

// CardsToggleFavorite - deprecated, use FavoritesHandler.ToggleCardFavorite
func CardsToggleFavorite(c echo.Context) error {
	// This is kept for backward compatibility but should not be used
	// Routes should be updated to use the new FavoritesHandler
	return c.String(http.StatusNotImplemented, "Use FavoritesHandler instead")
}

// VouchersToggleFavorite - deprecated, use FavoritesHandler.ToggleVoucherFavorite
func VouchersToggleFavorite(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Use FavoritesHandler instead")
}

// GiftCardsToggleFavorite - deprecated, use FavoritesHandler.ToggleGiftCardFavorite
func GiftCardsToggleFavorite(c echo.Context) error {
	return c.String(http.StatusNotImplemented, "Use FavoritesHandler instead")
}
