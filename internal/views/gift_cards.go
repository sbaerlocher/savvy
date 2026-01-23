// Package views contains view models for templates.
package views

import (
	"savvy/internal/models"
)

// GiftCardPermissions represents user permissions for a gift card
type GiftCardPermissions struct {
	CanEdit             bool
	CanDelete           bool
	CanEditTransactions bool
	IsFavorite          bool
}

// GiftCardShowView contains all data needed for gift_cards/show template
type GiftCardShowView struct {
	GiftCard        models.GiftCard
	Merchants       []models.Merchant
	Shares          []models.GiftCardShare
	User            *models.User
	Permissions     GiftCardPermissions
	IsImpersonating bool
}

// GiftCardEditView contains all data needed for gift_cards/edit template
type GiftCardEditView struct {
	GiftCard        models.GiftCard
	Merchants       []models.Merchant
	User            *models.User
	IsImpersonating bool
}

// GiftCardIndexView contains all data needed for gift_cards/index template
type GiftCardIndexView struct {
	GiftCards       []models.GiftCard
	User            *models.User
	IsImpersonating bool
}
