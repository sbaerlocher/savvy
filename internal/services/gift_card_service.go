// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GiftCardServiceInterface defines the interface for gift card business logic.
type GiftCardServiceInterface interface {
	CreateGiftCard(ctx context.Context, giftCard *models.GiftCard) error
	GetGiftCard(ctx context.Context, id uuid.UUID) (*models.GiftCard, error)
	GetUserGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error)
	GetUserGiftCardsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) (*repository.PaginatedResult[models.GiftCard], error)
	UpdateGiftCard(ctx context.Context, giftCard *models.GiftCard) error
	DeleteGiftCard(ctx context.Context, id uuid.UUID) error
	CountUserGiftCards(ctx context.Context, userID uuid.UUID) (int64, error)
	GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error)
	GetCurrentBalance(ctx context.Context, giftCardID uuid.UUID) (float64, error)
	CanUserAccessGiftCard(ctx context.Context, giftCardID, userID uuid.UUID) (bool, error)
	CreateTransaction(ctx context.Context, transaction *models.GiftCardTransaction) error
	GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error)
	DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error
	CheckDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID, excludeID *uuid.UUID) (*models.GiftCard, error)
	FindDeletedDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.GiftCard, error)
	RestoreGiftCard(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.GiftCard, error)
}

// GiftCardService implements GiftCardServiceInterface.
type GiftCardService struct {
	repo repository.GiftCardRepository
}

// NewGiftCardService creates a new gift card service.
func NewGiftCardService(repo repository.GiftCardRepository) GiftCardServiceInterface {
	return &GiftCardService{repo: repo}
}

// CreateGiftCard creates a new gift card.
func (s *GiftCardService) CreateGiftCard(ctx context.Context, giftCard *models.GiftCard) error {
	if giftCard.MerchantName == "" {
		return errors.New("merchant name is required")
	}

	if giftCard.CardNumber == "" {
		return errors.New("card number is required")
	}

	if giftCard.InitialBalance <= 0 {
		return errors.New("initial balance must be positive")
	}

	if giftCard.Currency == "" {
		giftCard.Currency = "CHF"
	}

	if err := s.repo.Create(ctx, giftCard); err != nil {
		return fmt.Errorf("create gift card: %w", err)
	}

	slog.Info("Gift card created", "gift_card_id", logsafe.UUID(giftCard.ID), "merchant", logsafe.String(giftCard.MerchantName), "balance", giftCard.InitialBalance)
	return nil
}

// GetGiftCard retrieves a gift card by ID.
func (s *GiftCardService) GetGiftCard(ctx context.Context, id uuid.UUID) (*models.GiftCard, error) {
	giftCard, err := s.repo.GetByID(ctx, id, "Merchant", "User", "Transactions")
	if err != nil {
		return nil, fmt.Errorf("get gift card %s: %w", id, err)
	}
	return giftCard, nil
}

// GetUserGiftCards retrieves all gift cards for a user (owned + shared).
func (s *GiftCardService) GetUserGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	ownedGiftCards, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get owned gift cards: %w", err)
	}

	sharedGiftCards, err := s.repo.GetSharedWithUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get shared gift cards: %w", err)
	}

	return append(ownedGiftCards, sharedGiftCards...), nil
}

// GetUserGiftCardsPaginated retrieves paginated gift cards for a user (owned + shared).
func (s *GiftCardService) GetUserGiftCardsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) (*repository.PaginatedResult[models.GiftCard], error) {
	params := repository.PaginationParams{Page: page, PerPage: perPage}
	result, err := s.repo.GetAllForUserPaginated(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("get paginated gift cards: %w", err)
	}
	return result, nil
}

// UpdateGiftCard updates a gift card.
func (s *GiftCardService) UpdateGiftCard(ctx context.Context, giftCard *models.GiftCard) error {
	if giftCard.MerchantName == "" {
		return errors.New("merchant name is required")
	}

	if giftCard.CardNumber == "" {
		return errors.New("card number is required")
	}

	if giftCard.InitialBalance <= 0 {
		return errors.New("initial balance must be positive")
	}

	if err := s.repo.Update(ctx, giftCard); err != nil {
		return fmt.Errorf("update gift card %s: %w", giftCard.ID, err)
	}

	slog.Info("Gift card updated", "gift_card_id", giftCard.ID)
	return nil
}

// DeleteGiftCard deletes a gift card.
func (s *GiftCardService) DeleteGiftCard(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete gift card %s: %w", id, err)
	}

	slog.Info("Gift card deleted", "gift_card_id", id)
	return nil
}

