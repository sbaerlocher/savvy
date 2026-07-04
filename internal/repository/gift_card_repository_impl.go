// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormGiftCardRepository implements GiftCardRepository using GORM.
type GormGiftCardRepository struct {
	db *gorm.DB
}

// NewGiftCardRepository creates a new gift card repository.
func NewGiftCardRepository(db *gorm.DB) GiftCardRepository {
	return &GormGiftCardRepository{db: db}
}

// Create creates a new gift card in the database.
func (r *GormGiftCardRepository) Create(ctx context.Context, giftCard *models.GiftCard) error {
	return r.db.WithContext(ctx).Create(giftCard).Error
}

// GetByID retrieves a gift card by its ID with optional preloads.
func (r *GormGiftCardRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.GiftCard, error) {
	var giftCard models.GiftCard
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.First(&giftCard, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &giftCard, nil
}

// GetByUserID retrieves all gift cards owned by a specific user.
func (r *GormGiftCardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	var giftCards []models.GiftCard
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date DESC")
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&giftCards).Error

	return giftCards, err
}

// GetSharedWithUser retrieves all gift cards shared with a specific user.
func (r *GormGiftCardRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	var giftCards []models.GiftCard
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date DESC")
		}).
		Joins("INNER JOIN gift_card_shares ON gift_card_shares.gift_card_id = gift_cards.id").
		Where("gift_card_shares.shared_with_id = ? AND gift_card_shares.deleted_at IS NULL", userID).
		Order("gift_cards.created_at DESC").
		Find(&giftCards).Error

	return giftCards, err
}

// Update updates an existing gift card in the database.
func (r *GormGiftCardRepository) Update(ctx context.Context, giftCard *models.GiftCard) error {
	return r.db.WithContext(ctx).Save(giftCard).Error
}

// Delete soft-deletes a gift card from the database.
func (r *GormGiftCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.GiftCard{}, "id = ?", id).Error
}

// Count returns the total number of gift cards owned by a user.
func (r *GormGiftCardRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.GiftCard{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

// GetTotalBalance calculates the total balance across all gift cards for a user.
func (r *GormGiftCardRepository) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	giftCards, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	var totalBalance float64
	for _, gc := range giftCards {
		totalBalance += gc.CurrentBalance
	}

	return totalBalance, nil
}

// CreateTransaction creates a new gift card transaction in the database.
func (r *GormGiftCardRepository) CreateTransaction(ctx context.Context, transaction *models.GiftCardTransaction) error {
	// Serialize concurrent transactions on the same gift card. The
	// check_gift_card_balance BEFORE-trigger guards against overdraw, but it
	// reads SUM(amount) without locking the card row — two concurrent spends
	// in separate DB transactions both read the old balance, both pass the
	// check, and both commit → negative balance (TOCTOU). Taking a row lock
	// on the gift card first forces the second transaction to wait until the
	// first commits, so its trigger sees the up-to-date sum and rejects an
	// overdraw. Insert happens in the same DB transaction as the lock.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.GiftCard
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").
			First(&locked, "id = ?", transaction.GiftCardID).Error; err != nil {
			return err
		}
		return tx.Create(transaction).Error
	})
}

// GetTransaction retrieves a specific transaction for a gift card.
func (r *GormGiftCardRepository) GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error) {
	var transaction models.GiftCardTransaction
	err := r.db.WithContext(ctx).
		Where("id = ? AND gift_card_id = ?", transactionID, giftCardID).
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// DeleteTransaction soft-deletes a gift card transaction.
func (r *GormGiftCardRepository) DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.GiftCardTransaction{}, "id = ?", transactionID).Error
}

// GetExpiringGiftCards retrieves gift cards expiring within the given number of days.
func (r *GormGiftCardRepository) GetExpiringGiftCards(ctx context.Context, withinDays int) ([]models.GiftCard, error) {
	// Use date-based comparison in UTC since dates are stored as end-of-day UTC
	// (T23:59:59Z). Timestamp-based comparison with NOW() caused timezone issues
	// where converting to local timezone shifted the date forward by one day.
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	windowEnd := startOfToday.AddDate(0, 0, withinDays+1)

	var giftCards []models.GiftCard
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Where("user_id IS NOT NULL").
		Where("expires_at IS NOT NULL").
		Where("expires_at >= ?", startOfToday).
		Where("expires_at < ?", windowEnd).
		Where("current_balance > 0").
		Find(&giftCards).Error
	return giftCards, err
}

// GetAllForUserPaginated retrieves all gift cards (owned + shared) with pagination.
func (r *GormGiftCardRepository) GetAllForUserPaginated(ctx context.Context, userID uuid.UUID, params PaginationParams) (*PaginatedResult[models.GiftCard], error) {
	params = NormalizePagination(params)

	// Build the base condition: owned OR shared with user
	baseQuery := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("User").
		Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date DESC")
		}).
		Where(
			r.db.Where("user_id = ?", userID).Or(
				"id IN (SELECT gift_card_id FROM gift_card_shares WHERE shared_with_id = ? AND deleted_at IS NULL)", userID,
			),
		)

	// Count total
	var total int64
	countQuery := r.db.WithContext(ctx).Model(&models.GiftCard{}).
		Where(
			r.db.Where("user_id = ?", userID).Or(
				"id IN (SELECT gift_card_id FROM gift_card_shares WHERE shared_with_id = ? AND deleted_at IS NULL)", userID,
			),
		)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Fetch paginated items
	var items []models.GiftCard
	if err := ApplyPagination(baseQuery, params).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &PaginatedResult[models.GiftCard]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: CalculateTotalPages(total, params.PerPage),
	}, nil
}

// FindByCardNumber finds a gift card by card number for a specific user.
func (r *GormGiftCardRepository) FindByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.GiftCard, error) {
	var giftCard models.GiftCard
	err := r.db.WithContext(ctx).
		Where("card_number = ? AND user_id = ?", cardNumber, userID).
		First(&giftCard).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No duplicate found
		}
		return nil, err
	}

	return &giftCard, nil
}

// FindDeletedByCardNumber finds a soft-deleted gift card by card number for a specific user.
func (r *GormGiftCardRepository) FindDeletedByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.GiftCard, error) {
	var giftCard models.GiftCard
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("card_number = ? AND user_id = ? AND deleted_at IS NOT NULL", cardNumber, userID).
		First(&giftCard).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &giftCard, nil
}

// RestoreByID clears deleted_at for a soft-deleted gift card owned by the user.
func (r *GormGiftCardRepository) RestoreByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&models.GiftCard{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil).Error
}

// Search searches gift cards by query (merchant name, card number, notes).
