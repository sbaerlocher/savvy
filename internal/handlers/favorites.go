// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"context"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/templates"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// FavoritesHandler handles favorite toggling operations
type FavoritesHandler struct {
	authzService    services.AuthzServiceInterface
	favoriteService services.FavoriteServiceInterface
}

// NewFavoritesHandler creates a new favorites handler
func NewFavoritesHandler(authzService services.AuthzServiceInterface, favoriteService services.FavoriteServiceInterface) *FavoritesHandler {
	return &FavoritesHandler{
		authzService:    authzService,
		favoriteService: favoriteService,
	}
}

// toggleFavoriteHandler is a generic handler for toggling favorites with authorization check
func (h *FavoritesHandler) toggleFavoriteHandler(c echo.Context, resourceType string, checkAccess func(context.Context, uuid.UUID, uuid.UUID) (*services.ResourcePermissions, error)) error {
	user := c.Get("current_user").(*models.User)
	resourceID := c.Param("id")

	resourceUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	_, err = checkAccess(c.Request().Context(), user.ID, resourceUUID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
	}

	isFavorite := h.toggleFavorite(user.ID, resourceType, resourceUUID)

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}
	return templates.FavoriteButton(c.Request().Context(), resourceID, isFavorite, csrfToken).Render(c.Request().Context(), c.Response().Writer)
}

// ToggleCardFavorite toggles the favorite status of a card for the current user
func (h *FavoritesHandler) ToggleCardFavorite(c echo.Context) error {
	return h.toggleFavoriteHandler(c, "card", h.authzService.CheckCardAccess)
}

// ToggleVoucherFavorite toggles the favorite status of a voucher for the current user
func (h *FavoritesHandler) ToggleVoucherFavorite(c echo.Context) error {
	return h.toggleFavoriteHandler(c, "voucher", h.authzService.CheckVoucherAccess)
}

// ToggleGiftCardFavorite toggles the favorite status of a gift card for the current user
func (h *FavoritesHandler) ToggleGiftCardFavorite(c echo.Context) error {
	return h.toggleFavoriteHandler(c, "gift_card", h.authzService.CheckGiftCardAccess)
}

// toggleFavorite is a helper function that handles the favorite toggle logic.
// Returns true if the resource is now favorited, false if unfavorited.
func (h *FavoritesHandler) toggleFavorite(userID uuid.UUID, resourceType string, resourceID uuid.UUID) bool {
	ctx := context.Background()
	if err := h.favoriteService.ToggleFavorite(ctx, userID, resourceType, resourceID); err != nil {
		return false
	}

	isFavorite, err := h.favoriteService.IsFavorite(ctx, userID, resourceType, resourceID)
	if err != nil {
		return false
	}

	return isFavorite
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
