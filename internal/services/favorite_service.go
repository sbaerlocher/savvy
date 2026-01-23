// Package services contains business logic.
package services

import (
	"context"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

// FavoriteServiceInterface defines the interface for favorite business logic.
type FavoriteServiceInterface interface {
	// ToggleFavorite toggles a favorite (create or soft-delete/restore).
	ToggleFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) error

	// GetUserFavorites retrieves all favorites for a user.
	GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]models.UserFavorite, error)

	// IsFavorite checks if a resource is favorited by user.
	IsFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (bool, error)

	// GetFavoriteCards retrieves all favorited cards for a user.
	GetFavoriteCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error)

	// GetFavoriteVouchers retrieves all favorited vouchers for a user.
	GetFavoriteVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error)

	// GetFavoriteGiftCards retrieves all favorited gift cards for a user.
	GetFavoriteGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error)
}

// FavoriteService implements favorite business logic.
type FavoriteService struct {
	repo         repository.FavoriteRepository
	cardRepo     repository.CardRepository
	voucherRepo  repository.VoucherRepository
	giftCardRepo repository.GiftCardRepository
}

// NewFavoriteService creates a new favorite service.
func NewFavoriteService(
	repo repository.FavoriteRepository,
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
) FavoriteServiceInterface {
	return &FavoriteService{
		repo:         repo,
		cardRepo:     cardRepo,
		voucherRepo:  voucherRepo,
		giftCardRepo: giftCardRepo,
	}
}

// ToggleFavorite toggles a favorite (create or soft-delete/restore).
func (s *FavoriteService) ToggleFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	// Check if favorite exists (including soft-deleted)
	existing, err := s.repo.GetByUserAndResource(ctx, userID, resourceType, resourceID)

	if err != nil {
		// Favorite doesn't exist, create new
		favorite := &models.UserFavorite{
			UserID:       userID,
			ResourceType: resourceType,
			ResourceID:   resourceID,
		}
		return s.repo.Create(ctx, favorite)
	}

	// Favorite exists
	if existing.DeletedAt.Valid {
		// Restore soft-deleted favorite
		return s.repo.Restore(ctx, existing)
	}

	// Soft-delete active favorite
	return s.repo.Delete(ctx, existing)
}

// GetUserFavorites retrieves all favorites for a user.
func (s *FavoriteService) GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]models.UserFavorite, error) {
	return s.repo.GetByUser(ctx, userID)
}

// IsFavorite checks if a resource is favorited by user.
func (s *FavoriteService) IsFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (bool, error) {
	return s.repo.IsFavorite(ctx, userID, resourceType, resourceID)
}

// GetFavoriteCards retrieves all favorited cards for a user.
func (s *FavoriteService) GetFavoriteCards(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	favorites, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var cards []models.Card
	for _, fav := range favorites {
		if fav.ResourceType == "card" {
			card, err := s.cardRepo.GetByID(ctx, fav.ResourceID, "Merchant", "User")
			if err == nil && card != nil {
				cards = append(cards, *card)
			}
		}
	}

	return cards, nil
}

// GetFavoriteVouchers retrieves all favorited vouchers for a user.
func (s *FavoriteService) GetFavoriteVouchers(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	favorites, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var vouchers []models.Voucher
	for _, fav := range favorites {
		if fav.ResourceType == "voucher" {
			voucher, err := s.voucherRepo.GetByID(ctx, fav.ResourceID, "Merchant", "User")
			if err == nil && voucher != nil {
				vouchers = append(vouchers, *voucher)
			}
		}
	}

	return vouchers, nil
}

// GetFavoriteGiftCards retrieves all favorited gift cards for a user.
func (s *FavoriteService) GetFavoriteGiftCards(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	favorites, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var giftCards []models.GiftCard
	for _, fav := range favorites {
		if fav.ResourceType == "gift_card" {
			giftCard, err := s.giftCardRepo.GetByID(ctx, fav.ResourceID, "Merchant", "User")
			if err == nil && giftCard != nil {
				giftCards = append(giftCards, *giftCard)
			}
		}
	}

	return giftCards, nil
}
