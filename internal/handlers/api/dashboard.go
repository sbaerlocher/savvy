// Package api contains JSON API handlers for the dashboard.
package api //nolint:revive // "api" is a meaningful package name for API handlers

import (
	"context"
	"errors"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/services"

	"github.com/labstack/echo/v5"
)

// DashboardHandler handles dashboard API endpoints.
type DashboardHandler struct {
	dashboardService services.DashboardServiceInterface
	favoriteService  services.FavoriteServiceInterface
}

// NewDashboardHandler creates a new dashboard API handler.
func NewDashboardHandler(
	dashboardService services.DashboardServiceInterface,
	favoriteService services.FavoriteServiceInterface,
) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		favoriteService:  favoriteService,
	}
}

// Get returns dashboard data (stats, recent items, favorites)
// GET /api/v1/dashboard
func (h *DashboardHandler) Get(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	// Get dashboard data from service
	data, err := h.dashboardService.GetDashboardData(c.Request().Context(), user.ID)
	if err != nil {
		// Ignore context cancelation (client closed connection)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}

		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to load dashboard data",
		})
	}

	// Get favorite IDs for recent items
	cardFavorites := make(map[string]bool)
	for _, card := range data.RecentCards {
		isFav, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", card.ID)
		if isFav {
			cardFavorites[card.ID.String()] = true
		}
	}

	voucherFavorites := make(map[string]bool)
	for _, voucher := range data.RecentVouchers {
		isFav, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "voucher", voucher.ID)
		if isFav {
			voucherFavorites[voucher.ID.String()] = true
		}
	}

	giftCardFavorites := make(map[string]bool)
	for _, giftCard := range data.RecentGiftCards {
		isFav, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "gift_card", giftCard.ID)
		if isFav {
			giftCardFavorites[giftCard.ID.String()] = true
		}
	}

	// Build stats
	stats := DashboardStats{
		CardsCount:     int(data.Stats.CardsOwned + data.Stats.CardsShared),
		VouchersCount:  int(data.Stats.VouchersOwned + data.Stats.VouchersShared),
		GiftCardsCount: int(data.Stats.GiftCardsOwned + data.Stats.GiftCardsShared),
		SharedCount:    int(data.Stats.CardsShared + data.Stats.VouchersShared + data.Stats.GiftCardsShared),
		TotalBalance:   data.Stats.TotalBalance,
		FavoriteCounts: make(map[string]int),
	}

	// Get actual favorite counts from FavoriteService
	if data.HasCardFavorites {
		favoriteCards, err := h.favoriteService.GetFavoriteCards(c.Request().Context(), user.ID)
		if err == nil {
			stats.FavoriteCounts["card"] = len(favoriteCards)
		}
	}
	if data.HasVoucherFavorites {
		favoriteVouchers, err := h.favoriteService.GetFavoriteVouchers(c.Request().Context(), user.ID)
		if err == nil {
			stats.FavoriteCounts["voucher"] = len(favoriteVouchers)
		}
	}
	if data.HasGiftCardFavorites {
		favoriteGiftCards, err := h.favoriteService.GetFavoriteGiftCards(c.Request().Context(), user.ID)
		if err == nil {
			stats.FavoriteCounts["gift_card"] = len(favoriteGiftCards)
		}
	}

	// Build response
	response := DashboardResponse{
		Stats:                stats,
		RecentCards:          ToCardDTOs(data.RecentCards, cardFavorites),
		RecentVouchers:       ToVoucherDTOs(data.RecentVouchers, voucherFavorites),
		RecentGiftCards:      ToGiftCardDTOs(data.RecentGiftCards, giftCardFavorites),
		HasFavorites:         data.HasFavorites,
		HasCardFavorites:     data.HasCardFavorites,
		HasVoucherFavorites:  data.HasVoucherFavorites,
		HasGiftCardFavorites: data.HasGiftCardFavorites,
	}

	return c.JSON(http.StatusOK, response)
}
