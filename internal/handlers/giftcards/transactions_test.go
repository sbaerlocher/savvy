package giftcards

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestTransactionNew_Success(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/transactions/new", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	c.Set("csrf", "test-csrf-token")

	handler := &Handler{}

	err := handler.TransactionNew(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTransactionCancel_Success(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/transactions/cancel", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	handler := &Handler{}

	err := handler.TransactionCancel(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTransactionCreate_InvalidGiftCardID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/gift-cards/invalid-id/transactions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-id")

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	handler := &Handler{}

	err := handler.TransactionCreate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransactionCreate_Forbidden(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: false, // Not allowed to edit transactions
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	handler := &Handler{
		authzService: mockAuthz,
	}

	err := handler.TransactionCreate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestTransactionCreate_GiftCardNotFound(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(nil, assert.AnError)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	err := handler.TransactionCreate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestTransactionCreate_InvalidAmount(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()

	formData := url.Values{}
	formData.Set("amount", "invalid")
	formData.Set("description", "Test transaction")
	formData.Set("transaction_date", time.Now().Format("2006-01-02"))

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		CurrentBalance: 100.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	err := handler.TransactionCreate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestTransactionCreate_NegativeAmount(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()

	formData := url.Values{}
	formData.Set("amount", "-10.00")
	formData.Set("description", "Test transaction")
	formData.Set("transaction_date", time.Now().Format("2006-01-02"))

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		CurrentBalance: 100.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	err := handler.TransactionCreate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // Returns form with error
	assert.Contains(t, rec.Body.String(), "Der Betrag muss positiv sein")
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestTransactionCreate_InsufficientBalance(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()

	formData := url.Values{}
	formData.Set("amount", "150.00") // More than balance
	formData.Set("description", "Test transaction")
	formData.Set("transaction_date", time.Now().Format("2006-01-02"))

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		CurrentBalance: 100.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	err := handler.TransactionCreate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // Returns form with error
	assert.Contains(t, rec.Body.String(), "Nicht genügend Guthaben")
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestTransactionCreate_InvalidDate(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()

	formData := url.Values{}
	formData.Set("amount", "10.00")
	formData.Set("description", "Test transaction")
	formData.Set("transaction_date", "invalid-date")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		CurrentBalance: 100.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	err := handler.TransactionCreate(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestTransactionDelete_InvalidGiftCardID(t *testing.T) {
	e := echo.New()
	transactionID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/invalid-id/transactions/"+transactionID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "transaction_id")
	c.SetParamValues("invalid-id", transactionID.String())

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	handler := &Handler{}

	err := handler.TransactionDelete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransactionDelete_InvalidTransactionID(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String()+"/transactions/invalid-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "transaction_id")
	c.SetParamValues(giftCardID.String(), "invalid-id")

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	handler := &Handler{}

	err := handler.TransactionDelete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransactionDelete_Forbidden(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	transactionID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String()+"/transactions/"+transactionID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "transaction_id")
	c.SetParamValues(giftCardID.String(), transactionID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: false, // Not allowed to edit transactions
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	handler := &Handler{
		authzService: mockAuthz,
	}

	err := handler.TransactionDelete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestTransactionDelete_GiftCardNotFound(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	transactionID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String()+"/transactions/"+transactionID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "transaction_id")
	c.SetParamValues(giftCardID.String(), transactionID.String())

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEditTransactions: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(nil, assert.AnError)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	err := handler.TransactionDelete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}
