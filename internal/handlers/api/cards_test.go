package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupCardsTest() (*CardsHandler, *mocks.MockCardServiceInterface, *mocks.MockAuthzServiceInterface, *mocks.MockMerchantServiceInterface, *mocks.MockUserServiceInterface, *mocks.MockFavoriteServiceInterface, *mocks.MockShareServiceInterface, *mocks.MockTransferServiceInterface, *mocks.MockAdminServiceInterface) {
	mockCardService := new(mocks.MockCardServiceInterface)
	mockAuthzService := new(mocks.MockAuthzServiceInterface)
	mockMerchantService := new(mocks.MockMerchantServiceInterface)
	mockUserService := new(mocks.MockUserServiceInterface)
	mockFavoriteService := new(mocks.MockFavoriteServiceInterface)
	mockShareService := new(mocks.MockShareServiceInterface)
	mockTransferService := new(mocks.MockTransferServiceInterface)
	mockAdminService := new(mocks.MockAdminServiceInterface)

	handler := NewCardsHandler(
		mockCardService,
		mockAuthzService,
		mockMerchantService,
		mockUserService,
		mockFavoriteService,
		mockShareService,
		mockTransferService,
		mockAdminService,
	)

	return handler, mockCardService, mockAuthzService, mockMerchantService, mockUserService, mockFavoriteService, mockShareService, mockTransferService, mockAdminService
}

