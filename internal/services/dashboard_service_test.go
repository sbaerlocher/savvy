package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"savvy/internal/models"
	"savvy/internal/repository"
)

func TestDashboardService_GetDashboardData_EmptyUser(t *testing.T) {
	db := setupDirectTestDB(t)
	service := NewDashboardService(repository.NewDashboardRepository(db))
	ctx := context.Background()

	// Create a user with no data
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-test@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// Get dashboard data
	data, err := service.GetDashboardData(ctx, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotNil(t, data.Stats)
	assert.Equal(t, int64(0), data.Stats.CardsOwned)
	assert.Equal(t, int64(0), data.Stats.VouchersOwned)
	assert.Equal(t, int64(0), data.Stats.GiftCardsOwned)
	assert.Equal(t, 0.0, data.Stats.TotalBalance)
	assert.False(t, data.HasFavorites)
	assert.Empty(t, data.RecentCards)
	assert.Empty(t, data.RecentVouchers)
	assert.Empty(t, data.RecentGiftCards)
}

func TestDashboardService_GetDashboardData_WithOwnedItems(t *testing.T) {
	db := setupDirectTestDB(t)
	service := NewDashboardService(repository.NewDashboardRepository(db))
	ctx := context.Background()

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-owned@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// Create owned items
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "DASH-CARD-1",
		MerchantName: "Test Merchant",
	}
	db.Create(card)

	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "DASH-VOUCHER-1",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)

	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "DASH-GIFT-1",
		MerchantName:   "Test",
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Currency:       "CHF",
	}
	db.Create(giftCard)

	// Get dashboard data
	data, err := service.GetDashboardData(ctx, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotNil(t, data.Stats)
	assert.Equal(t, int64(1), data.Stats.CardsOwned)
	assert.Equal(t, int64(1), data.Stats.VouchersOwned)
	assert.Equal(t, int64(1), data.Stats.GiftCardsOwned)
	assert.GreaterOrEqual(t, data.Stats.TotalBalance, 75.0) // At least our test card
	assert.False(t, data.HasFavorites)                      // No favorites yet
	assert.GreaterOrEqual(t, len(data.RecentCards), 1)
	assert.GreaterOrEqual(t, len(data.RecentVouchers), 1)
	assert.GreaterOrEqual(t, len(data.RecentGiftCards), 1)
}

func TestDashboardService_GetDashboardData_WithSharedItems(t *testing.T) {
	db := setupDirectTestDB(t)
	service := NewDashboardService(repository.NewDashboardRepository(db))
	ctx := context.Background()

	// Create owner
	ownerID := uuid.New()
	owner := &models.User{
		ID:           ownerID,
		Email:        "dashboard-owner@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "Owner",
	}
	db.Create(owner)

	// Create shared user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-shared@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// Create owned items for owner
	card := &models.Card{
		UserID:       &ownerID,
		CardNumber:   "DASH-SHARED-CARD",
		MerchantName: "Test",
	}
	db.Create(card)

	// Share with user
	share := &models.CardShare{
		CardID:       card.ID,
		SharedWithID: userID,
		CanEdit:      true,
		CanDelete:    false,
	}
	db.Create(share)

	// Get dashboard data
	data, err := service.GetDashboardData(ctx, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotNil(t, data.Stats)
	assert.Equal(t, int64(0), data.Stats.CardsOwned)           // No owned cards
	assert.GreaterOrEqual(t, data.Stats.CardsShared, int64(1)) // At least 1 shared
}

func TestDashboardService_GetDashboardData_WithFavorites(t *testing.T) {
	db := setupDirectTestDB(t)
	service := NewDashboardService(repository.NewDashboardRepository(db))
	ctx := context.Background()

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-fav@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// Create a card
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "DASH-FAV-CARD",
		MerchantName: "Test",
	}
	db.Create(card)

	// Mark as favorite
	favorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   card.ID,
	}
	db.Create(favorite)

	// Get dashboard data
	data, err := service.GetDashboardData(ctx, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.True(t, data.HasFavorites)
	assert.True(t, data.HasCardFavorites)
	assert.False(t, data.HasVoucherFavorites)
	assert.False(t, data.HasGiftCardFavorites)
}

func TestDashboardService_GetDashboardData_FavoritesExcludeRecentOfOtherTypes(t *testing.T) {
	db := setupDirectTestDB(t)
	service := NewDashboardService(repository.NewDashboardRepository(db))
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-fav-mixed@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// A card and a gift card that are NOT favorited
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "DASH-NOFAV-CARD",
		MerchantName: "Test",
	}
	db.Create(card)
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "DASH-NOFAV-GC",
		MerchantName:   "Test",
		InitialBalance: 50,
		CurrentBalance: 50,
		Currency:       "CHF",
	}
	db.Create(giftCard)

	// A voucher that IS favorited
	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "DASH-FAV-VOUCHER",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "single_use",
	}
	db.Create(voucher)
	db.Create(&models.UserFavorite{
		UserID:       userID,
		ResourceType: "voucher",
		ResourceID:   voucher.ID,
	})

	data, err := service.GetDashboardData(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.True(t, data.HasFavorites)
	// Favorites exist, so the section must contain ONLY favorites: the
	// non-favorited card must not fall back to "recent cards".
	assert.Empty(t, data.RecentCards)
	assert.Empty(t, data.RecentGiftCards)
	assert.Len(t, data.RecentVouchers, 1)
}

func TestDashboardService_GetDashboardData_MixedScenario(t *testing.T) {
	db := setupDirectTestDB(t)
	service := NewDashboardService(repository.NewDashboardRepository(db))
	ctx := context.Background()

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-mixed@example.com",
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	db.Create(user)

	// Create multiple cards
	for i := 0; i < 3; i++ {
		card := &models.Card{
			UserID:       &userID,
			CardNumber:   "DASH-MIX-" + string(rune('A'+i)),
			MerchantName: "Test",
		}
		db.Create(card)
	}

	// Create gift cards with balance
	giftCard1 := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "DASH-MIX-GC1",
		MerchantName:   "Test",
		InitialBalance: 100,
		CurrentBalance: 80,
		Currency:       "CHF",
	}
	db.Create(giftCard1)

	giftCard2 := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "DASH-MIX-GC2",
		MerchantName:   "Test",
		InitialBalance: 50,
		CurrentBalance: 20,
		Currency:       "CHF",
	}
	db.Create(giftCard2)

	// Get dashboard data
	data, err := service.GetDashboardData(ctx, userID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.NotNil(t, data.Stats)
	assert.Equal(t, int64(3), data.Stats.CardsOwned)
	assert.Equal(t, int64(2), data.Stats.GiftCardsOwned)
	assert.GreaterOrEqual(t, data.Stats.TotalBalance, 100.0) // 80 + 20 = 100
	assert.GreaterOrEqual(t, len(data.RecentCards), 1)
	assert.GreaterOrEqual(t, len(data.RecentGiftCards), 1)
}
