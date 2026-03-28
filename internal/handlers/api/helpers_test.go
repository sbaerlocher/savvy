package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
)

// ==================== parsePaginationParams Tests ====================

func TestParsePaginationParams_NoPagination(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, perPage, isPaginated := parsePaginationParams(c)

	assert.Equal(t, 0, page)
	assert.Equal(t, 0, perPage)
	assert.False(t, isPaginated)
}

func TestParsePaginationParams_WithPageOnly(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?page=3", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, perPage, isPaginated := parsePaginationParams(c)

	assert.Equal(t, 3, page)
	assert.Equal(t, 25, perPage) // Default
	assert.True(t, isPaginated)
}

func TestParsePaginationParams_WithPerPageOnly(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?per_page=50", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, perPage, isPaginated := parsePaginationParams(c)

	assert.Equal(t, 1, page) // Default
	assert.Equal(t, 50, perPage)
	assert.True(t, isPaginated)
}

func TestParsePaginationParams_BothParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?page=2&per_page=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, perPage, isPaginated := parsePaginationParams(c)

	assert.Equal(t, 2, page)
	assert.Equal(t, 10, perPage)
	assert.True(t, isPaginated)
}

func TestParsePaginationParams_CapsPerPageAt100(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?per_page=999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, perPage, isPaginated := parsePaginationParams(c)

	assert.Equal(t, 100, perPage)
	assert.True(t, isPaginated)
}

func TestParsePaginationParams_InvalidValues(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?page=abc&per_page=-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, perPage, isPaginated := parsePaginationParams(c)

	assert.Equal(t, 1, page)     // Default when invalid
	assert.Equal(t, 25, perPage) // Default when invalid (negative)
	assert.True(t, isPaginated)
}

func TestParsePaginationParams_ZeroValues(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards?page=0&per_page=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, perPage, isPaginated := parsePaginationParams(c)

	assert.Equal(t, 1, page)     // Default when zero
	assert.Equal(t, 25, perPage) // Default when zero
	assert.True(t, isPaginated)
}

// ==================== capitalizeFirst Tests ====================

func TestCapitalizeFirst(t *testing.T) {
	assert.Equal(t, "Card", capitalizeFirst("card"))
	assert.Equal(t, "Voucher", capitalizeFirst("voucher"))
	assert.Equal(t, "Gift_card", capitalizeFirst("gift_card"))
	assert.Equal(t, "A", capitalizeFirst("a"))
	assert.Equal(t, "", capitalizeFirst(""))
	assert.Equal(t, "Already", capitalizeFirst("Already"))
}

// ==================== parseResourceID Tests ====================

func TestParseResourceID_Valid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/:id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	expectedID := "550e8400-e29b-41d4-a716-446655440000"
	c.SetParamValues(expectedID)

	id, err := parseResourceID(c, "card")

	assert.NoError(t, err)
	assert.Equal(t, expectedID, id.String())
}

func TestParseResourceID_Invalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/:id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	_, err := parseResourceID(c, "card")

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ==================== parseUUIDParam Tests ====================

func TestParseUUIDParam_Valid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("sharedWithID")
	expectedID := "550e8400-e29b-41d4-a716-446655440000"
	c.SetParamValues(expectedID)

	id, err := parseUUIDParam(c, "sharedWithID", "invalid_user_id", "Invalid user ID")

	assert.NoError(t, err)
	assert.Equal(t, expectedID, id.String())
}

func TestParseUUIDParam_Invalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("sharedWithID")
	c.SetParamValues("invalid")

	_, err := parseUUIDParam(c, "sharedWithID", "invalid_user_id", "Invalid user ID")

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ==================== validateEnum Tests ====================

func TestValidateEnum_Valid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := validateEnum(c, "active", map[string]bool{"active": true, "inactive": true}, "status")

	assert.Nil(t, err)
}

func TestValidateEnum_Invalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := validateEnum(c, "unknown", map[string]bool{"active": true, "inactive": true}, "status")

	assert.Nil(t, err) // validateEnum returns the c.JSON result, not an error
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ==================== stringPtrValue Tests ====================

func TestStringPtrValue(t *testing.T) {
	val := "hello"
	assert.Equal(t, "hello", stringPtrValue(&val))
	assert.Equal(t, "", stringPtrValue(nil))
}

// ==================== FormatTime Tests ====================

func TestFormatTime(t *testing.T) {
	tm := time.Date(2026, 3, 4, 12, 30, 0, 0, time.UTC)
	result := FormatTime(tm)
	assert.Contains(t, result, "2026-03-04")
}

func TestFormatTimePtr(t *testing.T) {
	tm := time.Date(2026, 3, 4, 12, 30, 0, 0, time.UTC)
	result := FormatTimePtr(&tm)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "2026-03-04")

	nilResult := FormatTimePtr(nil)
	assert.Nil(t, nilResult)
}

// ==================== resolveMerchantUpdate Tests ====================

func TestResolveMerchantUpdate_EmptyID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockMerchant := new(mocks.MockMerchantServiceInterface)
	mid, name, err := resolveMerchantUpdate(c, mockMerchant, "")
	assert.Nil(t, mid)
	assert.Equal(t, "", name)
	assert.Nil(t, err)
}

func TestResolveMerchantUpdate_InvalidUUID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockMerchant := new(mocks.MockMerchantServiceInterface)
	_, _, err := resolveMerchantUpdate(c, mockMerchant, "not-a-uuid")
	assert.Nil(t, err) // error is written to response, returns nil
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestResolveMerchantUpdate_MerchantNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	merchantID := uuid.New()
	mockMerchant := new(mocks.MockMerchantServiceInterface)
	mockMerchant.On("GetMerchantByID", mock.Anything, merchantID).
		Return((*models.Merchant)(nil), assert.AnError)

	_, _, err := resolveMerchantUpdate(c, mockMerchant, merchantID.String())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockMerchant.AssertExpectations(t)
}

func TestResolveMerchantUpdate_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	merchantID := uuid.New()
	merchant := &models.Merchant{ID: merchantID, Name: "IKEA"}
	mockMerchant := new(mocks.MockMerchantServiceInterface)
	mockMerchant.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	mid, name, err := resolveMerchantUpdate(c, mockMerchant, merchantID.String())
	assert.Nil(t, err)
	assert.NotNil(t, mid)
	assert.Equal(t, merchantID, *mid)
	assert.Equal(t, "IKEA", name)
	mockMerchant.AssertExpectations(t)
}
