// Package middleware provides Echo middleware for feature toggling.
package middleware

import (
	"net/http"
	"savvy/internal/config"

	"github.com/labstack/echo/v4"
)

// RequireCardsEnabled middleware requires cards feature to be enabled
func RequireCardsEnabled(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !cfg.EnableCards {
				return echo.NewHTTPError(http.StatusNotFound, "Cards feature is disabled")
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
				return echo.NewHTTPError(http.StatusNotFound, "Vouchers feature is disabled")
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
				return echo.NewHTTPError(http.StatusNotFound, "Gift cards feature is disabled")
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
				return echo.NewHTTPError(http.StatusNotFound, "Local login is disabled")
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
				// Redirect to login page instead of showing error
				return c.Redirect(http.StatusSeeOther, "/auth/login")
			}
			return next(c)
		}
	}
}
