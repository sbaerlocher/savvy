package services

import (
	"context"
	"errors"
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

func (m *MockGiftCardRepository) GetAllForUserPaginated(ctx context.Context, userID uuid.UUID, params repository.PaginationParams) (*repository.PaginatedResult[models.GiftCard], error) {
	args := m.Called(ctx, userID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PaginatedResult[models.GiftCard]), args.Error(1)
}

func (m *MockGiftCardRepository) GetExpiringGiftCards(_ context.Context, _ int) ([]models.GiftCard, error) {
	return nil, nil
}

func (m *MockGiftCardRepository) FindByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.GiftCard, error) {
	args := m.Called(ctx, cardNumber, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCard), args.Error(1)
}

func (m *MockGiftCardRepository) Search(ctx context.Context, userID uuid.UUID, query string) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
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

// ============================================================================
// ADDITIONAL TESTS FOR TRANSACTION COVERAGE
// ============================================================================

func TestGiftCardService_CreateTransaction_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	transaction := &models.GiftCardTransaction{
		GiftCardID:  uuid.New(),
		Amount:      50.00,
		Description: "Test purchase",
	}

	mockRepo.On("CreateTransaction", ctx, transaction).Return(nil)

	err := service.CreateTransaction(ctx, transaction)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_CreateTransaction_InvalidAmount(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	transaction := &models.GiftCardTransaction{
		GiftCardID:  uuid.New(),
		Amount:      -10.00, // Invalid negative amount
		Description: "Test",
	}

	err := service.CreateTransaction(ctx, transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")
	mockRepo.AssertNotCalled(t, "CreateTransaction")
}

func TestGiftCardService_CreateTransaction_ZeroAmount(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	transaction := &models.GiftCardTransaction{
		GiftCardID:  uuid.New(),
		Amount:      0, // Invalid zero amount
		Description: "Test",
	}

	err := service.CreateTransaction(ctx, transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")
	mockRepo.AssertNotCalled(t, "CreateTransaction")
}

func TestGiftCardService_GetTransaction_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	transactionID := uuid.New()
	giftCardID := uuid.New()
	expectedTransaction := &models.GiftCardTransaction{
		ID:          transactionID,
		GiftCardID:  giftCardID,
		Amount:      25.50,
		Description: "Test transaction",
	}

	mockRepo.On("GetTransaction", ctx, transactionID, giftCardID).Return(expectedTransaction, nil)

	transaction, err := service.GetTransaction(ctx, transactionID, giftCardID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransaction, transaction)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_GetTransaction_NotFound(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	transactionID := uuid.New()
	giftCardID := uuid.New()

	mockRepo.On("GetTransaction", ctx, transactionID, giftCardID).Return(nil, gorm.ErrRecordNotFound)

	transaction, err := service.GetTransaction(ctx, transactionID, giftCardID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, transaction)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_DeleteTransaction_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	transactionID := uuid.New()

	mockRepo.On("DeleteTransaction", ctx, transactionID).Return(nil)

	err := service.DeleteTransaction(ctx, transactionID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_DeleteTransaction_Error(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	transactionID := uuid.New()
	deleteError := errors.New("delete failed")

	mockRepo.On("DeleteTransaction", ctx, transactionID).Return(deleteError)

	err := service.DeleteTransaction(ctx, transactionID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, deleteError)
	mockRepo.AssertExpectations(t)
}

// Tests for GetCurrentBalance error path
func TestGiftCardService_GetCurrentBalance_NotFound(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()

	mockRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User", "Transactions"}).Return(nil, gorm.ErrRecordNotFound)

	balance, err := service.GetCurrentBalance(ctx, giftCardID)

	assert.Error(t, err)
	assert.Equal(t, 0.0, balance)
	mockRepo.AssertExpectations(t)
}

// Tests for CanUserAccessGiftCard error paths
func TestGiftCardService_CanUserAccessGiftCard_NotFound(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()
	userID := uuid.New()

	mockRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User", "Transactions"}).Return(nil, gorm.ErrRecordNotFound)

	canAccess, err := service.CanUserAccessGiftCard(ctx, giftCardID, userID)

	assert.NoError(t, err)
	assert.False(t, canAccess)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_CanUserAccessGiftCard_NotOwner(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCardID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	giftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &otherUserID, // Different owner
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	mockRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User", "Transactions"}).Return(giftCard, nil)

	canAccess, err := service.CanUserAccessGiftCard(ctx, giftCardID, userID)

	assert.NoError(t, err)
	assert.False(t, canAccess) // Not owner and no share
	mockRepo.AssertExpectations(t)
}

// Tests for GetUserGiftCards error paths
func TestGiftCardService_GetUserGiftCards_OwnedError(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, repoError)

	giftCards, err := service.GetUserGiftCards(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, giftCards)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_GetUserGiftCards_SharedError(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	ownedGiftCards := []models.GiftCard{
		{ID: uuid.New(), CardNumber: "1111"},
	}
	repoError := errors.New("shared query failed")

	mockRepo.On("GetByUserID", ctx, userID).Return(ownedGiftCards, nil)
	mockRepo.On("GetSharedWithUser", ctx, userID).Return(nil, repoError)

	giftCards, err := service.GetUserGiftCards(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, giftCards)
	mockRepo.AssertExpectations(t)
}

// Tests for UpdateGiftCard validation and error paths
func TestGiftCardService_UpdateGiftCard_ValidationMissingMerchant(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCard := &models.GiftCard{
		ID:             uuid.New(),
		CardNumber:     "1234",
		MerchantName:   "", // Missing merchant
		InitialBalance: 100.0,
	}

	err := service.UpdateGiftCard(ctx, giftCard)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestGiftCardService_UpdateGiftCard_ValidationMissingCardNumber(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCard := &models.GiftCard{
		ID:             uuid.New(),
		CardNumber:     "", // Missing card number
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	err := service.UpdateGiftCard(ctx, giftCard)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card number is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestGiftCardService_UpdateGiftCard_ValidationNegativeBalance(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	giftCard := &models.GiftCard{
		ID:             uuid.New(),
		CardNumber:     "1234",
		MerchantName:   "Test Store",
		InitialBalance: -50.0, // Negative balance
	}

	err := service.UpdateGiftCard(ctx, giftCard)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "initial balance must be positive")
	mockRepo.AssertNotCalled(t, "Update")
}

// ============================================================================
// CHECK DUPLICATE TESTS
// ============================================================================

func TestGiftCardService_CheckDuplicate_DuplicateFound(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	existingGiftCard := &models.GiftCard{
		ID:             uuid.New(),
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	mockRepo.On("FindByCardNumber", ctx, "1234567890", userID).Return(existingGiftCard, nil)

	result, err := service.CheckDuplicate(ctx, "1234567890", userID, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingGiftCard.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_CheckDuplicate_NoDuplicate(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("FindByCardNumber", ctx, "9999999999", userID).Return(nil, nil)

	result, err := service.CheckDuplicate(ctx, "9999999999", userID, nil)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_CheckDuplicate_ExcludeSameGiftCard(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	giftCardID := uuid.New()
	existingGiftCard := &models.GiftCard{
		ID:             giftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	mockRepo.On("FindByCardNumber", ctx, "1234567890", userID).Return(existingGiftCard, nil)

	// Exclude the same gift card (update case) - should not be considered a duplicate
	result, err := service.CheckDuplicate(ctx, "1234567890", userID, &giftCardID)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_CheckDuplicate_ExcludeDifferentGiftCard(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	existingGiftCardID := uuid.New()
	differentGiftCardID := uuid.New()
	existingGiftCard := &models.GiftCard{
		ID:             existingGiftCardID,
		UserID:         &userID,
		CardNumber:     "1234567890",
		MerchantName:   "Test Store",
		InitialBalance: 100.0,
	}

	mockRepo.On("FindByCardNumber", ctx, "1234567890", userID).Return(existingGiftCard, nil)

	// Exclude a different gift card - should still be considered a duplicate
	result, err := service.CheckDuplicate(ctx, "1234567890", userID, &differentGiftCardID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingGiftCardID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_CheckDuplicate_RepositoryError(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	mockRepo.On("FindByCardNumber", ctx, "1234567890", userID).Return(nil, repoError)

	result, err := service.CheckDuplicate(ctx, "1234567890", userID, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "check duplicate")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// GET USER GIFT CARDS PAGINATED TESTS
// ============================================================================

func TestGiftCardService_GetUserGiftCardsPaginated_Success(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedResult := &repository.PaginatedResult[models.GiftCard]{
		Items: []models.GiftCard{
			{ID: uuid.New(), CardNumber: "1111", MerchantName: "Store 1", InitialBalance: 50.0},
			{ID: uuid.New(), CardNumber: "2222", MerchantName: "Store 2", InitialBalance: 100.0},
		},
		Total:      6,
		Page:       1,
		PerPage:    2,
		TotalPages: 3,
	}

	params := repository.PaginationParams{Page: 1, PerPage: 2}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(expectedResult, nil)

	result, err := service.GetUserGiftCardsPaginated(ctx, userID, 1, 2)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(6), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.PerPage)
	assert.Equal(t, 3, result.TotalPages)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_GetUserGiftCardsPaginated_EmptyResult(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedResult := &repository.PaginatedResult[models.GiftCard]{
		Items:      []models.GiftCard{},
		Total:      0,
		Page:       1,
		PerPage:    10,
		TotalPages: 0,
	}

	params := repository.PaginationParams{Page: 1, PerPage: 10}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(expectedResult, nil)

	result, err := service.GetUserGiftCardsPaginated(ctx, userID, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
	mockRepo.AssertExpectations(t)
}

func TestGiftCardService_GetUserGiftCardsPaginated_RepositoryError(t *testing.T) {
	mockRepo := new(MockGiftCardRepository)
	service := NewGiftCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	params := repository.PaginationParams{Page: 1, PerPage: 10}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(nil, repoError)

	result, err := service.GetUserGiftCardsPaginated(ctx, userID, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get paginated gift cards")
	mockRepo.AssertExpectations(t)
}
