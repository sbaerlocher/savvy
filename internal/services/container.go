// Package services contains business logic.
package services

import (
	"savvy/internal/email"
	"savvy/internal/repository"

	"gorm.io/gorm"
)

// Container holds all service instances.
type Container struct {
	CardService         CardServiceInterface
	VoucherService      VoucherServiceInterface
	GiftCardService     GiftCardServiceInterface
	MerchantService     MerchantServiceInterface
	UserService         UserServiceInterface
	ShareService        ShareServiceInterface
	FavoriteService     FavoriteServiceInterface
	AuthzService        AuthzServiceInterface
	DashboardService    DashboardServiceInterface
	AdminService        AdminServiceInterface
	TransferService     TransferServiceInterface
	NotificationService NotificationServiceInterface
	EmailTokenService   EmailTokenServiceInterface
	ExportService       ExportServiceInterface
	AccountService      AccountServiceInterface
	ImportService       ImportServiceInterface
	PushService         PushServiceInterface
	ReminderService     ReminderServiceInterface
	TOTPService         TOTPServiceInterface
	SessionService      SessionServiceInterface
}

// NewContainer creates a new service container with all services initialized.
func NewContainer(db *gorm.DB, emailService email.ServiceInterface) *Container {
	// Initialize repositories
	cardRepo := repository.NewCardRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	giftCardRepo := repository.NewGiftCardRepository(db)
	merchantRepo := repository.NewMerchantRepository(db)
	userRepo := repository.NewUserRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	cardShareRepo := repository.NewCardShareRepository(db)
	voucherShareRepo := repository.NewVoucherShareRepository(db)
	giftCardShareRepo := repository.NewGiftCardShareRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	transferRepo := repository.NewTransferRepository(db)

	// Initialize repositories that were added later
	emailTokenRepo := repository.NewEmailTokenRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Initialize notification service first (needed by ShareService and TransferService)
	notificationService := NewNotificationService(notificationRepo, userRepo)

	// Initialize user service (needed by ExportService)
	userService := NewUserService(userRepo)

	// Initialize services needed by multiple consumers
	cardService := NewCardService(cardRepo)
	voucherService := NewVoucherService(voucherRepo)
	giftCardService := NewGiftCardService(giftCardRepo)
	merchantService := NewMerchantService(merchantRepo)

	// Initialize services
	return &Container{
		CardService:     cardService,
		VoucherService:  voucherService,
		GiftCardService: giftCardService,
		MerchantService: merchantService,
		UserService:     userService,
		ShareService: NewShareService(
			db, cardRepo, voucherRepo, giftCardRepo,
			cardShareRepo, voucherShareRepo, giftCardShareRepo,
			userRepo, auditLogRepo, notificationService,
		),
		FavoriteService: NewFavoriteService(favoriteRepo, cardRepo, voucherRepo, giftCardRepo),
		AuthzService: NewAuthzService(
			cardRepo, voucherRepo, giftCardRepo,
			cardShareRepo, voucherShareRepo, giftCardShareRepo,
		),
		DashboardService: NewDashboardService(dashboardRepo),
		AdminService:     NewAdminService(userRepo, auditLogRepo),
		TransferService: NewTransferService(
			cardRepo, voucherRepo, giftCardRepo,
			userRepo, transferRepo, auditLogRepo, notificationService,
		),
		NotificationService: notificationService,
		EmailTokenService:   NewEmailTokenService(emailTokenRepo, userRepo),
		ExportService:       NewExportService(userService, cardRepo, voucherRepo, giftCardRepo, favoriteRepo),
		AccountService:      NewAccountService(db, userService, emailService),
		ImportService:       NewImportService(cardService, voucherService, giftCardService, merchantService),
		SessionService:      NewSessionService(sessionRepo),
	}
}
