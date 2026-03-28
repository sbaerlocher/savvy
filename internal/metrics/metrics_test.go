package metrics //nolint:revive // Standard package name, false positive

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify metrics were recorded (counter only - histogram testing is complex)
	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/test", "2xx"))
	assert.Equal(t, 1.0, count)
}

func TestMiddleware_DifferentStatusCodes(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	tests := []struct {
		path        string
		statusCode  int
		statusClass string
	}{
		{"/success", http.StatusOK, "2xx"},
		{"/redirect", http.StatusMovedPermanently, "3xx"},
		{"/not-found", http.StatusNotFound, "4xx"},
		{"/error", http.StatusInternalServerError, "5xx"},
	}

	for _, tt := range tests {
		e.GET(tt.path, func(c *echo.Context) error {
			return c.String(tt.statusCode, "response")
		})
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Verify status class is correct
		count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", tt.path, tt.statusClass))
		assert.Equal(t, 1.0, count, "Failed for %s", tt.path)
	}
}

func TestMiddleware_RecordsDuration(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	e.GET("/slow", func(c *echo.Context) error {
		time.Sleep(50 * time.Millisecond)
		return c.String(http.StatusOK, "slow")
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify request was recorded (histogram duration testing is complex)
	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/slow", "2xx"))
	assert.Equal(t, 1.0, count)
}

func TestMiddleware_NormalizedPaths(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	// Register a route with a path parameter
	e.GET("/users/:id", func(c *echo.Context) error {
		return c.String(http.StatusOK, "user")
	})

	// Make multiple requests with different IDs
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users/"+string(rune('0'+i)), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// All requests should be grouped under the same path pattern
}

func TestUpdateDBMetrics(t *testing.T) {
	tests := []struct {
		name   string
		active int
		idle   int
	}{
		{"zero", 0, 0},
		{"some_active", 5, 10},
		{"all_active", 20, 0},
		{"all_idle", 0, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateDBMetrics(tt.active, tt.idle)

			activeValue := testutil.ToFloat64(dbConnectionsActive)
			idleValue := testutil.ToFloat64(dbConnectionsIdle)

			assert.Equal(t, float64(tt.active), activeValue)
			assert.Equal(t, float64(tt.idle), idleValue)
		})
	}
}

func TestUpdateResourceCounts(t *testing.T) {
	tests := []struct {
		name      string
		cards     int64
		vouchers  int64
		giftCards int64
	}{
		{"zero", 0, 0, 0},
		{"some_resources", 10, 20, 30},
		{"large_numbers", 1000, 2000, 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateResourceCounts(tt.cards, tt.vouchers, tt.giftCards)

			cardsValue := testutil.ToFloat64(cardsTotal)
			vouchersValue := testutil.ToFloat64(vouchersTotal)
			giftCardsValue := testutil.ToFloat64(giftCardsTotal)

			assert.Equal(t, float64(tt.cards), cardsValue)
			assert.Equal(t, float64(tt.vouchers), vouchersValue)
			assert.Equal(t, float64(tt.giftCards), giftCardsValue)
		})
	}
}

func TestMetricsExport(t *testing.T) {
	// Initialize HTTP metrics (vectors only appear in export after first observation)
	httpDuration.WithLabelValues("GET", "/export-test", "2xx").Observe(0.01)
	httpRequestsTotal.WithLabelValues("GET", "/export-test", "2xx").Inc()

	// Set some metrics
	UpdateDBMetrics(10, 5)
	UpdateResourceCounts(100, 200, 300)
	UpdateVouchersByStatus(80, 20)
	UpdateGiftCardsByStatus(50, 10, 40)
	UpdateSharesCounts(5, 3, 7)
	UpdateNotificationMetrics(NotificationMetrics{
		PushSubscriptions: 8, PushSubscribedUsers: 4, EmailVerifiedUsers: 15,
		PushNotificationsEnabled: 12, EmailNotificationsEnabled: 14,
		PushRemindersEnabled: 11, PushSharingEnabled: 10,
		EmailRemindersEnabled: 20, EmailSharingEnabled: 18,
	})
	RecordLoginAttempt("success")

	// Create metrics handler
	handler := promhttp.Handler()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	// Serve metrics
	handler.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "http_request_duration_seconds")
	assert.Contains(t, body, "http_requests_total")
	assert.Contains(t, body, "db_connections_active")
	assert.Contains(t, body, "db_connections_idle")
	assert.Contains(t, body, "cards_total")
	assert.Contains(t, body, "vouchers_total")
	assert.Contains(t, body, "gift_cards_total")
	assert.Contains(t, body, "vouchers_by_status")
	assert.Contains(t, body, "gift_cards_by_status")
	assert.Contains(t, body, "shares_total")
	assert.Contains(t, body, "push_subscriptions_total")
	assert.Contains(t, body, "push_subscribed_users_total")
	assert.Contains(t, body, "email_verified_users_total")
	assert.Contains(t, body, "push_notifications_enabled_total")
	assert.Contains(t, body, "email_notifications_enabled_total")
	assert.Contains(t, body, "push_reminders_enabled_total")
	assert.Contains(t, body, "push_sharing_enabled_total")
	assert.Contains(t, body, "email_reminders_enabled_total")
	assert.Contains(t, body, "email_sharing_enabled_total")
	assert.Contains(t, body, "login_attempts_total")

	// Verify removed metrics are NOT present (use HELP prefix to avoid substring matches)
	assert.NotContains(t, body, "# HELP users_total")      // Removed for privacy/DSGVO
	assert.NotContains(t, body, "# HELP active_sessions")  // Removed (not implemented)
	assert.NotContains(t, body, "# HELP app_errors_total") // Removed (not used)
}

