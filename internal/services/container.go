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
	UserService      UserServiceInterface
	ShareService     ShareServiceInterface
	FavoriteService  FavoriteServiceInterface
	AuthzService     AuthzServiceInterface
	DashboardService DashboardServiceInterface
	AdminService     AdminServiceInterface
	TransferService  TransferServiceInterface
}

// NewContainer creates a new service container with all services initialized.
func NewContainer(db *gorm.DB) *Container {
	// Initialize repositories
	cardRepo := repository.NewCardRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	giftCardRepo := repository.NewGiftCardRepository(db)
	merchantRepo := repository.NewMerchantRepository(db)
	userRepo := repository.NewUserRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)

	// Initialize services
	return &Container{
		CardService:      NewCardService(cardRepo),
		VoucherService:   NewVoucherService(voucherRepo),
		GiftCardService:  NewGiftCardService(giftCardRepo),
		MerchantService:  NewMerchantService(merchantRepo),
		UserService:      NewUserService(userRepo),
		ShareService:     NewShareService(cardRepo, voucherRepo, giftCardRepo, db),
		FavoriteService:  NewFavoriteService(favoriteRepo, cardRepo, voucherRepo, giftCardRepo),
		AuthzService:     NewAuthzService(db),
		DashboardService: NewDashboardService(db),
		AdminService:     NewAdminService(db),
		TransferService:  NewTransferService(db, cardRepo, voucherRepo, giftCardRepo),
	}
}
