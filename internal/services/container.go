// Package services contains business logic.
package services

import (
	"savvy/internal/repository"

	"gorm.io/gorm"
)

// Container holds all service instances.
type Container struct {
	CardService      CardServiceInterface
	VoucherService   VoucherServiceInterface
	GiftCardService  GiftCardServiceInterface
	MerchantService  MerchantServiceInterface
	ShareService     ShareServiceInterface
	FavoriteService  FavoriteServiceInterface
	AuthzService     AuthzServiceInterface
	DashboardService DashboardServiceInterface
}

// NewContainer creates a new service container with all services initialized.
func NewContainer(db *gorm.DB) *Container {
	// Initialize repositories
	cardRepo := repository.NewCardRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	giftCardRepo := repository.NewGiftCardRepository(db)
	merchantRepo := repository.NewMerchantRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)

	// Initialize services
	return &Container{
		CardService:      NewCardService(cardRepo),
		VoucherService:   NewVoucherService(voucherRepo),
		GiftCardService:  NewGiftCardService(giftCardRepo),
		MerchantService:  NewMerchantService(merchantRepo),
		ShareService:     NewShareService(cardRepo, voucherRepo, giftCardRepo),
		FavoriteService:  NewFavoriteService(favoriteRepo, cardRepo, voucherRepo, giftCardRepo),
		AuthzService:     NewAuthzService(db),
		DashboardService: NewDashboardService(db),
	}
}
