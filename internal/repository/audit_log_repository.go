// Package repository defines data access interfaces.
package repository

import (
	"context"
	"savvy/internal/models"

	"github.com/google/uuid"
)

// AuditLogFilters represents filters for audit log queries.
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

// AuditLogRepository defines data access operations for audit logs and resource restoration.
type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	GetFiltered(ctx context.Context, filters AuditLogFilters) ([]models.AuditLog, int64, error)
	RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error
}
