// Package services contains business logic.
package services

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DashboardStats represents aggregated statistics for the dashboard
type DashboardStats struct {
	CardsOwned      int64
	CardsShared     int64
	VouchersOwned   int64
	VouchersShared  int64
	GiftCardsOwned  int64
	GiftCardsShared int64
	TotalBalance    float64
}

// DashboardData contains all data needed for dashboard rendering
type DashboardData struct {
	Stats                *DashboardStats
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
	db *gorm.DB
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(db *gorm.DB) DashboardServiceInterface {
	return &DashboardService{db: db}
}

// GetDashboardData fetches all dashboard data with optimized queries
func (s *DashboardService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*DashboardData, error) {
	stats, err := s.getStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check all favorite types in ONE query
	favoriteCounts, err := s.getFavoriteCounts(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Load favorites and recent items in parallel using goroutines
	type itemsResult struct {
		cards      []models.Card
		vouchers   []models.Voucher
		giftCards  []models.GiftCard
		err        error
		resultType string
	}

	resultsChan := make(chan itemsResult, 3)

	// Load cards
	go func() {
		cards, err := s.loadCards(ctx, userID, favoriteCounts["card"] > 0)
		resultsChan <- itemsResult{cards: cards, err: err, resultType: "cards"}
	}()

	// Load vouchers
	go func() {
		vouchers, err := s.loadVouchers(ctx, userID, favoriteCounts["voucher"] > 0)
		resultsChan <- itemsResult{vouchers: vouchers, err: err, resultType: "vouchers"}
	}()

	// Load gift cards
	go func() {
		giftCards, err := s.loadGiftCards(ctx, userID, favoriteCounts["gift_card"] > 0)
		resultsChan <- itemsResult{giftCards: giftCards, err: err, resultType: "gift_cards"}
	}()

	// Collect results
	var recentCards []models.Card
	var recentVouchers []models.Voucher
	var recentGiftCards []models.GiftCard

	for range 3 {
		result := <-resultsChan
		if result.err != nil {
			return nil, result.err
		}
		switch result.resultType {
		case "cards":
			recentCards = result.cards
		case "vouchers":
			recentVouchers = result.vouchers
		case "gift_cards":
			recentGiftCards = result.giftCards
		}
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

// getStats fetches all statistics with optimized queries
func (s *DashboardService) getStats(ctx context.Context, userID uuid.UUID) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Batch all COUNT queries in parallel using goroutines
	type countResult struct {
		field string
		count int64
		err   error
	}

	countChan := make(chan countResult, 6)

	// Cards owned
	go func() {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.Card{}).Where("user_id = ?", userID).Count(&count).Error
		countChan <- countResult{"cards_owned", count, err}
	}()

	// Cards shared
	go func() {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.CardShare{}).Where("shared_with_id = ?", userID).Count(&count).Error
		countChan <- countResult{"cards_shared", count, err}
	}()

	// Vouchers owned
	go func() {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.Voucher{}).Where("user_id = ?", userID).Count(&count).Error
		countChan <- countResult{"vouchers_owned", count, err}
	}()

	// Vouchers shared
	go func() {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.VoucherShare{}).Where("shared_with_id = ?", userID).Count(&count).Error
		countChan <- countResult{"vouchers_shared", count, err}
	}()

	// Gift cards owned
	go func() {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.GiftCard{}).Where("user_id = ?", userID).Count(&count).Error
		countChan <- countResult{"gift_cards_owned", count, err}
	}()

	// Gift cards shared
	go func() {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.GiftCardShare{}).Where("shared_with_id = ?", userID).Count(&count).Error
		countChan <- countResult{"gift_cards_shared", count, err}
	}()

	// Collect results
	for range 6 {
		result := <-countChan
		if result.err != nil {
			return nil, result.err
		}
		switch result.field {
		case "cards_owned":
			stats.CardsOwned = result.count
		case "cards_shared":
			stats.CardsShared = result.count
		case "vouchers_owned":
			stats.VouchersOwned = result.count
		case "vouchers_shared":
			stats.VouchersShared = result.count
		case "gift_cards_owned":
			stats.GiftCardsOwned = result.count
		case "gift_cards_shared":
			stats.GiftCardsShared = result.count
		}
	}

	// Calculate total balance using cached current_balance column
	err := s.db.WithContext(ctx).
		Model(&models.GiftCard{}).
		Where("user_id = ? AND status = ?", userID, "active").
		Select("COALESCE(SUM(current_balance), 0)").
		Scan(&stats.TotalBalance).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// getFavoriteCounts fetches favorite counts for all resource types in ONE query
