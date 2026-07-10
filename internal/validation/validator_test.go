package validation

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStruct_LoginRequest(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		wantErr bool
	}{
		{
			name: "valid login request",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			request: LoginRequest{
				Email:    "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			request: LoginRequest{
				Email:    "not-an-email",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "email too long (>255 chars)",
			request: LoginRequest{
				Email:    string(make([]byte, 256)) + "@example.com",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "password at max length (255 chars)",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: string(make([]byte, 255)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStruct_RegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		request RegisterRequest
		wantErr bool
	}{
		{
			name: "valid registration",
			request: RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "Password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "password too short (<8 chars)",
			request: RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "Pass1",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "missing first name",
			request: RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "Password123",
				FirstName: "",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "missing last name",
			request: RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "Password123",
				FirstName: "John",
				LastName:  "",
			},
			wantErr: true,
		},
		{
			name: "first name too long (>100 chars)",
			request: RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "Password123",
				FirstName: string(make([]byte, 101)),
				LastName:  "Doe",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStruct_CardRequest(t *testing.T) {
	tests := []struct {
		name    string
		request CardRequest
		wantErr bool
	}{
		{
			name: "valid card request with merchant name",
			request: CardRequest{
				MerchantName: "IKEA",
				Program:      "IKEA Family",
				CardNumber:   "1234567890",
				BarcodeType:  "CODE128",
				Status:       "active",
			},
			wantErr: false,
		},
		{
			name: "valid card request with merchant ID",
			request: CardRequest{
				MerchantID:  "550e8400-e29b-41d4-a716-446655440000",
				Program:     "IKEA Family",
				CardNumber:  "1234567890",
				BarcodeType: "CODE128",
				Status:      "active",
			},
			wantErr: false,
		},
		{
			name: "invalid barcode type",
			request: CardRequest{
				MerchantName: "IKEA",
				Program:      "IKEA Family",
				CardNumber:   "1234567890",
				BarcodeType:  "INVALID",
				Status:       "active",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			request: CardRequest{
				MerchantName: "IKEA",
				Program:      "IKEA Family",
				CardNumber:   "1234567890",
				BarcodeType:  "CODE128",
				Status:       "invalid_status",
			},
			wantErr: true,
		},
		{
			name: "missing both merchant ID and name",
			request: CardRequest{
				Program:     "IKEA Family",
				CardNumber:  "1234567890",
				BarcodeType: "CODE128",
				Status:      "active",
			},
			wantErr: true,
		},
		{
			name: "notes too long (>1000 chars)",
			request: CardRequest{
				MerchantName: "IKEA",
				Program:      "IKEA Family",
				CardNumber:   "1234567890",
				BarcodeType:  "CODE128",
				Status:       "active",
				Notes:        string(make([]byte, 1001)),
			},
			wantErr: true,
		},
		{
			name: "all valid barcode types",
			request: CardRequest{
				MerchantName: "IKEA",
				Program:      "IKEA Family",
				CardNumber:   "1234567890",
				BarcodeType:  "QR",
				Status:       "active",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStruct_VoucherRequest(t *testing.T) {
	tests := []struct {
		name    string
		request VoucherRequest
		wantErr bool
	}{
		{
			name: "valid voucher request",
			request: VoucherRequest{
				MerchantName:      "Target",
				Code:              "SAVE20",
				VoucherType:       "percentage",
				Value:             20.0,
				MinPurchaseAmount: 50.0,
				UsageLimitType:    "single_use",
				MaxUses:           1,
				BarcodeType:       "QR",
				Status:            "active",
			},
			wantErr: false,
		},
		{
			name: "valid bonus_points voucher",
			request: VoucherRequest{
				MerchantName:      "Coop",
				Code:              "SUPER222",
				VoucherType:       "bonus_points",
				Value:             222.0,
				MinPurchaseAmount: 22.0,
				UsageLimitType:    "multiple_use_with_card",
				BarcodeType:       "CODE128",
				Status:            "active",
			},
			wantErr: false,
		},
		{
			name: "invalid voucher type",
			request: VoucherRequest{
				MerchantName:   "Target",
				Code:           "SAVE20",
				VoucherType:    "invalid_type",
				Value:          20.0,
				UsageLimitType: "single_use",
				BarcodeType:    "QR",
				Status:         "active",
			},
			wantErr: true,
		},
		{
			// Struct-level validation accepts value 0 (omitempty,gte=0); the
			// type-dependent "must be positive" rule lives in VoucherValueRequired,
			// enforced in the service/handler, not this tag.
			name: "zero value",
			request: VoucherRequest{
				MerchantName:   "Target",
				Code:           "SAVE20",
				VoucherType:    "percentage",
				Value:          0,
				UsageLimitType: "single_use",
				BarcodeType:    "QR",
				Status:         "active",
			},
			wantErr: false,
		},
		{
			name: "free voucher with zero value",
			request: VoucherRequest{
				MerchantName:   "Target",
				Code:           "FREEBIE",
				VoucherType:    "free",
				Value:          0,
				UsageLimitType: "single_use",
				BarcodeType:    "QR",
				Status:         "active",
			},
			wantErr: false,
		},
		{
			name: "negative value",
			request: VoucherRequest{
				MerchantName:   "Target",
				Code:           "SAVE20",
				VoucherType:    "percentage",
				Value:          -10.0,
				UsageLimitType: "single_use",
				BarcodeType:    "QR",
				Status:         "active",
			},
			wantErr: true,
		},
		{
			name: "invalid usage limit type",
			request: VoucherRequest{
				MerchantName:   "Target",
				Code:           "SAVE20",
				VoucherType:    "percentage",
				Value:          20.0,
				UsageLimitType: "invalid",
				BarcodeType:    "QR",
				Status:         "active",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStruct_GiftCardRequest(t *testing.T) {
	tests := []struct {
		name    string
		request GiftCardRequest
		wantErr bool
	}{
		{
			name: "valid gift card request",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: 100.0,
				Currency:       "USD",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: false,
		},
		{
			name: "zero initial balance is actually treated as missing by validator",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: 0,
				Currency:       "USD",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: true, // Go validator treats 0 as zero value, fails required check
		},
		{
			name: "negative initial balance",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: -10.0,
				Currency:       "USD",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: true,
		},
		{
			name: "invalid currency code (not 3 chars)",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: 100.0,
				Currency:       "US",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: true,
		},
		{
			name: "valid currency EUR",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: 100.0,
				Currency:       "EUR",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: false,
		},
		{
			name: "with PIN",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: 100.0,
				Currency:       "USD",
				PIN:            "1234",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: false,
		},
		{
			name: "initial balance exceeds upper bound",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: 1_000_001,
				Currency:       "USD",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: true,
		},
		{
			name: "initial balance at upper bound",
			request: GiftCardRequest{
				MerchantName:   "Amazon",
				CardNumber:     "AMZN-1234-5678",
				InitialBalance: 1_000_000,
				Currency:       "USD",
				BarcodeType:    "CODE128",
				Status:         "active",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStruct_TransactionRequest(t *testing.T) {
	tests := []struct {
		name    string
		request TransactionRequest
		wantErr bool
	}{
		{
			name: "positive amount",
			request: TransactionRequest{
				Amount:      50.0,
				Description: "Added funds",
			},
			wantErr: false,
		},
		{
			name: "negative amount",
			request: TransactionRequest{
				Amount:      -25.0,
				Description: "Purchase",
			},
			wantErr: false,
		},
		{
			name: "zero amount not allowed",
			request: TransactionRequest{
				Amount:      0,
				Description: "Zero transaction",
			},
			wantErr: true,
		},
		{
			name: "description too long (>500 chars)",
			request: TransactionRequest{
				Amount:      50.0,
				Description: string(make([]byte, 501)),
			},
			wantErr: true,
		},
		{
			name: "no description",
			request: TransactionRequest{
				Amount: 50.0,
			},
			wantErr: false,
		},
		{
			name: "amount exceeds upper bound",
			request: TransactionRequest{
				Amount:      1_000_001,
				Description: "Too large",
			},
			wantErr: true,
		},
		{
			name: "negative amount exceeds lower bound",
			request: TransactionRequest{
				Amount:      -1_000_001,
				Description: "Too large negative",
			},
			wantErr: true,
		},
		{
			name: "amount at upper bound",
			request: TransactionRequest{
				Amount:      1_000_000,
				Description: "At limit",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMonetaryAmount(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		wantErr bool
		errMsg  string
	}{
		{name: "valid amount", value: 100.0, wantErr: false},
		{name: "zero is valid", value: 0, wantErr: false},
		{name: "at upper bound", value: 1_000_000, wantErr: false},
		{name: "exceeds upper bound", value: 1_000_001, wantErr: true, errMsg: "exceeds maximum"},
		{name: "negative", value: -1, wantErr: true, errMsg: "must not be negative"},
		{name: "NaN", value: math.NaN(), wantErr: true, errMsg: "must be a valid number"},
		{name: "positive infinity", value: math.Inf(1), wantErr: true, errMsg: "must be a valid number"},
		{name: "negative infinity", value: math.Inf(-1), wantErr: true, errMsg: "must be a valid number"},
		{name: "very large float", value: 1e308, wantErr: true, errMsg: "exceeds maximum"},
		{name: "small valid amount", value: 0.01, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMonetaryAmount(tt.value, "balance")
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTransactionAmount(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		wantErr bool
		errMsg  string
	}{
		{name: "valid positive", value: 50.0, wantErr: false},
		{name: "valid negative", value: -25.0, wantErr: false},
		{name: "zero not allowed", value: 0, wantErr: true, errMsg: "cannot be zero"},
		{name: "at upper bound", value: 1_000_000, wantErr: false},
		{name: "at lower bound", value: -1_000_000, wantErr: false},
		{name: "exceeds upper bound", value: 1_000_001, wantErr: true, errMsg: "exceeds maximum"},
		{name: "exceeds lower bound", value: -1_000_001, wantErr: true, errMsg: "exceeds maximum"},
		{name: "NaN", value: math.NaN(), wantErr: true, errMsg: "must be a valid number"},
		{name: "positive infinity", value: math.Inf(1), wantErr: true, errMsg: "must be a valid number"},
		{name: "negative infinity", value: math.Inf(-1), wantErr: true, errMsg: "must be a valid number"},
		{name: "very large float", value: 1e308, wantErr: true, errMsg: "exceeds maximum"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransactionAmount(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidateDate(t *testing.T) {
	tests := []struct {
		name      string
		dateStr   string
		allowPast bool
		wantErr   bool
		checkDate func(t *testing.T, result time.Time)
	}{
		{
			name:      "valid future date",
			dateStr:   "2030-12-31",
			allowPast: false,
			wantErr:   false,
			checkDate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2030, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 31, result.Day())
				assert.Equal(t, 0, result.Hour())
				assert.Equal(t, 0, result.Minute())
				assert.Equal(t, 0, result.Second())
				assert.Equal(t, time.UTC, result.Location())
			},
		},
		{
			name:      "valid past date when allowed",
			dateStr:   "2020-01-01",
			allowPast: true,
			wantErr:   false,
		},
		{
			name:      "past date when not allowed",
			dateStr:   "2020-01-01",
			allowPast: false,
			wantErr:   true,
		},
		{
			name:      "empty date string",
			dateStr:   "",
			allowPast: false,
			wantErr:   true,
		},
		{
			name:      "invalid date format",
			dateStr:   "31-12-2030",
			allowPast: false,
			wantErr:   true,
		},
		{
			name:      "invalid date format (with time)",
			dateStr:   "2030-12-31 12:00:00",
			allowPast: false,
			wantErr:   true,
		},
		{
			name:      "today's date should be valid",
			dateStr:   time.Now().Format("2006-01-02"),
			allowPast: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAndValidateDate(tt.dateStr, tt.allowPast)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkDate != nil {
					tt.checkDate(t, result)
				}
			}
		})
	}
}

func TestParseAndValidateDateRange(t *testing.T) {
	tests := []struct {
		name       string
		fromStr    string
		untilStr   string
		allowPast  bool
		wantErr    bool
		checkDates func(t *testing.T, from, until time.Time)
	}{
		{
			name:      "valid date range",
			fromStr:   "2030-01-01",
			untilStr:  "2030-12-31",
			allowPast: false,
			wantErr:   false,
			checkDates: func(t *testing.T, from, until time.Time) {
				assert.Equal(t, 2030, from.Year())
				assert.Equal(t, time.January, from.Month())
				assert.Equal(t, 1, from.Day())
				assert.Equal(t, 0, from.Hour())

				assert.Equal(t, 2030, until.Year())
				assert.Equal(t, time.December, until.Month())
				assert.Equal(t, 31, until.Day())
				assert.Equal(t, 23, until.Hour())
				assert.Equal(t, 59, until.Minute())
				assert.Equal(t, 59, until.Second())
			},
		},
		{
			name:      "until before from",
			fromStr:   "2030-12-31",
			untilStr:  "2030-01-01",
			allowPast: false,
			wantErr:   true,
		},
		{
			name:      "same date for from and until",
			fromStr:   "2030-06-15",
			untilStr:  "2030-06-15",
			allowPast: false,
			wantErr:   false,
		},
		{
			name:      "invalid from date",
			fromStr:   "invalid",
			untilStr:  "2030-12-31",
			allowPast: false,
			wantErr:   true,
		},
		{
			name:      "invalid until date",
			fromStr:   "2030-01-01",
			untilStr:  "invalid",
			allowPast: false,
			wantErr:   true,
		},
		{
			name:      "past from date when not allowed",
			fromStr:   "2020-01-01",
			untilStr:  "2030-12-31",
			allowPast: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, until, err := ParseAndValidateDateRange(tt.fromStr, tt.untilStr, tt.allowPast)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkDates != nil {
					tt.checkDates(t, from, until)
				}
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "too long",
			password: string(make([]byte, 129)) + "Pass123",
			wantErr:  true,
			errMsg:   "at most 128 characters",
		},
		{
			name:     "no uppercase letter",
			password: "password123",
			wantErr:  true,
			errMsg:   "at least one uppercase letter",
		},
		{
			name:     "no lowercase letter",
			password: "PASSWORD123",
			wantErr:  true,
			errMsg:   "at least one lowercase letter",
		},
		{
			name:     "no digit",
			password: "PasswordOnly",
			wantErr:  true,
			errMsg:   "at least one digit",
		},
		{
			name:     "minimum valid password",
			password: "Passw0rd",
			wantErr:  false,
		},
		{
			name:     "with special characters",
			password: "P@ssw0rd!",
			wantErr:  false,
		},
		{
			name:     "complex password",
			password: "MyC0mpl3xP@ssw0rd!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
