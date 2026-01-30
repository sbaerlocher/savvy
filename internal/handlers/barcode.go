// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"image/png"
	"net/http"
	"savvy/internal/database"
	"savvy/internal/models"
	"savvy/internal/security"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/codabar"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/code93"
	"github.com/boombuler/barcode/datamatrix"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/pdf417"
	"github.com/boombuler/barcode/qr"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BarcodeGenerate generates a barcode image using a secure token
// The token contains encrypted resource information and expires after 60 seconds
func BarcodeGenerate(c echo.Context) error {
	// Get token from URL path parameter
	token := c.Param("token")
	if token == "" {
		return c.String(http.StatusBadRequest, "Missing token")
	}

	// Validate and decode token
	claims, err := security.ValidateBarcodeToken(token)
	if err != nil {
		if err == security.ErrTokenExpired {
			return c.String(http.StatusUnauthorized, "Token expired - please refresh the page")
		}
		c.Logger().Warnf("Invalid barcode token attempt: %v", err)
		return c.String(http.StatusForbidden, "Invalid or tampered token")
	}

	// Get current user from context
	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.String(http.StatusUnauthorized, "Authentication required")
	}

	// Verify the token belongs to this user
	if claims.UserID != user.ID {
		c.Logger().Warnf("User %s attempted to access barcode for user %s", user.ID, claims.UserID)
		return c.String(http.StatusForbidden, "Access denied")
	}

	// Fetch resource and extract barcode data
	var barcodeType, data string
	var dbErr error
	switch claims.ResourceType {
	case "card":
		var card models.Card
		if dbErr = database.DB.Where("id = ?", claims.ResourceID).First(&card).Error; dbErr != nil {
			return c.String(http.StatusNotFound, "Card not found")
		}
		// Verify user has access to card (owner or shared with)
		if !hasCardAccess(user.ID, &card) {
			return c.String(http.StatusForbidden, "Access denied")
		}
		barcodeType = card.BarcodeType
		data = card.CardNumber

	case "voucher":
		var voucher models.Voucher
		if dbErr = database.DB.Where("id = ?", claims.ResourceID).First(&voucher).Error; dbErr != nil {
			return c.String(http.StatusNotFound, "Voucher not found")
		}
		// Verify user has access to voucher (owner or shared with)
		if !hasVoucherAccess(user.ID, &voucher) {
			return c.String(http.StatusForbidden, "Access denied")
		}
		barcodeType = voucher.BarcodeType
		data = voucher.Code

	case "gift_card":
		var giftCard models.GiftCard
		if dbErr = database.DB.Where("id = ?", claims.ResourceID).First(&giftCard).Error; dbErr != nil {
			return c.String(http.StatusNotFound, "Gift card not found")
		}
		// Verify user has access to gift card (owner or shared with)
		if !hasGiftCardAccess(user.ID, &giftCard) {
			return c.String(http.StatusForbidden, "Access denied")
		}
		barcodeType = giftCard.BarcodeType
		data = giftCard.CardNumber

	default:
		return c.String(http.StatusBadRequest, "Invalid resource type")
	}

	if data == "" {
		return c.String(http.StatusBadRequest, "No barcode data available")
	}

	var barcodeImage barcode.Barcode

	switch barcodeType {
	case "CODE128":
		barcodeImage, err = code128.Encode(data)
		if err != nil {
			c.Logger().Errorf("CODE128 encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige Barcode-Daten")
		}

	case "CODE39":
		barcodeImage, err = code39.Encode(data, true, true) // includeChecksum, fullASCII
		if err != nil {
			c.Logger().Errorf("CODE39 encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige CODE39-Daten")
		}

	case "CODE93":
		barcodeImage, err = code93.Encode(data, true, true) // includeChecksum, fullASCII
		if err != nil {
			c.Logger().Errorf("CODE93 encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige CODE93-Daten")
		}

	case "CODABAR":
		barcodeImage, err = codabar.Encode(data)
		if err != nil {
			c.Logger().Errorf("CODABAR encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige CODABAR-Daten")
		}

	case "QR":
		barcodeImage, err = qr.Encode(data, qr.M, qr.Auto)
		if err != nil {
			c.Logger().Errorf("QR encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige QR-Code-Daten")
		}

	case "EAN13", "ISBN13":
		// Try EAN13, fallback to CODE128 if it fails (invalid checksum)
		barcodeImage, err = ean.Encode(data)
		if err != nil {
			// Fallback to CODE128 for invalid EAN numbers
			barcodeImage, err = code128.Encode(data)
			if err != nil {
				c.Logger().Errorf("EAN13/CODE128 encoding failed: %v", err)
				return c.String(http.StatusBadRequest, "Ungültige Barcode-Daten")
			}
		}

	case "EAN8":
		// Try EAN8, fallback to CODE128 if it fails (invalid checksum)
		barcodeImage, err = ean.Encode(data)
		if err != nil {
			// Fallback to CODE128 for invalid EAN numbers
			barcodeImage, err = code128.Encode(data)
			if err != nil {
				c.Logger().Errorf("EAN8/CODE128 encoding failed: %v", err)
				return c.String(http.StatusBadRequest, "Ungültige Barcode-Daten")
			}
		}

	case "PDF417":
		barcodeImage, err = pdf417.Encode(data, 2) // security level 2
		if err != nil {
			c.Logger().Errorf("PDF417 encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige PDF417-Daten")
		}

	case "DATAMATRIX":
		barcodeImage, err = datamatrix.Encode(data)
		if err != nil {
			c.Logger().Errorf("DataMatrix encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige DataMatrix-Daten")
		}

	case "AZTEC":
		barcodeImage, err = aztec.Encode([]byte(data), 50, 0) // 50% error correction, auto layers
		if err != nil {
			c.Logger().Errorf("Aztec encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige Aztec-Daten")
		}

	case "UPCA", "UPCE", "ITF", "ITF14", "MAXICODE":
		// These formats are not supported by boombuler/barcode
		// Fallback to CODE128
		c.Logger().Warnf("Barcode type %s not supported, falling back to CODE128", barcodeType)
		barcodeImage, err = code128.Encode(data)
		if err != nil {
			c.Logger().Errorf("CODE128 fallback encoding failed: %v", err)
			return c.String(http.StatusBadRequest, "Ungültige Barcode-Daten")
		}

	default:
		return c.String(http.StatusBadRequest, "Nicht unterstützter Barcode-Typ")
	}

	// Scale barcode to appropriate size
	var width, height int
	if barcodeType == "QR" {
		width = 300
		height = 300
	} else {
		width = 400
		height = 100
	}

	barcodeImage, err = barcode.Scale(barcodeImage, width, height)
	if err != nil {
		c.Logger().Errorf("Barcode scaling failed: %v", err)
		return c.String(http.StatusInternalServerError, "Fehler beim Generieren des Barcodes")
	}

	// Set response headers
	c.Response().Header().Set("Content-Type", "image/png")
	// Security: Don't cache barcodes containing sensitive card/voucher numbers
	// This prevents data leakage on shared computers or browser history
	c.Response().Header().Set("Cache-Control", "private, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Encode and write PNG
	return png.Encode(c.Response().Writer, barcodeImage)
}

