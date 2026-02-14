// Package validation provides input validation utilities.
package validation

import (
	"fmt"
	"math"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

// ==================== Enum Allowed Values ====================

// ValidStatuses defines the allowed values for card/resource status fields.
var ValidStatuses = map[string]bool{
	"active": true, "inactive": true, "expired": true,
}

// ValidBarcodeTypes defines the allowed barcode type values.
var ValidBarcodeTypes = map[string]bool{
	"CODE128": true, "CODE39": true, "CODE93": true, "CODABAR": true,
	"QR": true, "EAN13": true, "EAN8": true, "UPCA": true, "UPCE": true,
	"ITF": true, "ITF14": true, "ISBN13": true, "ISBN10": true, "ISSN": true,
	"PDF417": true, "DATAMATRIX": true, "AZTEC": true, "MAXICODE": true,
}

// ValidVoucherTypes defines the allowed voucher type values.
var ValidVoucherTypes = map[string]bool{
	"percentage": true, "fixed_amount": true, "points_multiplier": true,
}

// ValidUsageLimitTypes defines the allowed usage limit type values.
var ValidUsageLimitTypes = map[string]bool{
	"single_use": true, "one_per_customer": true,
	"multiple_use_with_card": true, "multiple_use_without_card": true,
}

// ValidCurrencies defines the allowed ISO 4217 currency codes.
var ValidCurrencies = map[string]bool{
	"EUR": true, "USD": true, "CHF": true, "GBP": true,
}

// ==================== Validation Functions ====================

// ValidateEnum checks if a value is in the allowed set.
// Returns an error if the value is not allowed, nil if valid.
func ValidateEnum(value string, allowed map[string]bool, fieldName string) error {
	if !allowed[value] {
		return fmt.Errorf("invalid %s value: %q", fieldName, value)
	}
	return nil
}

// MaxFieldLengths defines maximum allowed lengths for string fields.
var MaxFieldLengths = map[string]int{
	"name":        255,
	"email":       255,
	"program":     255,
	"card_number": 500,
	"code":        500,
	"pin":         50,
	"notes":       5000,
	"description": 5000,
	"color":       50,
	"logo_url":    2048,
	"website":     2048,
	"currency":    10,
}

// DefaultMaxFieldLength is the fallback for fields not in MaxFieldLengths.
const DefaultMaxFieldLength = 1000

// ValidateStringLength checks if a string exceeds the maximum allowed length for the given field.
func ValidateStringLength(value string, fieldName string) error {
	maxLen, ok := MaxFieldLengths[fieldName]
	if !ok {
		maxLen = DefaultMaxFieldLength
	}
	if utf8.RuneCountInString(value) > maxLen {
		return fmt.Errorf("%s exceeds maximum length of %d characters", fieldName, maxLen)
	}
	return nil
}

// ValidateStringLengthPtr is the pointer variant for optional fields.
func ValidateStringLengthPtr(value *string, fieldName string) error {
	if value == nil {
		return nil
	}
	return ValidateStringLength(*value, fieldName)
}

// Validator is the global validator instance
var Validator *validator.Validate

func init() {
	Validator = validator.New()
}

// LoginRequest represents the login form validation
type LoginRequest struct {
	Email    string `validate:"required,email,max=255"`
	Password string `validate:"required,max=255"` // #nosec G117 -- struct field name, not a hardcoded secret
}

// RegisterRequest represents the registration form validation
type RegisterRequest struct {
	Email     string `validate:"required,email,max=255"`
	Password  string `validate:"required,min=8,max=255"` // #nosec G117 -- struct field name, not a hardcoded secret
	FirstName string `validate:"required,min=1,max=100"`
	LastName  string `validate:"required,min=1,max=100"`
}

// CardRequest represents card creation/update validation
type CardRequest struct {
	MerchantID   string `validate:"omitempty,uuid"`
	MerchantName string `validate:"required_without=MerchantID,max=255"`
	Program      string `validate:"required,max=255"`
	CardNumber   string `validate:"required,max=255"`
	BarcodeType  string `validate:"required,oneof=CODE128 CODE39 CODE93 CODABAR QR EAN13 EAN8 UPCA UPCE ITF ITF14 ISBN13 ISBN10 ISSN PDF417 DATAMATRIX AZTEC MAXICODE"`
	Notes        string `validate:"max=1000"`
	Status       string `validate:"required,oneof=active inactive expired"`
}

// VoucherRequest represents voucher creation/update validation
type VoucherRequest struct {
	MerchantID        string  `validate:"omitempty,uuid"`
	MerchantName      string  `validate:"required_without=MerchantID,max=255"`
	Code              string  `validate:"required,max=255"`
	VoucherType       string  `validate:"required,oneof=percentage fixed_amount points_multiplier"`
	Value             float64 `validate:"required,gt=0"`
	MinPurchaseAmount float64 `validate:"omitempty,gte=0"`
	UsageLimitType    string  `validate:"required,oneof=single_use one_per_customer multiple_use_with_card multiple_use_without_card"`
	MaxUses           int     `validate:"omitempty,gte=1"`
	BarcodeType       string  `validate:"required,oneof=CODE128 CODE39 CODE93 CODABAR QR EAN13 EAN8 UPCA UPCE ITF ITF14 ISBN13 ISBN10 ISSN PDF417 DATAMATRIX AZTEC MAXICODE"`
	Status            string  `validate:"required,oneof=active inactive expired"`
}

// GiftCardRequest represents gift card creation/update validation
type GiftCardRequest struct {
	MerchantID     string  `validate:"omitempty,uuid"`
	MerchantName   string  `validate:"required_without=MerchantID,max=255"`
	CardNumber     string  `validate:"required,max=255"`
	InitialBalance float64 `validate:"required,gte=0,lte=1000000"`
	Currency       string  `validate:"required,len=3"` // ISO 4217 currency code
	PIN            string  `validate:"omitempty,max=50"`
	BarcodeType    string  `validate:"required,oneof=CODE128 CODE39 CODE93 CODABAR QR EAN13 EAN8 UPCA UPCE ITF ITF14 ISBN13 ISBN10 ISSN PDF417 DATAMATRIX AZTEC MAXICODE"`
	Status         string  `validate:"required,oneof=active inactive expired"`
}

// TransactionRequest represents transaction creation validation
type TransactionRequest struct {
	Amount      float64 `validate:"required,ne=0,gte=-1000000,lte=1000000"` // Can be positive or negative, but not zero
	Description string  `validate:"omitempty,max=500"`
}

// MaxMonetaryAmount is the upper bound for monetary values (balances, amounts).
const MaxMonetaryAmount = 1_000_000

// ValidateMonetaryAmount checks that a monetary value (e.g. balance) is finite and within bounds [0, MaxMonetaryAmount].
func ValidateMonetaryAmount(value float64, fieldName string) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("%s must be a valid number", fieldName)
	}
	if value < 0 {
		return fmt.Errorf("%s must not be negative", fieldName)
	}
	if value > MaxMonetaryAmount {
		return fmt.Errorf("%s exceeds maximum of %d", fieldName, MaxMonetaryAmount)
	}
	return nil
}

