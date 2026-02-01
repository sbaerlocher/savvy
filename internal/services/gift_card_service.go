// Package services contains business logic.
package services

import (
	"context"
	"errors"
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
	UpdateGiftCard(ctx context.Context, giftCard *models.GiftCard) error
	DeleteGiftCard(ctx context.Context, id uuid.UUID) error
	CountUserGiftCards(ctx context.Context, userID uuid.UUID) (int64, error)
	GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error)
	GetCurrentBalance(ctx context.Context, giftCardID uuid.UUID) (float64, error)
	CanUserAccessGiftCard(ctx context.Context, giftCardID, userID uuid.UUID) (bool, error)
	CreateTransaction(ctx context.Context, transaction *models.GiftCardTransaction) error
	GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error)
	DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error
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
	// Business logic: validate gift card
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

	return s.repo.Create(ctx, giftCard)
}

// GetGiftCard retrieves a gift card by ID.
func (s *GiftCardService) GetGiftCard(ctx context.Context, id uuid.UUID) (*models.GiftCard, error) {
	return s.repo.GetByID(ctx, id, "Merchant", "User", "Transactions")
}

// GetUserGiftCards retrieves all gift cards for a user (owned + shared).
func (s *GiftCardService) GetUserGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	// Get owned gift cards
	ownedGiftCards, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get shared gift cards
	sharedGiftCards, err := s.repo.GetSharedWithUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Combine and return
	return append(ownedGiftCards, sharedGiftCards...), nil
}

// UpdateGiftCard updates a gift card.
func (s *GiftCardService) UpdateGiftCard(ctx context.Context, giftCard *models.GiftCard) error {
	// Business logic: validate gift card
	if giftCard.MerchantName == "" {
		return errors.New("merchant name is required")
	}

	if giftCard.CardNumber == "" {
		return errors.New("card number is required")
	}

	if giftCard.InitialBalance <= 0 {
		return errors.New("initial balance must be positive")
	}

	return s.repo.Update(ctx, giftCard)
}

// DeleteGiftCard deletes a gift card.
func (s *GiftCardService) DeleteGiftCard(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// CountUserGiftCards counts gift cards for a user.
func (s *GiftCardService) CountUserGiftCards(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.Count(ctx, userID)
}

// GetTotalBalance calculates total balance for a user.
func (s *GiftCardService) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	return s.repo.GetTotalBalance(ctx, userID)
}

// GetCurrentBalance retrieves the current balance of a gift card.
func (s *GiftCardService) GetCurrentBalance(ctx context.Context, giftCardID uuid.UUID) (float64, error) {
	giftCard, err := s.GetGiftCard(ctx, giftCardID)
	if err != nil {
		return 0, err
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
	// Validate transaction
	if transaction.Amount <= 0 {
		return errors.New("transaction amount must be positive")
	}

	return s.repo.CreateTransaction(ctx, transaction)
}

// GetTransaction retrieves a transaction by ID, validating it belongs to the gift card.
func (s *GiftCardService) GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error) {
	return s.repo.GetTransaction(ctx, transactionID, giftCardID)
}

// DeleteTransaction deletes a transaction by ID.
func (s *GiftCardService) DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error {
	return s.repo.DeleteTransaction(ctx, transactionID)
}
