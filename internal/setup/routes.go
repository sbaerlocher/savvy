// Package setup contains setup logic for route registration.
package setup

import (
	"savvy/internal/config"
	"savvy/internal/debug"
	"savvy/internal/email"
	"savvy/internal/handlers"
	"savvy/internal/handlers/api"
	"savvy/internal/middleware"
	"savvy/internal/oauth"
	"savvy/internal/services"

	"github.com/labstack/echo/v5"
)

// RouteConfig holds dependencies needed for route registration.
type RouteConfig struct {
	Echo             *echo.Echo
	Config           *config.Config
	ServiceContainer *services.Container
	RateLimiters     *RateLimiters
	OAuthProvider    *oauth.Provider
	EmailService     email.ServiceInterface
	HealthService    services.HealthCheckServiceInterface
}

// RegisterRoutes registers all application routes.
func RegisterRoutes(rc *RouteConfig) {
	e := rc.Echo
	cfg := rc.Config
	serviceContainer := rc.ServiceContainer

	// ========================================
	// JSON API Routes (for SvelteKit Frontend)
	// ========================================
	registerAPIRoutes(e, cfg, serviceContainer, rc.RateLimiters, rc.EmailService, rc.HealthService)

	// ========================================
	// OAuth Routes (if OAuth is enabled)
	// ========================================
	if cfg.IsOAuthEnabled() {
		oauthHandler := handlers.NewOAuthHandler(serviceContainer.UserService, rc.OAuthProvider, cfg)
		e.GET("/auth/oauth/login", oauthHandler.Login)
		e.GET("/auth/oauth/callback", oauthHandler.Callback)
	}

	// ========================================
	// PWA Reset Page (Server-rendered, bypasses SPA + Service Worker)
	// ========================================
	resetHandler := handlers.NewResetHandler()
	e.GET("/reset", resetHandler.ServeResetPage)

	// ========================================
	// SvelteKit SPA Frontend (Production only - Dev uses Vite on :5173)
	// ========================================
	// In development, frontend runs separately on :5173 with Vite Dev Server
	// In production, frontend is embedded and served by Go backend
	if cfg.IsProduction() {
		spaHandler := handlers.NewSPAHandler()

		// ========================================
		// Static Assets (NO AUTH REQUIRED)
		// ========================================
		// Serve static assets (_app/*, favicon, manifest, etc.)
		e.GET("/_app/*", spaHandler.ServeStatic)
		e.GET("/logo.png", spaHandler.ServeStatic)
		e.GET("/favicon.png", spaHandler.ServeStatic)
		e.GET("/favicon-*.png", spaHandler.ServeStatic)
		e.GET("/icon-*.png", spaHandler.ServeStatic)
		e.GET("/apple-touch-icon.png", spaHandler.ServeStatic)
		e.GET("/manifest.webmanifest", spaHandler.ServeStatic)
		e.GET("/service-worker.js", spaHandler.ServeStatic)
		e.HEAD("/service-worker.js", spaHandler.ServeStatic)
		e.GET("/sw.js", spaHandler.ServeStatic)
		e.GET("/workbox-*.js", spaHandler.ServeStatic)
		e.GET("/registerSW.js", spaHandler.ServeStatic)
		e.GET("/*.wasm", spaHandler.ServeStatic)
		e.HEAD("/*.wasm", spaHandler.ServeStatic)

		// ========================================
		// Public SPA Routes (NO AUTH REQUIRED)
		// ========================================
		// Login & Registration & Verification & Password Reset
		e.GET("/login", spaHandler.ServeSPA)
		e.GET("/register", spaHandler.ServeSPA)
		e.GET("/verify-email", spaHandler.ServeSPA)
		e.GET("/forgot-password", spaHandler.ServeSPA)
		e.GET("/reset-password", spaHandler.ServeSPA)
		e.GET("/login/2fa", spaHandler.ServeSPA)
		e.GET("/unsubscribe", spaHandler.ServeSPA)

		// OAuth routes (if enabled - already registered above)
		// /auth/oauth/login and /auth/oauth/callback are handled separately

		// ========================================
		// Protected SPA Routes (AUTH REQUIRED) - SVL-001 & SVL-002 Fix
		// ========================================
		// Security Context:
		// - SVL-001: Client-Side Only Auth (CRITICAL) - Fixed by Go Backend Guards
		// - All protected routes require valid session (SetCurrentUserWithService + RequireAuth)
		// - Works in Production (no Node.js SSR needed)
		spaProtected := e.Group("")
		spaProtected.Use(middleware.SetCurrentUserWithService(serviceContainer.UserService)) // Load user from session
		spaProtected.Use(middleware.RequireAuth)                                             // Redirect to /login if not authenticated

		// Dashboard & Root
		spaProtected.GET("/", spaHandler.ServeSPA)
		spaProtected.GET("/dashboard", spaHandler.ServeSPA)

		// Settings & Account
		spaProtected.GET("/settings", spaHandler.ServeSPA)
		spaProtected.GET("/profile", spaHandler.ServeSPA)
		spaProtected.GET("/security", spaHandler.ServeSPA)
		spaProtected.GET("/notifications", spaHandler.ServeSPA)

		// Search

		// Cards Routes
		spaProtected.GET("/cards", spaHandler.ServeSPA)
		spaProtected.GET("/cards/new", spaHandler.ServeSPA)
		spaProtected.GET("/cards/:id", spaHandler.ServeSPA)
		spaProtected.GET("/cards/:id/edit", spaHandler.ServeSPA)

		// Vouchers Routes
		spaProtected.GET("/vouchers", spaHandler.ServeSPA)
		spaProtected.GET("/vouchers/new", spaHandler.ServeSPA)
		spaProtected.GET("/vouchers/:id", spaHandler.ServeSPA)
		spaProtected.GET("/vouchers/:id/edit", spaHandler.ServeSPA)

		// Gift Cards Routes
		spaProtected.GET("/gift-cards", spaHandler.ServeSPA)
		spaProtected.GET("/gift-cards/new", spaHandler.ServeSPA)
		spaProtected.GET("/gift-cards/:id", spaHandler.ServeSPA)
		spaProtected.GET("/gift-cards/:id/edit", spaHandler.ServeSPA)

		// Merchants Routes
		spaProtected.GET("/merchants", spaHandler.ServeSPA)
		spaProtected.GET("/merchants/:id", spaHandler.ServeSPA)

		// ========================================
		// Admin SPA Routes (ADMIN REQUIRED) - SVL-002 Fix
		// ========================================
		// Security Context:
		// - SVL-002: Admin Routes ohne Server-Side Guards (CRITICAL) - Fixed by RequireAdmin
		// - All admin routes require is_admin = true
		// - Non-admin users get 403 Forbidden or redirect
		spaAdmin := e.Group("/admin")
		spaAdmin.Use(middleware.SetCurrentUserWithService(serviceContainer.UserService)) // Load user from session
		spaAdmin.Use(middleware.RequireAdmin)                                            // Check admin role

		spaAdmin.GET("", spaHandler.ServeSPA)
		spaAdmin.GET("/", spaHandler.ServeSPA)
		spaAdmin.GET("/users", spaHandler.ServeSPA)
		spaAdmin.GET("/users/new", spaHandler.ServeSPA)
		spaAdmin.GET("/users/:id/edit", spaHandler.ServeSPA)
		spaAdmin.GET("/merchants", spaHandler.ServeSPA)
		spaAdmin.GET("/merchants/new", spaHandler.ServeSPA)
		spaAdmin.GET("/merchants/:id/edit", spaHandler.ServeSPA)
		spaAdmin.GET("/audit-log", spaHandler.ServeSPA)
		spaAdmin.GET("/system-health", spaHandler.ServeSPA)
		spaAdmin.GET("/email-templates", spaHandler.ServeSPA)
	}

	// ========================================
	// Development Debug Tools
	// ========================================
	if !cfg.IsProduction() {
		debug.PrintRoutes(e)
	}
}

