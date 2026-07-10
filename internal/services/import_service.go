// Package services contains business logic.
package services

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"savvy/internal/models"
	"savvy/internal/validation"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// maxImportItems is the maximum total number of items allowed in a single import request.
const maxImportItems = 5000

// ImportResult contains the result of an import operation.
type ImportResult struct {
	CardsImported     int           `json:"cards_imported"`
	VouchersImported  int           `json:"vouchers_imported"`
	GiftCardsImported int           `json:"gift_cards_imported"`
	Skipped           int           `json:"skipped"`
	Errors            []ImportError `json:"errors,omitempty"`
}

// ImportError describes a single import error.
type ImportError struct {
	Row     int    `json:"row,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// ImportPreview shows what will be imported without executing.
type ImportPreview struct {
	Cards     int `json:"cards"`
	Vouchers  int `json:"vouchers"`
	GiftCards int `json:"gift_cards"`
}

// ImportServiceInterface defines import operations.
type ImportServiceInterface interface {
	PreviewJSON(ctx context.Context, data *ExportData) (*ImportPreview, error)
	ImportJSON(ctx context.Context, userID uuid.UUID, data *ExportData) (*ImportResult, error)
	ImportCardsCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*ImportResult, error)
	ImportVouchersCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*ImportResult, error)
	ImportGiftCardsCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*ImportResult, error)
}

// ImportService implements ImportServiceInterface.
type ImportService struct {
	cardService     CardServiceInterface
	voucherService  VoucherServiceInterface
	giftCardService GiftCardServiceInterface
	merchantService MerchantServiceInterface
}

// NewImportService creates a new import service.
func NewImportService(
	cardService CardServiceInterface,
	voucherService VoucherServiceInterface,
	giftCardService GiftCardServiceInterface,
	merchantService MerchantServiceInterface,
) ImportServiceInterface {
	return &ImportService{
		cardService:     cardService,
		voucherService:  voucherService,
		giftCardService: giftCardService,
		merchantService: merchantService,
	}
}

// PreviewJSON returns counts of what would be imported without executing.
func (s *ImportService) PreviewJSON(_ context.Context, data *ExportData) (*ImportPreview, error) {
	if data == nil {
		return nil, errors.New("no data provided")
	}
	return &ImportPreview{
		Cards:     len(data.Cards),
		Vouchers:  len(data.Vouchers),
		GiftCards: len(data.GiftCards),
	}, nil
}

// ImportJSON imports data from the export JSON format.
func (s *ImportService) ImportJSON(ctx context.Context, userID uuid.UUID, data *ExportData) (*ImportResult, error) {
	if data == nil {
		return nil, errors.New("no data provided")
	}

	totalItems := len(data.Cards) + len(data.Vouchers) + len(data.GiftCards)
	if totalItems > maxImportItems {
		return nil, fmt.Errorf("import exceeds maximum of %d items (got %d)", maxImportItems, totalItems)
	}

	result := &ImportResult{}

	// Import cards
	for i, ec := range data.Cards {
		barcodeType := defaultIfEmpty(ec.BarcodeType, "CODE128")
		status := defaultIfEmpty(ec.Status, "active")

		if validationErrs := validateCard(barcodeType, status, ec.CardNumber, ec.Program, ec.Notes); len(validationErrs) > 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("Card %q: %s", ec.CardNumber, strings.Join(validationErrs, "; ")),
			})
			result.Skipped++
			continue
		}

		merchantID, err := s.resolveMerchant(ctx, ec.MerchantName)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Field:   "merchant_name",
				Message: fmt.Sprintf("Card %q: %v", ec.CardNumber, err),
			})
			result.Skipped++
			continue
		}

		card := &models.Card{
			UserID:       &userID,
			MerchantID:   merchantID,
			MerchantName: ec.MerchantName,
			Program:      ec.Program,
			CardNumber:   ec.CardNumber,
			BarcodeType:  barcodeType,
			Status:       status,
			Notes:        ec.Notes,
		}

		if err := s.cardService.CreateCard(ctx, card); err != nil {
			slog.WarnContext(ctx, "import card failed", "card_number", ec.CardNumber, "error", err)
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("Card %q: %s", ec.CardNumber, sanitizeImportError(err)),
			})
			result.Skipped++
			continue
		}
		result.CardsImported++
	}

	// Import vouchers
	for i, ev := range data.Vouchers {
		voucherType := defaultIfEmpty(ev.Type, "percentage")
		barcodeType := defaultIfEmpty(ev.BarcodeType, "CODE128")
		usageLimitType := defaultIfEmpty(ev.UsageLimitType, "single_use")

		if validationErrs := validateVoucher(ev.Code, voucherType, barcodeType, usageLimitType, ev.Description, ev.Value); len(validationErrs) > 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("Voucher %q: %s", ev.Code, strings.Join(validationErrs, "; ")),
			})
			result.Skipped++
			continue
		}

		validFrom, err := time.Parse(time.RFC3339, ev.ValidFrom)
		if err != nil && ev.ValidFrom != "" {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Field:   "valid_from",
				Message: fmt.Sprintf("Voucher %q: invalid valid_from date format (use RFC3339)", ev.Code),
			})
			result.Skipped++
			continue
		}

		validUntil, err := time.Parse(time.RFC3339, ev.ValidUntil)
		if err != nil && ev.ValidUntil != "" {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Field:   "valid_until",
				Message: fmt.Sprintf("Voucher %q: invalid valid_until date format (use RFC3339)", ev.Code),
			})
			result.Skipped++
			continue
		}

		merchantID, err := s.resolveMerchant(ctx, ev.MerchantName)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Field:   "merchant_name",
				Message: fmt.Sprintf("Voucher %q: %v", ev.Code, err),
			})
			result.Skipped++
			continue
		}

		voucher := &models.Voucher{
			UserID:            &userID,
			MerchantID:        merchantID,
			MerchantName:      ev.MerchantName,
			Code:              ev.Code,
			Type:              voucherType,
			Value:             ev.Value,
			Description:       ev.Description,
			MinPurchaseAmount: ev.MinPurchaseAmount,
			ValidFrom:         validFrom,
			ValidUntil:        validUntil,
			UsageLimitType:    usageLimitType,
			BarcodeType:       barcodeType,
		}

		if err := s.voucherService.CreateVoucher(ctx, voucher); err != nil {
			slog.WarnContext(ctx, "import voucher failed", "code", ev.Code, "error", err)
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("Voucher %q: %s", ev.Code, sanitizeImportError(err)),
			})
			result.Skipped++
			continue
		}
		result.VouchersImported++
	}

	// Import gift cards
	for i, egc := range data.GiftCards {
		currency := defaultIfEmpty(egc.Currency, "CHF")
		barcodeType := defaultIfEmpty(egc.BarcodeType, "CODE128")

		if validationErrs := validateGiftCard(egc.CardNumber, currency, barcodeType, egc.PIN, egc.Notes, egc.InitialBalance); len(validationErrs) > 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("Gift card %q: %s", egc.CardNumber, strings.Join(validationErrs, "; ")),
			})
			result.Skipped++
			continue
		}

		merchantID, err := s.resolveMerchant(ctx, egc.MerchantName)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Field:   "merchant_name",
				Message: fmt.Sprintf("Gift card %q: %v", egc.CardNumber, err),
			})
			result.Skipped++
			continue
		}

		giftCard := &models.GiftCard{
			UserID:         &userID,
			MerchantID:     merchantID,
			MerchantName:   egc.MerchantName,
			CardNumber:     egc.CardNumber,
			InitialBalance: egc.InitialBalance,
			Currency:       currency,
			PIN:            egc.PIN,
			BarcodeType:    barcodeType,
			Notes:          egc.Notes,
		}

		if egc.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, egc.ExpiresAt)
			if err != nil {
				result.Errors = append(result.Errors, ImportError{
					Row:     i + 1,
					Field:   "expires_at",
					Message: fmt.Sprintf("Gift card %q: invalid expires_at date format (use RFC3339)", egc.CardNumber),
				})
				result.Skipped++
				continue
			}
			giftCard.ExpiresAt = &t
		}

		if err := s.giftCardService.CreateGiftCard(ctx, giftCard); err != nil {
			slog.WarnContext(ctx, "import gift card failed", "card_number", egc.CardNumber, "error", err)
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("Gift card %q: %s", egc.CardNumber, sanitizeImportError(err)),
			})
			result.Skipped++
			continue
		}
		result.GiftCardsImported++
	}

	return result, nil
}

// ImportCardsCSV imports cards from a CSV reader.
// Expected columns: merchant_name,program,card_number,barcode_type,status,notes
func (s *ImportService) ImportCardsCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*ImportResult, error) {
	records, err := readCSV(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return &ImportResult{}, nil
	}

	header := normalizeCSVHeader(records[0])
	result := &ImportResult{}

	for i, record := range records[1:] {
		row := mapCSVRow(header, record)

		merchantName := row["merchant_name"]
		merchantID, err := s.resolveMerchant(ctx, merchantName)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "merchant_name", Message: err.Error()})
			result.Skipped++
			continue
		}

		cardNumber := row["card_number"]
		if cardNumber == "" {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "card_number", Message: "card_number is required"})
			result.Skipped++
			continue
		}

		barcodeType := defaultIfEmpty(row["barcode_type"], "CODE128")
		status := defaultIfEmpty(row["status"], "active")

		if validationErrs := validateCard(barcodeType, status, cardNumber, row["program"], row["notes"]); len(validationErrs) > 0 {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Message: fmt.Sprintf("Card %q: %s", cardNumber, strings.Join(validationErrs, "; "))})
			result.Skipped++
			continue
		}

		card := &models.Card{
			UserID:       &userID,
			MerchantID:   merchantID,
			MerchantName: merchantName,
			Program:      row["program"],
			CardNumber:   cardNumber,
			BarcodeType:  barcodeType,
			Status:       status,
			Notes:        row["notes"],
		}

		if err := s.cardService.CreateCard(ctx, card); err != nil {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Message: fmt.Sprintf("Card %q: %s", cardNumber, sanitizeImportError(err))})
			result.Skipped++
			continue
		}
		result.CardsImported++
	}

	return result, nil
}

// ImportVouchersCSV imports vouchers from a CSV reader.
// Expected columns: merchant_name,code,type,value,description,valid_from,valid_until,usage_limit_type,barcode_type
func (s *ImportService) ImportVouchersCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*ImportResult, error) {
	records, err := readCSV(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return &ImportResult{}, nil
	}

	header := normalizeCSVHeader(records[0])
	result := &ImportResult{}

	for i, record := range records[1:] {
		row := mapCSVRow(header, record)

		merchantName := row["merchant_name"]
		merchantID, err := s.resolveMerchant(ctx, merchantName)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "merchant_name", Message: err.Error()})
			result.Skipped++
			continue
		}

		code := row["code"]
		if code == "" {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "code", Message: "code is required"})
			result.Skipped++
			continue
		}

		voucherType := defaultIfEmpty(row["type"], "percentage")
		barcodeType := defaultIfEmpty(row["barcode_type"], "CODE128")
		usageLimitType := defaultIfEmpty(row["usage_limit_type"], "single_use")

		value, err := strconv.ParseFloat(row["value"], 64)
		if err != nil && row["value"] != "" {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "value", Message: fmt.Sprintf("invalid value: %q", row["value"])})
			result.Skipped++
			continue
		}

		minPurchase, _ := strconv.ParseFloat(row["min_purchase_amount"], 64)

		if validationErrs := validateVoucher(code, voucherType, barcodeType, usageLimitType, row["description"], value); len(validationErrs) > 0 {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Message: fmt.Sprintf("Voucher %q: %s", code, strings.Join(validationErrs, "; "))})
			result.Skipped++
			continue
		}

		validFrom := parseTimeOrNow(row["valid_from"])
		validUntil := parseTimeOrDefault(row["valid_until"], time.Now().AddDate(1, 0, 0))

		voucher := &models.Voucher{
			UserID:            &userID,
			MerchantID:        merchantID,
			MerchantName:      merchantName,
			Code:              code,
			Type:              voucherType,
			Value:             value,
			Description:       row["description"],
			MinPurchaseAmount: minPurchase,
			ValidFrom:         validFrom,
			ValidUntil:        validUntil,
			UsageLimitType:    usageLimitType,
			BarcodeType:       barcodeType,
		}

		if err := s.voucherService.CreateVoucher(ctx, voucher); err != nil {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Message: fmt.Sprintf("Voucher %q: %s", code, sanitizeImportError(err))})
			result.Skipped++
			continue
		}
		result.VouchersImported++
	}

	return result, nil
}

// ImportGiftCardsCSV imports gift cards from a CSV reader.
// Expected columns: merchant_name,card_number,initial_balance,currency,pin,expires_at,notes
func (s *ImportService) ImportGiftCardsCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*ImportResult, error) {
	records, err := readCSV(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return &ImportResult{}, nil
	}

	header := normalizeCSVHeader(records[0])
	result := &ImportResult{}

	for i, record := range records[1:] {
		row := mapCSVRow(header, record)

		merchantName := row["merchant_name"]
		merchantID, err := s.resolveMerchant(ctx, merchantName)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "merchant_name", Message: err.Error()})
			result.Skipped++
			continue
		}

		cardNumber := row["card_number"]
		if cardNumber == "" {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "card_number", Message: "card_number is required"})
			result.Skipped++
			continue
		}

		initialBalance, err := strconv.ParseFloat(row["initial_balance"], 64)
		if err != nil && row["initial_balance"] != "" {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "initial_balance", Message: fmt.Sprintf("invalid initial_balance: %q", row["initial_balance"])})
			result.Skipped++
			continue
		}

		currency := defaultIfEmpty(row["currency"], "CHF")
		barcodeType := defaultIfEmpty(row["barcode_type"], "CODE128")

		if validationErrs := validateGiftCard(cardNumber, currency, barcodeType, row["pin"], row["notes"], initialBalance); len(validationErrs) > 0 {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Message: fmt.Sprintf("Gift card %q: %s", cardNumber, strings.Join(validationErrs, "; "))})
			result.Skipped++
			continue
		}

		giftCard := &models.GiftCard{
			UserID:         &userID,
			MerchantID:     merchantID,
			MerchantName:   merchantName,
			CardNumber:     cardNumber,
			InitialBalance: initialBalance,
			Currency:       currency,
			PIN:            row["pin"],
			BarcodeType:    barcodeType,
			Notes:          row["notes"],
		}

		if row["expires_at"] != "" {
			t, err := time.Parse(time.RFC3339, row["expires_at"])
			if err != nil {
				result.Errors = append(result.Errors, ImportError{Row: i + 2, Field: "expires_at", Message: fmt.Sprintf("invalid expires_at date format: %q (use RFC3339)", row["expires_at"])})
				result.Skipped++
				continue
			}
			giftCard.ExpiresAt = &t
		}

		if err := s.giftCardService.CreateGiftCard(ctx, giftCard); err != nil {
			result.Errors = append(result.Errors, ImportError{Row: i + 2, Message: fmt.Sprintf("Gift card %q: %s", cardNumber, sanitizeImportError(err))})
			result.Skipped++
			continue
		}
		result.GiftCardsImported++
	}

	return result, nil
}

// validateCard validates card fields and returns collected errors.
func validateCard(barcodeType, status, cardNumber, program, notes string) []string {
	var errs []string
	if err := validation.ValidateStringLength(cardNumber, "card_number"); err != nil {
		errs = append(errs, err.Error())
	}
	if program != "" {
		if err := validation.ValidateStringLength(program, "program"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if notes != "" {
		if err := validation.ValidateStringLength(notes, "notes"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := validation.ValidateEnum(barcodeType, validation.ValidBarcodeTypes, "barcode_type"); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validation.ValidateEnum(status, validation.ValidStatuses, "status"); err != nil {
		errs = append(errs, err.Error())
	}
	return errs
}

// validateVoucher validates voucher fields and returns collected errors.
func validateVoucher(code, voucherType, barcodeType, usageLimitType, description string, value float64) []string {
	var errs []string
	if err := validation.ValidateStringLength(code, "code"); err != nil {
		errs = append(errs, err.Error())
	}
	if description != "" {
		if err := validation.ValidateStringLength(description, "description"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := validation.ValidateEnum(voucherType, validation.ValidVoucherTypes, "type"); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validation.ValidateEnum(barcodeType, validation.ValidBarcodeTypes, "barcode_type"); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validation.ValidateEnum(usageLimitType, validation.ValidUsageLimitTypes, "usage_limit_type"); err != nil {
		errs = append(errs, err.Error())
	}
	if validation.VoucherValueRequired(voucherType) && value <= 0 {
		errs = append(errs, "value must be greater than 0")
	}
	return errs
}

// validateGiftCard validates gift card fields and returns collected errors.
func validateGiftCard(cardNumber, currency, barcodeType, pin, notes string, initialBalance float64) []string {
	var errs []string
	if err := validation.ValidateStringLength(cardNumber, "card_number"); err != nil {
		errs = append(errs, err.Error())
	}
	if pin != "" {
		if err := validation.ValidateStringLength(pin, "pin"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if notes != "" {
		if err := validation.ValidateStringLength(notes, "notes"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := validation.ValidateEnum(currency, validation.ValidCurrencies, "currency"); err != nil {
		errs = append(errs, err.Error())
	}
	if barcodeType != "" {
		if err := validation.ValidateEnum(barcodeType, validation.ValidBarcodeTypes, "barcode_type"); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := validation.ValidateMonetaryAmount(initialBalance, "initial_balance"); err != nil {
		errs = append(errs, err.Error())
	}
	return errs
}

// resolveMerchant finds a merchant by name or creates a new one.
func (s *ImportService) resolveMerchant(ctx context.Context, name string) (*uuid.UUID, error) {
	if name == "" {
		return nil, nil
	}

	// Try to find existing merchant
	merchant, err := s.merchantService.GetMerchantByName(ctx, name)
	if err == nil && merchant != nil {
		return &merchant.ID, nil
	}

	// Create new merchant
	newMerchant := &models.Merchant{Name: name}
	if err := s.merchantService.CreateMerchant(ctx, newMerchant); err != nil {
		return nil, fmt.Errorf("failed to create merchant %q: %w", name, err)
	}
	return &newMerchant.ID, nil
}

// sanitizeImportError replaces raw database errors with user-friendly messages.
func sanitizeImportError(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "SQLSTATE 23505") || strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
		return "already exists (duplicate)"
	}
	if strings.Contains(msg, "SQLSTATE") || strings.Contains(msg, "pq:") || strings.Contains(msg, "sql:") {
		return "database error"
	}
	return msg
}

// ==================== CSV Helpers ====================

func readCSV(reader io.Reader) ([][]string, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	return r.ReadAll()
}

func normalizeCSVHeader(header []string) []string {
	normalized := make([]string, len(header))
	for i, h := range header {
		normalized[i] = strings.ToLower(strings.TrimSpace(h))
	}
	return normalized
}

func mapCSVRow(header, record []string) map[string]string {
	m := make(map[string]string, len(header))
	for i, h := range header {
		if i < len(record) {
			m[h] = strings.TrimSpace(record[i])
		}
	}
	return m
}

func defaultIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func parseTimeOrNow(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try date-only format
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return time.Now()
		}
	}
	return t
}

func parseTimeOrDefault(s string, def time.Time) time.Time {
	if s == "" {
		return def
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return def
		}
	}
	return t
}
