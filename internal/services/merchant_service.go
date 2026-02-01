// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

// MerchantServiceInterface defines the interface for merchant business logic.
type MerchantServiceInterface interface {
	// CreateMerchant creates a new merchant.
	CreateMerchant(ctx context.Context, merchant *models.Merchant) error

	// GetMerchantByID retrieves a merchant by ID.
	GetMerchantByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error)

	// GetMerchantByName retrieves a merchant by name.
	GetMerchantByName(ctx context.Context, name string) (*models.Merchant, error)

	// GetAllMerchants retrieves all merchants.
	GetAllMerchants(ctx context.Context) ([]models.Merchant, error)

	// SearchMerchants searches merchants by name.
	SearchMerchants(ctx context.Context, query string) ([]models.Merchant, error)

	// UpdateMerchant updates an existing merchant.
	UpdateMerchant(ctx context.Context, merchant *models.Merchant) error

	// DeleteMerchant deletes a merchant by ID.
	DeleteMerchant(ctx context.Context, id uuid.UUID) error

	// GetMerchantCount returns the total number of merchants.
	GetMerchantCount(ctx context.Context) (int64, error)
}

// MerchantService implements merchant business logic.
type MerchantService struct {
	repo repository.MerchantRepository
}

// NewMerchantService creates a new merchant service.
func NewMerchantService(repo repository.MerchantRepository) MerchantServiceInterface {
	return &MerchantService{repo: repo}
}

// CreateMerchant creates a new merchant with validation.
func (s *MerchantService) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	// Validate merchant
	if merchant.Name == "" {
		return errors.New("merchant name is required")
	}
	if merchant.Color == "" {
		merchant.Color = "#3B82F6" // Default blue color
	}

	return s.repo.Create(ctx, merchant)
}

// GetMerchantByID retrieves a merchant by ID.
func (s *MerchantService) GetMerchantByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error) {
	return s.repo.GetByID(ctx, id)
}

// GetMerchantByName retrieves a merchant by name.
func (s *MerchantService) GetMerchantByName(ctx context.Context, name string) (*models.Merchant, error) {
	return s.repo.GetByName(ctx, name)
}

// GetAllMerchants retrieves all merchants.
func (s *MerchantService) GetAllMerchants(ctx context.Context) ([]models.Merchant, error) {
	return s.repo.GetAll(ctx)
}

// SearchMerchants searches merchants by name.
func (s *MerchantService) SearchMerchants(ctx context.Context, query string) ([]models.Merchant, error) {
	if query == "" {
		return s.repo.GetAll(ctx)
	}
	return s.repo.Search(ctx, query)
}

// UpdateMerchant updates an existing merchant with validation.
func (s *MerchantService) UpdateMerchant(ctx context.Context, merchant *models.Merchant) error {
	// Validate merchant
	if merchant.Name == "" {
		return errors.New("merchant name is required")
	}
	if merchant.Color == "" {
		return errors.New("merchant color is required")
	}

	return s.repo.Update(ctx, merchant)
}

// DeleteMerchant deletes a merchant by ID.
func (s *MerchantService) DeleteMerchant(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// GetMerchantCount returns the total number of merchants.
func (s *MerchantService) GetMerchantCount(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}
