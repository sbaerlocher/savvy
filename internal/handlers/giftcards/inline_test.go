package giftcards

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestEditInline_Success(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/edit-inline", nil)
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
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	mockMerchantService := new(MockMerchantService)
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
		merchantService: mockMerchantService,
	}

	err := handler.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestEditInline_Forbidden(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/edit-inline", nil)
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
		CanView: true,
		CanEdit: false, // Not allowed to edit
		IsOwner: false,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	handler := &Handler{
		authzService: mockAuthz,
	}

	err := handler.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestEditInline_InvalidID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/invalid-id/edit-inline", nil)
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

	err := handler.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCancelEdit_Success(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/"+giftCardID.String()+"/cancel-edit", nil)
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
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	mockFavoriteService := new(MockFavoriteService)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "gift_card", giftCardID).Return(false, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
		favoriteService: mockFavoriteService,
	}

	err := handler.CancelEdit(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestCancelEdit_InvalidID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/gift-cards/invalid-id/cancel-edit", nil)
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

	err := handler.CancelEdit(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateInline_Success(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()
	merchantID := uuid.New()

	expiresAt := time.Now().Add(365 * 24 * time.Hour).Format("2006-01-02")

	formData := url.Values{}
	formData.Set("card_number", "9876543210")
	formData.Set("initial_balance", "150.00")
	formData.Set("currency", "CHF")
	formData.Set("pin", "1234")
	formData.Set("expires_at", expiresAt)
	formData.Set("status", "active")
	formData.Set("barcode_type", "CODE128")
	formData.Set("notes", "Updated notes")
	formData.Set("merchant_id", merchantID.String())

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/update-inline", strings.NewReader(formData.Encode()))
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
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil).Twice()
	mockGiftCardService.On("UpdateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(nil)

	mockMerchantService := new(MockMerchantService)
	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Test Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	mockFavoriteService := new(MockFavoriteService)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "gift_card", giftCardID).Return(false, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
		merchantService: mockMerchantService,
		favoriteService: mockFavoriteService,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestUpdateInline_Forbidden(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()

	formData := url.Values{}
	formData.Set("card_number", "9876543210")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/update-inline", strings.NewReader(formData.Encode()))
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
		CanView: true,
		CanEdit: false, // Not allowed to edit
		IsOwner: false,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	handler := &Handler{
		authzService: mockAuthz,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthz.AssertExpectations(t)
}

func TestUpdateInline_ServiceError(t *testing.T) {
	e := echo.New()
	giftCardID := uuid.New()

	formData := url.Values{}
	formData.Set("card_number", "9876543210")
	formData.Set("initial_balance", "150.00")

	req := httptest.NewRequest(http.MethodPost, "/gift-cards/"+giftCardID.String()+"/update-inline", strings.NewReader(formData.Encode()))
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
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckGiftCardAccess", mock.Anything, userID, giftCardID).Return(perms, nil)

	mockGiftCardService := new(MockGiftCardService)
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Currency:       "CHF",
	}
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)
	mockGiftCardService.On("UpdateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(assert.AnError)

	handler := &Handler{
		authzService:    mockAuthz,
		giftCardService: mockGiftCardService,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}
