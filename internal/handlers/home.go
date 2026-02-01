// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"savvy/internal/models"
	"savvy/internal/services"
	"savvy/internal/templates"

	"github.com/labstack/echo/v4"
)

// dashboardService is injected via InitDashboardService
var dashboardService services.DashboardServiceInterface

// InitDashboardService initializes the dashboard service
func InitDashboardService(service services.DashboardServiceInterface) {
	dashboardService = service
}

// HomeIndex shows the home page with optimized queries
func HomeIndex(c echo.Context) error {
	user := c.Get("current_user").(*models.User)
	isImpersonating := c.Get("is_impersonating") != nil

	// Get all dashboard data with optimized queries
	data, err := dashboardService.GetDashboardData(c.Request().Context(), user.ID)
	if err != nil {
		c.Logger().Errorf("Failed to load dashboard data: %v", err)
		return c.String(500, "Internal Server Error")
	}

	return templates.Home(
		c.Request().Context(),
		user,
		isImpersonating,
		data.Stats.CardsOwned+data.Stats.CardsShared,
		data.Stats.VouchersOwned+data.Stats.VouchersShared,
		data.Stats.GiftCardsOwned+data.Stats.GiftCardsShared,
		data.Stats.TotalBalance,
		data.RecentCards,
		data.RecentVouchers,
		data.RecentGiftCards,
		data.HasFavorites,
		data.HasCardFavorites,
		data.HasVoucherFavorites,
		data.HasGiftCardFavorites,
	).Render(c.Request().Context(), c.Response().Writer)
}

// OfflinePage shows the offline fallback page (PWA)
func OfflinePage(c echo.Context) error {
	// Try to get user from session (optional - page works without auth)
	var user *models.User
	if userVal := c.Get("current_user"); userVal != nil {
		user = userVal.(*models.User)
	}

	isImpersonating := c.Get("is_impersonating") != nil

	return templates.Offline(c.Request().Context(), user, isImpersonating).Render(c.Request().Context(), c.Response().Writer)
}
