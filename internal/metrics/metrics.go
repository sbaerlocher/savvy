package metrics

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Application Metrics
	errorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_errors_total",
			Help: "Total application errors",
		},
		[]string{"handler", "error_type"},
	)

	activeSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "Number of active user sessions",
		},
	)

	// Database Metrics
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// Resource Counts
	cardsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cards_total",
			Help: "Total number of savvy cards",
		},
	)

	vouchersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "vouchers_total",
			Help: "Total number of vouchers",
		},
	)

	giftCardsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gift_cards_total",
			Help: "Total number of gift cards",
		},
	)

	usersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_total",
			Help: "Total number of registered users",
		},
	)
)

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Call the next handler
			err := next(c)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := c.Response().Status
			method := c.Request().Method
			path := c.Path()

			// Normalize path to avoid high cardinality
			if path == "" {
				path = c.Request().URL.Path
			}

			// Determine status class (2xx, 3xx, 4xx, 5xx)
			statusClass := "2xx"
			if status >= 500 {
				statusClass = "5xx"
			} else if status >= 400 {
				statusClass = "4xx"
			} else if status >= 300 {
				statusClass = "3xx"
			}

			// Record histogram and counter
			httpDuration.WithLabelValues(method, path, statusClass).Observe(duration)
			httpRequestsTotal.WithLabelValues(method, path, statusClass).Inc()

			return err
		}
	}
}

// RecordError records an application error
func RecordError(handler, errorType string) {
	errorCount.WithLabelValues(handler, errorType).Inc()
}

// SetActiveSessions updates the active sessions gauge
func SetActiveSessions(count float64) {
	activeSessions.Set(count)
}

// IncrementActiveSessions increments the active sessions counter
func IncrementActiveSessions() {
	activeSessions.Inc()
}

// DecrementActiveSessions decrements the active sessions counter
func DecrementActiveSessions() {
	activeSessions.Dec()
}

// UpdateDBMetrics updates database connection pool metrics
func UpdateDBMetrics(active, idle int) {
	dbConnectionsActive.Set(float64(active))
	dbConnectionsIdle.Set(float64(idle))
}

// UpdateResourceCounts updates resource count gauges
func UpdateResourceCounts(cards, vouchers, giftCards, users int64) {
	cardsTotal.Set(float64(cards))
	vouchersTotal.Set(float64(vouchers))
	giftCardsTotal.Set(float64(giftCards))
	usersTotal.Set(float64(users))
}
