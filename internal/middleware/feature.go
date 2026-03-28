// Package middleware provides Echo middleware for feature toggling.
package middleware

import (
	"net/http"
	"savvy/internal/config"

	"github.com/labstack/echo/v5"
)

// RequireCardsEnabled middleware requires cards feature to be enabled
func RequireCardsEnabled(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !cfg.EnableCards {
				return echo.NewHTTPError(http.StatusForbidden, "Cards feature is disabled")
			}
			return next(c)
		}
	}
}

// RequireVouchersEnabled middleware requires vouchers feature to be enabled
func RequireVouchersEnabled(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !cfg.EnableVouchers {
				return echo.NewHTTPError(http.StatusForbidden, "Vouchers feature is disabled")
			}
			return next(c)
		}
	}
}

// RequireGiftCardsEnabled middleware requires gift cards feature to be enabled
func RequireGiftCardsEnabled(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !cfg.EnableGiftCards {
				return echo.NewHTTPError(http.StatusForbidden, "Gift cards feature is disabled")
			}
			return next(c)
		}
	}
}

// RequireLocalLoginEnabled middleware requires local login to be enabled
func RequireLocalLoginEnabled(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !cfg.EnableLocalLogin {
				return echo.NewHTTPError(http.StatusForbidden, "Local login is disabled")
			}
			return next(c)
		}
	}
}

// RequireRegistrationEnabled middleware requires registration to be enabled
func RequireRegistrationEnabled(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !cfg.EnableRegistration {
				return echo.NewHTTPError(http.StatusForbidden, "Registration is disabled")
			}
			return next(c)
		}
	}
}
