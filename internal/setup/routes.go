// Package setup contains setup logic for route registration.
package setup

import (
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/debug"
	"savvy/internal/handlers"
	"savvy/internal/handlers/cards"
	"savvy/internal/handlers/giftcards"
	"savvy/internal/handlers/merchants"
	"savvy/internal/handlers/vouchers"
	"savvy/internal/middleware"
	"savvy/internal/services"

	"github.com/labstack/echo/v5"
)

// RouteConfig holds dependencies needed for route registration.
type RouteConfig struct {
	Echo             *echo.Echo
	Config           *config.Config
	ServiceContainer *services.Container
}

// RegisterRoutes registers all application routes.
func RegisterRoutes(rc *RouteConfig) {
	e := rc.Echo
	cfg := rc.Config
	serviceContainer := rc.ServiceContainer

	// Initialize handlers
	handlers.InitDashboardService(serviceContainer.DashboardService)

	cardHandler := cards.NewHandler(
		serviceContainer.CardService,
		serviceContainer.AuthzService,
		serviceContainer.MerchantService,
		serviceContainer.UserService,
		serviceContainer.FavoriteService,
		serviceContainer.ShareService,
		serviceContainer.TransferService,
		database.DB,
	)

	voucherHandler := vouchers.NewHandler(
		serviceContainer.VoucherService,
		serviceContainer.AuthzService,
		serviceContainer.MerchantService,
		serviceContainer.UserService,
		serviceContainer.FavoriteService,
		serviceContainer.ShareService,
		serviceContainer.TransferService,
		database.DB,
	)

	giftCardHandler := giftcards.NewHandler(
		serviceContainer.GiftCardService,
		serviceContainer.AuthzService,
		serviceContainer.MerchantService,
		serviceContainer.UserService,
		serviceContainer.FavoriteService,
		serviceContainer.ShareService,
		serviceContainer.TransferService,
		database.DB,
	)

	cardSharesHandler := handlers.NewCardSharesHandler(database.DB, serviceContainer.AuthzService, serviceContainer.UserService, serviceContainer.NotificationService)
	voucherSharesHandler := handlers.NewVoucherSharesHandler(database.DB, serviceContainer.AuthzService, serviceContainer.UserService, serviceContainer.NotificationService)
	giftCardSharesHandler := handlers.NewGiftCardSharesHandler(database.DB, serviceContainer.AuthzService, serviceContainer.UserService, serviceContainer.NotificationService)
	favoritesHandler := handlers.NewFavoritesHandler(serviceContainer.AuthzService, serviceContainer.FavoriteService)

	barcodeHandler := handlers.NewBarcodeHandler(
		serviceContainer.AuthzService,
		serviceContainer.CardService,
		serviceContainer.VoucherService,
		serviceContainer.GiftCardService,
	)

	merchantsHandler := merchants.NewHandler(serviceContainer.MerchantService)
	authHandler := handlers.NewAuthHandler(serviceContainer.UserService)
	oauthHandler := handlers.NewOAuthHandler(serviceContainer.UserService)
	sharedUsersHandler := handlers.NewSharedUsersHandler(serviceContainer.ShareService)
	notificationHandler := handlers.NewNotificationHandler(serviceContainer.NotificationService)
	adminHandler := handlers.NewAdminHandler(serviceContainer.AdminService, serviceContainer.UserService)

	// Rate limiter for auth endpoints (5 requests per second, burst of 10)
	authLimiter := middleware.NewIPRateLimiter(5, 10)

	// ========================================
	// Authentication Routes (Public)
	// ========================================
	auth := e.Group("/auth")
	auth.GET("/login", handlers.AuthLoginGet)
	auth.POST("/login", authHandler.LoginPost, middleware.RateLimitMiddleware(authLimiter), middleware.RequireLocalLoginEnabled(cfg))
	auth.GET("/register", handlers.AuthRegisterGet, middleware.RequireRegistrationEnabled(cfg))
	auth.POST("/register", authHandler.RegisterPost, middleware.RateLimitMiddleware(authLimiter), middleware.RequireRegistrationEnabled(cfg))
	auth.GET("/logout", handlers.AuthLogout)
	// OAuth routes (public)
	auth.GET("/oauth/login", handlers.OAuthLogin)
	auth.GET("/oauth/callback", oauthHandler.Callback)

	// ========================================
	// Protected Routes (Authentication Required)
	// ========================================
	protected := e.Group("")
	protected.Use(middleware.RequireAuth)

	// Dashboard & Home
	protected.GET("/", handlers.HomeIndex)

	// Barcode generation (secure token-based access)
	protected.GET("/barcode/:token", barcodeHandler.Generate)

	// HTMX autocomplete endpoint (returns HTML fragment)
	protected.GET("/api/shared-users", sharedUsersHandler.Autocomplete)

	// ========================================
	// Notifications
	// ========================================
	protected.GET("/notifications", notificationHandler.ShowNotifications)
	protected.GET("/api/notifications/count", notificationHandler.GetUnreadCount)
	protected.GET("/api/notifications/preview", notificationHandler.GetNotificationsDropdown)
	protected.POST("/notifications/:id/read", notificationHandler.MarkAsRead)
	protected.POST("/notifications/mark-all-read", notificationHandler.MarkAllAsRead)
	protected.DELETE("/notifications/:id", notificationHandler.DeleteNotification)

	// ========================================
	// Merchants Routes (Read-Only for All Users)
	// ========================================
	merchantsGroup := protected.Group("/merchants")
	merchantsGroup.GET("", merchantsHandler.Index)
	merchantsGroup.GET("/search", merchantsHandler.Search)
	merchantsGroup.GET("/:id", merchantsHandler.Show)

	// ========================================
	// Merchants CRUD Routes (Admin or Impersonation)
	// ========================================
	merchantsCRUD := protected.Group("/merchants")
	merchantsCRUD.Use(middleware.RequireImpersonationOrAdmin)
	merchantsCRUD.GET("/new", merchantsHandler.New)
	merchantsCRUD.POST("", merchantsHandler.Create)
	merchantsCRUD.GET("/:id/edit", merchantsHandler.Edit)
	merchantsCRUD.POST("/:id", merchantsHandler.Update)
	merchantsCRUD.DELETE("/:id", merchantsHandler.Delete)

	// ========================================
	// Cards Resource
	// ========================================
	registerCardsRoutes(protected, cfg, cardHandler, cardSharesHandler, favoritesHandler)

	// ========================================
	// Vouchers Resource
	// ========================================
	registerVouchersRoutes(protected, cfg, voucherHandler, voucherSharesHandler, favoritesHandler)

	// ========================================
	// Gift Cards Resource
	// ========================================
	registerGiftCardsRoutes(protected, cfg, giftCardHandler, giftCardSharesHandler, favoritesHandler)

	// ========================================
	// Impersonation Management
	// ========================================
	protected.GET("/admin/stop-impersonate", handlers.AdminStopImpersonate)

	// ========================================
	// Admin Routes
	// ========================================
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAdmin)
	admin.GET("/users", adminHandler.UsersIndex)
	admin.GET("/users/create", adminHandler.CreateUserGet)
	admin.POST("/users/create", adminHandler.CreateUserPost)
	admin.POST("/users/:id/role", adminHandler.UpdateUserRole)
	admin.GET("/audit-log", adminHandler.AuditLogIndex)
	admin.POST("/audit-log/restore", adminHandler.RestoreResource)
	admin.GET("/impersonate/:id", authHandler.Impersonate)

	// ========================================
	// Development Debug Tools
	// ========================================
	if !cfg.IsProduction() {
		debug.PrintRoutes(e)
	}
}

