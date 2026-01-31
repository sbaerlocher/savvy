// Package main is the entry point for the savvy system server.
package main

import (
	"context"
	"flag"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"savvy/internal/assets"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/debug"
	"savvy/internal/handlers"
	"savvy/internal/handlers/cards"
	"savvy/internal/handlers/giftcards"
	"savvy/internal/handlers/vouchers"
	"savvy/internal/i18n"
	"savvy/internal/metrics"
	"savvy/internal/middleware"
	"savvy/internal/migrations"
	"savvy/internal/oauth"
	"savvy/internal/security"
	"savvy/internal/services"
	"savvy/internal/telemetry"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

var (
	healthCheck = flag.Bool("health", false, "perform health check and exit")
	healthPort  = flag.String("port", "3000", "server port for health check")
)

func main() {
	flag.Parse()

	// If health check flag is set, perform check and exit
	if *healthCheck {
		os.Exit(performHealthCheck(*healthPort))
	}

	os.Exit(run())
}

// performHealthCheck makes HTTP request to /health endpoint
func performHealthCheck(port string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := "http://127.0.0.1:" + port + "/health"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return 1
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return 1
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Health check returned status %d", resp.StatusCode)
		return 1
	}

	return 0
}

// updateMetrics updates Prometheus gauges for resource counts and DB stats
func updateMetrics() {
	// Update resource counts
	var cardsCount, vouchersCount, giftCardsCount, usersCount int64
	database.DB.Table("cards").Count(&cardsCount)
	database.DB.Table("vouchers").Count(&vouchersCount)
	database.DB.Table("gift_cards").Count(&giftCardsCount)
	database.DB.Table("users").Count(&usersCount)

	metrics.UpdateResourceCounts(cardsCount, vouchersCount, giftCardsCount, usersCount)

	// Cleanup inactive sessions (sessions are re-counted via middleware)
	middleware.CleanupInactiveSessions()

	// Update DB connection pool metrics
	sqlDB, err := database.DB.DB()
	if err == nil {
		stats := sqlDB.Stats()
		metrics.UpdateDBMetrics(stats.InUse, stats.Idle)
	}
}

