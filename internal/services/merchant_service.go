// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"savvy/internal/database"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrMerchantExists is returned when a merchant with the same name exists.
var ErrMerchantExists = errors.New("merchant with this name already exists")

// MerchantServiceInterface defines the interface for merchant business logic.
type MerchantServiceInterface interface {
	// CreateMerchant creates a new merchant.
	CreateMerchant(ctx context.Context, merchant *models.Merchant) error

	// GetOrCreateByName returns the merchant with this name, creating it or
	// restoring a soft-deleted one if needed.
	GetOrCreateByName(ctx context.Context, name string) (*models.Merchant, error)

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
		merchant.Color = models.DefaultMerchantColor
	}

	// Check if merchant with same name already exists
	existing, err := s.repo.GetByName(ctx, merchant.Name)
	if err == nil && existing != nil {
		return ErrMerchantExists
	}

	if err := s.repo.Create(ctx, merchant); err != nil {
		if database.IsDuplicateError(err) {
			return ErrMerchantExists
		}
		return fmt.Errorf("create merchant: %w", err)
	}

	slog.Info("Merchant created", "merchant_id", logsafe.UUID(merchant.ID), "name", logsafe.String(merchant.Name))
	return nil
}

// GetOrCreateByName returns the active merchant with this name or creates it.
// If a concurrent request creates it first, the existing active row is returned.
func (s *MerchantService) GetOrCreateByName(ctx context.Context, name string) (*models.Merchant, error) {
	if name == "" {
		return nil, errors.New("merchant name is required")
	}

	existing, err := s.repo.GetByName(ctx, name)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("get merchant by name %q: %w", name, err)
	}

	merchant := &models.Merchant{Name: name, Color: models.DefaultMerchantColor}
	if err := s.repo.Create(ctx, merchant); err != nil {
		if !database.IsDuplicateError(err) {
			return nil, fmt.Errorf("create merchant: %w", err)
		}
		existing, lookupErr := s.repo.GetByName(ctx, name)
		if lookupErr != nil {
			return nil, fmt.Errorf("get merchant after duplicate create %q: %w", name, lookupErr)
		}
		if existing == nil {
			return nil, fmt.Errorf("get merchant after duplicate create %q: %w", name, err)
		}
		return existing, nil
	}

	slog.Info("Merchant created", "merchant_id", logsafe.UUID(merchant.ID), "name", logsafe.String(merchant.Name))
	return merchant, nil
}

// GetMerchantByID retrieves a merchant by ID.
func (s *MerchantService) GetMerchantByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error) {
	merchant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get merchant %s: %w", id, err)
	}
	return merchant, nil
}

// GetMerchantByName retrieves a merchant by name.
func (s *MerchantService) GetMerchantByName(ctx context.Context, name string) (*models.Merchant, error) {
	merchant, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get merchant by name %q: %w", name, err)
	}
	return merchant, nil
}

// GetAllMerchants retrieves all merchants.
func (s *MerchantService) GetAllMerchants(ctx context.Context) ([]models.Merchant, error) {
	merchants, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all merchants: %w", err)
	}
	return merchants, nil
}

// SearchMerchants searches merchants by name.
func (s *MerchantService) SearchMerchants(ctx context.Context, query string) ([]models.Merchant, error) {
	if query == "" {
		return s.GetAllMerchants(ctx)
	}
	merchants, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search merchants: %w", err)
	}
	return merchants, nil
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

	// Check if another merchant with same name already exists
	existing, err := s.repo.GetByName(ctx, merchant.Name)
	if err == nil && existing != nil && existing.ID != merchant.ID {
		return ErrMerchantExists
	}

	if err := s.repo.Update(ctx, merchant); err != nil {
		if database.IsDuplicateError(err) {
			return ErrMerchantExists
		}
		return fmt.Errorf("update merchant %s: %w", merchant.ID, err)
	}

	slog.Info("Merchant updated", "merchant_id", logsafe.UUID(merchant.ID), "name", logsafe.String(merchant.Name))
	return nil
}

// DeleteMerchant deletes a merchant by ID.
func (s *MerchantService) DeleteMerchant(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete merchant %s: %w", id, err)
	}

	slog.Info("Merchant deleted", "merchant_id", id)
	return nil
}

// GetMerchantCount returns the total number of merchants.
func (s *MerchantService) GetMerchantCount(ctx context.Context) (int64, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count merchants: %w", err)
	}
	return count, nil
}
