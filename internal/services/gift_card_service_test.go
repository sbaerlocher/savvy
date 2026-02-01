package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/repository"
)

type MockGiftCardRepository struct {
	mock.Mock
}

func (m *MockGiftCardRepository) Create(ctx context.Context, giftCard *models.GiftCard) error {
	args := m.Called(ctx, giftCard)
	return args.Error(0)
}

func (m *MockGiftCardRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.GiftCard, error) {
	args := m.Called(ctx, id, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCard), args.Error(1)
}

func (m *MockGiftCardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

func (m *MockGiftCardRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

func (m *MockGiftCardRepository) Update(ctx context.Context, giftCard *models.GiftCard) error {
	args := m.Called(ctx, giftCard)
	return args.Error(0)
}

func (m *MockGiftCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGiftCardRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGiftCardRepository) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockGiftCardRepository) CreateTransaction(ctx context.Context, transaction *models.GiftCardTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockGiftCardRepository) GetTransaction(ctx context.Context, transactionID, giftCardID uuid.UUID) (*models.GiftCardTransaction, error) {
	args := m.Called(ctx, transactionID, giftCardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCardTransaction), args.Error(1)
}

func (m *MockGiftCardRepository) DeleteTransaction(ctx context.Context, transactionID uuid.UUID) error {
	args := m.Called(ctx, transactionID)
	return args.Error(0)
}

var _ repository.GiftCardRepository = (*MockGiftCardRepository)(nil)

func TestGiftCardService_CreateGiftCard_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
		Currency:       "CHF",
	}

	mockRepo.On("Create", ctx, giftCard).Return(nil)

	err := service.CreateGiftCard(ctx, giftCard)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_CreateGiftCard_DefaultCurrency(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	mockRepo.On("Create", ctx, giftCard).Return(nil)

	err := service.CreateGiftCard(ctx, giftCard)

	assert.NoError(t, err)
	assert.Equal(t, "CHF", giftCard.Currency)
}

func TestGiftCardService_CreateGiftCard_MissingMerchantName(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "1234567890",
		InitialBalance: 100.0,
	}

	err := service.CreateGiftCard(ctx, giftCard)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestGiftCardService_CreateGiftCard_MissingCardNumber(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	giftCard := &models.GiftCard{
		UserID:         &userID,
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	err := service.CreateGiftCard(ctx, giftCard)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card number is required")
}

func TestGiftCardService_CreateGiftCard_InvalidBalance(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	giftCard := &models.GiftCard{
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 0,
	}

	err := service.CreateGiftCard(ctx, giftCard)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "initial balance must be positive")
}

func TestGiftCardService_GetGiftCard_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()
	userID := uuid.New()
	expectedGiftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
		CurrentBalance: 75.0,
	}

	mockRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User", "Transactions"}).Return(expectedGiftCard, nil)

	giftCard, err := service.GetGiftCard(ctx, giftCardID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGiftCard, giftCard)
}

func TestGiftCardService_GetGiftCard_NotFound(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()

	mockRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User", "Transactions"}).Return(nil, gorm.ErrRecordNotFound)

	giftCard, err := service.GetGiftCard(ctx, giftCardID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, giftCard)
}

func TestGiftCardService_GetUserGiftCards_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	ownedCards := []models.GiftCard{
		{ID: uuid.New(), CardNumber: "1111", MerchantName: "Store 1"},
		{ID: uuid.New(), CardNumber: "2222", MerchantName: "Store 2"},
	}

	sharedCards := []models.GiftCard{
		{ID: uuid.New(), CardNumber: "3333", MerchantName: "Store 3"},
	}

	mockRepo.On("GetByUserID", ctx, userID).Return(ownedCards, nil)
	mockRepo.On("GetSharedWithUser", ctx, userID).Return(sharedCards, nil)

	cards, err := service.GetUserGiftCards(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, cards, 3)
}

func TestGiftCardService_UpdateGiftCard_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()
	userID := uuid.New()
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "9999",
		MerchantName:   "Updated Store",
		InitialBalance: 200.0,
	}

	mockRepo.On("Update", ctx, giftCard).Return(nil)

	err := service.UpdateGiftCard(ctx, giftCard)

	assert.NoError(t, err)
}

func TestGiftCardService_DeleteGiftCard_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()

	mockRepo.On("Delete", ctx, giftCardID).Return(nil)

	err := service.DeleteGiftCard(ctx, giftCardID)

	assert.NoError(t, err)
}

func TestGiftCardService_CountUserGiftCards_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedCount := int64(5)

	mockRepo.On("Count", ctx, userID).Return(expectedCount, nil)

	count, err := service.CountUserGiftCards(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
}

func TestGiftCardService_GetTotalBalance_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedBalance := 350.75

	mockRepo.On("GetTotalBalance", ctx, userID).Return(expectedBalance, nil)

	balance, err := service.GetTotalBalance(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
}

func TestGiftCardService_GetCurrentBalance_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()
	userID := uuid.New()
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CurrentBalance: 75.50,
	}

	mockRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User", "Transactions"}).Return(giftCard, nil)

	balance, err := service.GetCurrentBalance(ctx, giftCardID)

	assert.NoError(t, err)
	assert.Equal(t, 75.50, balance)
}

func TestGiftCardService_CanUserAccessGiftCard_Owner(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()
	userID := uuid.New()
	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	mockRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User", "Transactions"}).Return(giftCard, nil)

	canAccess, err := service.CanUserAccessGiftCard(ctx, giftCardID, userID)

	assert.NoError(t, err)
	assert.True(t, canAccess)
}