// hasCardAccess checks if user has access to a card (owner or shared with)
func hasCardAccess(userID uuid.UUID, card *models.Card) bool {
	// User is owner
	if card.UserID != nil && *card.UserID == userID {
		return true
	}

	// Check if shared with user
	var shareCount int64
	database.DB.Model(&models.CardShare{}).
		Where("card_id = ? AND shared_with_id = ?", card.ID, userID).
		Count(&shareCount)
	return shareCount > 0
}

// hasVoucherAccess checks if user has access to a voucher (owner or shared with)
func hasVoucherAccess(userID uuid.UUID, voucher *models.Voucher) bool {
	// User is owner
	if voucher.UserID != nil && *voucher.UserID == userID {
		return true
	}

	// Check if shared with user
	var shareCount int64
	database.DB.Model(&models.VoucherShare{}).
		Where("voucher_id = ? AND shared_with_id = ?", voucher.ID, userID).
		Count(&shareCount)
	return shareCount > 0
}

// hasGiftCardAccess checks if user has access to a gift card (owner or shared with)
func hasGiftCardAccess(userID uuid.UUID, giftCard *models.GiftCard) bool {
	// User is owner
	if giftCard.UserID != nil && *giftCard.UserID == userID {
		return true
	}

	// Check if shared with user
	var shareCount int64
	database.DB.Model(&models.GiftCardShare{}).
		Where("gift_card_id = ? AND shared_with_id = ?", giftCard.ID, userID).
		Count(&shareCount)
	return shareCount > 0
}
