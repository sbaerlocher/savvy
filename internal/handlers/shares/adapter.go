// Package shares provides unified share handler logic for Cards, Vouchers, and Gift Cards.
// This package eliminates 70% code duplication by abstracting common sharing operations
// into a single base handler with resource-specific adapters.
package shares

import (
	"context"
	"time"

	"github.com/google/uuid"
	"savvy/internal/models"
)

// ShareAdapter defines the interface for resource-specific share operations.
// Each resource type (Card, Voucher, GiftCard) implements this interface to provide
// custom behavior while sharing common handler logic.
//
// Design Pattern: Adapter Pattern
// - Base handler calls adapter methods for resource-specific operations
// - Adapters encapsulate differences between Cards, Vouchers, and Gift Cards
// - Enables 70% code reduction by moving common logic to base handler
type ShareAdapter interface {
	// Resource identification
	ResourceType() string // "cards", "vouchers", "gift_cards"
	ResourceName() string // "Card", "Voucher", "Gift Card" (for UI messages)

	// Ownership verification
	// Returns true if the given user owns the resource
	CheckOwnership(ctx context.Context, userID, resourceID uuid.UUID) (bool, error)

	// Share operations
	// ListShares returns all shares for a given resource (formatted for templates)
	ListShares(ctx context.Context, resourceID uuid.UUID) ([]ShareView, error)

	// CreateShare creates a new share with the given permissions
	CreateShare(ctx context.Context, req CreateShareRequest) error

	// UpdateShare updates share permissions (not supported for vouchers)
	UpdateShare(ctx context.Context, req UpdateShareRequest) error

	// DeleteShare removes a share
	DeleteShare(ctx context.Context, shareID uuid.UUID) error

	// Capability flags
	// SupportsEdit returns false for vouchers (read-only sharing)
	SupportsEdit() bool

	// HasTransactionPermission returns true only for gift cards
	HasTransactionPermission() bool
}

// ShareView represents a share for template rendering.
// This unified struct works for all resource types (Cards, Vouchers, Gift Cards).
type ShareView struct {
	ID                  uuid.UUID
	ResourceID          uuid.UUID
	SharedWith          *models.User // User who has access
	CanEdit             bool
	CanDelete           bool
	CanEditTransactions bool // Only populated for gift cards
	CreatedAt           time.Time
}

// CreateShareRequest encapsulates share creation parameters.
type CreateShareRequest struct {
	UserID              uuid.UUID // Owner creating the share
	ResourceID          uuid.UUID // Resource being shared
	SharedWithEmail     string    // Email of user to share with
	CanEdit             bool      // Permission: can edit metadata
	CanDelete           bool      // Permission: can delete resource
	CanEditTransactions bool      // Permission: can edit transactions (gift cards only)
}

// UpdateShareRequest encapsulates share update parameters.
type UpdateShareRequest struct {
	ShareID             uuid.UUID // Share being updated
	UserID              uuid.UUID // User requesting update (must be owner)
	ResourceID          uuid.UUID // Resource ID (for ownership verification)
	CanEdit             bool      // Updated permission
	CanDelete           bool      // Updated permission
	CanEditTransactions bool      // Updated permission (gift cards only)
}
