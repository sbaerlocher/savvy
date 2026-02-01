// Package repository contains data access implementations.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormMerchantRepository is a GORM implementation of MerchantRepository.
type GormMerchantRepository struct {
	db *gorm.DB
}

// NewMerchantRepository creates a new merchant repository.
func NewMerchantRepository(db *gorm.DB) MerchantRepository {
	return &GormMerchantRepository{db: db}
}

func (r *GormMerchantRepository) Create(ctx context.Context, merchant *models.Merchant) error {
	return r.db.WithContext(ctx).Create(merchant).Error
}

func (r *GormMerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.WithContext(ctx).First(&merchant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (r *GormMerchantRepository) GetByName(ctx context.Context, name string) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (r *GormMerchantRepository) GetAll(ctx context.Context) ([]models.Merchant, error) {
	var merchants []models.Merchant
	err := r.db.WithContext(ctx).Order("name ASC").Find(&merchants).Error
	return merchants, err
}

func (r *GormMerchantRepository) Search(ctx context.Context, query string) ([]models.Merchant, error) {
	var merchants []models.Merchant
	searchPattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE LOWER(?)", searchPattern).
		Order("name ASC").
		Find(&merchants).Error
	return merchants, err
}

func (r *GormMerchantRepository) Update(ctx context.Context, merchant *models.Merchant) error {
	return r.db.WithContext(ctx).Save(merchant).Error
}

func (r *GormMerchantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Merchant{}, "id = ?", id).Error
}

func (r *GormMerchantRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Merchant{}).Count(&count).Error
	return count, err
}
