package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupGiftCardsTest() (*GiftCardsHandler, *mocks.MockGiftCardServiceInterface, *mocks.MockAuthzServiceInterface, *mocks.MockMerchantServiceInterface, *mocks.MockUserServiceInterface, *mocks.MockFavoriteServiceInterface, *mocks.MockShareServiceInterface, *mocks.MockTransferServiceInterface) {
	mockGiftCardService := new(mocks.MockGiftCardServiceInterface)
	mockAuthzService := new(mocks.MockAuthzServiceInterface)
	mockMerchantService := new(mocks.MockMerchantServiceInterface)
	mockUserService := new(mocks.MockUserServiceInterface)
	mockFavoriteService := new(mocks.MockFavoriteServiceInterface)
	mockShareService := new(mocks.MockShareServiceInterface)
	mockTransferService := new(mocks.MockTransferServiceInterface)

	handler := NewGiftCardsHandler(
		mockGiftCardService,
		mockAuthzService,
		mockMerchantService,
		mockUserService,
		mockFavoriteService,
		mockShareService,
		mockTransferService,
	)

	return handler, mockGiftCardService, mockAuthzService, mockMerchantService, mockUserService, mockFavoriteService, mockShareService, mockTransferService
}

func createTestGiftCard() *models.GiftCard {
	giftCardID := uuid.New()
	userID := uuid.New()
	merchantID := uuid.New()
	return &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		MerchantID:     &merchantID,
		MerchantName:   "Test Merchant",
		CardNumber:     "1234567890",
		InitialBalance: 100.00,
		CurrentBalance: 75.50,
		Currency:       "EUR",
	}
}

func createTestTransaction() *models.GiftCardTransaction {
	txID := uuid.New()
	giftCardID := uuid.New()
	return &models.GiftCardTransaction{
		ID:              txID,
		GiftCardID:      giftCardID,
		Amount:          -25.50,
		Description:     "Purchase",
		TransactionDate: time.Now(),
	}
}

// ==================== List Tests ====================

