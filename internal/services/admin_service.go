// Package services contains business logic.
package services

import (
	"context"
	"errors"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogFilters represents filters for audit log queries
type AuditLogFilters struct {
	UserID       *uuid.UUID
	ResourceType string
	Action       string
	DateFrom     string
	DateTo       string
	SearchQuery  string
	Page         int
	PerPage      int
}

// AuditLogResult represents paginated audit log results
type AuditLogResult struct {
	Logs  []models.AuditLog
	Total int64
}

// AdminServiceInterface defines the interface for admin operations
type AdminServiceInterface interface {
	// User management
	GetAllUsers(ctx context.Context) ([]models.User, error)
	UpdateUserRole(ctx context.Context, userID uuid.UUID, newRole string) error
	CreateLocalUser(ctx context.Context, user *models.User) error

	// Audit log
	GetAuditLogs(ctx context.Context, filters AuditLogFilters) (*AuditLogResult, error)
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error

	// Resource restoration
	RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error
}

// AdminService implements AdminServiceInterface
type AdminService struct {
	db *gorm.DB
}

// NewAdminService creates a new admin service
func NewAdminService(db *gorm.DB) AdminServiceInterface {
	return &AdminService{
		db: db,
	}
}

// GetAllUsers retrieves all users ordered by creation date
func (s *AdminService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserRole updates a user's role
func (s *AdminService) UpdateUserRole(ctx context.Context, userID uuid.UUID, newRole string) error {
	if newRole != "user" && newRole != "admin" {
		return errors.New("invalid role")
	}

	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return err
	}

	user.Role = newRole
	return s.db.WithContext(ctx).Save(&user).Error
}

// CreateLocalUser creates a new local auth user
func (s *AdminService) CreateLocalUser(ctx context.Context, user *models.User) error {
	if user.AuthProvider != "local" {
		return errors.New("can only create local auth users")
	}
	return s.db.WithContext(ctx).Create(user).Error
}

// GetAuditLogs retrieves audit logs with filters and pagination
func (s *AdminService) GetAuditLogs(ctx context.Context, filters AuditLogFilters) (*AuditLogResult, error) {
	query := s.db.WithContext(ctx).Model(&models.AuditLog{}).
		Preload("User").
		Order("created_at DESC")

	// Apply filters
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

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated results
	var logs []models.AuditLog
	offset := (filters.Page - 1) * filters.PerPage
	if err := query.Limit(filters.PerPage).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}

	return &AuditLogResult{
		Logs:  logs,
		Total: total,
	}, nil
}

// CreateAuditLog creates a new audit log entry
func (s *AdminService) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return s.db.WithContext(ctx).Create(log).Error
}

// RestoreResource restores a soft-deleted resource
func (s *AdminService) RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	var model interface{}
	var deletedAt interface{}

	switch resourceType {
	case "cards":
		var card models.Card
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&card).Error; err != nil {
			return err
		}
		if !card.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &card
		deletedAt = nil

	case "card_shares":
		var share models.CardShare
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&share).Error; err != nil {
			return err
		}
		if !share.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &share
		deletedAt = nil

	case "vouchers":
		var voucher models.Voucher
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&voucher).Error; err != nil {
			return err
		}
		if !voucher.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &voucher
		deletedAt = nil

	case "voucher_shares":
		var share models.VoucherShare
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&share).Error; err != nil {
			return err
		}
		if !share.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &share
		deletedAt = nil

	case "gift_cards":
		var giftCard models.GiftCard
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&giftCard).Error; err != nil {
			return err
		}
		if !giftCard.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &giftCard
		deletedAt = nil

	case "gift_card_shares":
		var share models.GiftCardShare
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&share).Error; err != nil {
			return err
		}
		if !share.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &share
		deletedAt = nil

	case "gift_card_transactions":
		var transaction models.GiftCardTransaction
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&transaction).Error; err != nil {
			return err
		}
		if !transaction.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &transaction
		deletedAt = nil

	case "merchants":
		var merchant models.Merchant
		if err := s.db.WithContext(ctx).Unscoped().Where("id = ?", resourceID).First(&merchant).Error; err != nil {
			return err
		}
		if !merchant.DeletedAt.Valid {
			return errors.New("resource is not deleted")
		}
		model = &merchant
		deletedAt = nil

	default:
		return errors.New("unsupported resource type")
	}

	return s.db.WithContext(ctx).Unscoped().Model(model).Update("deleted_at", deletedAt).Error
}
