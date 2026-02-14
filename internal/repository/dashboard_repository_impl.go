// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormDashboardRepository implements DashboardRepository using GORM.
type GormDashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a new dashboard repository.
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &GormDashboardRepository{db: db}
}

// GetStats returns aggregated resource counts and total balance for a user.
func (r *GormDashboardRepository) GetStats(ctx context.Context, userID uuid.UUID) (*DashboardStats, error) {
	stats := &DashboardStats{}

	type countResult struct {
		field string
		count int64
		err   error
	}

	countChan := make(chan countResult, 6)

	go func() {
		var count int64
		err := r.db.WithContext(ctx).Model(&models.Card{}).Where("user_id = ?", userID).Count(&count).Error
		countChan <- countResult{"cards_owned", count, err}
	}()

	go func() {
		var count int64
		err := r.db.WithContext(ctx).Model(&models.CardShare{}).Where("shared_with_id = ?", userID).Count(&count).Error
		countChan <- countResult{"cards_shared", count, err}
	}()

	go func() {
		var count int64
		err := r.db.WithContext(ctx).Model(&models.Voucher{}).Where("user_id = ?", userID).Count(&count).Error
		countChan <- countResult{"vouchers_owned", count, err}
	}()

	go func() {
		var count int64
		err := r.db.WithContext(ctx).Model(&models.VoucherShare{}).Where("shared_with_id = ?", userID).Count(&count).Error
		countChan <- countResult{"vouchers_shared", count, err}
	}()

	go func() {
		var count int64
		err := r.db.WithContext(ctx).Model(&models.GiftCard{}).Where("user_id = ?", userID).Count(&count).Error
		countChan <- countResult{"gift_cards_owned", count, err}
	}()

	go func() {
		var count int64
		err := r.db.WithContext(ctx).Model(&models.GiftCardShare{}).Where("shared_with_id = ?", userID).Count(&count).Error
		countChan <- countResult{"gift_cards_shared", count, err}
	}()

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
	// Include both owned gift cards AND shared gift cards
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(current_balance), 0)
		FROM gift_cards
		WHERE status = 'active'
		  AND deleted_at IS NULL
		  AND (
		    user_id = ?
		    OR id IN (
		      SELECT gift_card_id
		      FROM gift_card_shares
		      WHERE shared_with_id = ?
		        AND deleted_at IS NULL
		    )
		  )
	`, userID, userID).Scan(&stats.TotalBalance).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetFavoriteCounts returns favorite counts grouped by resource type for a user.
func (r *GormDashboardRepository) GetFavoriteCounts(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	type CountRow struct {
		ResourceType string
		Count        int64
	}

	var rows []CountRow
	err := r.db.WithContext(ctx).
		Model(&models.UserFavorite{}).
		Select("resource_type, COUNT(*) as count").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Group("resource_type").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, row := range rows {
		counts[row.ResourceType] = row.Count
	}

	return counts, nil
}

// LoadFavoriteCards returns the user's favorite cards up to the given limit.
func (r *GormDashboardRepository) LoadFavoriteCards(ctx context.Context, userID uuid.UUID, limit int) ([]models.Card, error) {
	var cards []models.Card
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Joins("INNER JOIN user_favorites ON user_favorites.resource_id = cards.id AND user_favorites.resource_type = 'card' AND user_favorites.user_id = ? AND user_favorites.deleted_at IS NULL", userID).
		Order("cards.created_at DESC").
		Limit(limit).
		Find(&cards).Error
	return cards, err
}

// LoadRecentCards returns the user's most recently created cards up to the given limit.
func (r *GormDashboardRepository) LoadRecentCards(ctx context.Context, userID uuid.UUID, limit int) ([]models.Card, error) {
	var cards []models.Card
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Merchant").
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&cards).Error
	return cards, err
}

// LoadFavoriteVouchers returns the user's favorite vouchers up to the given limit.
func (r *GormDashboardRepository) LoadFavoriteVouchers(ctx context.Context, userID uuid.UUID, limit int) ([]models.Voucher, error) {
	var vouchers []models.Voucher
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Joins("INNER JOIN user_favorites ON user_favorites.resource_id = vouchers.id AND user_favorites.resource_type = 'voucher' AND user_favorites.user_id = ? AND user_favorites.deleted_at IS NULL", userID).
		Order("vouchers.created_at DESC").
		Limit(limit).
		Find(&vouchers).Error
	return vouchers, err
}

// LoadRecentVouchers returns the user's most recently created vouchers up to the given limit.
func (r *GormDashboardRepository) LoadRecentVouchers(ctx context.Context, userID uuid.UUID, limit int) ([]models.Voucher, error) {
	var vouchers []models.Voucher
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Merchant").
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&vouchers).Error
	return vouchers, err
}

// LoadFavoriteGiftCards returns the user's favorite gift cards up to the given limit.
func (r *GormDashboardRepository) LoadFavoriteGiftCards(ctx context.Context, userID uuid.UUID, limit int) ([]models.GiftCard, error) {
	var giftCards []models.GiftCard
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date DESC").Limit(10)
		}).
		Joins("INNER JOIN user_favorites ON user_favorites.resource_id = gift_cards.id AND user_favorites.resource_type = 'gift_card' AND user_favorites.user_id = ? AND user_favorites.deleted_at IS NULL", userID).
		Order("gift_cards.created_at DESC").
		Limit(limit).
		Find(&giftCards).Error
	return giftCards, err
}

// LoadRecentGiftCards returns the user's most recently created gift cards up to the given limit.
func (r *GormDashboardRepository) LoadRecentGiftCards(ctx context.Context, userID uuid.UUID, limit int) ([]models.GiftCard, error) {
	var giftCards []models.GiftCard
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Merchant").
		Preload("User").
		Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date DESC").Limit(10)
		}).
		Order("created_at DESC").
		Limit(limit).
		Find(&giftCards).Error
	return giftCards, err
}