func TestGiftCardsHandler_List_Success(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, mockFavoriteService, mockShareService, _ := setupGiftCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/gift-cards", "")
	user := createTestUser()
	c.Set("current_user", user)

	giftCards := []models.GiftCard{*createTestGiftCard()}
	mockGiftCardService.On("GetUserGiftCards", mock.Anything, user.ID).Return(giftCards, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "gift_card", giftCards[0].ID).Return(false, nil)
	mockShareService.On("GetGiftCardShareCounts", mock.Anything, mock.Anything).Return(map[uuid.UUID]int64{}, nil)

	// Mock authorization check for each gift card
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCards[0].ID).Return(&services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}, nil)

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response GiftCardListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response.GiftCards, 1)
	mockGiftCardService.AssertExpectations(t)
	mockAuthzService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestGiftCardsHandler_List_ServiceError(t *testing.T) {
	handler, mockGiftCardService, _, _, _, _, _, _ := setupGiftCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/gift-cards", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockGiftCardService.On("GetUserGiftCards", mock.Anything, user.ID).Return([]models.GiftCard(nil), errors.New("database error"))

	err := handler.List(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockGiftCardService.AssertExpectations(t)
}

// ==================== Show Tests ====================

func TestGiftCardsHandler_Show_Success(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, mockFavoriteService, mockShareService, _ := setupGiftCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/gift-cards/:id", "")
	user := createTestUser()
	giftCard := createTestGiftCard()
	giftCard.Transactions = []models.GiftCardTransaction{*createTestTransaction()}
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCard.ID.String()}})

	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCard.ID).Return(perms, nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCard.ID).Return(giftCard, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "gift_card", giftCard.ID).Return(false, nil)
	mockShareService.On("GetGiftCardShares", mock.Anything, giftCard.ID).Return([]models.GiftCardShare{}, nil)

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response GiftCardDetailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, giftCard.ID.String(), response.GiftCard.ID)
	assert.True(t, response.Permissions.IsOwner)
	assert.True(t, response.Permissions.CanEditTransactions)
	assert.Len(t, response.Transactions, 1)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestGiftCardsHandler_Show_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/gift-cards/:id", "")
	user := createTestUser()
	giftCardID := uuid.New()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return((*services.ResourcePermissions)(nil), errors.New("forbidden"))

	err := handler.Show(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== Create Tests ====================

func TestGiftCardsHandler_Create_Success(t *testing.T) {
	handler, mockGiftCardService, _, mockMerchantService, _, _, _, _ := setupGiftCardsTest()
	merchantID := uuid.New()
	body := `{"merchant_id":"` + merchantID.String() + `","card_number":"1234567890","initial_balance":100,"currency":"EUR"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	mockGiftCardService.On("CheckDuplicate", mock.Anything, "1234567890", user.ID, (*uuid.UUID)(nil)).Return((*models.GiftCard)(nil), nil)
	mockGiftCardService.On("FindDeletedDuplicate", mock.Anything, "1234567890", user.ID).Return((*models.GiftCard)(nil), nil)
	mockGiftCardService.On("CreateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(createTestGiftCard(), nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockMerchantService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestGiftCardsHandler_Create_MissingFields(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupGiftCardsTest()
	body := `{"card_number":"1234567890"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGiftCardsHandler_Create_InvalidExpiresAt(t *testing.T) {
	handler, _, _, mockMerchantService, _, _, _, _ := setupGiftCardsTest()
	merchantID := uuid.New()
	body := `{"merchant_id":"` + merchantID.String() + `","card_number":"1234567890","initial_balance":100,"currency":"EUR","expires_at":"invalid-date"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	// Mock merchant lookup (happens before date validation)
	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockMerchantService.AssertExpectations(t)
}

// ==================== Update Tests ====================

func TestGiftCardsHandler_Update_Success(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, mockFavoriteService, mockShareService, _ := setupGiftCardsTest()
	giftCard := createTestGiftCard()
	body := `{"card_number":"0987654321"}`
	c, rec := createTestContext(http.MethodPut, "/api/v1/gift-cards/:id", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCard.ID.String()}})

	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
		IsOwner:             true,
	}

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCard.ID).Return(perms, nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCard.ID).Return(giftCard, nil).Twice()
	mockGiftCardService.On("CheckDuplicate", mock.Anything, "0987654321", user.ID, &giftCard.ID).Return((*models.GiftCard)(nil), nil)
	mockGiftCardService.On("UpdateGiftCard", mock.Anything, mock.AnythingOfType("*models.GiftCard")).Return(nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "gift_card", giftCard.ID).Return(false, nil)
	mockShareService.On("GetGiftCardShares", mock.Anything, giftCard.ID).Return([]models.GiftCardShare{}, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestGiftCardsHandler_Update_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/gift-cards/:id", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{CanView: true, CanEdit: false}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== Delete Tests ====================

func TestGiftCardsHandler_Delete_Success(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/gift-cards/:id", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{
		CanView:   true,
		CanDelete: true,
	}

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("DeleteGiftCard", mock.Anything, giftCardID).Return(nil)

	err := handler.Delete(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

// ==================== ToggleFavorite Tests ====================

func TestGiftCardsHandler_ToggleFavorite_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, mockFavoriteService, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/favorite", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{CanView: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockFavoriteService.On("ToggleFavorite", mock.Anything, user.ID, "gift_card", giftCardID).Return(nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "gift_card", giftCardID).Return(true, nil)

	err := handler.ToggleFavorite(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

// ==================== CreateShare Tests ====================

func TestGiftCardsHandler_CreateShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, mockShareService, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	sharedUserID := uuid.New()
	body := `{"emails":["shared@example.com"],"can_edit":true,"can_delete":false,"can_edit_transactions":true}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "shared@example.com").Return(sharedUser, nil)
	mockShareService.On("CreateGiftCardShare", mock.Anything, mock.Anything, giftCardID, sharedUserID, true, false, true).Return(nil)
	mockShareService.On("GetGiftCardShares", mock.Anything, giftCardID).Return([]models.GiftCardShare{}, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp ShareCreateResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, 1, resp.SuccessCount)
	assert.Empty(t, resp.Failed)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

func TestGiftCardsHandler_CreateShare_NotOwner(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	body := `{"email":"shared@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: false}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

// ==================== UpdateShare Tests ====================

func TestGiftCardsHandler_UpdateShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	sharedWithID := uuid.New()
	body := `{"can_edit":true,"can_delete":true,"can_edit_transactions":false}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/gift-cards/:id/share/:sharedWithID", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockShareService.On("UpdateGiftCardShare", mock.Anything, mock.Anything, giftCardID, sharedWithID, true, true, false).Return(nil)
	mockShareService.On("GetGiftCardShares", mock.Anything, giftCardID).Return([]models.GiftCardShare{}, nil)

	err := handler.UpdateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

// ==================== DeleteShare Tests ====================

func TestGiftCardsHandler_DeleteShare_Success(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	sharedWithID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/gift-cards/:id/share/:sharedWithID", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockShareService.On("DeleteGiftCardShare", mock.Anything, mock.Anything, giftCardID, sharedWithID).Return(nil)

	err := handler.DeleteShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockShareService.AssertExpectations(t)
}

// ==================== Transfer Tests ====================

func TestGiftCardsHandler_Transfer_Success(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, _, mockTransferService := setupGiftCardsTest()
	giftCardID := uuid.New()
	newOwnerID := uuid.New()
	body := `{"new_owner_email":"newowner@example.com"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/transfer", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	newOwner := &models.User{ID: newOwnerID, Email: "newowner@example.com"}

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "newowner@example.com").Return(newOwner, nil)
	mockTransferService.On("TransferGiftCardOwnership", mock.Anything, giftCardID, newOwnerID, user.ID).Return(nil)

	err := handler.Transfer(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockTransferService.AssertExpectations(t)
}

// ==================== Transaction Tests ====================

func TestGiftCardsHandler_ListTransactions_Success(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCard := createTestGiftCard()
	giftCard.Transactions = []models.GiftCardTransaction{*createTestTransaction()}
	c, rec := createTestContext(http.MethodGet, "/api/v1/gift-cards/:id/transactions", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCard.ID.String()}})

	perms := &services.ResourcePermissions{CanView: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCard.ID).Return(perms, nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCard.ID).Return(giftCard, nil)

	err := handler.ListTransactions(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string][]GiftCardTransactionDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Len(t, response["transactions"], 1)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestGiftCardsHandler_ListTransactions_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodGet, "/api/v1/gift-cards/:id/transactions", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return((*services.ResourcePermissions)(nil), errors.New("forbidden"))

	err := handler.ListTransactions(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestGiftCardsHandler_CreateTransaction_Success(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	body := `{"amount":-25.50,"description":"Purchase"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/transactions", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{CanEditTransactions: true}
	giftCard := createTestGiftCard()
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.GiftCardTransaction")).Return(nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	err := handler.CreateTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestGiftCardsHandler_CreateTransaction_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	body := `{"amount":-25.50}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/transactions", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{CanView: true, CanEditTransactions: false}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)

	err := handler.CreateTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestGiftCardsHandler_CreateTransaction_ZeroAmount(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	body := `{"amount":0}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/transactions", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{CanEditTransactions: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)

	err := handler.CreateTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestGiftCardsHandler_DeleteTransaction_Success(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	transactionID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/gift-cards/:id/transactions/:transactionID", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}, {Name: "transactionID", Value: transactionID.String()}})

	perms := &services.ResourcePermissions{CanEditTransactions: true}
	transaction := &models.GiftCardTransaction{ID: transactionID, GiftCardID: giftCardID}
	giftCard := createTestGiftCard()

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("GetTransaction", mock.Anything, transactionID, giftCardID).Return(transaction, nil)
	mockGiftCardService.On("DeleteTransaction", mock.Anything, transactionID).Return(nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return(giftCard, nil)

	err := handler.DeleteTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestGiftCardsHandler_DeleteTransaction_Forbidden(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	transactionID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/gift-cards/:id/transactions/:transactionID", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}, {Name: "transactionID", Value: transactionID.String()}})

	perms := &services.ResourcePermissions{CanView: true, CanEditTransactions: false}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)

	err := handler.DeleteTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
}

func TestGiftCardsHandler_DeleteTransaction_WrongGiftCard(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	transactionID := uuid.New()
	wrongGiftCardID := uuid.New()
	c, rec := createTestContext(http.MethodDelete, "/api/v1/gift-cards/:id/transactions/:transactionID", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}, {Name: "transactionID", Value: transactionID.String()}})

	perms := &services.ResourcePermissions{CanEditTransactions: true}
	transaction := &models.GiftCardTransaction{ID: transactionID, GiftCardID: wrongGiftCardID}

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("GetTransaction", mock.Anything, transactionID, giftCardID).Return(transaction, nil)

	err := handler.DeleteTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

// ==================== Additional Update Tests for Coverage ====================

func TestGiftCardsHandler_Update_InvalidID(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupGiftCardsTest()
	c, rec := createTestContext(http.MethodPut, "/api/v1/gift-cards/invalid", `{}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid"}})

	err := handler.Update(c)

	// parseResourceID writes JSON response AND returns error
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_id", response.Error)
}

func TestGiftCardsHandler_Update_NotFound(t *testing.T) {
	handler, mockGiftCardService, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/gift-cards/"+giftCardID.String(), `{"card_number":"123"}`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{CanEdit: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockGiftCardService.On("GetGiftCard", mock.Anything, giftCardID).Return((*models.GiftCard)(nil), errors.New("not found"))

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "not_found", response.Error)
	mockAuthzService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestGiftCardsHandler_Update_InvalidRequestBody(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodPut, "/api/v1/gift-cards/"+giftCardID.String(), `invalid json`)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{CanEdit: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)

	err := handler.Update(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "invalid_request", response.Error)
	mockAuthzService.AssertExpectations(t)
}

// ==================== Error Message Sanitization Tests ====================

func TestGiftCardsHandler_CreateShare_ServiceError_NoLeakedDetails(t *testing.T) {
	handler, _, mockAuthzService, _, mockUserService, _, mockShareService, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	sharedUserID := uuid.New()
	body := `{"emails":["shared@example.com"],"can_edit":true,"can_delete":false,"can_edit_transactions":true}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/share", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	sharedUser := &models.User{ID: sharedUserID, Email: "shared@example.com"}

	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "shared@example.com").Return(sharedUser, nil)
	// Simulate a GORM error with internal DB details
	mockShareService.On("CreateGiftCardShare", mock.Anything, mock.Anything, giftCardID, sharedUserID, true, false, true).
		Return(errors.New("ERROR: duplicate key value violates unique constraint \"gift_card_shares_pkey\" (SQLSTATE 23505)"))
	mockShareService.On("GetGiftCardShares", mock.Anything, giftCardID).Return([]models.GiftCardShare{}, nil)

	err := handler.CreateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	// Failed entries carry a generic reason, never the raw DB error.
	body2 := rec.Body.String()
	assert.NotContains(t, body2, "duplicate key")
	assert.NotContains(t, body2, "gift_card_shares_pkey")
	assert.NotContains(t, body2, "SQLSTATE")
	var resp ShareCreateResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, 0, resp.SuccessCount)
	assert.Len(t, resp.Failed, 1)
	assert.Equal(t, "share failed", resp.Failed[0].Error)
	mockShareService.AssertExpectations(t)
}

func TestGiftCardsHandler_UpdateShare_ServiceError_NoLeakedDetails(t *testing.T) {
	handler, _, mockAuthzService, _, _, _, mockShareService, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	sharedWithID := uuid.New()
	body := `{"can_edit":true,"can_delete":false,"can_edit_transactions":true}`
	c, rec := createTestContext(http.MethodPatch, "/api/v1/gift-cards/:id/share/:sharedWithID", body)
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}, {Name: "sharedWithID", Value: sharedWithID.String()}})

	perms := &services.ResourcePermissions{IsOwner: true}
	mockAuthzService.On("CheckGiftCardAccess", mock.Anything, user.ID, giftCardID).Return(perms, nil)
	// Simulate a GORM error with table/column details
	mockShareService.On("UpdateGiftCardShare", mock.Anything, mock.Anything, giftCardID, sharedWithID, true, false, true).
		Return(errors.New("record not found: SELECT * FROM \"gift_card_shares\" WHERE gift_card_id = $1 AND shared_with_id = $2"))

	err := handler.UpdateShare(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to update share permissions", response.Message)
	assert.NotContains(t, response.Message, "gift_card_shares")
	assert.NotContains(t, response.Message, "SELECT")
	mockShareService.AssertExpectations(t)
}

// ==================== checkGiftCardDuplicate Tests ====================

func TestCheckGiftCardDuplicate_NilCardNumber(t *testing.T) {
	result := checkGiftCardDuplicate(nil, nil, nil, uuid.New(), nil)
	assert.Nil(t, result)
}

func TestCheckGiftCardDuplicate_EmptyCardNumber(t *testing.T) {
	empty := ""
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)
	result := checkGiftCardDuplicate(c, nil, &empty, uuid.New(), nil)
	assert.Nil(t, result)
}

func TestCheckGiftCardDuplicate_NoDuplicate(t *testing.T) {
	mockSvc := new(mocks.MockGiftCardServiceInterface)
	cardNum := "GC123"
	userID := uuid.New()
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)

	mockSvc.On("CheckDuplicate", mock.Anything, cardNum, userID, (*uuid.UUID)(nil)).Return((*models.GiftCard)(nil), nil)

	result := checkGiftCardDuplicate(c, mockSvc, &cardNum, userID, nil)
	assert.Nil(t, result)
	mockSvc.AssertExpectations(t)
}

func TestCheckGiftCardDuplicate_Found(t *testing.T) {
	mockSvc := new(mocks.MockGiftCardServiceInterface)
	cardNum := "GC123"
	userID := uuid.New()
	dupID := uuid.New()
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)

	dup := &models.GiftCard{CardNumber: "GC123"}
	dup.ID = dupID
	dup.MerchantName = "TestMerchant"
	mockSvc.On("CheckDuplicate", mock.Anything, cardNum, userID, (*uuid.UUID)(nil)).Return(dup, nil)

	result := checkGiftCardDuplicate(c, mockSvc, &cardNum, userID, nil)
	assert.NotNil(t, result)
	assert.True(t, result.HasDuplicate)
	assert.Equal(t, "TestMerchant", result.MerchantName)
	assert.Equal(t, "GC123", result.ResourceNumber)
	mockSvc.AssertExpectations(t)
}

func TestCheckGiftCardDuplicate_ServiceError(t *testing.T) {
	mockSvc := new(mocks.MockGiftCardServiceInterface)
	cardNum := "GC123"
	userID := uuid.New()
	e := echo.New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	c := e.NewContext(req, nil)

	mockSvc.On("CheckDuplicate", mock.Anything, cardNum, userID, (*uuid.UUID)(nil)).Return((*models.GiftCard)(nil), errors.New("db error"))

	result := checkGiftCardDuplicate(c, mockSvc, &cardNum, userID, nil)
	assert.Nil(t, result)
	mockSvc.AssertExpectations(t)
}

// ==================== Restore Tests ====================

func TestGiftCardsHandler_Create_DeletedDuplicate_Returns409WithDeletedFlag(t *testing.T) {
	handler, mockGiftCardService, _, mockMerchantService, _, _, _, _ := setupGiftCardsTest()
	merchantID := uuid.New()
	body := `{"merchant_id":"` + merchantID.String() + `","card_number":"1234567890","initial_balance":100,"currency":"EUR"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards", body)
	user := createTestUser()
	c.Set("current_user", user)

	deletedGiftCard := createTestGiftCard()
	deletedGiftCard.ID = uuid.New()
	deletedGiftCard.UserID = &user.ID
	deletedGiftCard.CardNumber = "1234567890"
	deletedGiftCard.MerchantName = "Test Merchant"

	merchant := &models.Merchant{ID: merchantID, Name: "Test Merchant"}
	mockMerchantService.On("GetMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
	// No active duplicate
	mockGiftCardService.On("CheckDuplicate", mock.Anything, "1234567890", user.ID, (*uuid.UUID)(nil)).Return((*models.GiftCard)(nil), nil)
	// Soft-deleted duplicate found
	mockGiftCardService.On("FindDeletedDuplicate", mock.Anything, "1234567890", user.ID).Return(deletedGiftCard, nil)

	err := handler.Create(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response DuplicateErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "duplicate_barcode", response.Error)
	assert.NotNil(t, response.Duplicate)
	assert.True(t, response.Duplicate.Deleted, "expected duplicate.deleted == true")
	assert.Equal(t, deletedGiftCard.ID.String(), response.Duplicate.ExistingID)
	mockMerchantService.AssertExpectations(t)
	mockGiftCardService.AssertExpectations(t)
}

func TestGiftCardsHandler_Restore_Success(t *testing.T) {
	handler, mockGiftCardService, _, _, _, mockFavoriteService, _, _ := setupGiftCardsTest()
	restoredGiftCard := createTestGiftCard()
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: restoredGiftCard.ID.String()}})

	mockGiftCardService.On("RestoreGiftCard", mock.Anything, restoredGiftCard.ID, user.ID).Return(restoredGiftCard, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "gift_card", restoredGiftCard.ID).Return(false, nil)

	err := handler.Restore(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response GiftCardDetailResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, restoredGiftCard.ID.String(), response.GiftCard.ID)
	mockGiftCardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestGiftCardsHandler_Restore_NotFound(t *testing.T) {
	handler, mockGiftCardService, _, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	mockGiftCardService.On("RestoreGiftCard", mock.Anything, giftCardID, user.ID).Return((*models.GiftCard)(nil), nil)

	err := handler.Restore(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockGiftCardService.AssertExpectations(t)
}

func TestGiftCardsHandler_Restore_InvalidID(t *testing.T) {
	handler, _, _, _, _, _, _, _ := setupGiftCardsTest()
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "invalid-uuid"}})

	err := handler.Restore(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGiftCardsHandler_Restore_ServiceError(t *testing.T) {
	handler, mockGiftCardService, _, _, _, _, _, _ := setupGiftCardsTest()
	giftCardID := uuid.New()
	c, rec := createTestContext(http.MethodPost, "/api/v1/gift-cards/:id/restore", "")
	user := createTestUser()
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: giftCardID.String()}})

	mockGiftCardService.On("RestoreGiftCard", mock.Anything, giftCardID, user.ID).Return((*models.GiftCard)(nil), errors.New("db error"))

	err := handler.Restore(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockGiftCardService.AssertExpectations(t)
}
