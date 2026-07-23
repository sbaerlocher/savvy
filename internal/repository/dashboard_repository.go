// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// DashboardStats represents aggregated statistics for the dashboard.
type DashboardStats struct {
	CardsOwned      int64
	CardsShared     int64
	VouchersOwned   int64
	VouchersShared  int64
	GiftCardsOwned  int64
	GiftCardsShared int64
	TotalBalance    float64
}

// DashboardRepository defines data access operations for dashboard aggregations.
type DashboardRepository interface {
	GetStats(ctx context.Context, userID uuid.UUID) (*DashboardStats, error)
	GetFavoriteCounts(ctx context.Context, userID uuid.UUID) (map[string]int64, error)
	LoadFavoriteCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error)
	LoadRecentCards(ctx context.Context, userID uuid.UUID, limit int) ([]models.Card, error)
	LoadFavoriteVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error)
	LoadRecentVouchers(ctx context.Context, userID uuid.UUID, limit int) ([]models.Voucher, error)
	LoadFavoriteGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error)
	LoadRecentGiftCards(ctx context.Context, userID uuid.UUID, limit int) ([]models.GiftCard, error)
}
