// Package cards contains HTTP handlers for card management.
package cards

import (
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/services"

	"gorm.io/gorm"
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

// Helper methods

// getMerchants retrieves all merchants for dropdown.
func (h *Handler) getMerchants() ([]models.Merchant, error) {
	var merchants []models.Merchant
	err := database.DB.Order("name ASC").Find(&merchants).Error
	return merchants, err
}
