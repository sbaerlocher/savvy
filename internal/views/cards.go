// Package views contains view models for templates.
package views

import (
	"savvy/internal/models"
)

// CardPermissions represents user permissions for a card
type CardPermissions struct {
	CanEdit    bool
	CanDelete  bool
	IsFavorite bool
}

// CardShowView contains all data needed for cards/show template
type CardShowView struct {
	Card            models.Card
	Merchants       []models.Merchant
	Shares          []models.CardShare
	User            *models.User
	Permissions     CardPermissions
	IsImpersonating bool
}

// CardEditView contains all data needed for cards/edit template
type CardEditView struct {
	Card            models.Card
	Merchants       []models.Merchant
	User            *models.User
	IsImpersonating bool
}

// CardIndexView contains all data needed for cards/index template
type CardIndexView struct {
	Cards           []models.Card
	User            *models.User
	IsImpersonating bool
}