func run() int {
	// Load config
	cfg := config.Load()

	// Initialize structured logging
	logLevel := slog.LevelInfo
	if !cfg.IsProduction() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting savvy system",
		"version", cfg.ServiceVersion,
		"environment", cfg.Environment,
		"otel_enabled", cfg.OTelEnabled,
	)

	// Initialize OpenTelemetry (enabled by default in production)
	otelEnabled := cfg.OTelEnabled
	if cfg.IsProduction() && !cfg.OTelEnabled {
		slog.Warn("OTel is disabled in production - metrics and tracing unavailable")
	}

	shutdown, err := telemetry.InitTracer(
		cfg.ServiceName,
		cfg.ServiceVersion,
		cfg.OTelEndpoint,
		otelEnabled,
	)
	if err != nil {
		slog.Error("Failed to initialize telemetry", "error", err)
		return 1
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down telemetry: %v", err)
		}
	}()

	// Init session store with secure flag based on environment
	middleware.InitSessionStore(cfg.SessionSecret, cfg.IsProduction())

	// Initialize i18n with embedded locale files
	if err := i18n.Init(assets.Locales); err != nil {
		log.Printf("Failed to initialize i18n: %v", err)
		return 1
	}

	// Initialize security package for token-based barcode access
	security.Init(cfg.SessionSecret)

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return 1
	}

	// Enable database telemetry if OTel is enabled
	if cfg.OTelEnabled {
		if err := database.EnableTelemetry(cfg.ServiceName); err != nil {
			log.Printf("Warning: Failed to enable database telemetry: %v", err)
		}
	}

	// Run migrations using Gormigrate (Laravel-like experience)
	if cfg.AutoMigrate {
		log.Println("🚀 Running database migrations...")
		m := gormigrate.New(database.DB, gormigrate.DefaultOptions, migrations.GetMigrations())
		if err := m.Migrate(); err != nil {
			log.Printf("❌ Migration failed: %v", err)
			return 1
		}
		log.Println("✅ Migrations applied successfully")
	} else {
		log.Println("⚠️  AutoMigrate disabled (AUTO_MIGRATE=false)")
		log.Println("   Run migrations manually: make migrate-up")
	}

	// Enable audit logging for all deletions
	if err := database.EnableAuditLogging(); err != nil {
		log.Printf("Warning: Failed to enable audit logging: %v", err)
	}

	// Initialize OAuth if configured
	if cfg.IsOAuthEnabled() {
		oauthProvider, err := oauth.NewProvider(cfg)
		if err != nil {
			log.Printf("Warning: Failed to initialize OAuth: %v", err)
			handlers.InitOAuth(nil, cfg) // Disable OAuth but pass config
		} else {
			handlers.InitOAuth(oauthProvider, cfg)
			log.Printf("✅ OAuth initialized successfully")
			if len(cfg.OAuthAdminEmails) > 0 {
				log.Printf("   Admin emails: %v", cfg.OAuthAdminEmails)
			}
			if cfg.OAuthAdminGroup != "" {
				log.Printf("   Admin group: %s", cfg.OAuthAdminGroup)
			}
		}
	} else {
		handlers.InitOAuth(nil, cfg)
		log.Printf("ℹ️  OAuth not configured (set OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_ISSUER to enable)")
	}

	// Create Echo instance
	e := echo.New()

	// OpenTelemetry Middleware (must be first for proper tracing)
	if cfg.OTelEnabled {
		e.Use(otelecho.Middleware(cfg.ServiceName))
		e.Use(middleware.OTelLogger()) // Add trace IDs to logs
	}

	// Middleware
	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		LogLatency:  true,
		HandleError: true,
		LogValuesFunc: func(_ echo.Context, v echomiddleware.RequestLoggerValues) error {
			if v.Error != nil {
				slog.Error("request error",
					"uri", v.URI,
					"method", v.Method,
					"status", v.Status,
					"latency", v.Latency,
					"error", v.Error,
				)
			} else {
				slog.Info("request",
					"uri", v.URI,
					"method", v.Method,
					"status", v.Status,
					"latency", v.Latency,
				)
			}
			return nil
		},
	}))
	e.Use(echomiddleware.Recover())
	e.Use(metrics.Middleware()) // Prometheus metrics

	// CSRF Protection (only for non-GET requests)
	e.Use(echomiddleware.CSRFWithConfig(echomiddleware.CSRFConfig{
		TokenLookup:    "form:csrf_token,header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookieHTTPOnly: true,
		CookieSecure:   cfg.IsProduction(), // true in production, false in dev
		CookieSameSite: 2,                  // SameSiteLaxMode
		ContextKey:     "csrf",             // Store token in context under "csrf" key
	}))

	// Observability endpoints (public, BEFORE auth middleware)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "healthy",
			"version": cfg.ServiceVersion,
			"service": "savvy",
		})
	})
	e.GET("/ready", handlers.Ready)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.Use(middleware.SetCurrentUser)
	e.Use(middleware.SessionTracking) // Track active sessions for metrics
	e.Use(middleware.LanguageDetection)

	// Set service version and config in context for health endpoint and templates
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("service_version", cfg.ServiceVersion)
			c.Set("config", cfg) // Make config available in Echo context

			// Inject config into Request Context for templates
			ctx := context.WithValue(c.Request().Context(), config.ConfigContextKey, cfg)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	})

	// Static files (embedded)
	staticFS, err := fs.Sub(assets.Static, "static")
	if err != nil {
		log.Printf("Failed to create static filesystem: %v", err)
		return 1
	}
	e.StaticFS("/static", staticFS)

	// Language switcher endpoint (public)
	e.GET("/set-language", middleware.SetLanguage)

	// PWA endpoints (public)
	e.GET("/offline", handlers.OfflinePage)                                   // Offline fallback page
	e.FileFS("/service-worker.js", "static/service-worker.js", assets.Static) // Service Worker
	e.FileFS("/manifest.json", "static/manifest.json", assets.Static)         // PWA Manifest

	// Initialize service container and handlers
	serviceContainer := services.NewContainer(database.DB)
	handlers.InitDashboardService(serviceContainer.DashboardService)
	cardHandler := cards.NewHandler(serviceContainer.CardService, serviceContainer.AuthzService, database.DB)
	voucherHandler := vouchers.NewHandler(serviceContainer.VoucherService, serviceContainer.AuthzService, database.DB)
	giftCardHandler := giftcards.NewHandler(serviceContainer.GiftCardService, serviceContainer.AuthzService, database.DB)
	cardSharesHandler := handlers.NewCardSharesHandler(serviceContainer.AuthzService)
	voucherSharesHandler := handlers.NewVoucherSharesHandler(serviceContainer.AuthzService)
	giftCardSharesHandler := handlers.NewGiftCardSharesHandler(serviceContainer.AuthzService)
	favoritesHandler := handlers.NewFavoritesHandler(serviceContainer.AuthzService)

	// ========================================
	// ROUTE REGISTRATION
	// ========================================

	// Rate limiter for auth endpoints (5 requests per second, burst of 10)
	authLimiter := middleware.NewIPRateLimiter(5, 10)

	// ========================================
	// Authentication Routes (Public)
	// Includes: Local login, OAuth, Registration
	// Middleware: Rate limiting on POST endpoints
	// ========================================
	auth := e.Group("/auth")
	auth.GET("/login", handlers.AuthLoginGet) // Handler manages redirect to OAuth if local login disabled
	auth.POST("/login", handlers.AuthLoginPost, middleware.RateLimitMiddleware(authLimiter), middleware.RequireLocalLoginEnabled(cfg))
	auth.GET("/register", handlers.AuthRegisterGet, middleware.RequireRegistrationEnabled(cfg))
	auth.POST("/register", handlers.AuthRegisterPost, middleware.RateLimitMiddleware(authLimiter), middleware.RequireRegistrationEnabled(cfg))
	auth.GET("/logout", handlers.AuthLogout)
	// OAuth routes (public)
	auth.GET("/oauth/login", handlers.OAuthLogin)
	auth.GET("/oauth/callback", handlers.OAuthCallback)

	// ========================================
	// Protected Routes (Authentication Required)
	// Middleware: RequireAuth
	// ========================================
	protected := e.Group("")
	protected.Use(middleware.RequireAuth)

	// Dashboard & Home
	protected.GET("/", handlers.HomeIndex)

	// Barcode generation (secure token-based access)
	protected.GET("/barcode/:token", handlers.BarcodeGenerate)

	// HTMX autocomplete endpoint (returns HTML fragment)
	protected.GET("/api/shared-users", handlers.SharedUsersAutocomplete)

	// ========================================
	// Merchants Routes (Read-Only for All Users)
	// ========================================
	merchants := protected.Group("/merchants")
	merchants.GET("", handlers.MerchantsIndex)
	merchants.GET("/search", handlers.MerchantsSearch) // Autocomplete search

	// ========================================
	// Merchants CRUD Routes (Admin or Impersonation)
	// Middleware: RequireImpersonationOrAdmin
	// Operations: Create, Update, Delete merchants
	// Accessible by: Admins (always) + Regular users (when impersonating)
	// ========================================
	merchantsCRUD := protected.Group("/merchants")
	merchantsCRUD.Use(middleware.RequireImpersonationOrAdmin)
	merchantsCRUD.GET("/new", handlers.MerchantsNew)
	merchantsCRUD.POST("", handlers.MerchantsCreate)
	merchantsCRUD.GET("/:id/edit", handlers.MerchantsEdit)
	merchantsCRUD.POST("/:id", handlers.MerchantsUpdate)
	merchantsCRUD.DELETE("/:id", handlers.MerchantsDelete)

	// ========================================
	// Cards Resource (Customer Loyalty Cards)
	// Feature Toggle: ENABLE_CARDS
	// Middleware: RequireCardsEnabled
	// Features: CRUD, Inline Editing (HTMX), Sharing, Favorites, Barcode Scanning
	// ========================================
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
	// Sharing (HTMX inline forms, granular permissions: CanEdit, CanDelete)
	cardsGroup.POST("/:id/shares", cardSharesHandler.Create)
	cardsGroup.PATCH("/:id/shares/:share_id", cardSharesHandler.Update)
	cardsGroup.DELETE("/:id/shares/:share_id", cardSharesHandler.Delete)
	cardsGroup.GET("/:id/shares/new-inline", cardSharesHandler.NewInline)
	cardsGroup.GET("/:id/shares/cancel", cardSharesHandler.Cancel)
	cardsGroup.GET("/:id/shares/:share_id/edit-inline", cardSharesHandler.EditInline)
	cardsGroup.GET("/:id/shares/:share_id/cancel-edit", cardSharesHandler.CancelEdit)
	// Favorites (polymorphic pinning)
	cardsGroup.POST("/:id/favorite", favoritesHandler.ToggleCardFavorite)

	// ========================================
	// Vouchers Resource (Time-Limited Vouchers)
	// Feature Toggle: ENABLE_VOUCHERS
	// Middleware: RequireVouchersEnabled
	// Features: CRUD, Inline Editing, Redemption Tracking, Read-Only Sharing, Favorites
	// ========================================
	vouchersGroup := protected.Group("/vouchers")
	vouchersGroup.Use(middleware.RequireVouchersEnabled(cfg))
	vouchersGroup.GET("", voucherHandler.Index)
	vouchersGroup.GET("/new", voucherHandler.New)
	vouchersGroup.POST("", voucherHandler.Create)
	vouchersGroup.GET("/:id", voucherHandler.Show)
	vouchersGroup.GET("/:id/edit", voucherHandler.Edit)
	vouchersGroup.POST("/:id", voucherHandler.Update)
	vouchersGroup.DELETE("/:id", voucherHandler.Delete)
	vouchersGroup.POST("/:id/redeem", voucherHandler.Redeem)
	// Inline editing (HTMX-powered)
	vouchersGroup.GET("/:id/edit-inline", voucherHandler.EditInline)
	vouchersGroup.GET("/:id/cancel-edit", voucherHandler.CancelEdit)
	vouchersGroup.PATCH("/:id", voucherHandler.UpdateInline)
	// Sharing (HTMX inline forms, read-only only - no permission editing)
	vouchersGroup.POST("/:id/shares", voucherSharesHandler.Create)
	vouchersGroup.DELETE("/:id/shares/:share_id", voucherSharesHandler.Delete)
	vouchersGroup.GET("/:id/shares/new-inline", voucherSharesHandler.NewInline)
	vouchersGroup.GET("/:id/shares/cancel", voucherSharesHandler.Cancel)
	// Favorites (polymorphic pinning)
	vouchersGroup.POST("/:id/favorite", favoritesHandler.ToggleVoucherFavorite)

	// ========================================
	// Gift Cards Resource (Rechargeable Gift Cards)
	// Feature Toggle: ENABLE_GIFT_CARDS
	// Middleware: RequireGiftCardsEnabled
	// Features: CRUD, Inline Editing, Transaction History, Sharing (CanEdit, CanDelete, CanEditTransactions), Favorites
	// Balance: Auto-calculated via database trigger (recalculate_gift_card_balance)
	// ========================================
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
	// Transaction management (HTMX, triggers balance recalculation)
	giftCardsGroup.GET("/:id/transactions/new", giftCardHandler.TransactionNew)
	giftCardsGroup.GET("/:id/transactions/cancel", giftCardHandler.TransactionCancel)
	giftCardsGroup.POST("/:id/transactions", giftCardHandler.TransactionCreate)
	giftCardsGroup.DELETE("/:id/transactions/:transaction_id", giftCardHandler.TransactionDelete)
	// Sharing (HTMX inline forms, includes CanEditTransactions permission)
	giftCardsGroup.POST("/:id/shares", giftCardSharesHandler.Create)
	giftCardsGroup.PATCH("/:id/shares/:share_id", giftCardSharesHandler.Update)
	giftCardsGroup.DELETE("/:id/shares/:share_id", giftCardSharesHandler.Delete)
	giftCardsGroup.GET("/:id/shares/new-inline", giftCardSharesHandler.NewInline)
	giftCardsGroup.GET("/:id/shares/cancel", giftCardSharesHandler.Cancel)
	giftCardsGroup.GET("/:id/shares/:share_id/edit-inline", giftCardSharesHandler.EditInline)
	giftCardsGroup.GET("/:id/shares/:share_id/cancel-edit", giftCardSharesHandler.CancelEdit)
	// Favorites (polymorphic pinning)
	giftCardsGroup.POST("/:id/favorite", favoritesHandler.ToggleGiftCardFavorite)

	// ========================================
	// Impersonation Management
	// Stop impersonation route (must be outside admin group - impersonated users aren't admins)
	// ========================================
	protected.GET("/admin/stop-impersonate", handlers.AdminStopImpersonate)

	// ========================================
	// Admin Routes (Admin-Only Panel)
	// Middleware: RequireAdmin
	// Features: User Management, Audit Log, Impersonation, Role Management
	// ========================================
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAdmin)
	admin.GET("/users", handlers.AdminUsersIndex)
	admin.GET("/users/create", handlers.AdminCreateUserGet)   // Only available if local login enabled
	admin.POST("/users/create", handlers.AdminCreateUserPost) // Only available if local login enabled
	admin.POST("/users/:id/role", handlers.AdminUpdateUserRole)
	admin.GET("/audit-log", handlers.AdminAuditLogIndex)
	admin.POST("/audit-log/restore", handlers.AdminRestoreResource)
	admin.GET("/impersonate/:id", handlers.AdminImpersonate)

	// ========================================
	// Development Debug Tools
	// ========================================
	// Print all registered routes in development mode
	if !cfg.IsProduction() {
		debug.PrintRoutes(e)
	}

	// Start metrics collector goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			updateMetrics()
		}
	}()

	// Start server with graceful shutdown
	slog.Info("Server starting", "port", cfg.ServerPort)

	// Start server in a goroutine
	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil {
			slog.Info("Server shutdown", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
		return 1
	}

	log.Println("Server gracefully stopped")
	return 0
}
