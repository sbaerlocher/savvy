// Package merchants contains HTTP request handlers for merchant management.
package merchants

import (
	"savvy/internal/services"
)

// Handler handles HTTP requests for merchant operations.
type Handler struct {
	merchantService services.MerchantServiceInterface
}

// NewHandler creates a new merchant handler with the provided services.
func NewHandler(
	merchantService services.MerchantServiceInterface,
) *Handler {
	return &Handler{
		merchantService: merchantService,
	}
}