// registerCardsRoutes registers all card-related routes.
func registerCardsRoutes(
	protected *echo.Group,
	cfg *config.Config,
	cardHandler *cards.Handler,
	cardSharesHandler *handlers.CardSharesHandler,
	favoritesHandler *handlers.FavoritesHandler,
) {
	cardsGroup := protected.Group("/cards")
	cardsGroup.Use(middleware.RequireCardsEnabled(cfg))
	cardsGroup.GET("", cardHandler.Index)
	cardsGroup.GET("/new", cardHandler.New)
	cardsGroup.POST("", cardHandler.Create)
	cardsGroup.GET("/:id", cardHandler.Show)
	cardsGroup.GET("/:id/edit", cardHandler.Edit)
	cardsGroup.POST("/:id", cardHandler.Update)
	cardsGroup.DELETE("/:id", cardHandler.Delete)
	// Inline editing (HTMX-powered)
	cardsGroup.GET("/:id/edit-inline", cardHandler.EditInline)
	cardsGroup.GET("/:id/cancel-edit", cardHandler.CancelEdit)
	cardsGroup.PATCH("/:id", cardHandler.UpdateInline)
	// Sharing
	cardsGroup.POST("/:id/shares", cardSharesHandler.Create)
	cardsGroup.PATCH("/:id/shares/:share_id", cardSharesHandler.Update)
	cardsGroup.DELETE("/:id/shares/:share_id", cardSharesHandler.Delete)
	cardsGroup.GET("/:id/shares/new-inline", cardSharesHandler.NewInline)
	cardsGroup.GET("/:id/shares/cancel", cardSharesHandler.Cancel)
	cardsGroup.GET("/:id/shares/:share_id/edit-inline", cardSharesHandler.EditInline)
	cardsGroup.GET("/:id/shares/:share_id/cancel-edit", cardSharesHandler.CancelEdit)
	// Transfer
	cardsGroup.GET("/:id/transfer/inline", cardHandler.TransferInline)
	cardsGroup.GET("/:id/transfer/cancel", cardHandler.CancelTransfer)
	cardsGroup.POST("/:id/transfer", cardHandler.Transfer)
	// Favorites
	cardsGroup.POST("/:id/favorite", favoritesHandler.ToggleCardFavorite)
}

