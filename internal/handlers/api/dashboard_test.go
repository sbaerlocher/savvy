package api //nolint:revive // "api" is a meaningful package name for API handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"savvy/internal/mocks"
	"savvy/internal/models"
	"savvy/internal/repository"
	"savvy/internal/services"
)

// ==================== Helper Functions ====================

func setupDashboardTest() (*DashboardHandler, *mocks.MockDashboardServiceInterface, *mocks.MockFavoriteServiceInterface) {
	mockDashboardService := new(mocks.MockDashboardServiceInterface)
	mockFavoriteService := new(mocks.MockFavoriteServiceInterface)

	handler := NewDashboardHandler(
		mockDashboardService,
		mockFavoriteService,
	)

	return handler, mockDashboardService, mockFavoriteService
}

// ==================== Get Tests ====================

func TestDashboardHandler_Get_Success(t *testing.T) {
	handler, mockDashboardService, mockFavoriteService := setupDashboardTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/dashboard", "")
	user := createTestUser()
	c.Set("current_user", user)

	card := createTestCard()
	voucher := createTestVoucher()
	giftCard := createTestGiftCard()

	dashboardData := &services.DashboardData{
		Stats: &repository.DashboardStats{
			CardsOwned:      1,
			CardsShared:     0,
			VouchersOwned:   1,
			VouchersShared:  0,
			GiftCardsOwned:  1,
			GiftCardsShared: 0,
			TotalBalance:    75.50,
		},
		RecentCards:          []models.Card{*card},
		RecentVouchers:       []models.Voucher{*voucher},
		RecentGiftCards:      []models.GiftCard{*giftCard},
		HasFavorites:         true,
		HasCardFavorites:     true,
		HasVoucherFavorites:  false,
		HasGiftCardFavorites: true,
	}

	mockDashboardService.On("GetDashboardData", mock.Anything, user.ID).Return(dashboardData, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "card", card.ID).Return(false, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "voucher", voucher.ID).Return(false, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "gift_card", giftCard.ID).Return(false, nil)
	mockFavoriteService.On("GetFavoriteCards", mock.Anything, user.ID).Return([]models.Card{*card}, nil)
	mockFavoriteService.On("GetFavoriteGiftCards", mock.Anything, user.ID).Return([]models.GiftCard{*giftCard}, nil)

	err := handler.Get(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response DashboardResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.Stats.CardsCount)
	assert.Equal(t, 1, response.Stats.VouchersCount)
	assert.Equal(t, 1, response.Stats.GiftCardsCount)
	assert.Equal(t, 75.50, response.Stats.TotalBalance)
	assert.True(t, response.HasFavorites)
	assert.True(t, response.HasCardFavorites)
	assert.False(t, response.HasVoucherFavorites)
	assert.True(t, response.HasGiftCardFavorites)
	assert.Len(t, response.RecentCards, 1)
	assert.Len(t, response.RecentVouchers, 1)
	assert.Len(t, response.RecentGiftCards, 1)

	mockDashboardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestDashboardHandler_Get_ServiceError(t *testing.T) {
	handler, mockDashboardService, mockFavoriteService := setupDashboardTest()
	_ = mockFavoriteService // Unused but needed for test setup
	c, rec := createTestContext(http.MethodGet, "/api/v1/dashboard", "")
	user := createTestUser()
	c.Set("current_user", user)

	mockDashboardService.On("GetDashboardData", mock.Anything, user.ID).Return((*services.DashboardData)(nil), errors.New("database error"))

	err := handler.Get(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "server_error", response.Error)
	assert.Equal(t, "Failed to load dashboard data", response.Message)

	mockDashboardService.AssertExpectations(t)
}

func TestDashboardHandler_Get_EmptyDashboard(t *testing.T) {
	handler, mockDashboardService, mockFavoriteService := setupDashboardTest()
	_ = mockFavoriteService // Unused but needed for test setup
	c, rec := createTestContext(http.MethodGet, "/api/v1/dashboard", "")
	user := createTestUser()
	c.Set("current_user", user)

	dashboardData := &services.DashboardData{
		Stats: &repository.DashboardStats{
			CardsOwned:      0,
			CardsShared:     0,
			VouchersOwned:   0,
			VouchersShared:  0,
			GiftCardsOwned:  0,
			GiftCardsShared: 0,
			TotalBalance:    0,
		},
		RecentCards:          []models.Card{},
		RecentVouchers:       []models.Voucher{},
		RecentGiftCards:      []models.GiftCard{},
		HasFavorites:         false,
		HasCardFavorites:     false,
		HasVoucherFavorites:  false,
		HasGiftCardFavorites: false,
	}

	mockDashboardService.On("GetDashboardData", mock.Anything, user.ID).Return(dashboardData, nil)

	err := handler.Get(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response DashboardResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 0, response.Stats.CardsCount)
	assert.Equal(t, 0, response.Stats.VouchersCount)
	assert.Equal(t, 0, response.Stats.GiftCardsCount)
	assert.Equal(t, 0.0, response.Stats.TotalBalance)
	assert.False(t, response.HasFavorites)
	assert.Empty(t, response.RecentCards)
	assert.Empty(t, response.RecentVouchers)
	assert.Empty(t, response.RecentGiftCards)

	mockDashboardService.AssertExpectations(t)
}

func TestDashboardHandler_Get_OnlySharedItems(t *testing.T) {
	handler, mockDashboardService, mockFavoriteService := setupDashboardTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/dashboard", "")
	user := createTestUser()
	c.Set("current_user", user)

	card := createTestCard()
	voucher := createTestVoucher()
	giftCard := createTestGiftCard()

	dashboardData := &services.DashboardData{
		Stats: &repository.DashboardStats{
			CardsOwned:      0,
			CardsShared:     1,
			VouchersOwned:   0,
			VouchersShared:  1,
			GiftCardsOwned:  0,
			GiftCardsShared: 1,
			TotalBalance:    0,
		},
		RecentCards:          []models.Card{*card},
		RecentVouchers:       []models.Voucher{*voucher},
		RecentGiftCards:      []models.GiftCard{*giftCard},
		HasFavorites:         false,
		HasCardFavorites:     false,
		HasVoucherFavorites:  false,
		HasGiftCardFavorites: false,
	}

	mockDashboardService.On("GetDashboardData", mock.Anything, user.ID).Return(dashboardData, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "card", card.ID).Return(false, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "voucher", voucher.ID).Return(false, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "gift_card", giftCard.ID).Return(false, nil)

	err := handler.Get(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response DashboardResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, 1, response.Stats.CardsCount)
	assert.Equal(t, 1, response.Stats.VouchersCount)
	assert.Equal(t, 1, response.Stats.GiftCardsCount)
	assert.Equal(t, 3, response.Stats.SharedCount)

	mockDashboardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}

func TestDashboardHandler_Get_FavoriteServiceError(t *testing.T) {
	handler, mockDashboardService, mockFavoriteService := setupDashboardTest()
	c, rec := createTestContext(http.MethodGet, "/api/v1/dashboard", "")
	user := createTestUser()
	c.Set("current_user", user)

	card := createTestCard()

	dashboardData := &services.DashboardData{
		Stats: &repository.DashboardStats{
			CardsOwned:  1,
			CardsShared: 0,
		},
		RecentCards:      []models.Card{*card},
		HasCardFavorites: true,
	}

	mockDashboardService.On("GetDashboardData", mock.Anything, user.ID).Return(dashboardData, nil)
	mockFavoriteService.On("IsFavorite", mock.Anything, user.ID, "card", card.ID).Return(false, nil)
	mockFavoriteService.On("GetFavoriteCards", mock.Anything, user.ID).Return([]models.Card(nil), errors.New("database error"))

	err := handler.Get(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response DashboardResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	// Should still return successfully, just without favorite counts
	assert.Equal(t, 0, response.Stats.FavoriteCounts["card"])

	mockDashboardService.AssertExpectations(t)
	mockFavoriteService.AssertExpectations(t)
}
