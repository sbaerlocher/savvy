// Package giftcards contains HTTP request handlers for gift card management.
package giftcards

import (
	"savvy/internal/services"

	"gorm.io/gorm"
)

const (
	newMerchantValue = "new"
	trueStringValue  = "true"
)

// Handler handles HTTP requests for gift card operations.
type Handler struct {
	giftCardService services.GiftCardServiceInterface
	authzService    services.AuthzServiceInterface
	db              *gorm.DB
}

// NewHandler creates a new gift card handler with the provided services.
func NewHandler(giftCardService services.GiftCardServiceInterface, authzService services.AuthzServiceInterface, db *gorm.DB) *Handler {
	return &Handler{
		giftCardService: giftCardService,
		authzService:    authzService,
		db:              db,
	}
}
