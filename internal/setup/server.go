// Package setup contains setup logic for initializing the Echo server.
package setup

import (
	"context"
	"io/fs"
	"log"
	"log/slog"
	"savvy/internal/assets"
	"savvy/internal/config"
	"savvy/internal/database"
	"savvy/internal/handlers"
	"savvy/internal/metrics"
	"savvy/internal/middleware"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// ServerConfig holds the configuration for server setup.
type ServerConfig struct {
	Config       *config.Config
	HealthHandler *handlers.HealthHandler
}

// NewEchoServer creates and configures a new Echo server instance.
func NewEchoServer(sc *ServerConfig) *echo.Echo {
	e := echo.New()

	// Configure middleware
	configureMiddleware(e, sc)

	// Configure observability endpoints (public, before auth middleware)
	configureObservabilityEndpoints(e, sc)

	// Configure authentication middleware
	configureAuthMiddleware(e, sc)

	// Configure static files and public endpoints
	configureStaticFiles(e)
	configurePublicEndpoints(e)

	return e
}

// configureMiddleware sets up all middleware for the Echo server.
func configureMiddleware(e *echo.Echo, sc *ServerConfig) {
	cfg := sc.Config

	// OpenTelemetry Middleware (must be first for proper tracing)
	if cfg.OTelEnabled {
		e.Use(otelecho.Middleware(cfg.ServiceName))
		e.Use(middleware.OTelLogger()) // Add trace IDs to logs
	}

	// Request logging
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

	// Recovery middleware
	e.Use(echomiddleware.Recover())

	// Prometheus metrics
	e.Use(metrics.Middleware())

	// CSRF Protection (only for non-GET requests)
	e.Use(echomiddleware.CSRFWithConfig(echomiddleware.CSRFConfig{
		TokenLookup:    "form:csrf_token,header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookieHTTPOnly: true,
		CookieSecure:   cfg.IsProduction(), // true in production, false in dev
		CookieSameSite: 2,                  // SameSiteLaxMode
		ContextKey:     "csrf",             // Store token in context under "csrf" key
	}))
}

// configureObservabilityEndpoints registers health and metrics endpoints.
func configureObservabilityEndpoints(e *echo.Echo, sc *ServerConfig) {
	e.GET("/health", sc.HealthHandler.Health)
	e.GET("/ready", sc.HealthHandler.Ready)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}

// configureAuthMiddleware sets up authentication and session middleware.
func configureAuthMiddleware(e *echo.Echo, sc *ServerConfig) {
	cfg := sc.Config

	e.Use(middleware.SetCurrentUser)
	e.Use(middleware.SessionTracking) // Track active sessions for metrics
	e.Use(middleware.LanguageDetection)

	// Set service version and config in context
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
}

// configureStaticFiles sets up static file serving.
func configureStaticFiles(e *echo.Echo) {
	staticFS, err := fs.Sub(assets.Static, "static")
	if err != nil {
		log.Fatalf("Failed to create static filesystem: %v", err)
	}
	e.StaticFS("/static", staticFS)
}

// configurePublicEndpoints registers public endpoints (language switcher, PWA).
func configurePublicEndpoints(e *echo.Echo) {
	// Language switcher endpoint (public)
	e.GET("/set-language", middleware.SetLanguage)

	// PWA endpoints (public)
	e.GET("/offline", handlers.OfflinePage)                                   // Offline fallback page
	e.FileFS("/service-worker.js", "static/service-worker.js", assets.Static) // Service Worker
	e.FileFS("/manifest.json", "static/manifest.json", assets.Static)         // PWA Manifest
}

// StartMetricsCollector starts a goroutine that periodically updates metrics.
func StartMetricsCollector() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			updateMetrics()
		}
	}()
}

// updateMetrics updates Prometheus gauges for resource counts and DB stats.
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