func (s *DashboardService) getFavoriteCounts(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	type CountRow struct {
		ResourceType string
		Count        int64
	}

	var rows []CountRow
	err := s.db.WithContext(ctx).
		Model(&models.UserFavorite{}).
		Select("resource_type, COUNT(*) as count").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Group("resource_type").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	// Convert to map
	counts := make(map[string]int64)
	for _, row := range rows {
		counts[row.ResourceType] = row.Count
	}

	return counts, nil
}

// loadCards fetches recent or favorite cards (hasFavorites is pre-computed)
func (s *DashboardService) loadCards(ctx context.Context, userID uuid.UUID, hasFavorites bool) ([]models.Card, error) {
	var cards []models.Card
	var err error

	if hasFavorites {
		// Show ONLY favorites (using INNER JOIN to filter)
		err = s.db.WithContext(ctx).
			Preload("Merchant").
			Preload("User").
			Joins("INNER JOIN user_favorites ON user_favorites.resource_id = cards.id AND user_favorites.resource_type = 'card' AND user_favorites.user_id = ? AND user_favorites.deleted_at IS NULL", userID).
			Order("cards.created_at DESC").
			Limit(5).
			Find(&cards).Error
	} else {
		// Fallback: Show recent owned cards
		err = s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Preload("Merchant").
			Preload("User").
			Order("created_at DESC").
			Limit(5).
			Find(&cards).Error
	}

	return cards, err
}

// loadVouchers fetches recent or favorite vouchers (hasFavorites is pre-computed)
func (s *DashboardService) loadVouchers(ctx context.Context, userID uuid.UUID, hasFavorites bool) ([]models.Voucher, error) {
	var vouchers []models.Voucher
	var err error

	if hasFavorites {
		// Show ONLY favorites (using INNER JOIN to filter)
		err = s.db.WithContext(ctx).
			Preload("Merchant").
			Preload("User").
			Joins("INNER JOIN user_favorites ON user_favorites.resource_id = vouchers.id AND user_favorites.resource_type = 'voucher' AND user_favorites.user_id = ? AND user_favorites.deleted_at IS NULL", userID).
			Order("vouchers.created_at DESC").
			Limit(5).
			Find(&vouchers).Error
	} else {
		// Fallback: Show recent owned vouchers
		err = s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Preload("Merchant").
			Preload("User").
			Order("created_at DESC").
			Limit(5).
			Find(&vouchers).Error
	}

	return vouchers, err
}

// loadGiftCards fetches recent or favorite gift cards (hasFavorites is pre-computed)
func (s *DashboardService) loadGiftCards(ctx context.Context, userID uuid.UUID, hasFavorites bool) ([]models.GiftCard, error) {
	var giftCards []models.GiftCard
	var err error

	if hasFavorites {
		// Show ONLY favorites (using INNER JOIN to filter)
		err = s.db.WithContext(ctx).
			Preload("Merchant").
			Preload("User").
			Preload("Transactions", func(db *gorm.DB) *gorm.DB {
				return db.Order("transaction_date DESC").Limit(10)
			}).
			Joins("INNER JOIN user_favorites ON user_favorites.resource_id = gift_cards.id AND user_favorites.resource_type = 'gift_card' AND user_favorites.user_id = ? AND user_favorites.deleted_at IS NULL", userID).
			Order("gift_cards.created_at DESC").
			Limit(5).
			Find(&giftCards).Error
	} else {
		// Fallback: Show recent owned gift cards
		err = s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Preload("Merchant").
			Preload("User").
			Preload("Transactions", func(db *gorm.DB) *gorm.DB {
				return db.Order("transaction_date DESC").Limit(10)
			}).
			Order("created_at DESC").
			Limit(5).
			Find(&giftCards).Error
	}

	return giftCards, err
}
