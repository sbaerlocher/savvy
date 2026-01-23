// Package handlers contains HTTP request handlers for the savvy system.
package handlers

import (
	"image/png"
	"net/http"

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
	"github.com/labstack/echo/v4"
)

// BarcodeGenerate generates a barcode image based on type and data
func BarcodeGenerate(c echo.Context) error {
	barcodeType := c.QueryParam("type")
	data := c.QueryParam("data")

	if data == "" {
		return c.String(http.StatusBadRequest, "Missing data parameter")
	}

	var barcodeImage barcode.Barcode
	var err error

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