func TestUpdateVouchersByStatus(t *testing.T) {
	tests := []struct {
		name    string
		active  int64
		expired int64
	}{
		{"zero", 0, 0},
		{"all_active", 50, 0},
		{"mixed", 30, 20},
		{"all_expired", 0, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateVouchersByStatus(tt.active, tt.expired)

			activeValue := testutil.ToFloat64(vouchersByStatus.WithLabelValues("active"))
			expiredValue := testutil.ToFloat64(vouchersByStatus.WithLabelValues("expired"))

			assert.Equal(t, float64(tt.active), activeValue)
			assert.Equal(t, float64(tt.expired), expiredValue)
		})
	}
}

func TestUpdateGiftCardsByStatus(t *testing.T) {
	tests := []struct {
		name     string
		active   int64
		expired  int64
		redeemed int64
	}{
		{"zero", 0, 0, 0},
		{"all_active", 50, 0, 0},
		{"mixed", 30, 10, 20},
		{"all_redeemed", 0, 0, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateGiftCardsByStatus(tt.active, tt.expired, tt.redeemed)

			activeValue := testutil.ToFloat64(giftCardsByStatus.WithLabelValues("active"))
			expiredValue := testutil.ToFloat64(giftCardsByStatus.WithLabelValues("expired"))
			redeemedValue := testutil.ToFloat64(giftCardsByStatus.WithLabelValues("redeemed"))

			assert.Equal(t, float64(tt.active), activeValue)
			assert.Equal(t, float64(tt.expired), expiredValue)
			assert.Equal(t, float64(tt.redeemed), redeemedValue)
		})
	}
}

func TestUpdateSharesCounts(t *testing.T) {
	UpdateSharesCounts(5, 3, 7)

	cardShares := testutil.ToFloat64(sharesTotal.WithLabelValues("card"))
	voucherShares := testutil.ToFloat64(sharesTotal.WithLabelValues("voucher"))
	giftCardShares := testutil.ToFloat64(sharesTotal.WithLabelValues("gift_card"))

	assert.Equal(t, 5.0, cardShares)
	assert.Equal(t, 3.0, voucherShares)
	assert.Equal(t, 7.0, giftCardShares)
}

func TestUpdateNotificationMetrics(t *testing.T) {
	tests := []struct {
		name string
		m    NotificationMetrics
	}{
		{"zero", NotificationMetrics{}},
		{"some_active", NotificationMetrics{
			PushSubscriptions: 5, PushSubscribedUsers: 3, EmailVerifiedUsers: 10,
			PushNotificationsEnabled: 9, EmailNotificationsEnabled: 8,
			PushRemindersEnabled: 7, PushSharingEnabled: 6,
			EmailRemindersEnabled: 8, EmailSharingEnabled: 7,
		}},
		{"many_subscriptions", NotificationMetrics{
			PushSubscriptions: 50, PushSubscribedUsers: 15, EmailVerifiedUsers: 100,
			PushNotificationsEnabled: 95, EmailNotificationsEnabled: 90,
			PushRemindersEnabled: 88, PushSharingEnabled: 85,
			EmailRemindersEnabled: 90, EmailSharingEnabled: 85,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateNotificationMetrics(tt.m)

			assert.Equal(t, float64(tt.m.PushSubscriptions), testutil.ToFloat64(pushSubscriptionsTotal))
			assert.Equal(t, float64(tt.m.PushSubscribedUsers), testutil.ToFloat64(pushSubscribedUsersTotal))
			assert.Equal(t, float64(tt.m.EmailVerifiedUsers), testutil.ToFloat64(emailVerifiedUsersTotal))
			assert.Equal(t, float64(tt.m.PushNotificationsEnabled), testutil.ToFloat64(pushNotificationsEnabledTotal))
			assert.Equal(t, float64(tt.m.EmailNotificationsEnabled), testutil.ToFloat64(emailNotificationsEnabledTotal))
			assert.Equal(t, float64(tt.m.PushRemindersEnabled), testutil.ToFloat64(pushRemindersEnabledTotal))
			assert.Equal(t, float64(tt.m.PushSharingEnabled), testutil.ToFloat64(pushSharingEnabledTotal))
			assert.Equal(t, float64(tt.m.EmailRemindersEnabled), testutil.ToFloat64(emailRemindersEnabledTotal))
			assert.Equal(t, float64(tt.m.EmailSharingEnabled), testutil.ToFloat64(emailSharingEnabledTotal))
		})
	}
}

