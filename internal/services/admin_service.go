// Package services contains business logic.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"savvy/internal/audit"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

// AuditLogResult represents paginated audit log results
type AuditLogResult struct {
	Logs  []models.AuditLog
	Total int64
}

// AdminServiceInterface defines the interface for admin operations
type AdminServiceInterface interface {
	// User management
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUserRole(ctx context.Context, userID uuid.UUID, newRole string) error
	UpdateUser(ctx context.Context, userID uuid.UUID, email, firstName, lastName, role string) error
	CreateLocalUser(ctx context.Context, user *models.User) error

	// Audit log
	GetAuditLogs(ctx context.Context, filters repository.AuditLogFilters) (*AuditLogResult, error)
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error

	// Resource restoration
	RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error

	// Impersonation
	ValidateImpersonation(ctx context.Context, adminID, targetUserID uuid.UUID) error
	StartImpersonation(ctx context.Context, adminID, targetUserID uuid.UUID, resourceData interface{}) error
	StopImpersonation(ctx context.Context, adminID, targetUserID uuid.UUID, resourceData interface{}) error
}

// AdminService implements AdminServiceInterface
type AdminService struct {
	userRepo     repository.UserRepository
	auditLogRepo repository.AuditLogRepository
}

// NewAdminService creates a new admin service
func NewAdminService(
	userRepo repository.UserRepository,
	auditLogRepo repository.AuditLogRepository,
) AdminServiceInterface {
	return &AdminService{
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
	}
}

// GetAllUsers retrieves all users ordered by creation date
func (s *AdminService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	return users, nil
}

// GetUserByID retrieves a single user by ID
func (s *AdminService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", userID, err)
	}
	return user, nil
}

// UpdateUserRole updates a user's role
func (s *AdminService) UpdateUserRole(ctx context.Context, userID uuid.UUID, newRole string) error {
	if newRole != "user" && newRole != "admin" {
		return errors.New("invalid role")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user for role update: %w", err)
	}

	user.Role = newRole
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user role: %w", err)
	}

	slog.Info("User role updated", "user_id", userID, "new_role", newRole)
	return nil
}

// UpdateUser updates a user's complete profile (email, name, role)
func (s *AdminService) UpdateUser(ctx context.Context, userID uuid.UUID, email, firstName, lastName, role string) error {
	if role != "user" && role != "admin" {
		return errors.New("invalid role")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user for update: %w", err)
	}

	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.Role = role

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user %s: %w", userID, err)
	}

	slog.Info("User updated by admin", "user_id", userID, "email", logsafe.String(email), "role", logsafe.String(role))
	return nil
}

// CreateLocalUser creates a new local auth user
func (s *AdminService) CreateLocalUser(ctx context.Context, user *models.User) error {
	if user.AuthProvider != "local" {
		return errors.New("can only create local auth users")
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create local user: %w", err)
	}

	slog.Info("Local user created by admin", "user_id", user.ID, "email", logsafe.String(user.Email))
	return nil
}

// GetAuditLogs retrieves audit logs with filters and pagination
func (s *AdminService) GetAuditLogs(ctx context.Context, filters repository.AuditLogFilters) (*AuditLogResult, error) {
	logs, total, err := s.auditLogRepo.GetFiltered(ctx, filters)
	if err != nil {
		return nil, err
	}
	return &AuditLogResult{Logs: logs, Total: total}, nil
}

// CreateAuditLog creates a new audit log entry
func (s *AdminService) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// RestoreResource restores a soft-deleted resource
func (s *AdminService) RestoreResource(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	return s.auditLogRepo.RestoreResource(ctx, resourceType, resourceID)
}

// ValidateImpersonation checks if impersonation is allowed
func (s *AdminService) ValidateImpersonation(ctx context.Context, adminID, targetUserID uuid.UUID) error {
	admin, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return errors.New("admin not found")
	}
	if !admin.IsAdmin() {
		return errors.New("only admins can impersonate")
	}

	if _, err := s.userRepo.GetByID(ctx, targetUserID); err != nil {
		return errors.New("target user not found")
	}

	if adminID == targetUserID {
		return errors.New("cannot impersonate yourself")
	}

	return nil
}

// StartImpersonation creates audit log for impersonation start
func (s *AdminService) StartImpersonation(ctx context.Context, adminID, targetUserID uuid.UUID, resourceData interface{}) error {
	ipAddress, userAgent := audit.ExtractAuditInfo(ctx)

	dataJSON, err := json.Marshal(resourceData)
	if err != nil {
		return fmt.Errorf("marshal impersonation data: %w", err)
	}

	auditLog := &models.AuditLog{
		UserID:       &adminID,
		Action:       "impersonate_start",
		ResourceType: "users",
		ResourceID:   targetUserID,
		ResourceData: string(dataJSON),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	slog.Info("Impersonation started", "admin_id", adminID, "target_user_id", targetUserID)
	return s.CreateAuditLog(ctx, auditLog)
}

// StopImpersonation creates audit log for impersonation stop
func (s *AdminService) StopImpersonation(ctx context.Context, adminID, targetUserID uuid.UUID, resourceData interface{}) error {
	ipAddress, userAgent := audit.ExtractAuditInfo(ctx)

	dataJSON, err := json.Marshal(resourceData)
	if err != nil {
		return fmt.Errorf("marshal impersonation data: %w", err)
	}

	auditLog := &models.AuditLog{
		UserID:       &adminID,
		Action:       "impersonate_stop",
		ResourceType: "users",
		ResourceID:   targetUserID,
		ResourceData: string(dataJSON),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	slog.Info("Impersonation stopped", "admin_id", adminID, "target_user_id", targetUserID)
	return s.CreateAuditLog(ctx, auditLog)
}
