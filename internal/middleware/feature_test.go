package middleware

import (
	"net/http"
	"net/http/httptest"
	"savvy/internal/config"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRequireCardsEnabled_Enabled(t *testing.T) {
	cfg := &config.Config{
		EnableCards: true,
	}

	e := echo.New()
	e.Use(RequireCardsEnabled(cfg))
	e.GET("/cards", func(c echo.Context) error {
		return c.String(http.StatusOK, "Cards page")
	})

	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Cards page", rec.Body.String())
}

func TestRequireCardsEnabled_Disabled(t *testing.T) {
	cfg := &config.Config{
		EnableCards: false,
	}

	e := echo.New()
	e.Use(RequireCardsEnabled(cfg))
	e.GET("/cards", func(c echo.Context) error {
		return c.String(http.StatusOK, "Cards page")
	})

	req := httptest.NewRequest(http.MethodGet, "/cards", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Cards feature is disabled")
}

func TestRequireVouchersEnabled_Enabled(t *testing.T) {
	cfg := &config.Config{
		EnableVouchers: true,
	}

	e := echo.New()
	e.Use(RequireVouchersEnabled(cfg))
	e.GET("/vouchers", func(c echo.Context) error {
		return c.String(http.StatusOK, "Vouchers page")
	})

	req := httptest.NewRequest(http.MethodGet, "/vouchers", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Vouchers page", rec.Body.String())
}

func TestRequireVouchersEnabled_Disabled(t *testing.T) {
	cfg := &config.Config{
		EnableVouchers: false,
	}

	e := echo.New()
	e.Use(RequireVouchersEnabled(cfg))
	e.GET("/vouchers", func(c echo.Context) error {
		return c.String(http.StatusOK, "Vouchers page")
	})

	req := httptest.NewRequest(http.MethodGet, "/vouchers", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Vouchers feature is disabled")
}

func TestRequireGiftCardsEnabled_Enabled(t *testing.T) {
	cfg := &config.Config{
		EnableGiftCards: true,
	}

	e := echo.New()
	e.Use(RequireGiftCardsEnabled(cfg))
	e.GET("/gift-cards", func(c echo.Context) error {
		return c.String(http.StatusOK, "Gift cards page")
	})

	req := httptest.NewRequest(http.MethodGet, "/gift-cards", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Gift cards page", rec.Body.String())
}

func TestRequireGiftCardsEnabled_Disabled(t *testing.T) {
	cfg := &config.Config{
		EnableGiftCards: false,
	}

	e := echo.New()
	e.Use(RequireGiftCardsEnabled(cfg))
	e.GET("/gift-cards", func(c echo.Context) error {
		return c.String(http.StatusOK, "Gift cards page")
	})

	req := httptest.NewRequest(http.MethodGet, "/gift-cards", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Gift cards feature is disabled")
}

func TestRequireLocalLoginEnabled_Enabled(t *testing.T) {
	cfg := &config.Config{
		EnableLocalLogin: true,
	}

	e := echo.New()
	e.Use(RequireLocalLoginEnabled(cfg))
	e.GET("/login", func(c echo.Context) error {
		return c.String(http.StatusOK, "Login page")
	})

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Login page", rec.Body.String())
}

func TestRequireLocalLoginEnabled_Disabled(t *testing.T) {
	cfg := &config.Config{
		EnableLocalLogin: false,
	}

	e := echo.New()
	e.Use(RequireLocalLoginEnabled(cfg))
	e.GET("/login", func(c echo.Context) error {
		return c.String(http.StatusOK, "Login page")
	})

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Local login is disabled")
}

func TestRequireRegistrationEnabled_Enabled(t *testing.T) {
	cfg := &config.Config{
		EnableRegistration: true,
	}

	e := echo.New()
	e.Use(RequireRegistrationEnabled(cfg))
	e.GET("/register", func(c echo.Context) error {
		return c.String(http.StatusOK, "Registration page")
	})

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Registration page", rec.Body.String())
}

func TestRequireRegistrationEnabled_Disabled(t *testing.T) {
	cfg := &config.Config{
		EnableRegistration: false,
	}

	e := echo.New()
	e.Use(RequireRegistrationEnabled(cfg))
	e.GET("/register", func(c echo.Context) error {
		return c.String(http.StatusOK, "Registration page")
	})

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Registration disabled should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestMultipleFeatureToggles(t *testing.T) {
	cfg := &config.Config{
		EnableCards:     true,
		EnableVouchers:  false,
		EnableGiftCards: true,
	}

	e := echo.New()

	// Cards route (enabled)
	cardsGroup := e.Group("/cards", RequireCardsEnabled(cfg))
	cardsGroup.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Cards")
	})

	// Vouchers route (disabled)
	vouchersGroup := e.Group("/vouchers", RequireVouchersEnabled(cfg))
	vouchersGroup.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Vouchers")
	})

	// Gift cards route (enabled)
	giftCardsGroup := e.Group("/gift-cards", RequireGiftCardsEnabled(cfg))
	giftCardsGroup.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Gift Cards")
	})

	// Test cards (should work)
	req1 := httptest.NewRequest(http.MethodGet, "/cards", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Test vouchers (should fail)
	req2 := httptest.NewRequest(http.MethodGet, "/vouchers", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusForbidden, rec2.Code)

	// Test gift cards (should work)
	req3 := httptest.NewRequest(http.MethodGet, "/gift-cards", nil)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code)
}

func TestFeatureToggle_AllDisabled(t *testing.T) {
	cfg := &config.Config{
		EnableCards:        false,
		EnableVouchers:     false,
		EnableGiftCards:    false,
		EnableLocalLogin:   false,
		EnableRegistration: false,
	}

	e := echo.New()

	routes := []struct {
		path       string
		middleware echo.MiddlewareFunc
	}{
		{"/cards", RequireCardsEnabled(cfg)},
		{"/vouchers", RequireVouchersEnabled(cfg)},
		{"/gift-cards", RequireGiftCardsEnabled(cfg)},
		{"/login", RequireLocalLoginEnabled(cfg)},
		{"/register", RequireRegistrationEnabled(cfg)},
	}

	for _, route := range routes {
		e.GET(route.path, func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}, route.middleware)

		req := httptest.NewRequest(http.MethodGet, route.path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	}
}

func TestFeatureToggle_AllEnabled(t *testing.T) {
	cfg := &config.Config{
		EnableCards:        true,
		EnableVouchers:     true,
		EnableGiftCards:    true,
		EnableLocalLogin:   true,
		EnableRegistration: true,
	}

	e := echo.New()

	routes := []struct {
		path       string
		middleware echo.MiddlewareFunc
	}{
		{"/cards", RequireCardsEnabled(cfg)},
		{"/vouchers", RequireVouchersEnabled(cfg)},
		{"/gift-cards", RequireGiftCardsEnabled(cfg)},
		{"/login", RequireLocalLoginEnabled(cfg)},
		{"/register", RequireRegistrationEnabled(cfg)},
	}

	for _, route := range routes {
		e.GET(route.path, func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		}, route.middleware)

		req := httptest.NewRequest(http.MethodGet, route.path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
