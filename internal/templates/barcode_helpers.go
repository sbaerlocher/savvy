// Package templates provides helper functions for barcode template rendering.
package templates

import (
	"context"
	"savvy/internal/middleware"
	"savvy/internal/models"
	"savvy/internal/security"
	"time"

	"github.com/google/uuid"
)

// GenerateBarcodeToken creates a secure token for barcode access
// The token expires after 7 days (matches Service Worker cache duration for offline support)
func GenerateBarcodeToken(ctx context.Context, resourceID uuid.UUID, resourceType string) string {
	// Get user from context
	user, ok := ctx.Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		// If no user in context, return empty string (barcode won't render)
		return ""
	}

	// Generate token valid for 7 days (matches PWA cache duration)
	token, err := security.GenerateBarcodeToken(resourceID, resourceType, user.ID, 7*24*time.Hour)
	if err != nil {
		// Log error but don't expose to template
		return ""
	}

	return token
}

// GenerateCardBarcodeToken generates a token for a card barcode
func GenerateCardBarcodeToken(ctx context.Context, cardID uuid.UUID) string {
	return GenerateBarcodeToken(ctx, cardID, "card")
}

// GenerateVoucherBarcodeToken generates a token for a voucher barcode
func GenerateVoucherBarcodeToken(ctx context.Context, voucherID uuid.UUID) string {
	return GenerateBarcodeToken(ctx, voucherID, "voucher")
}

// GenerateGiftCardBarcodeToken generates a token for a gift card barcode
func GenerateGiftCardBarcodeToken(ctx context.Context, giftCardID uuid.UUID) string {
	return GenerateBarcodeToken(ctx, giftCardID, "gift_card")
}
