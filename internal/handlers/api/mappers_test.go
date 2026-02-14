package api

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"savvy/internal/models"
	"savvy/internal/services"
)

// ==================== User Mapper Tests ====================

func TestToUserDTO(t *testing.T) {
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "admin",
	}

	dto := ToUserDTO(user)

	assert.Equal(t, user.ID.String(), dto.ID)
	assert.Equal(t, "test@example.com", dto.Email)
	assert.Equal(t, "John", dto.FirstName)
	assert.Equal(t, "Doe", dto.LastName)
	assert.True(t, dto.IsAdmin)
}

func TestToUserDTO_RegularUser(t *testing.T) {
	user := &models.User{
		ID:   uuid.New(),
		Role: "user",
	}

	dto := ToUserDTO(user)

	assert.False(t, dto.IsAdmin)
}

// ==================== Merchant Mapper Tests ====================

func TestToMerchantDTO(t *testing.T) {
	merchant := &models.Merchant{
		ID:      uuid.New(),
		Name:    "IKEA",
		Color:   "#0051BA",
		LogoURL: "https://example.com/logo.png",
		Website: "https://ikea.com",
	}

	dto := ToMerchantDTO(merchant)

	assert.Equal(t, merchant.ID.String(), dto.ID)
	assert.Equal(t, "IKEA", dto.Name)
	assert.NotNil(t, dto.Color)
	assert.Equal(t, "#0051BA", *dto.Color)
	assert.NotNil(t, dto.LogoURL)
	assert.Equal(t, "https://example.com/logo.png", *dto.LogoURL)
	assert.NotNil(t, dto.Website)
	assert.Equal(t, "https://ikea.com", *dto.Website)
}

func TestToMerchantDTO_EmptyOptionalFields(t *testing.T) {
	merchant := &models.Merchant{
		ID:   uuid.New(),
		Name: "Simple",
	}

	dto := ToMerchantDTO(merchant)

	assert.Equal(t, "Simple", dto.Name)
	assert.Nil(t, dto.Color)
	assert.Nil(t, dto.LogoURL)
	assert.Nil(t, dto.Website)
}

func TestToMerchantDTOs(t *testing.T) {
	merchants := []models.Merchant{
		{ID: uuid.New(), Name: "IKEA"},
		{ID: uuid.New(), Name: "Amazon"},
	}

	dtos := ToMerchantDTOs(merchants)

	assert.Len(t, dtos, 2)
	assert.Equal(t, "IKEA", dtos[0].Name)
	assert.Equal(t, "Amazon", dtos[1].Name)
}

func TestToMerchantDTOs_Empty(t *testing.T) {
	dtos := ToMerchantDTOs([]models.Merchant{})
	assert.Len(t, dtos, 0)
}

// ==================== Permission Mapper Tests ====================

func TestToPermissionDTO(t *testing.T) {
	perms := &services.ResourcePermissions{
		CanView:             true,
		CanEdit:             true,
		CanDelete:           false,
		CanEditTransactions: true,
		IsOwner:             true,
	}

	dto := ToPermissionDTO(perms)

	assert.True(t, dto.CanView)
	assert.True(t, dto.CanEdit)
	assert.False(t, dto.CanDelete)
	assert.True(t, dto.CanEditTransactions)
	assert.True(t, dto.IsOwner)
}

// ==================== Card Mapper Tests ====================

func TestToCardDTO(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()
	card := &models.Card{
		ID:           uuid.New(),
		UserID:       &userID,
		MerchantID:   &merchantID,
		CardNumber:   "1234567890",
		BarcodeType:  "CODE128",
		Program:      "Loyalty",
		Notes:        "Test notes",
		Status:       "active",
		MerchantName: "IKEA",
	}

	dto := ToCardDTO(card, true)

	assert.Equal(t, card.ID.String(), dto.ID)
	assert.NotNil(t, dto.MerchantID)
	assert.Equal(t, "1234567890", dto.CardNumber)
	assert.NotNil(t, dto.BarcodeType)
	assert.Equal(t, "CODE128", *dto.BarcodeType)
	assert.NotNil(t, dto.Program)
	assert.Equal(t, "Loyalty", *dto.Program)
	assert.NotNil(t, dto.Notes)
	assert.Equal(t, "Test notes", *dto.Notes)
	assert.True(t, dto.IsFavorite)
	assert.False(t, dto.IsShared) // Has owner
}

func TestToCardDTO_SharedCard(t *testing.T) {
	card := &models.Card{
		ID:         uuid.New(),
		UserID:     nil, // No owner = shared
		CardNumber: "123",
	}

	dto := ToCardDTO(card, false)

	assert.True(t, dto.IsShared)
	assert.False(t, dto.IsFavorite)
	assert.Nil(t, dto.MerchantID)
	assert.Nil(t, dto.Program)
	assert.Nil(t, dto.BarcodeType)
	assert.Nil(t, dto.Notes)
}