// registerVouchersRoutes registers all voucher-related routes.
func registerVouchersRoutes(
	protected *echo.Group,
	cfg *config.Config,
	voucherHandler *vouchers.Handler,
	voucherSharesHandler *handlers.VoucherSharesHandler,
	favoritesHandler *handlers.FavoritesHandler,
) {
	vouchersGroup := protected.Group("/vouchers")
	vouchersGroup.Use(middleware.RequireVouchersEnabled(cfg))
	vouchersGroup.GET("", voucherHandler.Index)
	vouchersGroup.GET("/new", voucherHandler.New)
	vouchersGroup.POST("", voucherHandler.Create)
	vouchersGroup.GET("/:id", voucherHandler.Show)
	vouchersGroup.GET("/:id/edit", voucherHandler.Edit)
	vouchersGroup.POST("/:id", voucherHandler.Update)
	vouchersGroup.DELETE("/:id", voucherHandler.Delete)
	// Inline editing (HTMX-powered)
	vouchersGroup.GET("/:id/edit-inline", voucherHandler.EditInline)
	vouchersGroup.GET("/:id/cancel-edit", voucherHandler.CancelEdit)
	vouchersGroup.PATCH("/:id", voucherHandler.UpdateInline)
	// Sharing (read-only only)
	vouchersGroup.POST("/:id/shares", voucherSharesHandler.Create)
	vouchersGroup.DELETE("/:id/shares/:share_id", voucherSharesHandler.Delete)
	vouchersGroup.GET("/:id/shares/new-inline", voucherSharesHandler.NewInline)
	vouchersGroup.GET("/:id/shares/cancel", voucherSharesHandler.Cancel)
	// Transfer
	vouchersGroup.GET("/:id/transfer/inline", voucherHandler.TransferInline)
	vouchersGroup.GET("/:id/transfer/cancel", voucherHandler.CancelTransfer)
	vouchersGroup.POST("/:id/transfer", voucherHandler.Transfer)
	// Favorites
	vouchersGroup.POST("/:id/favorite", favoritesHandler.ToggleVoucherFavorite)
}

// registerGiftCardsRoutes registers all gift card-related routes.
func registerGiftCardsRoutes(
	protected *echo.Group,
	cfg *config.Config,
	giftCardHandler *giftcards.Handler,
	giftCardSharesHandler *handlers.GiftCardSharesHandler,
	favoritesHandler *handlers.FavoritesHandler,
) {
	giftCardsGroup := protected.Group("/gift-cards")
	giftCardsGroup.Use(middleware.RequireGiftCardsEnabled(cfg))
	giftCardsGroup.GET("", giftCardHandler.Index)
	giftCardsGroup.GET("/new", giftCardHandler.New)
	giftCardsGroup.POST("", giftCardHandler.Create)
	giftCardsGroup.GET("/:id", giftCardHandler.Show)
	giftCardsGroup.GET("/:id/edit", giftCardHandler.Edit)
	giftCardsGroup.POST("/:id", giftCardHandler.Update)
	giftCardsGroup.DELETE("/:id", giftCardHandler.Delete)
	// Inline editing (HTMX-powered)
	giftCardsGroup.GET("/:id/edit-inline", giftCardHandler.EditInline)
	giftCardsGroup.GET("/:id/cancel-edit", giftCardHandler.CancelEdit)
	giftCardsGroup.PATCH("/:id", giftCardHandler.UpdateInline)
	// Transaction management
	giftCardsGroup.GET("/:id/transactions/new", giftCardHandler.TransactionNew)
	giftCardsGroup.GET("/:id/transactions/cancel", giftCardHandler.TransactionCancel)
	giftCardsGroup.POST("/:id/transactions", giftCardHandler.TransactionCreate)
	giftCardsGroup.DELETE("/:id/transactions/:transaction_id", giftCardHandler.TransactionDelete)
	// Sharing
	giftCardsGroup.POST("/:id/shares", giftCardSharesHandler.Create)
	giftCardsGroup.PATCH("/:id/shares/:share_id", giftCardSharesHandler.Update)
	giftCardsGroup.DELETE("/:id/shares/:share_id", giftCardSharesHandler.Delete)
	giftCardsGroup.GET("/:id/shares/new-inline", giftCardSharesHandler.NewInline)
	giftCardsGroup.GET("/:id/shares/cancel", giftCardSharesHandler.Cancel)
	giftCardsGroup.GET("/:id/shares/:share_id/edit-inline", giftCardSharesHandler.EditInline)
	giftCardsGroup.GET("/:id/shares/:share_id/cancel-edit", giftCardSharesHandler.CancelEdit)
	// Transfer
	giftCardsGroup.GET("/:id/transfer/inline", giftCardHandler.TransferInline)
	giftCardsGroup.GET("/:id/transfer/cancel", giftCardHandler.CancelTransfer)
	giftCardsGroup.POST("/:id/transfer", giftCardHandler.Transfer)
	// Favorites
	giftCardsGroup.POST("/:id/favorite", favoritesHandler.ToggleGiftCardFavorite)
}
