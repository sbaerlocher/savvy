// Package gift_cards contains HTTP request handlers for gift card management.
package gift_cards

import (
	"savvy/internal/services"

	"gorm.io/gorm"
)

type Handler struct {
	giftCardService services.GiftCardServiceInterface
	authzService    services.AuthzServiceInterface
	db              *gorm.DB
}

func NewHandler(giftCardService services.GiftCardServiceInterface, authzService services.AuthzServiceInterface, db *gorm.DB) *Handler {
	return &Handler{
		giftCardService: giftCardService,
		authzService:    authzService,
		db:              db,
	}
}
