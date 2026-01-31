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

// resourceData holds barcode data extracted from a resource
type resourceData struct {
	barcodeType string
	data        string
}

// fetchResourceData fetches resource and extracts barcode data based on resource type
func fetchResourceData(claims *security.BarcodeTokenClaims, userID uuid.UUID) (*resourceData, error) {
	switch claims.ResourceType {
	case "card":
		var card models.Card
		if err := database.DB.Where("id = ?", claims.ResourceID).First(&card).Error; err != nil {
			return nil, err
		}
		if !hasCardAccess(userID, &card) {
			return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
		return &resourceData{barcodeType: card.BarcodeType, data: card.CardNumber}, nil

	case "voucher":
		var voucher models.Voucher
		if err := database.DB.Where("id = ?", claims.ResourceID).First(&voucher).Error; err != nil {
			return nil, err
		}
		if !hasVoucherAccess(userID, &voucher) {
			return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
		return &resourceData{barcodeType: voucher.BarcodeType, data: voucher.Code}, nil

	case "gift_card":
		var giftCard models.GiftCard
		if err := database.DB.Where("id = ?", claims.ResourceID).First(&giftCard).Error; err != nil {
			return nil, err
		}
		if !hasGiftCardAccess(userID, &giftCard) {
			return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
		return &resourceData{barcodeType: giftCard.BarcodeType, data: giftCard.CardNumber}, nil

	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid resource type")
	}
}

// encodeBarcode creates a barcode image from data and type
func encodeBarcode(barcodeType, data string) (barcode.Barcode, error) {
	switch barcodeType {
	case "CODE128":
		return code128.Encode(data)

	case "CODE39":
		return code39.Encode(data, true, true)

	case "CODE93":
		return code93.Encode(data, true, true)

	case "CODABAR":
		return codabar.Encode(data)

	case "QR":
		return qr.Encode(data, qr.M, qr.Auto)

	case "EAN13", "ISBN13":
		barcodeImage, err := ean.Encode(data)
		if err != nil {
			// Fallback to CODE128 for invalid EAN numbers
			return code128.Encode(data)
		}
		return barcodeImage, nil

	case "EAN8":
		barcodeImage, err := ean.Encode(data)
		if err != nil {
			// Fallback to CODE128 for invalid EAN numbers
			return code128.Encode(data)
		}
		return barcodeImage, nil

	case "PDF417":
		return pdf417.Encode(data, 2)

	case "DATAMATRIX":
		return datamatrix.Encode(data)

	case "AZTEC":
		return aztec.Encode([]byte(data), 50, 0)

	case "UPCA", "UPCE", "ITF", "ITF14", "MAXICODE":
		// These formats are not supported - fallback to CODE128
		return code128.Encode(data)

	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Nicht unterstützter Barcode-Typ")
	}
}

// BarcodeGenerate generates a barcode image using a secure token
// The token contains encrypted resource information and expires after 7 days (matches PWA cache)
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
	resData, err := fetchResourceData(claims, user.ID)
	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		return c.String(http.StatusNotFound, "Resource not found")
	}

	if resData.data == "" {
		return c.String(http.StatusBadRequest, "No barcode data available")
	}

	// Encode barcode
	barcodeImage, err := encodeBarcode(resData.barcodeType, resData.data)
	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		c.Logger().Errorf("Barcode encoding failed (%s): %v", resData.barcodeType, err)
		return c.String(http.StatusBadRequest, "Ungültige Barcode-Daten")
	}

	// Scale barcode to appropriate size
	var width, height int
	if resData.barcodeType == "QR" {
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
