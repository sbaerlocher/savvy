// Package repository defines data access interfaces.
package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormAuditLogRepository implements AuditLogRepository using GORM.
type GormAuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &GormAuditLogRepository{db: db}
}

// Create persists a new audit log entry.
func (r *GormAuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetFiltered returns audit logs matching the given filters with pagination.
func (r *GormAuditLogRepository) GetFiltered(ctx context.Context, filters AuditLogFilters) ([]models.AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Preload("User").
		Order("created_at DESC")

	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.ResourceType != "" {
		query = query.Where("resource_type = ?", filters.ResourceType)
	}
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	if filters.DateFrom != "" {
		query = query.Where("created_at >= ?", filters.DateFrom)
	}
	if filters.DateTo != "" {
		query = query.Where("created_at <= ?", filters.DateTo+" 23:59:59")
	}
	if filters.SearchQuery != "" {
		query = query.Where("resource_data::text ILIKE ?", "%"+filters.SearchQuery+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	var logs []models.AuditLog
	offset := (filters.Page - 1) * filters.PerPage
	if err := query.Limit(filters.PerPage).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("get audit logs: %w", err)
	}

	return logs, total, nil
}

// RestoreResource undeletes a soft-deleted resource by clearing its deleted_at timestamp.
func (r *GormAuditLogRepository) RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	var model any

	switch resourceType {
	case "cards":
		var card models.Card
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&card).Error; err != nil {
			return fmt.Errorf("find deleted card: %w", err)
		}
		if !card.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &card

	case "card_shares":
		var share models.CardShare
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&share).Error; err != nil {
			return fmt.Errorf("find deleted card share: %w", err)
		}
		if !share.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &share

	case "vouchers":
		var voucher models.Voucher
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&voucher).Error; err != nil {
			return fmt.Errorf("find deleted voucher: %w", err)
		}
		if !voucher.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &voucher

	case "voucher_shares":
		var share models.VoucherShare
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&share).Error; err != nil {
			return fmt.Errorf("find deleted voucher share: %w", err)
		}
		if !share.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &share

	case "gift_cards":
		var giftCard models.GiftCard
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&giftCard).Error; err != nil {
			return fmt.Errorf("find deleted gift card: %w", err)
		}
		if !giftCard.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &giftCard

	case "gift_card_shares":
		var share models.GiftCardShare
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&share).Error; err != nil {
			return fmt.Errorf("find deleted gift card share: %w", err)
		}
		if !share.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &share

	case "gift_card_transactions":
		var transaction models.GiftCardTransaction
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&transaction).Error; err != nil {
			return fmt.Errorf("find deleted gift card transaction: %w", err)
		}
		if !transaction.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &transaction

	case "merchants":
		var merchant models.Merchant
		if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&merchant).Error; err != nil {
			return fmt.Errorf("find deleted merchant: %w", err)
		}
		if !merchant.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &merchant

	default:
		return errors.New("unsupported resource type")
	}

	if err := r.db.WithContext(ctx).Unscoped().Model(model).Update("deleted_at", nil).Error; err != nil {
		return fmt.Errorf("restore %s %s: %w", resourceType, resourceID, err)
	}

	slog.Info("Resource restored", "resource_type", resourceType, "resource_id", resourceID)
	return nil
}