// ValidateTransactionAmount checks that a transaction amount is finite, non-zero, and within bounds.
func ValidateTransactionAmount(value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("amount must be a valid number")
	}
	if value == 0 {
		return fmt.Errorf("amount cannot be zero")
	}
	if value > MaxMonetaryAmount || value < -MaxMonetaryAmount {
		return fmt.Errorf("amount exceeds maximum of %d", MaxMonetaryAmount)
	}
	return nil
}

// ValidateStruct validates a struct using the validator
func ValidateStruct(s interface{}) error {
	return Validator.Struct(s)
}

// ParseAndValidateDate parses a date string and validates it
// Returns the parsed time in UTC at start of day (00:00:00)
func ParseAndValidateDate(dateStr string, allowPast bool) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date cannot be empty")
	}

	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format (expected YYYY-MM-DD)")
	}

	// Check if date is not in the past (if required)
	if !allowPast {
		now := time.Now().UTC()
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if parsed.Before(nowDate) {
			return time.Time{}, fmt.Errorf("date cannot be in the past")
		}
	}

	// Return date at start of day in UTC
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
}

// ParseAndValidateDateRange parses and validates a date range
// Returns from (00:00:00) and until (23:59:59) in UTC
func ParseAndValidateDateRange(fromStr, untilStr string, allowPast bool) (time.Time, time.Time, error) {
	from, err := ParseAndValidateDate(fromStr, allowPast)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}

	until, err := ParseAndValidateDate(untilStr, true) // until can be any date
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
	}

	// Check if until is after from
	if until.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("end date must be after start date")
	}

	// Set until to end of day (23:59:59)
	until = time.Date(until.Year(), until.Month(), until.Day(), 23, 59, 59, 0, time.UTC)

	return from, until, nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must be at most 128 characters")
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}
