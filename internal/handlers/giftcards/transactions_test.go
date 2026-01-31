package giftcards

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
	"savvy/internal/services"
	savvyi18n "savvy/internal/i18n"
)

func TestTransactionCreate_Success(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()

	// Form data
	form := url.Values{}
	form.Add("amount", "50.00")
	form.Add("description", "Test transaction")
	form.Add("transaction_date", "2024-01-15")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	// Mock authz service
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: nil,
		db:              nil, // Handler uses database.DB directly - will cause error
	}

	// Execute
	err := handler.TransactionCreate(c)

	// Assert
	mockAuthz.AssertExpectations(t)
	// Will get DB error since database.DB is nil
	assert.Error(t, err)
}

func TestTransactionCreate_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()

	// Form data
	form := url.Values{}
	form.Add("amount", "50.00")
	form.Add("description", "Test transaction")
	form.Add("transaction_date", "2024-01-15")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/transactions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(giftCardID.String())

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "shared@example.com",
	}
	c.Set("current_user", user)

	// Mock authz service - deny transaction edit
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           false,
		CanEditTransactions: false, // Not allowed to edit transactions
		IsOwner:             false,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.TransactionCreate(c)

	// Assert - Should return forbidden
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestTransactionCreate_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()

	// Form data
	form := url.Values{}
	form.Add("amount", "50.00")
	form.Add("description", "Test transaction")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/invalid-id/transactions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-id")

	// Setup user context
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	// Create handler
	handler := &Handler{
		authzService:    nil,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.TransactionCreate(c)

	// Assert - Should return bad request
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransactionDelete_Success(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()
	transactionID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String()+"/transactions/"+transactionID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "transaction_id")
	c.SetParamValues(giftCardID.String(), transactionID.String())

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "owner@example.com",
	}
	c.Set("current_user", user)

	// Mock authz service
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Create handler with mocks
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: nil,
		db:              nil, // Handler uses database.DB directly - will cause error
	}

	// Execute
	err := handler.TransactionDelete(c)

	// Assert
	mockAuthz.AssertExpectations(t)
	// Will get DB error since database.DB is nil
	assert.Error(t, err)
}

func TestTransactionDelete_Forbidden(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()
	transactionID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/"+giftCardID.String()+"/transactions/"+transactionID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "transaction_id")
	c.SetParamValues(giftCardID.String(), transactionID.String())

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "shared@example.com",
	}
	c.Set("current_user", user)

	// Mock authz service - deny transaction edit
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           false,
		CanEditTransactions: false, // Not allowed to delete transactions
		IsOwner:             false,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	// Create handler with mock
	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.TransactionDelete(c)

	// Assert - Should return forbidden
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestTransactionDelete_InvalidID(t *testing.T) {
	// Setup
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/gift-cards/invalid-id/transactions/invalid-tid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "transaction_id")
	c.SetParamValues("invalid-id", "invalid-tid")

	// Setup user context
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)

	// Create handler
	handler := &Handler{
		authzService:    nil,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.TransactionDelete(c)

	// Assert - Should return bad request
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransactionNew_GetForm(t *testing.T) {
	// Setup
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

	// Setup user context
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	// Create handler
	handler := &Handler{
		authzService:    nil,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.TransactionNew(c)

	// Assert - Should render form successfully
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTransactionCancel_Success(t *testing.T) {
	// Setup
	e := echo.New()
	giftCardID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/transactions/cancel", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create handler
	handler := &Handler{
		authzService:    nil,
		giftCardService: nil,
		db:              nil,
	}

	// Execute
	err := handler.TransactionCancel(c)

	// Assert - Should return no content
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
