// Package metrics provides Prometheus metrics collection for HTTP requests, database connections, and application resources.
package metrics //nolint:revive // "metrics" does not conflict with any Go standard library package

import (
	"time"

	"github.com/labstack/echo/v5"
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

	// Note: app_errors_total and active_sessions removed (unused/not implemented)

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

	// Status Breakdown Metrics
	vouchersByStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vouchers_by_status",
			Help: "Number of vouchers by status (active, expired)",
		},
		[]string{"status"},
	)

	giftCardsByStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gift_cards_by_status",
			Help: "Number of gift cards by computed status (active, expired, redeemed)",
		},
		[]string{"status"},
	)

	// Sharing Metrics
	sharesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shares_total",
			Help: "Total number of active shares by resource type",
		},
		[]string{"resource_type"},
	)

	// Notification Metrics
	pushSubscriptionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "push_subscriptions_total",
			Help: "Total number of active push subscriptions",
		},
	)

	pushSubscribedUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "push_subscribed_users_total",
			Help: "Number of users with at least one push subscription",
		},
	)

	emailVerifiedUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "email_verified_users_total",
			Help: "Number of users with verified email addresses",
		},
	)

	// Channel toggle metrics
	pushNotificationsEnabledTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "push_notifications_enabled_total",
			Help: "Number of users with push notifications channel enabled",
		},
	)

	emailNotificationsEnabledTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "email_notifications_enabled_total",
			Help: "Number of users with email notifications channel enabled",
		},
	)

	// Per-channel category metrics
	pushRemindersEnabledTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "push_reminders_enabled_total",
			Help: "Number of users with push reminders (expiry/validity) enabled",
		},
	)

	pushSharingEnabledTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "push_sharing_enabled_total",
			Help: "Number of users with push sharing notifications enabled",
		},
	)

	emailRemindersEnabledTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "email_reminders_enabled_total",
			Help: "Number of users with email reminders (expiry/validity) enabled",
		},
	)

	emailSharingEnabledTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "email_sharing_enabled_total",
			Help: "Number of users with email sharing notifications enabled",
		},
	)

	// Notification Email Delivery Metrics
	notificationEmailsSentTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_emails_sent_total",
			Help: "Total number of notification emails delivered successfully",
		},
	)

	notificationEmailsFailedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_emails_failed_total",
			Help: "Total number of notification email delivery attempts that failed",
		},
	)

	// A rising pending gauge is the only signal that separates a stalled
	// dispatcher from an idle one — both send zero mails per minute.
	notificationEmailsPending = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_emails_pending",
			Help: "Number of notification emails waiting to be delivered",
		},
	)

	// 'failed' is terminal and nothing requeues it, so a parked row is silent
	// data loss unless it is counted. Any non-zero value needs an operator.
	notificationEmailsParked = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_emails_parked",
			Help: "Number of notification emails parked as permanently failed",
		},
	)

	// Authentication Metrics
	loginAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_attempts_total",
			Help: "Total number of login attempts by result",
		},
		[]string{"result"},
	)
)

// Middleware records HTTP request metrics
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			// Call the next handler
			handlerErr := next(c)

			// Record metrics
			duration := time.Since(start).Seconds()
			_, status := echo.ResolveResponseStatus(c.Response(), handlerErr)
			method := c.Request().Method
			path := c.Path()

			// Normalize path to avoid high cardinality
			// c.Path() returns the route pattern (e.g. "/api/v1/cards/:id")
			// If empty (unmatched route), use a generic label instead of the raw URL
			if path == "" {
				path = "/unmatched"
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

			return handlerErr
		}
	}
}

// UpdateDBMetrics updates database connection pool metrics
func UpdateDBMetrics(active, idle int) {
	dbConnectionsActive.Set(float64(active))
	dbConnectionsIdle.Set(float64(idle))
}

// UpdateResourceCounts updates resource count gauges
// Note: users count removed for privacy/DSGVO compliance
func UpdateResourceCounts(cards, vouchers, giftCards int64) {
	cardsTotal.Set(float64(cards))
	vouchersTotal.Set(float64(vouchers))
	giftCardsTotal.Set(float64(giftCards))
}

// UpdateVouchersByStatus updates voucher counts by status
func UpdateVouchersByStatus(active, expired int64) {
	vouchersByStatus.WithLabelValues("active").Set(float64(active))
	vouchersByStatus.WithLabelValues("expired").Set(float64(expired))
}

// UpdateGiftCardsByStatus updates gift card counts by computed status
func UpdateGiftCardsByStatus(active, expired, redeemed int64) {
	giftCardsByStatus.WithLabelValues("active").Set(float64(active))
	giftCardsByStatus.WithLabelValues("expired").Set(float64(expired))
	giftCardsByStatus.WithLabelValues("redeemed").Set(float64(redeemed))
}

// UpdateSharesCounts updates share counts by resource type
func UpdateSharesCounts(cards, vouchers, giftCards int64) {
	sharesTotal.WithLabelValues("card").Set(float64(cards))
	sharesTotal.WithLabelValues("voucher").Set(float64(vouchers))
	sharesTotal.WithLabelValues("gift_card").Set(float64(giftCards))
}

// NotificationMetrics holds all notification-related metric values.
type NotificationMetrics struct {
	PushSubscriptions         int64
	PushSubscribedUsers       int64
	EmailVerifiedUsers        int64
	PushNotificationsEnabled  int64
	EmailNotificationsEnabled int64
	PushRemindersEnabled      int64
	PushSharingEnabled        int64
	EmailRemindersEnabled     int64
	EmailSharingEnabled       int64
}

// UpdateNotificationMetrics updates notification-related gauges
func UpdateNotificationMetrics(m NotificationMetrics) {
	pushSubscriptionsTotal.Set(float64(m.PushSubscriptions))
	pushSubscribedUsersTotal.Set(float64(m.PushSubscribedUsers))
	emailVerifiedUsersTotal.Set(float64(m.EmailVerifiedUsers))
	pushNotificationsEnabledTotal.Set(float64(m.PushNotificationsEnabled))
	emailNotificationsEnabledTotal.Set(float64(m.EmailNotificationsEnabled))
	pushRemindersEnabledTotal.Set(float64(m.PushRemindersEnabled))
	pushSharingEnabledTotal.Set(float64(m.PushSharingEnabled))
	emailRemindersEnabledTotal.Set(float64(m.EmailRemindersEnabled))
	emailSharingEnabledTotal.Set(float64(m.EmailSharingEnabled))
}

// RecordNotificationEmailResult counts one delivery attempt by outcome.
func RecordNotificationEmailResult(sent, failed int) {
	if sent > 0 {
		notificationEmailsSentTotal.Add(float64(sent))
	}
	if failed > 0 {
		notificationEmailsFailedTotal.Add(float64(failed))
	}
}

// UpdateNotificationEmailsPending sets the delivery backlog gauge.
func UpdateNotificationEmailsPending(pending int64) {
	notificationEmailsPending.Set(float64(pending))
}

// UpdateNotificationEmailsParked sets the permanently-failed gauge.
func UpdateNotificationEmailsParked(parked int64) {
	notificationEmailsParked.Set(float64(parked))
}

// RecordLoginAttempt increments the login attempts counter
func RecordLoginAttempt(result string) {
	loginAttemptsTotal.WithLabelValues(result).Inc()
}