func TestToCardDTO_WithMerchantAndOwner(t *testing.T) {
	userID := uuid.New()
	merchant := &models.Merchant{ID: uuid.New(), Name: "IKEA", Color: "#0051BA"}
	user := &models.User{ID: userID, Email: "test@example.com", Role: "user"}
	card := &models.Card{
		ID:         uuid.New(),
		UserID:     &userID,
		Merchant:   merchant,
		User:       user,
		CardNumber: "123",
	}

	dto := ToCardDTO(card, false)

	assert.NotNil(t, dto.Merchant)
	assert.Equal(t, "IKEA", dto.Merchant.Name)
	assert.NotNil(t, dto.Owner)
	assert.Equal(t, "test@example.com", dto.Owner.Email)
}

func TestToCardDTOs(t *testing.T) {
	userID := uuid.New()
	cards := []models.Card{
		{ID: uuid.New(), UserID: &userID, CardNumber: "111"},
		{ID: uuid.New(), UserID: &userID, CardNumber: "222"},
	}
	favIDs := map[string]bool{cards[0].ID.String(): true}

	dtos := ToCardDTOs(cards, favIDs)

	assert.Len(t, dtos, 2)
	assert.True(t, dtos[0].IsFavorite)
	assert.False(t, dtos[1].IsFavorite)
}

// ==================== Voucher Mapper Tests ====================

func TestToVoucherDTO_ActiveVoucher(t *testing.T) {
	userID := uuid.New()
	voucher := &models.Voucher{
		ID:             uuid.New(),
		UserID:         &userID,
		Code:           "SAVE10",
		Type:           "percentage",
		Value:          10.0,
		Currency:       "EUR",
		Description:    "10% off",
		ValidFrom:      time.Now().Add(-24 * time.Hour),
		ValidUntil:     time.Now().Add(30 * 24 * time.Hour),
		UsageLimitType: "single_use",
		BarcodeType:    "QR_CODE",
	}

	dto := ToVoucherDTO(voucher, false)

	assert.Equal(t, "valid", dto.Status)
	assert.Equal(t, "SAVE10", dto.Code)
	assert.NotNil(t, dto.Description)
	assert.Equal(t, "10% off", *dto.Description)
}

func TestToVoucherDTO_ExpiredVoucher(t *testing.T) {
	userID := uuid.New()
	voucher := &models.Voucher{
		ID:         uuid.New(),
		UserID:     &userID,
		Code:       "EXPIRED",
		ValidFrom:  time.Now().Add(-60 * 24 * time.Hour),
		ValidUntil: time.Now().Add(-1 * 24 * time.Hour),
	}

	dto := ToVoucherDTO(voucher, false)

	assert.Equal(t, "expired", dto.Status)
}

func TestToVoucherDTO_InactiveVoucher(t *testing.T) {
	userID := uuid.New()
	voucher := &models.Voucher{
		ID:         uuid.New(),
		UserID:     &userID,
		Code:       "FUTURE",
		ValidFrom:  time.Now().Add(30 * 24 * time.Hour),
		ValidUntil: time.Now().Add(60 * 24 * time.Hour),
	}

	dto := ToVoucherDTO(voucher, false)

	assert.Equal(t, "inactive", dto.Status)
}

// ==================== Gift Card Mapper Tests ====================

func TestToGiftCardDTO(t *testing.T) {
	userID := uuid.New()
	expiresAt := time.Now().Add(365 * 24 * time.Hour)
	giftCard := &models.GiftCard{
		ID:             uuid.New(),
		UserID:         &userID,
		CardNumber:     "GC-123",
		InitialBalance: 100.00,
		CurrentBalance: 75.50,
		Currency:       "EUR",
		PIN:            "1234",
		ExpiresAt:      &expiresAt,
		Notes:          "Birthday gift",
	}

	dto := ToGiftCardDTO(giftCard, true)

	assert.Equal(t, "GC-123", dto.CardNumber)
	assert.Equal(t, 100.00, dto.InitialBalance)
	assert.Equal(t, 75.50, dto.CurrentBalance)
	assert.Equal(t, "EUR", dto.Currency)
	assert.NotNil(t, dto.PIN)
	assert.Equal(t, "1234", *dto.PIN)
	assert.NotNil(t, dto.ExpiresAt)
	assert.NotNil(t, dto.Notes)
	assert.Equal(t, "Birthday gift", *dto.Notes)
	assert.True(t, dto.IsFavorite)
}

func TestToGiftCardDTO_MinimalFields(t *testing.T) {
	giftCard := &models.GiftCard{
		ID:         uuid.New(),
		CardNumber: "GC-999",
		Currency:   "CHF",
	}

	dto := ToGiftCardDTO(giftCard, false)

	assert.Nil(t, dto.PIN)
	assert.Nil(t, dto.ExpiresAt)
	assert.Nil(t, dto.Notes)
	assert.True(t, dto.IsShared) // No UserID
}

// ==================== Transaction Mapper Tests ====================