// CountUserGiftCards counts gift cards for a user.
func (s *GiftCardService) CountUserGiftCards(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.Count(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count user gift cards: %w", err)
	}
	return count, nil
}

// GetTotalBalance calculates total balance for a user.
func (s *GiftCardService) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	balance, err := s.repo.GetTotalBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get total balance: %w", err)
	}
	return balance, nil
}

// GetCurrentBalance retrieves the current balance of a gift card.
func (s *GiftCardService) GetCurrentBalance(ctx context.Context, giftCardID uuid.UUID) (float64, error) {
	giftCard, err := s.GetGiftCard(ctx, giftCardID)
	if err != nil {
		return 0, fmt.Errorf("get current balance: %w", err)
	}

	return giftCard.CurrentBalance, nil
}

// CanUserAccessGiftCard checks if a user can access a gift card (owner or shared).
func (s *GiftCardService) CanUserAccessGiftCard(ctx context.Context, giftCardID, userID uuid.UUID) (bool, error) {
	giftCard, err := s.GetGiftCard(ctx, giftCardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// Check if user is owner
	if giftCard.UserID != nil && *giftCard.UserID == userID {
		return true, nil
	}

	// Check if shared (simplified - in real implementation check gift_card_shares table)
	return false, nil
}

// CreateTransaction creates a new transaction for a gift card.
func (s *GiftCardService) CreateTransaction(ctx context.Context, transaction *models.GiftCardTransaction) error {
	if transaction.Amount <= 0 {
		return errors.New("transaction amount must be positive")
	}

	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	slog.Info("Gift card transaction created", "transaction_id", logsafe.UUID(transaction.ID), "gift_card_id", logsafe.UUID(transaction.GiftCardID), "amount", transaction.Amount)
	return nil
}

// GetTransaction retrieves a transaction by ID, validating it belongs to the gift card.
func (s *GiftCardService) GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error) {
	tx, err := s.repo.GetTransaction(ctx, transactionID, giftCardID)
	if err != nil {
		return nil, fmt.Errorf("get transaction %s: %w", transactionID, err)
	}
	return tx, nil
}

// DeleteTransaction deletes a transaction by ID.
func (s *GiftCardService) DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error {
	if err := s.repo.DeleteTransaction(ctx, transactionID); err != nil {
		return fmt.Errorf("delete transaction %s: %w", transactionID, err)
	}

	slog.Info("Gift card transaction deleted", "transaction_id", transactionID)
	return nil
}

// FindDeletedDuplicate returns a soft-deleted gift card with the same number owned by the user, or nil.
func (s *GiftCardService) FindDeletedDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.GiftCard, error) {
	return s.repo.FindDeletedByCardNumber(ctx, cardNumber, userID)
}

// RestoreGiftCard clears deleted_at for the user's soft-deleted gift card and returns the restored gift card.
// Returns (nil, nil) when there is no restorable twin for this user (id unknown or not owned by user);
// (restoredGiftCard, nil) on success; (nil, err) on real DB error.
// Zero-row-restore guard: after RestoreByID, GetByID is called; if the record is still
// not found (nothing was actually undeleted), we return (nil, nil) instead of an error.
func (s *GiftCardService) RestoreGiftCard(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.GiftCard, error) {
	if err := s.repo.RestoreByID(ctx, id, userID); err != nil {
		return nil, fmt.Errorf("restore gift card: %w", err)
	}
	restored, err := s.repo.GetByID(ctx, id, "Merchant", "User", "Transactions")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Nothing was restored (wrong id or not owned by this user) — signal 404.
			return nil, nil
		}
		return nil, fmt.Errorf("load restored gift card: %w", err)
	}
	// Guard against cross-user reads: RestoreByID is user-scoped and no-ops on a
	// foreign id, but GetByID fetches by id only. Without this check a user could
	// read another user's active gift card via the restore endpoint. Signal 404 instead.
	if restored.UserID == nil || *restored.UserID != userID {
		return nil, nil
	}
	return restored, nil
}

// CheckDuplicate checks if a gift card with the same card number already exists for the user.
// Returns the existing gift card if found, nil otherwise.
// excludeID is used during updates to exclude the gift card being updated.
func (s *GiftCardService) CheckDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID, excludeID *uuid.UUID) (*models.GiftCard, error) {
	existing, err := s.repo.FindByCardNumber(ctx, cardNumber, userID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}

	// No duplicate found
	if existing == nil {
		return nil, nil
	}

	// If excludeID is provided (update case), check if it's the same gift card
	if excludeID != nil && existing.ID == *excludeID {
		return nil, nil // Same gift card, not a duplicate
	}

	// Duplicate found
	return existing, nil
}