func createTestContext(method, path string, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func createTestUser() *models.User {
	userID := uuid.New()
	return &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
}

func createTestCard() *models.Card {
	cardID := uuid.New()
	userID := uuid.New()
	merchantID := uuid.New()
	return &models.Card{
		ID:           cardID,
		UserID:       &userID,
		MerchantID:   &merchantID,
		MerchantName: "Test Merchant",
		CardNumber:   "1234567890",
		BarcodeType:  "CODE128",
		Status:       "active",
	}
}

// ==================== List Tests ====================

func TestCardsHandler_List_Success(t *testing.T) {
	handler, mockCardService, _, _, _, mockFavoriteService, mockShareService, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/cards", "")
	user := createTestUser()
	c.Set("current_user", user)

	cards := []models.Card{*createTestCard()}
	mockCardService.On("GetUserCards", mock.Anything, user.ID).Return(cards, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "card", cards[0].ID).Return(false, nil)
	mockShareService.On("GetCardShareCounts", mock.Anything, mock.Anything).Return(map[uuid.UUID]int64{}, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response CardListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response.Cards, 1)
	mockCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestCardsHandler_List_ServiceError(t *testing.T) {
	handler, mockCardService, _, _, _, _, _, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/cards", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockCardService.On("GetUserCards", mock.Anything, user.ID).Return([]models.Card(nil), errors.New("database error"))

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockCardService.AssertExpectations(t)
}

// ==================== Show Tests ====================

func TestCardsHandler_Show_Success(t *testing.T) {
	handler, mockCardService, mockAuthzService, _, _, mockFavoriteService, mockShareService, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/cards/:id", "")
	user := createTestUser()
	card := createTestCard()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: card.ID.String()}})

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, card.ID).Return(perms, nil)
	mockCardService.On("GetCard", mock.Anything, card.ID).Return(card, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "card", card.ID).Return(false, nil)
	mockShareService.On("GetCardShares", mock.Anything, card.ID).Return([]models.CardShare{}, nil)

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response CardDetailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, card.ID.String(), response.Card.ID)
	assert.True(t, response.Permissions.IsOwner)
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestCardsHandler_Show_InvalidID(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/cards/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid-uuid"}})

	err := handler.Show(c)

	// parseResourceID writes JSON response AND returns error
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCardsHandler_Show_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/cards/:id", "")
	user := createTestUser()
	cardID := uuid.New()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return((*services.ResourcePermissions)(nil), errors.New("forbidden"))

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestCardsHandler_Show_NotFound(t *testing.T) {
	handler, mockCardService, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/cards/:id", "")
	user := createTestUser()
	cardID := uuid.New()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{CanView: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockCardService.On("GetCard", mock.Anything, cardID).Return((*models.Card)(nil), errors.New("not found"))

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

// ==================== Create Tests ====================

func TestCardsHandler_Create_Success(t *testing.T) {
	handler, mockCardService, _, mockMerchantService, _, _, _, _, _ := setupCardsTest()
	merchantID := uuid.New()
	body := `{"merchant_id":"` + merchantID.String() + `","card_number":"1234567890","barcode_type":"CODE128"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockCardService.On("CheckDuplicate", mock.Anything, "1234567890", user.ID, (*uuid.UUID)(nil)).Return((*models.Card)(nil), nil)
	mockCardService.On("CreateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)
	mockCardService.On("GetCard", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(createTestCard(), nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockMerchantService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

func TestCardsHandler_Create_InvalidRequest(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", "invalid-json")
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCardsHandler_Create_MissingCardNumber(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupCardsTest()
	merchantID := uuid.New()
	body := `{"merchant_id":"` + merchantID.String() + `"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCardsHandler_Create_WithNewMerchant(t *testing.T) {
	handler, mockCardService, _, mockMerchantService, _, _, _, _, _ := setupCardsTest()
	body := `{"new_merchant_name":"New Merchant","card_number":"1234567890"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockMerchantService.On("CreateMerchant", mock.Anything, mock.AnythingOfType("*models.Merchant")).Return(nil)
	mockCardService.On("CheckDuplicate", mock.Anything, "1234567890", user.ID, (*uuid.UUID)(nil)).Return((*models.Card)(nil), nil)
	mockCardService.On("CreateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)
	mockCardService.On("GetCard", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(createTestCard(), nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockMerchantService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

func TestCardsHandler_Create_MissingMerchant(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupCardsTest()
	body := `{"card_number":"1234567890"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCardsHandler_Create_InvalidMerchantID(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupCardsTest()
	body := `{"merchant_id":"invalid-uuid","card_number":"1234567890"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCardsHandler_Create_MerchantNotFound(t *testing.T) {
	handler, _, _, mockMerchantService, _, _, _, _, _ := setupCardsTest()
	merchantID := uuid.New()
	body := `{"merchant_id":"` + merchantID.String() + `","card_number":"1234567890"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return((*models.Merchant)(nil), errors.New("not found"))

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

// ==================== Update Tests ====================

func TestCardsHandler_Update_Success(t *testing.T) {
	handler, mockCardService, mockAuthzService, _, _, mockFavoriteService, mockShareService, _, _ := setupCardsTest()
	card := createTestCard()
	body := `{"card_number":"0987654321"}`
	c, rec := createTestContext(http.MethodPut, "/api/v1/cards/:id", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: card.ID.String()}})

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanEdit:   true,
		CanDelete: true,
		IsOwner:   true,
	}

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, card.ID).Return(perms, nil)
	mockCardService.On("GetCard", mock.Anything, card.ID).Return(card, nil).Twice()
	mockCardService.On("CheckDuplicate", mock.Anything, "0987654321", user.ID, &card.ID).Return((*models.Card)(nil), nil)
	mockCardService.On("UpdateCard", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "card", card.ID).Return(false, nil)
	mockShareService.On("GetCardShares", mock.Anything, card.ID).Return([]models.CardShare{}, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestCardsHandler_Update_InvalidID(t *testing.T) {
	handler, _, _, _, _, _, _, _, _ := setupCardsTest()
	c, rec := createTestContext(http.MethodPut, "/api/v1/cards/:id", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid-uuid"}})

	err := handler.Update(c)

	// parseResourceID writes JSON response AND returns error
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCardsHandler_Update_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/cards/:id", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{CanView: true, CanEdit: false}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestCardsHandler_Update_InvalidRequest(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/cards/:id", "invalid-json")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{CanEdit: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestCardsHandler_Update_NotFound(t *testing.T) {
	handler, mockCardService, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/cards/:id", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{CanEdit: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockCardService.On("GetCard", mock.Anything, cardID).Return((*models.Card)(nil), errors.New("not found"))

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

// ==================== Delete Tests ====================

func TestCardsHandler_Delete_Success(t *testing.T) {
	handler, mockCardService, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/cards/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanDelete: true,
	}

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockCardService.On("DeleteCard", mock.Anything, cardID).Return(nil)

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockCardService.AssertExpectations(t)
}

func TestCardsHandler_Delete_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/cards/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{CanView: true, CanDelete: false}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== ToggleFavorite Tests ====================

func TestCardsHandler_ToggleFavorite_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, mockFavoriteService, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/favorite", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{CanView: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockFavoriteService.On("ToggleFavorite", mock.Anything, user.ID, "card", cardID).Return(nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "card", cardID).Return(true, nil)

	err := handler.ToggleFavorite(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]bool
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.True(t, response["is_favorite"])
	mockAuthzService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestCardsHandler_ToggleFavorite_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/favorite", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return((*services.ResourcePermissions)(nil), errors.New("forbidden"))

	err := handler.ToggleFavorite(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== CreateShare Tests ====================

func TestCardsHandler_CreateShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, mockShareService, _, _ := setupCardsTest()
	cardID := uuid.New()
	sharedUserID := uuid.New()
	body := `{"email":"shared@example.com","can_edit":true,"can_delete":false}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "shared@example.com").Return(sharedUser, nil)
	mockShareService.On("CreateCardShare", mock.Anything, mock.Anything, cardID, sharedUserID, true, false).Return(nil)
	mockShareService.On("GetCardShares", mock.Anything, cardID).Return([]models.CardShare{}, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestCardsHandler_CreateShare_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	body := `{"email":"shared@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestCardsHandler_CreateShare_UserNotFound(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	body := `{"email":"notfound@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return((*models.User)(nil), errors.New("not found"))

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "share_failed", response.Error)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestCardsHandler_CreateShare_SelfShare(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	body := `{"email":"test@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// ==================== UpdateShare Tests ====================

func TestCardsHandler_UpdateShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _, mockAdminService := setupCardsTest()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	body := `{"can_edit":true,"can_delete":true}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/cards/:id/share/:sharedWithID", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockShareService.On("UpdateCardShare", mock.Anything, mock.Anything, cardID, sharedWithID, true, true).Return(nil)
	mockShareService.On("GetCardShares", mock.Anything, cardID).Return([]models.CardShare{}, nil)
	mockAdminService.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil)

	err := handler.UpdateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
	mockAdminService.AssertExpectations(t)
}

func TestCardsHandler_UpdateShare_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := createTestContext(http.MethodPatch, "/api/v1/cards/:id/share/:sharedWithID", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)

	err := handler.UpdateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== DeleteShare Tests ====================

func TestCardsHandler_DeleteShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _, _ := setupCardsTest()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/cards/:id/share/:sharedWithID", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockShareService.On("DeleteCardShare", mock.Anything, mock.Anything, cardID, sharedWithID).Return(nil)

	err := handler.DeleteShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestCardsHandler_DeleteShare_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/cards/:id/share/:sharedWithID", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)

	err := handler.DeleteShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== Transfer Tests ====================

func TestCardsHandler_Transfer_Success(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, _, mockTransferService, _ := setupCardsTest()
	cardID := uuid.New()
	newOwnerID := uuid.New()
	body := `{"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockTransferService.On("TransferCardOwnership", mock.Anything, cardID, newOwnerID, user.ID).Return(nil)

	err := handler.Transfer(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockTransferService.AssertExpectations(t)
}

func TestCardsHandler_Transfer_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/transfer", `{"new_owner_email":"test@example.com"}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)

	err := handler.Transfer(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestCardsHandler_Transfer_SelfTransfer(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, _, _, _ := setupCardsTest()
	cardID := uuid.New()
	body := `{"new_owner_email":"test@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)

	err := handler.Transfer(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// ==================== Error Message Sanitization Tests ====================

func TestCardsHandler_CreateShare_ServiceError_NoLeakedDetails(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, mockShareService, _, _ := setupCardsTest()
	cardID := uuid.New()
	sharedUserID := uuid.New()
	body := `{"email":"shared@example.com","can_edit":true,"can_delete":false}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}

	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "shared@example.com").Return(sharedUser, nil)
	// Simulate a GORM error with internal DB details
	mockShareService.On("CreateCardShare", mock.Anything, mock.Anything, cardID, sharedUserID, true, false).
		Return(errors.New("ERROR: duplicate key value violates unique constraint \"card_shares_pkey\" (SQLSTATE 23505)"))

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to share card", response.Message)
	assert.NotContains(t, response.Message, "duplicate key")
	assert.NotContains(t, response.Message, "card_shares_pkey")
	assert.NotContains(t, response.Message, "SQLSTATE")
	mockShareService.AssertExpectations(t)
}

func TestCardsHandler_UpdateShare_ServiceError_NoLeakedDetails(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _, _ := setupCardsTest()
	cardID := uuid.New()
	sharedWithID := uuid.New()
	body := `{"can_edit":true,"can_delete":false}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/cards/:id/share/:sharedWithID", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: cardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckCardAccess", mock.Anything, user.ID, cardID).Return(perms, nil)
	// Simulate a GORM error with table/column details
	mockShareService.On("UpdateCardShare", mock.Anything, mock.Anything, cardID, sharedWithID, true, false).
		Return(errors.New("record not found: SELECT * FROM \"card_shares\" WHERE card_id = $1 AND shared_with_id = $2"))

	err := handler.UpdateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to update share permissions", response.Message)
	assert.NotContains(t, response.Message, "card_shares")
	assert.NotContains(t, response.Message, "SELECT")
	mockShareService.AssertExpectations(t)
}
