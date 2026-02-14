// Package services contains business logic.
package services

import (
	"context"
	"fmt"
	"savvy/internal/models"
	"savvy/internal/repository"
	"time"

	"github.com/google/uuid"
)

// DashboardData contains all data needed for dashboard rendering
type DashboardData struct {
	Stats                *repository.DashboardStats
	RecentCards          []models.Card
	RecentVouchers       []models.Voucher
	RecentGiftCards      []models.GiftCard
	HasFavorites         bool
	HasCardFavorites     bool
	HasVoucherFavorites  bool
	HasGiftCardFavorites bool
}

// DashboardServiceInterface defines the interface for dashboard operations
type DashboardServiceInterface interface {
	GetDashboardData(ctx context.Context, userID uuid.UUID) (*DashboardData, error)
}

// DashboardService handles dashboard-related business logic
type DashboardService struct {
	dashboardRepo repository.DashboardRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(dashboardRepo repository.DashboardRepository) DashboardServiceInterface {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

// GetDashboardData fetches all dashboard data with optimized queries
func (s *DashboardService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*DashboardData, error) {
	stats, err := s.dashboardRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get dashboard stats: %w", err)
	}

	// Check all favorite types in ONE query
	favoriteCounts, err := s.dashboardRepo.GetFavoriteCounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get favorite counts: %w", err)
	}

	// Add timeout protection for goroutines to prevent hanging on slow DB queries
	loadCtx, loadCancel := context.WithTimeout(ctx, 10*time.Second)
	defer loadCancel()

	type itemsResult struct {
		cards      []models.Card
		vouchers   []models.Voucher
		giftCards  []models.GiftCard
		err        error
		resultType string
	}

	resultsChan := make(chan itemsResult, 3)

	go func() {
		cards, err := s.loadCards(loadCtx, userID, favoriteCounts["card"] > 0)
		resultsChan <- itemsResult{cards: cards, err: err, resultType: "cards"}
	}()

	go func() {
		vouchers, err := s.loadVouchers(loadCtx, userID, favoriteCounts["voucher"] > 0)
		resultsChan <- itemsResult{vouchers: vouchers, err: err, resultType: "vouchers"}
	}()

	go func() {
		giftCards, err := s.loadGiftCards(loadCtx, userID, favoriteCounts["gift_card"] > 0)
		resultsChan <- itemsResult{giftCards: giftCards, err: err, resultType: "gift_cards"}
	}()

	var recentCards []models.Card
	var recentVouchers []models.Voucher
	var recentGiftCards []models.GiftCard

	var firstErr error
	for range 3 {
		select {
		case result := <-resultsChan:
			if result.err != nil && firstErr == nil {
				firstErr = result.err
			}
			switch result.resultType {
			case "cards":
				recentCards = result.cards
			case "vouchers":
				recentVouchers = result.vouchers
			case "gift_cards":
				recentGiftCards = result.giftCards
			}
		case <-loadCtx.Done():
			return nil, fmt.Errorf("dashboard load timed out: %w", loadCtx.Err())
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}

	return &DashboardData{
		Stats:                stats,
		RecentCards:          recentCards,
		RecentVouchers:       recentVouchers,
		RecentGiftCards:      recentGiftCards,
		HasFavorites:         favoriteCounts["card"] > 0 || favoriteCounts["voucher"] > 0 || favoriteCounts["gift_card"] > 0,
		HasCardFavorites:     favoriteCounts["card"] > 0,
		HasVoucherFavorites:  favoriteCounts["voucher"] > 0,
		HasGiftCardFavorites: favoriteCounts["gift_card"] > 0,
	}, nil
}

func (s *DashboardService) loadCards(ctx context.Context, userID uuid.UUID, hasFavorites bool) ([]models.Card, error) {
	if hasFavorites {
		return s.dashboardRepo.LoadFavoriteCards(ctx, userID, 5)
	}
	return s.dashboardRepo.LoadRecentCards(ctx, userID, 5)
}

func (s *DashboardService) loadVouchers(ctx context.Context, userID uuid.UUID, hasFavorites bool) ([]models.Voucher, error) {
	if hasFavorites {
		return s.dashboardRepo.LoadFavoriteVouchers(ctx, userID, 5)
	}
	return s.dashboardRepo.LoadRecentVouchers(ctx, userID, 5)
}

func (s *DashboardService) loadGiftCards(ctx context.Context, userID uuid.UUID, hasFavorites bool) ([]models.GiftCard, error) {
	if hasFavorites {
		return s.dashboardRepo.LoadFavoriteGiftCards(ctx, userID, 5)
	}
	return s.dashboardRepo.LoadRecentGiftCards(ctx, userID, 5)
}
