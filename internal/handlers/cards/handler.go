// Package cards contains HTTP handlers for card management.
package cards

import (
	"savvy/internal/services"

	"gorm.io/gorm"
)

const (
	newMerchantValue = "new"
)

// Handler contains card service and shared dependencies.
type Handler struct {
	cardService  services.CardServiceInterface
	authzService services.AuthzServiceInterface
	db           *gorm.DB
}

// NewHandler creates a new card handler.
func NewHandler(cardService services.CardServiceInterface, authzService services.AuthzServiceInterface, db *gorm.DB) *Handler {
	return &Handler{
		cardService:  cardService,
		authzService: authzService,
		db:           db,
	}
}
