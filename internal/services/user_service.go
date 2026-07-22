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

// UserServiceInterface defines the interface for user business logic.
type UserServiceInterface interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
}

// UserService implements UserServiceInterface.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service.
func NewUserService(repo repository.UserRepository) UserServiceInterface {
	return &UserService{repo: repo}
}

// GetUserByID retrieves a user by ID.
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email (case-insensitive).
// Returns ErrUserNotFound if user doesn't exist.
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

// CreateUser creates a new user.
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	// Business logic: validate user
	if user.Email == "" {
		return errors.New("email is required")
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	slog.Info("User created", "user_id", logsafe.UUID(user.ID), "email", logsafe.String(user.Email))
	return nil
}

// UpdateUser updates a user.
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user %s: %w", user.ID, err)
	}
	return nil
}

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")