// registerAPIRoutes registers all JSON API routes for the SvelteKit frontend
func registerAPIRoutes(e *echo.Echo, cfg *config.Config, serviceContainer *services.Container, rateLimiters *RateLimiters, emailService email.ServiceInterface, healthService services.HealthCheckServiceInterface) {
	// Set configured timezone for date-based calculations (voucher status, etc.)
	api.SetAppLocation(cfg.Location)
	// ========================================
	// API v1 Group
	// ========================================
	apiV1 := e.Group("/api/v1")

	// Global Rate Limiting (applies to all API routes)
	if rateLimiters != nil && rateLimiters.Global != nil {
		apiV1.Use(middleware.RateLimitMiddleware(rateLimiters.Global))
	}

	// CORS Middleware (Development only)
	if !cfg.IsProduction() {
		apiV1.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig(cfg.CORSAllowedOrigins)))
	}

	// CSRF Protection (Double Submit Cookie Pattern)
	apiV1.Use(middleware.CSRFApiMiddleware)

	// ========================================
	// Initialize API Handlers
	// ========================================
	configAPIHandler := api.NewConfigHandler(cfg)
	authAPIHandler := api.NewAuthHandler(
		serviceContainer.UserService,
		serviceContainer.EmailTokenService,
		emailService,
		cfg.FrontendURL,
	)
	profileAPIHandler := api.NewProfileHandler(serviceContainer.UserService, serviceContainer.AccountService)
	profileAPIHandler.SetSessionService(serviceContainer.SessionService)
	sessionsAPIHandler := api.NewSessionsHandler(serviceContainer.SessionService)
	exportAPIHandler := api.NewExportHandler(serviceContainer.ExportService)
	dashboardAPIHandler := api.NewDashboardHandler(serviceContainer.DashboardService, serviceContainer.FavoriteService)
	merchantsAPIHandler := api.NewMerchantsHandler(serviceContainer.MerchantService)
	sharedUsersAPIHandler := api.NewSharedUsersHandler(serviceContainer.ShareService)

	cardsAPIHandler := api.NewCardsHandler(
		serviceContainer.CardService,
		serviceContainer.AuthzService,
		serviceContainer.MerchantService,
		serviceContainer.UserService,
		serviceContainer.FavoriteService,
		serviceContainer.ShareService,
		serviceContainer.TransferService,
		serviceContainer.AdminService,
	)

	vouchersAPIHandler := api.NewVouchersHandler(
		serviceContainer.VoucherService,
		serviceContainer.AuthzService,
		serviceContainer.MerchantService,
		serviceContainer.UserService,
		serviceContainer.FavoriteService,
		serviceContainer.ShareService,
		serviceContainer.TransferService,
	)

	giftCardsAPIHandler := api.NewGiftCardsHandler(
		serviceContainer.GiftCardService,
		serviceContainer.AuthzService,
		serviceContainer.MerchantService,
		serviceContainer.UserService,
		serviceContainer.FavoriteService,
		serviceContainer.ShareService,
		serviceContainer.TransferService,
	)

	batchAPIHandler := api.NewBatchHandler(
		serviceContainer.CardService,
		serviceContainer.VoucherService,
		serviceContainer.GiftCardService,
		serviceContainer.AuthzService,
		serviceContainer.ShareService,
		serviceContainer.TransferService,
		serviceContainer.UserService,
		serviceContainer.ExportService,
	)

	adminAPIHandler := api.NewAdminHandler(
		serviceContainer.AdminService,
		serviceContainer.UserService,
		healthService,
		emailService,
		serviceContainer.PushService,
	)

	importAPIHandler := api.NewImportHandler(serviceContainer.ImportService)

	notificationsAPIHandler := api.NewNotificationsHandler(
		serviceContainer.NotificationService,
	)

	pushAPIHandler := api.NewPushHandler(serviceContainer.PushService)

	// Set session service on auth handler for session revocation on password reset
	if serviceContainer.SessionService != nil {
		authAPIHandler.SetSessionService(serviceContainer.SessionService)
	}

	var totpAPIHandler *api.TOTPHandler
	if serviceContainer.TOTPService != nil {
		totpAPIHandler = api.NewTOTPHandler(serviceContainer.TOTPService)
		authAPIHandler.SetTOTPService(serviceContainer.TOTPService)
	}

	otelProxyHandler := handlers.NewOTelProxyHandler(cfg)

	// ========================================
	// Public API Routes (No Auth Required)
	// ========================================
	apiV1.GET("/config", configAPIHandler.GetConfig) // App configuration (OAuth, features, etc.)

	// Auth endpoints with stricter rate limiting (5 req/s vs 100 req/s global)
	authGroup := apiV1.Group("/auth")
	if rateLimiters != nil && rateLimiters.Auth != nil {
		authGroup.Use(middleware.RateLimitMiddleware(rateLimiters.Auth))
	}
	authGroup.POST("/login", authAPIHandler.Login, middleware.RequireLocalLoginEnabled(cfg))
	authGroup.POST("/register", authAPIHandler.Register, middleware.RequireRegistrationEnabled(cfg))
	authGroup.POST("/logout", authAPIHandler.Logout)
	authGroup.POST("/verify-email", authAPIHandler.VerifyEmail)
	authGroup.POST("/unsubscribe-notifications", authAPIHandler.UnsubscribeNotifications)
	authGroup.POST("/unsubscribe-reminders", authAPIHandler.UnsubscribeReminders)
	// Password reset endpoints with dedicated stricter rate limiter (1 req/min per IP)
	passwordResetGroup := apiV1.Group("/auth")
	if rateLimiters != nil && rateLimiters.PasswordReset != nil {
		passwordResetGroup.Use(middleware.RateLimitMiddleware(rateLimiters.PasswordReset))
	}
	passwordResetGroup.POST("/forgot-password", authAPIHandler.RequestPasswordReset)
	passwordResetGroup.POST("/reset-password", authAPIHandler.ResetPassword)

	// 2FA challenge (public - used during login before full auth)
	// Uses a dedicated stricter rate limiter (1 req/3s per IP, burst 2) separate from the
	// general auth limiter to slow brute-force attempts against the 6-digit TOTP window.
	if totpAPIHandler != nil {
		twoFAChallengeGroup := apiV1.Group("/auth")
		if rateLimiters != nil && rateLimiters.TwoFAChallenge != nil {
			twoFAChallengeGroup.Use(middleware.RateLimitMiddleware(rateLimiters.TwoFAChallenge))
		}
		twoFAChallengeGroup.POST("/2fa/challenge", totpAPIHandler.Challenge)
	}

	// ========================================
	// Protected API Routes (Auth Required)
	// ========================================
	apiProtected := apiV1.Group("")
	apiProtected.Use(middleware.SetCurrentUserWithService(serviceContainer.UserService)) // Load user from session
	apiProtected.Use(middleware.RequireAuth)                                             // Check if user is authenticated

	// Per-user rate limiting (after auth, so user ID is available)
	if rateLimiters != nil && rateLimiters.User != nil {
		apiProtected.Use(middleware.UserRateLimitMiddleware(rateLimiters.User))
	}

	// Auth
	apiProtected.GET("/auth/me", authAPIHandler.Me)
	apiProtected.POST("/auth/request-verification", authAPIHandler.RequestVerification)

	// 2FA management (protected - requires full auth)
	if totpAPIHandler != nil {
		// New enrollment (setup + verify-and-enable) is gated on Is2FAEnabled:
		// when the operator turns 2FA off, no one can newly enroll.
		if cfg.Is2FAEnabled() {
			apiProtected.POST("/auth/2fa/setup", totpAPIHandler.Setup)
			apiProtected.POST("/auth/2fa/verify", totpAPIHandler.Verify)
		}
		// Existing-2FA management stays available whenever the TOTP service
		// runs (key present), independent of the enrollment flag: users must
		// still be able to disable, rotate backup codes, and read status for
		// 2FA they already have — otherwise flipping ENABLE_2FA=false would
		// trap them with an unmanageable second factor.
		apiProtected.POST("/auth/2fa/disable", totpAPIHandler.Disable)
		apiProtected.POST("/auth/2fa/backup-codes", totpAPIHandler.RegenerateBackupCodes)
		apiProtected.GET("/auth/2fa/status", totpAPIHandler.Status)
	}

	// Profile
	apiProtected.GET("/profile", profileAPIHandler.GetProfile)
	apiProtected.PATCH("/profile", profileAPIHandler.UpdateProfile)
	apiProtected.POST("/profile/change-password", profileAPIHandler.ChangePassword)
	apiProtected.POST("/profile/delete-account", profileAPIHandler.DeleteAccount)

	// Sessions
	apiProtected.GET("/profile/sessions", sessionsAPIHandler.List)
	apiProtected.DELETE("/profile/sessions/:id", sessionsAPIHandler.Revoke)
	apiProtected.POST("/profile/sessions/revoke-others", sessionsAPIHandler.RevokeOthers)

	// Export
	apiProtected.GET("/export", exportAPIHandler.ExportData)

	// Import
	apiProtected.POST("/import/json", importAPIHandler.ImportJSON)
	apiProtected.POST("/import/json/preview", importAPIHandler.PreviewJSON)
	apiProtected.POST("/import/csv/cards", importAPIHandler.ImportCardsCSV)
	apiProtected.POST("/import/csv/vouchers", importAPIHandler.ImportVouchersCSV)
	apiProtected.POST("/import/csv/gift-cards", importAPIHandler.ImportGiftCardsCSV)

	// Dashboard
	apiProtected.GET("/dashboard", dashboardAPIHandler.Get)

	// Search

	// OpenTelemetry Proxy Endpoints (Frontend → OTEL Collector)
	apiProtected.POST("/otel/traces", otelProxyHandler.ProxyTraces)
	apiProtected.POST("/otel/logs", otelProxyHandler.ProxyLogs)
	apiProtected.POST("/otel/metrics", otelProxyHandler.ProxyMetrics)

	// Push Notifications
	if serviceContainer.PushService != nil {
		apiProtected.POST("/push/subscribe", pushAPIHandler.Subscribe)
		apiProtected.POST("/push/unsubscribe", pushAPIHandler.Unsubscribe)
		apiProtected.GET("/push/vapid-key", pushAPIHandler.GetVAPIDKey)
	}

	// Notifications
	apiProtected.GET("/notifications", notificationsAPIHandler.List)
	apiProtected.GET("/notifications/unread-count", notificationsAPIHandler.GetUnreadCount)
	apiProtected.POST("/notifications/:id/read", notificationsAPIHandler.MarkAsRead)
	apiProtected.POST("/notifications/read-all", notificationsAPIHandler.MarkAllAsRead)
	apiProtected.DELETE("/notifications/:id", notificationsAPIHandler.Delete)

	// Merchants (Read-Only)
	apiProtected.GET("/merchants", merchantsAPIHandler.List)
	apiProtected.GET("/merchants/search", merchantsAPIHandler.Search)
	apiProtected.GET("/merchants/:id", merchantsAPIHandler.Show)

	// Shared Users (for autocomplete)
	apiProtected.GET("/shared-users", sharedUsersAPIHandler.Search)

	// Cards API
	cardsAPI := apiProtected.Group("/cards")
	cardsAPI.Use(middleware.RequireCardsEnabled(cfg))
	cardsAPI.GET("", cardsAPIHandler.List)
	cardsAPI.POST("", cardsAPIHandler.Create)
	cardsAPI.GET("/:id", cardsAPIHandler.Show)
	cardsAPI.PATCH("/:id", cardsAPIHandler.Update)
	cardsAPI.DELETE("/:id", cardsAPIHandler.Delete)
	cardsAPI.POST("/:id/restore", cardsAPIHandler.Restore)
	cardsAPI.POST("/:id/favorite", cardsAPIHandler.ToggleFavorite)
	cardsAPI.POST("/:id/share", cardsAPIHandler.CreateShare)
	cardsAPI.PATCH("/:id/share/:sharedWithID", cardsAPIHandler.UpdateShare)
	cardsAPI.DELETE("/:id/share/:sharedWithID", cardsAPIHandler.DeleteShare)
	cardsAPI.POST("/:id/transfer", cardsAPIHandler.Transfer)
	// Batch operations
	cardsAPI.POST("/batch/delete", batchAPIHandler.DeleteCards)
	cardsAPI.POST("/batch/share", batchAPIHandler.ShareCards)
	cardsAPI.POST("/batch/transfer", batchAPIHandler.TransferCards)
	cardsAPI.POST("/batch/export", batchAPIHandler.ExportCards)

	// Vouchers API
	vouchersAPI := apiProtected.Group("/vouchers")
	vouchersAPI.Use(middleware.RequireVouchersEnabled(cfg))
	vouchersAPI.GET("", vouchersAPIHandler.List)
	vouchersAPI.POST("", vouchersAPIHandler.Create)
	vouchersAPI.GET("/:id", vouchersAPIHandler.Show)
	vouchersAPI.PATCH("/:id", vouchersAPIHandler.Update)
	vouchersAPI.DELETE("/:id", vouchersAPIHandler.Delete)
	vouchersAPI.POST("/:id/restore", vouchersAPIHandler.Restore)
	vouchersAPI.POST("/:id/favorite", vouchersAPIHandler.ToggleFavorite)
	vouchersAPI.POST("/:id/share", vouchersAPIHandler.CreateShare)
	vouchersAPI.DELETE("/:id/share/:sharedWithID", vouchersAPIHandler.DeleteShare)
	vouchersAPI.POST("/:id/transfer", vouchersAPIHandler.Transfer)
	// Batch operations
	vouchersAPI.POST("/batch/delete", batchAPIHandler.DeleteVouchers)
	vouchersAPI.POST("/batch/share", batchAPIHandler.ShareVouchers)
	vouchersAPI.POST("/batch/transfer", batchAPIHandler.TransferVouchers)
	vouchersAPI.POST("/batch/export", batchAPIHandler.ExportVouchers)

	// Gift Cards API
	giftCardsAPI := apiProtected.Group("/gift-cards")
	giftCardsAPI.Use(middleware.RequireGiftCardsEnabled(cfg))
	giftCardsAPI.GET("", giftCardsAPIHandler.List)
	giftCardsAPI.POST("", giftCardsAPIHandler.Create)
	giftCardsAPI.GET("/:id", giftCardsAPIHandler.Show)
	giftCardsAPI.PATCH("/:id", giftCardsAPIHandler.Update)
	giftCardsAPI.DELETE("/:id", giftCardsAPIHandler.Delete)
	giftCardsAPI.POST("/:id/restore", giftCardsAPIHandler.Restore)
	giftCardsAPI.POST("/:id/favorite", giftCardsAPIHandler.ToggleFavorite)
	giftCardsAPI.POST("/:id/share", giftCardsAPIHandler.CreateShare)
	giftCardsAPI.PATCH("/:id/share/:sharedWithID", giftCardsAPIHandler.UpdateShare)
	giftCardsAPI.DELETE("/:id/share/:sharedWithID", giftCardsAPIHandler.DeleteShare)
	giftCardsAPI.POST("/:id/transfer", giftCardsAPIHandler.Transfer)
	// Batch operations
	giftCardsAPI.POST("/batch/delete", batchAPIHandler.DeleteGiftCards)
	giftCardsAPI.POST("/batch/share", batchAPIHandler.ShareGiftCards)
	giftCardsAPI.POST("/batch/transfer", batchAPIHandler.TransferGiftCards)
	giftCardsAPI.POST("/batch/export", batchAPIHandler.ExportGiftCards)
	// Transactions
	giftCardsAPI.GET("/:id/transactions", giftCardsAPIHandler.ListTransactions)
	giftCardsAPI.POST("/:id/transactions", giftCardsAPIHandler.CreateTransaction)
	giftCardsAPI.DELETE("/:id/transactions/:transactionID", giftCardsAPIHandler.DeleteTransaction)

	// Admin API (requires admin role)
	adminAPI := apiProtected.Group("/admin")
	adminAPI.Use(middleware.RequireAdmin)
	// Users
	adminAPI.GET("/users", adminAPIHandler.ListUsers)
	adminAPI.GET("/users/:id", adminAPIHandler.GetUser)
	adminAPI.POST("/users", adminAPIHandler.CreateUser)
	adminAPI.PATCH("/users/:id", adminAPIHandler.UpdateUser)
	// Merchants (Admin CRUD)
	adminAPI.POST("/merchants", merchantsAPIHandler.Create)
	adminAPI.PATCH("/merchants/:id", merchantsAPIHandler.Update)
	adminAPI.DELETE("/merchants/:id", merchantsAPIHandler.Delete)
	// Audit Log
	adminAPI.GET("/audit-log", adminAPIHandler.GetAuditLogs)
	// Resource Restoration
	adminAPI.POST("/restore/:resource_type/:resource_id", adminAPIHandler.RestoreResource)
	// System Health
	adminAPI.GET("/system-health", adminAPIHandler.GetSystemHealth)
	adminAPI.POST("/test-email", adminAPIHandler.SendTestEmail)
	adminAPI.POST("/test-push", adminAPIHandler.SendTestPush)
	// Email Template Preview (development only)
	if !cfg.IsProduction() {
		adminAPI.POST("/preview-email", adminAPIHandler.SendPreviewEmail)
	}
	// Impersonation endpoints
	adminAPI.POST("/users/:id/impersonate", adminAPIHandler.StartImpersonation)

	// Stop impersonation - NO RequireAdmin middleware (impersonated user is not admin)
	apiProtected.POST("/admin/impersonate/stop", adminAPIHandler.StopImpersonation)
}
