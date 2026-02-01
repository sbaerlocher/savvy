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
	cardService     services.CardServiceInterface
	authzService    services.AuthzServiceInterface
	merchantService services.MerchantServiceInterface
	userService     services.UserServiceInterface
	favoriteService services.FavoriteServiceInterface
	shareService    services.ShareServiceInterface
	db              *gorm.DB
}

// NewHandler creates a new card handler.
func NewHandler(
	cardService services.CardServiceInterface,
	authzService services.AuthzServiceInterface,
	merchantService services.MerchantServiceInterface,
	userService services.UserServiceInterface,
	favoriteService services.FavoriteServiceInterface,
	shareService services.ShareServiceInterface,
	db *gorm.DB,
) *Handler {
	return &Handler{
		cardService:     cardService,
		authzService:    authzService,
		merchantService: merchantService,
		userService:     userService,
		favoriteService: favoriteService,
		shareService:    shareService,
		db:              db,
	}
}
