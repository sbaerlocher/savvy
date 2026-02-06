package cards

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	savvyi18n "savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/services"
)

func TestEditInline_Success(t *testing.T) {
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/"+cardID.String()+"/edit-inline", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

	// Setup i18n context
	localizer := savvyi18n.NewLocalizer("de")
	ctx := savvyi18n.SetLocalizer(c.Request().Context(), localizer)
	c.SetRequest(c.Request().WithContext(ctx))

	// Setup user context
	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set("current_user", user)
	c.Set("csrf", "test-csrf-token")

	// Mock services
	mockAuthz := new(MockAuthzService)
	perms := &services.ResourcePermissions{
		CanView: true,
		CanEdit: true,
		IsOwner: true,
	}
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	mockCardService := new(MockCardService)
	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil)

	mockMerchantService := new(MockMerchantServiceNew)
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant"},
	}
	mockMerchantService.On("GetAllMerchants", mock.Anything).Return(merchants, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		cardService:     mockCardService,
		merchantService: mockMerchantService,
	}

	err := handler.EditInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
}

func TestEditInline_Forbidden(t *testing.T) {
	e := echo.New()
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/"+cardID.String()+"/edit-inline", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

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
	req := httptest.NewRequest(http.MethodGet, "/cards/invalid-id/edit-inline", nil)
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
	cardID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/"+cardID.String()+"/cancel-edit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	mockCardService := new(MockCardService)
	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil)

	mockFavoriteService := new(MockFavoriteService)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "card", cardID).Return(false, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		cardService:     mockCardService,
		favoriteService: mockFavoriteService,
	}

	err := handler.CancelEdit(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestCancelEdit_InvalidID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/cards/invalid-id/cancel-edit", nil)
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
	cardID := uuid.New()
	merchantID := uuid.New()

	formData := url.Values{}
	formData.Set("program", "Gold")
	formData.Set("card_number", "5678")
	formData.Set("barcode_type", "CODE128")
	formData.Set("status", "active")
	formData.Set("notes", "Updated notes")
	formData.Set("merchant_id", merchantID.String())

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String()+"/update-inline", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	mockCardService := new(MockCardService)
	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil).Twice()
	mockCardService.On("UpdateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)

	mockMerchantService := new(MockMerchantServiceNew)
	merchant := &models.Merchant{
		ID:   merchantID,
		Name: "Test Merchant",
	}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	mockFavoriteService := new(MockFavoriteService)
	mockFavoriteService.On("IsFavorite", mock.Anything, userID, "card", cardID).Return(false, nil)

	handler := &Handler{
		authzService:    mockAuthz,
		cardService:     mockCardService,
		merchantService: mockMerchantService,
		favoriteService: mockFavoriteService,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
	mockMerchantService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestUpdateInline_Forbidden(t *testing.T) {
	e := echo.New()
	cardID := uuid.New()

	formData := url.Values{}
	formData.Set("card_number", "5678")

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String()+"/update-inline", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

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
	cardID := uuid.New()

	formData := url.Values{}
	formData.Set("card_number", "5678")

	req := httptest.NewRequest(http.MethodPost, "/cards/"+cardID.String()+"/update-inline", strings.NewReader(formData.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cardID.String())

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
	mockAuthz.On("CheckCardAccess", mock.Anything, userID, cardID).Return(perms, nil)

	mockCardService := new(MockCardService)
	card := &models.Card{
		ID:         cardID,
		UserID:     &userID,
		CardNumber: "1234",
	}
	mockCardService.On("GetCard", mock.Anything, cardID).Return(card, nil)
	mockCardService.On("UpdateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(assert.AnError)

	handler := &Handler{
		authzService: mockAuthz,
		cardService:  mockCardService,
	}

	err := handler.UpdateInline(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockAuthz.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}
