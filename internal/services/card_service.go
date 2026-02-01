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

// CardServiceInterface defines the interface for card business logic.
type CardServiceInterface interface {
	CreateCard(ctx context.Context, card *models.Card) error
	GetCard(ctx context.Context, id uuid.UUID) (*models.Card, error)
	GetUserCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error)
	UpdateCard(ctx context.Context, card *models.Card) error
	DeleteCard(ctx context.Context, id uuid.UUID) error
	CountUserCards(ctx context.Context, userID uuid.UUID) (int64, error)
	CanUserAccessCard(ctx context.Context, cardID, userID uuid.UUID) (bool, error)
}

// CardService implements CardServiceInterface.
type CardService struct {
	repo repository.CardRepository
}

// NewCardService creates a new card service.
func NewCardService(repo repository.CardRepository) CardServiceInterface {
	return &CardService{repo: repo}
}

// CreateCard creates a new card.
func (s *CardService) CreateCard(ctx context.Context, card *models.Card) error {
	if card.MerchantName == "" {
		return errors.New("merchant name is required")
	}

	if card.CardNumber == "" {
		return errors.New("card number is required")
	}

	return s.repo.Create(ctx, card)
}

// GetCard retrieves a card by ID.
func (s *CardService) GetCard(ctx context.Context, id uuid.UUID) (*models.Card, error) {
	return s.repo.GetByID(ctx, id, "Merchant", "User")
}

// GetUserCards retrieves all cards for a user (owned + shared).
func (s *CardService) GetUserCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	ownedCards, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sharedCards, err := s.repo.GetSharedWithUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return append(ownedCards, sharedCards...), nil
}

// UpdateCard updates a card.
func (s *CardService) UpdateCard(ctx context.Context, card *models.Card) error {
	if card.MerchantName == "" {
		return errors.New("merchant name is required")
	}

	if card.CardNumber == "" {
		return errors.New("card number is required")
	}

	return s.repo.Update(ctx, card)
}

// DeleteCard deletes a card.
func (s *CardService) DeleteCard(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// CountUserCards counts cards for a user.
func (s *CardService) CountUserCards(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.Count(ctx, userID)
}

// CanUserAccessCard checks if a user can access a card (owner or shared).
func (s *CardService) CanUserAccessCard(ctx context.Context, cardID, userID uuid.UUID) (bool, error) {
	card, err := s.GetCard(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// Check if user is owner
	if card.UserID != nil && *card.UserID == userID {
		return true, nil
	}

	// Check if shared (simplified - in real implementation check card_shares table)
	return false, nil
}