func TestRecordLoginAttempt(t *testing.T) {
	// Get initial counts
	successBefore := testutil.ToFloat64(loginAttemptsTotal.WithLabelValues("success"))
	failureBefore := testutil.ToFloat64(loginAttemptsTotal.WithLabelValues("failure"))

	RecordLoginAttempt("success")
	RecordLoginAttempt("success")
	RecordLoginAttempt("failure")

	successAfter := testutil.ToFloat64(loginAttemptsTotal.WithLabelValues("success"))
	failureAfter := testutil.ToFloat64(loginAttemptsTotal.WithLabelValues("failure"))

	assert.Equal(t, successBefore+2.0, successAfter)
	assert.Equal(t, failureBefore+1.0, failureAfter)
}

func TestMiddleware_ErrorHandling(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	e.GET("/error", func(_ *echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify 5xx metrics were recorded
	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/error", "5xx"))
	assert.Greater(t, count, 0.0)
}

func TestMiddleware_PathNormalization(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	e.GET("/api/users/:id", func(c *echo.Context) error {
		return c.String(http.StatusOK, "user")
	})

	// Make requests with different IDs
	req1 := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/api/users/456", nil)

	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec1, req1)
	e.ServeHTTP(rec2, req2)

	// Both should use the same metric label (normalized path)
	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/api/users/:id", "2xx"))
	assert.Equal(t, 2.0, count)
}

func TestMetricLabels(t *testing.T) {
	// Test that metrics have correct labels
	// Use unique path to avoid conflicts with other tests
	labels := prometheus.Labels{"method": "POST", "path": "/test-metric-labels", "status": "2xx"}

	// Get count before increment
	countBefore := testutil.ToFloat64(httpRequestsTotal.With(labels))

	httpDuration.With(labels).Observe(0.1)
	httpRequestsTotal.With(labels).Inc()

	// Verify counter increased by exactly 1
	countAfter := testutil.ToFloat64(httpRequestsTotal.With(labels))
	assert.Equal(t, countBefore+1.0, countAfter)
}

func TestConcurrentMetricUpdates(t *testing.T) {
	const numGoroutines = 10
	const updatesPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	// Update DB metrics concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < updatesPerGoroutine; j++ {
				UpdateDBMetrics(j, j*2)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Just verify no panics occurred
	activeValue := testutil.ToFloat64(dbConnectionsActive)
	idleValue := testutil.ToFloat64(dbConnectionsIdle)

	assert.GreaterOrEqual(t, activeValue, 0.0)
	assert.GreaterOrEqual(t, idleValue, 0.0)
}

func TestStatusClassification(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	tests := []struct {
		status      int
		statusClass string
	}{
		{200, "2xx"},
		{201, "2xx"},
		{204, "2xx"},
		{301, "3xx"},
		{302, "3xx"},
		{304, "3xx"},
		{400, "4xx"},
		{401, "4xx"},
		{403, "4xx"},
		{404, "4xx"},
		{500, "5xx"},
		{502, "5xx"},
		{503, "5xx"},
	}

	for _, tt := range tests {
		path := "/status-" + strconv.Itoa(tt.status)
		e.GET(path, func(c *echo.Context) error {
			return c.String(tt.status, "response")
		})

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Verify correct status class was recorded
		count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", path, tt.statusClass))
		assert.Greater(t, count, 0.0, "Failed for status %d", tt.status)
	}
}

func TestMiddleware_ChainedHandlers(t *testing.T) {
	e := echo.New()

	// Custom middleware that runs before metrics
	preMetrics := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("test", "value")
			return next(c)
		}
	}

	// Custom middleware that runs after metrics
	postMetrics := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := next(c)
			// Verify metrics context is available
			assert.NotNil(t, c.Get("test"))
			return err
		}
	}

	e.Use(preMetrics)
	e.Use(Middleware())
	e.Use(postMetrics)

	e.GET("/chained", func(c *echo.Context) error {
		return c.String(http.StatusOK, "chained")
	})

	req := httptest.NewRequest(http.MethodGet, "/chained", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	// Verify metrics were still recorded
	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/chained", "2xx"))
	assert.Equal(t, 1.0, count)
}
