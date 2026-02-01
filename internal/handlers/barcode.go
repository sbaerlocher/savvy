// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"image/png"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/security"
	"savvy/internal/services"

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

// BarcodeHandler handles barcode generation with authorization checks.
type BarcodeHandler struct {
	authzService    services.AuthzServiceInterface
	cardService     services.CardServiceInterface
	voucherService  services.VoucherServiceInterface
	giftCardService services.GiftCardServiceInterface
}

// NewBarcodeHandler creates a new barcode handler.
func NewBarcodeHandler(
	authzService services.AuthzServiceInterface,
	cardService services.CardServiceInterface,
	voucherService services.VoucherServiceInterface,
	giftCardService services.GiftCardServiceInterface,
) *BarcodeHandler {
	return &BarcodeHandler{
		authzService:    authzService,
		cardService:     cardService,
		voucherService:  voucherService,
		giftCardService: giftCardService,
	}
}

// resourceData holds barcode data extracted from a resource
type resourceData struct {
	barcodeType string
	data        string
}

// fetchResourceData fetches resource and extracts barcode data based on resource type
func (h *BarcodeHandler) fetchResourceData(ctx echo.Context, claims *security.BarcodeTokenClaims, userID uuid.UUID) (*resourceData, error) {
	switch claims.ResourceType {
	case "card":
		card, err := h.cardService.GetCard(ctx.Request().Context(), claims.ResourceID)
		if err != nil {
			return nil, err
		}
		_, err = h.authzService.CheckCardAccess(ctx.Request().Context(), userID, claims.ResourceID)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
		return &resourceData{barcodeType: card.BarcodeType, data: card.CardNumber}, nil

	case "voucher":
		voucher, err := h.voucherService.GetVoucher(ctx.Request().Context(), claims.ResourceID)
		if err != nil {
			return nil, err
		}
		_, err = h.authzService.CheckVoucherAccess(ctx.Request().Context(), userID, claims.ResourceID)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
		return &resourceData{barcodeType: voucher.BarcodeType, data: voucher.Code}, nil

	case "gift_card":
		giftCard, err := h.giftCardService.GetGiftCard(ctx.Request().Context(), claims.ResourceID)
		if err != nil {
			return nil, err
		}
		_, err = h.authzService.CheckGiftCardAccess(ctx.Request().Context(), userID, claims.ResourceID)
		if err != nil {
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
			return code128.Encode(data)
		}
		return barcodeImage, nil

	case "EAN8":
		barcodeImage, err := ean.Encode(data)
		if err != nil {
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
		return code128.Encode(data)

	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Nicht unterstützter Barcode-Typ")
	}
}

// Generate generates a barcode image using a secure token.
// The token contains encrypted resource information and expires after 7 days (matches PWA cache).
func (h *BarcodeHandler) Generate(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return c.String(http.StatusBadRequest, "Missing token")
	}

	claims, err := security.ValidateBarcodeToken(token)
	if err != nil {
		if err == security.ErrTokenExpired {
			return c.String(http.StatusUnauthorized, "Token expired - please refresh the page")
		}
		c.Logger().Warnf("Invalid barcode token attempt: %v", err)
		return c.String(http.StatusForbidden, "Invalid or tampered token")
	}

	user, ok := c.Get("current_user").(*models.User)
	if !ok || user == nil {
		return c.String(http.StatusUnauthorized, "Authentication required")
	}

	if claims.UserID != user.ID {
		c.Logger().Warnf("User %s attempted to access barcode for user %s", user.ID, claims.UserID)
		return c.String(http.StatusForbidden, "Access denied")
	}

	resData, err := h.fetchResourceData(c, claims, user.ID)
	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		return c.String(http.StatusNotFound, "Resource not found")
	}

	if resData.data == "" {
		return c.String(http.StatusBadRequest, "No barcode data available")
	}

	barcodeImage, err := encodeBarcode(resData.barcodeType, resData.data)
	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		c.Logger().Errorf("Barcode encoding failed (%s): %v", resData.barcodeType, err)
		return c.String(http.StatusBadRequest, "Ungültige Barcode-Daten")
	}

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

	c.Response().Header().Set("Content-Type", "image/png")
	// Cache for 7 days (matches token validity) - enables Service Worker offline support
	// private: only user's browser can cache (no shared/proxy caches)
	// immutable: browser won't revalidate (token rotation handles freshness)
	c.Response().Header().Set("Cache-Control", "private, max-age=604800, immutable")

	return png.Encode(c.Response().Writer, barcodeImage)
}
