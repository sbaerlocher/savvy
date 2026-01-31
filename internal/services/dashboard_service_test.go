package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"savvy/internal/models"
)

// setupDashboardTestDB creates a test database for dashboard tests
func setupDashboardTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://savvy:savvy_dev_password@localhost:5432/savvy?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
		return nil
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Merchant{},
		&models.Card{},
		&models.CardShare{},
		&models.Voucher{},
		&models.VoucherShare{},
		&models.GiftCard{},
		&models.GiftCardShare{},
		&models.UserFavorite{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestDashboardService_GetDashboardData_EmptyUser(t *testing.T) {
	db := setupDashboardTestDB(t)
	service := NewDashboardService(db)
	ctx := context.Background()

	// Create a user with no data
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-test@example.com",
		PasswordHash: "hashed",
	}
	db.Create(user)
	defer db.Exec("DELETE FROM users WHERE id = ?", userID)

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
	db := setupDashboardTestDB(t)
	service := NewDashboardService(db)
	ctx := context.Background()

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-owned@example.com",
		PasswordHash: "hashed",
	}
	db.Create(user)
	defer db.Exec("DELETE FROM users WHERE id = ?", userID)

	// Create owned items
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "DASH-CARD-1",
		MerchantName: "Test Merchant",
	}
	db.Create(card)
	defer db.Exec("DELETE FROM cards WHERE id = ?", card.ID)

	voucher := &models.Voucher{
		UserID:         &userID,
		Code:           "DASH-VOUCHER-1",
		MerchantName:   "Test",
		ValidFrom:      time.Now(),
		ValidUntil:     time.Now().Add(24 * time.Hour),
		UsageLimitType: "unlimited",
	}
	db.Create(voucher)
	defer db.Exec("DELETE FROM vouchers WHERE id = ?", voucher.ID)

	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "DASH-GIFT-1",
		MerchantName:   "Test",
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
		Currency:       "CHF",
	}
	db.Create(giftCard)
	defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCard.ID)

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
	db := setupDashboardTestDB(t)
	service := NewDashboardService(db)
	ctx := context.Background()

	// Create owner
	ownerID := uuid.New()
	owner := &models.User{
		ID:           ownerID,
		Email:        "dashboard-owner@example.com",
		PasswordHash: "hashed",
	}
	db.Create(owner)
	defer db.Exec("DELETE FROM users WHERE id = ?", ownerID)

	// Create shared user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-shared@example.com",
		PasswordHash: "hashed",
	}
	db.Create(user)
	defer db.Exec("DELETE FROM users WHERE id = ?", userID)

	// Create owned items for owner
	card := &models.Card{
		UserID:       &ownerID,
		CardNumber:   "DASH-SHARED-CARD",
		MerchantName: "Test",
	}
	db.Create(card)
	defer db.Exec("DELETE FROM cards WHERE id = ?", card.ID)

	// Share with user
	share := &models.CardShare{
		CardID:       card.ID,
		SharedWithID: userID,
		CanEdit:      true,
		CanDelete:    false,
	}
	db.Create(share)
	defer db.Exec("DELETE FROM card_shares WHERE id = ?", share.ID)

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
	db := setupDashboardTestDB(t)
	service := NewDashboardService(db)
	ctx := context.Background()

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-fav@example.com",
		PasswordHash: "hashed",
	}
	db.Create(user)
	defer db.Exec("DELETE FROM users WHERE id = ?", userID)

	// Create a card
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "DASH-FAV-CARD",
		MerchantName: "Test",
	}
	db.Create(card)
	defer db.Exec("DELETE FROM cards WHERE id = ?", card.ID)

	// Mark as favorite
	favorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   card.ID,
	}
	db.Create(favorite)
	defer db.Exec("DELETE FROM user_favorites WHERE id = ?", favorite.ID)

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

func TestDashboardService_GetDashboardData_MixedScenario(t *testing.T) {
	db := setupDashboardTestDB(t)
	service := NewDashboardService(db)
	ctx := context.Background()

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "dashboard-mixed@example.com",
		PasswordHash: "hashed",
	}
	db.Create(user)
	defer db.Exec("DELETE FROM users WHERE id = ?", userID)

	// Create multiple cards
	for i := 0; i < 3; i++ {
		card := &models.Card{
			UserID:       &userID,
			CardNumber:   "DASH-MIX-" + string(rune('A'+i)),
			MerchantName: "Test",
		}
		db.Create(card)
		defer db.Exec("DELETE FROM cards WHERE id = ?", card.ID)
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
	defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCard1.ID)

	giftCard2 := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "DASH-MIX-GC2",
		MerchantName:   "Test",
		InitialBalance: 50,
		CurrentBalance: 20,
		Currency:       "CHF",
	}
	db.Create(giftCard2)
	defer db.Exec("DELETE FROM gift_cards WHERE id = ?", giftCard2.ID)

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
