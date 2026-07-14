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

// CardServiceInterface defines the interface for card business logic.
type CardServiceInterface interface {
	CreateCard(ctx context.Context, card *models.Card) error
	GetCard(ctx context.Context, id uuid.UUID) (*models.Card, error)
	GetUserCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error)
	GetUserCardsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) (*repository.PaginatedResult[models.Card], error)
	UpdateCard(ctx context.Context, card *models.Card) error
	DeleteCard(ctx context.Context, id uuid.UUID) error
	CountUserCards(ctx context.Context, userID uuid.UUID) (int64, error)
	CanUserAccessCard(ctx context.Context, cardID, userID uuid.UUID) (bool, error)
	CheckDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID, excludeID *uuid.UUID) (*models.Card, error)
	CheckSharedDuplicate(ctx context.Context, cardNumber string, merchantID *uuid.UUID, userID uuid.UUID) (*models.Card, error)
	FindDeletedDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error)
	RestoreCard(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Card, error)
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

	if err := s.repo.Create(ctx, card); err != nil {
		return fmt.Errorf("create card: %w", err)
	}

	slog.Info("Card created", "card_id", card.ID, "merchant", logsafe.String(card.MerchantName))
	return nil
}

// GetCard retrieves a card by ID.
func (s *CardService) GetCard(ctx context.Context, id uuid.UUID) (*models.Card, error) {
	card, err := s.repo.GetByID(ctx, id, "Merchant", "User")
	if err != nil {
		return nil, fmt.Errorf("get card %s: %w", id, err)
	}
	return card, nil
}

// GetUserCards retrieves all cards for a user (owned + shared).
func (s *CardService) GetUserCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	ownedCards, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get owned cards: %w", err)
	}

	sharedCards, err := s.repo.GetSharedWithUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get shared cards: %w", err)
	}

	return append(ownedCards, sharedCards...), nil
}

// GetUserCardsPaginated retrieves paginated cards for a user (owned + shared).
func (s *CardService) GetUserCardsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) (*repository.PaginatedResult[models.Card], error) {
	params := repository.PaginationParams{Page: page, PerPage: perPage}
	result, err := s.repo.GetAllForUserPaginated(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("get paginated cards: %w", err)
	}
	return result, nil
}

// UpdateCard updates a card.
func (s *CardService) UpdateCard(ctx context.Context, card *models.Card) error {
	// MerchantName validation: only required if MerchantID is set
	// (allows removing merchant by setting both to empty)
	if card.MerchantID != nil && card.MerchantName == "" {
		return errors.New("merchant name is required when merchant ID is set")
	}

	if card.CardNumber == "" {
		return errors.New("card number is required")
	}

	if err := s.repo.Update(ctx, card); err != nil {
		return fmt.Errorf("update card %s: %w", card.ID, err)
	}

	slog.Info("Card updated", "card_id", card.ID)
	return nil
}

// DeleteCard deletes a card.
func (s *CardService) DeleteCard(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete card %s: %w", id, err)
	}

	slog.Info("Card deleted", "card_id", id)
	return nil
}

// CountUserCards counts cards for a user.
func (s *CardService) CountUserCards(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.Count(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count user cards: %w", err)
	}
	return count, nil
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

// FindDeletedDuplicate returns a soft-deleted card with the same number owned by the user, or nil.
func (s *CardService) FindDeletedDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error) {
	return s.repo.FindDeletedByCardNumber(ctx, cardNumber, userID)
}

// RestoreCard clears deleted_at for the user's soft-deleted card and returns the restored card.
// Returns (nil, nil) when there is no restorable twin for this user (id unknown or not owned by user);
// (restoredCard, nil) on success; (nil, err) on real DB error.
// Zero-row-restore guard (approach a): after RestoreByID, GetByID is called; if the record is still
// not found (nothing was actually undeleted), we return (nil, nil) instead of an error.
func (s *CardService) RestoreCard(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Card, error) {
	if err := s.repo.RestoreByID(ctx, id, userID); err != nil {
		return nil, fmt.Errorf("restore card: %w", err)
	}
	restored, err := s.repo.GetByID(ctx, id, "Merchant", "User")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Nothing was restored (wrong id or not owned by this user) — signal 404.
			return nil, nil
		}
		return nil, fmt.Errorf("load restored card: %w", err)
	}
	// Guard against cross-user reads: RestoreByID is user-scoped and no-ops on a
	// foreign id, but GetByID fetches by id only. Without this check a user could
	// read another user's active card via the restore endpoint. Signal 404 instead.
	if restored.UserID == nil || *restored.UserID != userID {
		return nil, nil
	}
	return restored, nil
}

// CheckDuplicate checks if a card with the same card number already exists for the user.
// Returns the existing card if found, nil otherwise.
// excludeID is used during updates to exclude the card being updated.
func (s *CardService) CheckDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID, excludeID *uuid.UUID) (*models.Card, error) {
	existing, err := s.repo.FindByCardNumber(ctx, cardNumber, userID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}

	// No duplicate found
	if existing == nil {
		return nil, nil
	}

	// If excludeID is provided (update case), check if it's the same card
	if excludeID != nil && existing.ID == *excludeID {
		return nil, nil // Same card, not a duplicate
	}

	// Duplicate found
	return existing, nil
}

// CheckSharedDuplicate checks whether a card with the same number and merchant was
// already shared with the user (owned by someone else). Returns the shared card with
// its owner preloaded (as User) if found, nil otherwise.
func (s *CardService) CheckSharedDuplicate(ctx context.Context, cardNumber string, merchantID *uuid.UUID, userID uuid.UUID) (*models.Card, error) {
	shared, err := s.repo.FindSharedByCardNumber(ctx, cardNumber, merchantID, userID)
	if err != nil {
		return nil, fmt.Errorf("check shared duplicate: %w", err)
	}
	return shared, nil
}