func TestToTransactionDTO(t *testing.T) {
	tx := &models.GiftCardTransaction{
		ID:              uuid.New(),
		GiftCardID:      uuid.New(),
		Amount:          -25.50,
		Description:     "Coffee purchase",
		TransactionDate: time.Now(),
	}

	dto := ToTransactionDTO(tx)

	assert.Equal(t, tx.ID.String(), dto.ID)
	assert.Equal(t, tx.GiftCardID.String(), dto.GiftCardID)
	assert.Equal(t, -25.50, dto.Amount)
	assert.NotNil(t, dto.Description)
	assert.Equal(t, "Coffee purchase", *dto.Description)
}

func TestToTransactionDTO_NoDescription(t *testing.T) {
	tx := &models.GiftCardTransaction{
		ID:         uuid.New(),
		GiftCardID: uuid.New(),
		Amount:     -10.00,
	}

	dto := ToTransactionDTO(tx)

	assert.Nil(t, dto.Description)
}

func TestToTransactionDTOs(t *testing.T) {
	txs := []models.GiftCardTransaction{
		{ID: uuid.New(), GiftCardID: uuid.New(), Amount: -10},
		{ID: uuid.New(), GiftCardID: uuid.New(), Amount: -20},
	}

	dtos := ToTransactionDTOs(txs)

	assert.Len(t, dtos, 2)
}

// ==================== Share Mapper Tests ====================

func TestToCardShareDTO(t *testing.T) {
	share := &models.CardShare{
		ID:             uuid.New(),
		SharedWithUser: &models.User{ID: uuid.New(), Email: "shared@example.com", Role: "user"},
		CanEdit:        true,
		CanDelete:      false,
	}

	dto := ToCardShareDTO(share)

	assert.Equal(t, share.ID.String(), dto.ID)
	assert.Equal(t, "shared@example.com", dto.SharedWithUser.Email)
	assert.True(t, dto.CanEdit)
	assert.False(t, dto.CanDelete)
}

func TestToVoucherShareDTO(t *testing.T) {
	share := &models.VoucherShare{
		ID:             uuid.New(),
		SharedWithUser: &models.User{ID: uuid.New(), Email: "shared@example.com", Role: "user"},
	}

	dto := ToVoucherShareDTO(share)

	assert.False(t, dto.CanEdit)   // Always read-only
	assert.False(t, dto.CanDelete) // Always read-only
}

func TestToGiftCardShareDTO(t *testing.T) {
	share := &models.GiftCardShare{
		ID:                  uuid.New(),
		SharedWithUser:      &models.User{ID: uuid.New(), Email: "shared@example.com", Role: "user"},
		CanEdit:             true,
		CanDelete:           true,
		CanEditTransactions: true,
	}

	dto := ToGiftCardShareDTO(share)

	assert.True(t, dto.CanEdit)
	assert.True(t, dto.CanDelete)
	assert.True(t, dto.CanEditTransactions)
}

// ==================== Admin Mapper Tests ====================

func TestToAdminUserDTO(t *testing.T) {
	user := &models.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		FirstName:    "Admin",
		LastName:     "User",
		Role:         "admin",
		AuthProvider: "local",
	}

	dto := ToAdminUserDTO(user)

	assert.Equal(t, "admin@example.com", dto.Email)
	assert.Equal(t, "admin", dto.Role)
	assert.Equal(t, "local", dto.AuthProvider)
}

func TestToAdminUserDTOs(t *testing.T) {
	users := []models.User{
		{ID: uuid.New(), Email: "user1@example.com", Role: "user"},
		{ID: uuid.New(), Email: "user2@example.com", Role: "admin"},
	}

	dtos := ToAdminUserDTOs(users)

	assert.Len(t, dtos, 2)
}

// ==================== Audit Log Mapper Tests ====================

func TestToAuditLogDTO(t *testing.T) {
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "admin@example.com", Role: "admin"}
	log := &models.AuditLog{
		ID:           uuid.New(),
		UserID:       &userID,
		User:         user,
		Action:       "delete",
		ResourceType: "cards",
		ResourceID:   uuid.New(),
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
	}

	dto := ToAuditLogDTO(log)

	assert.NotNil(t, dto.UserID)
	assert.Equal(t, userID.String(), *dto.UserID)
	assert.NotNil(t, dto.User)
	assert.Equal(t, "admin@example.com", dto.User.Email)
	assert.Equal(t, "delete", dto.Action)
	assert.Equal(t, "cards", dto.ResourceType)
}

func TestToAuditLogDTO_NilUser(t *testing.T) {
	log := &models.AuditLog{
		ID:           uuid.New(),
		UserID:       nil,
		User:         nil,
		Action:       "delete",
		ResourceType: "cards",
		ResourceID:   uuid.New(),
	}

	dto := ToAuditLogDTO(log)

	assert.Nil(t, dto.UserID)
	assert.Nil(t, dto.User)
}

// ==================== Time Formatting Tests ====================

func TestSetAppLocation(t *testing.T) {
	// Save original
	origLoc := appLocation

	loc, _ := time.LoadLocation("Europe/Zurich")
	SetAppLocation(loc)
	assert.Equal(t, loc, appLocation)

	// nil should not change
	SetAppLocation(nil)
	assert.Equal(t, loc, appLocation)

	// Restore
	appLocation = origLoc
}
