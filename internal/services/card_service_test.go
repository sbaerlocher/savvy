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

// MockCardRepository is a manual mock for CardRepository
type MockCardRepository struct {
	mock.Mock
}

func (m *MockCardRepository) Create(ctx context.Context, card *models.Card) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockCardRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Card, error) {
	args := m.Called(ctx, id, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}

func (m *MockCardRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Card), args.Error(1)
}

func (m *MockCardRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Card), args.Error(1)
}

func (m *MockCardRepository) Update(ctx context.Context, card *models.Card) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCardRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCardRepository) GetAllForUserPaginated(ctx context.Context, userID uuid.UUID, params repository.PaginationParams) (*repository.PaginatedResult[models.Card], error) {
	args := m.Called(ctx, userID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PaginatedResult[models.Card]), args.Error(1)
}

func (m *MockCardRepository) FindByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error) {
	args := m.Called(ctx, cardNumber, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}

func (m *MockCardRepository) Search(ctx context.Context, userID uuid.UUID, query string) ([]models.Card, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Card), args.Error(1)
}

// Ensure MockCardRepository implements CardRepository
var _ repository.CardRepository = (*MockCardRepository)(nil)

// ============================================================================
// TESTS
// ============================================================================

func TestCardService_CreateCard_Success(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	card := &models.Card{
		UserID:       &userID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}

	mockRepo.On("Create", ctx, card).Return(nil)

	err := service.CreateCard(ctx, card)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCardService_CreateCard_ValidationError_MissingMerchant(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	card := &models.Card{
		UserID:     &userID,
		CardNumber: "1234567890",
	}

	err := service.CreateCard(ctx, card)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCardService_CreateCard_ValidationError_MissingCardNumber(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	card := &models.Card{
		UserID:       &userID,
		MerchantName: "Test Merchant",
	}

	err := service.CreateCard(ctx, card)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card number is required")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCardService_GetCard_Success(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	cardID := uuid.New()
	userID := uuid.New()
	expectedCard := &models.Card{
		ID:           cardID,
		UserID:       &userID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}

	mockRepo.On("GetByID", ctx, cardID, []string{"Merchant", "User"}).Return(expectedCard, nil)

	card, err := service.GetCard(ctx, cardID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCard, card)
}

func TestCardService_GetCard_NotFound(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	cardID := uuid.New()

	mockRepo.On("GetByID", ctx, cardID, []string{"Merchant", "User"}).Return(nil, gorm.ErrRecordNotFound)

	card, err := service.GetCard(ctx, cardID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, card)
}

func TestCardService_GetUserCards_Success(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	ownedCards := []models.Card{
		{ID: uuid.New(), CardNumber: "1111", MerchantName: "Merchant 1"},
		{ID: uuid.New(), CardNumber: "2222", MerchantName: "Merchant 2"},
	}

	sharedCards := []models.Card{
		{ID: uuid.New(), CardNumber: "3333", MerchantName: "Merchant 3"},
	}

	mockRepo.On("GetByUserID", ctx, userID).Return(ownedCards, nil)
	mockRepo.On("GetSharedWithUser", ctx, userID).Return(sharedCards, nil)

	cards, err := service.GetUserCards(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, cards, 3)
}

func TestCardService_UpdateCard_Success(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	cardID := uuid.New()
	userID := uuid.New()
	card := &models.Card{
		ID:           cardID,
		UserID:       &userID,
		CardNumber:   "9999",
		MerchantName: "Updated Merchant",
	}

	mockRepo.On("Update", ctx, card).Return(nil)

	err := service.UpdateCard(ctx, card)

	assert.NoError(t, err)
}

func TestCardService_DeleteCard_Success(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	cardID := uuid.New()

	mockRepo.On("Delete", ctx, cardID).Return(nil)

	err := service.DeleteCard(ctx, cardID)

	assert.NoError(t, err)
}

func TestCardService_CountUserCards_Success(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedCount := int64(42)

	mockRepo.On("Count", ctx, userID).Return(expectedCount, nil)

	count, err := service.CountUserCards(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
}

func TestCardService_CanUserAccessCard_Owner(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	cardID := uuid.New()
	userID := uuid.New()

	card := &models.Card{
		ID:           cardID,
		UserID:       &userID,
		CardNumber:   "123456",
		MerchantName: "Test",
	}

	mockRepo.On("GetByID", ctx, cardID, mock.Anything).Return(card, nil)

	canAccess, err := service.CanUserAccessCard(ctx, cardID, userID)

	assert.NoError(t, err)
	assert.True(t, canAccess)
	mockRepo.AssertExpectations(t)
}

func TestCardService_CanUserAccessCard_NotFound(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	cardID := uuid.New()
	userID := uuid.New()

	mockRepo.On("GetByID", ctx, cardID, mock.Anything).Return(nil, gorm.ErrRecordNotFound)

	canAccess, err := service.CanUserAccessCard(ctx, cardID, userID)

	assert.NoError(t, err)
	assert.False(t, canAccess)
	mockRepo.AssertExpectations(t)
}

func TestCardService_CanUserAccessCard_NotOwner(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	cardID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	card := &models.Card{
		ID:           cardID,
		UserID:       &otherUserID, // Different user
		CardNumber:   "123456",
		MerchantName: "Test",
	}

	mockRepo.On("GetByID", ctx, cardID, mock.Anything).Return(card, nil)

	canAccess, err := service.CanUserAccessCard(ctx, cardID, userID)

	assert.NoError(t, err)
	assert.False(t, canAccess) // Not owner, no share check implemented
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// ADDITIONAL TESTS FOR ERROR PATH COVERAGE
// ============================================================================

func TestCardService_GetUserCards_OwnedError(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, repoError)

	cards, err := service.GetUserCards(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, cards)
	mockRepo.AssertExpectations(t)
}

func TestCardService_GetUserCards_SharedError(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	ownedCards := []models.Card{
		{ID: uuid.New(), CardNumber: "1111"},
	}
	repoError := errors.New("shared query failed")

	mockRepo.On("GetByUserID", ctx, userID).Return(ownedCards, nil)
	mockRepo.On("GetSharedWithUser", ctx, userID).Return(nil, repoError)

	cards, err := service.GetUserCards(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, cards)
	mockRepo.AssertExpectations(t)
}

func TestCardService_UpdateCard_ValidationMissingMerchant(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	merchantID := uuid.New()
	card := &models.Card{
		ID:           uuid.New(),
		CardNumber:   "1234",
		MerchantID:   &merchantID, // MerchantID is set
		MerchantName: "",          // But MerchantName is missing
	}

	err := service.UpdateCard(ctx, card)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestCardService_UpdateCard_ValidationMissingCardNumber(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	card := &models.Card{
		ID:           uuid.New(),
		CardNumber:   "", // Missing card number
		MerchantName: "Test Merchant",
	}

	err := service.UpdateCard(ctx, card)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card number is required")
	mockRepo.AssertNotCalled(t, "Update")
}

// ============================================================================
// CHECK DUPLICATE TESTS
// ============================================================================

func TestCardService_CheckDuplicate_DuplicateFound(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	existingCard := &models.Card{
		ID:           uuid.New(),
		UserID:       &userID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}

	mockRepo.On("FindByCardNumber", ctx, "1234567890", userID).Return(existingCard, nil)

	result, err := service.CheckDuplicate(ctx, "1234567890", userID, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingCard.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestCardService_CheckDuplicate_NoDuplicate(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("FindByCardNumber", ctx, "9999999999", userID).Return(nil, nil)

	result, err := service.CheckDuplicate(ctx, "9999999999", userID, nil)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCardService_CheckDuplicate_ExcludeSameCard(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	cardID := uuid.New()
	existingCard := &models.Card{
		ID:           cardID,
		UserID:       &userID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}

	mockRepo.On("FindByCardNumber", ctx, "1234567890", userID).Return(existingCard, nil)

	// Exclude the same card (update case) - should not be considered a duplicate
	result, err := service.CheckDuplicate(ctx, "1234567890", userID, &cardID)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCardService_CheckDuplicate_ExcludeDifferentCard(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	existingCardID := uuid.New()
	differentCardID := uuid.New()
	existingCard := &models.Card{
		ID:           existingCardID,
		UserID:       &userID,
		CardNumber:   "1234567890",
		MerchantName: "Test Merchant",
	}

	mockRepo.On("FindByCardNumber", ctx, "1234567890", userID).Return(existingCard, nil)

	// Exclude a different card - should still be considered a duplicate
	result, err := service.CheckDuplicate(ctx, "1234567890", userID, &differentCardID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingCardID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestCardService_CheckDuplicate_RepositoryError(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
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
// GET USER CARDS PAGINATED TESTS
// ============================================================================

func TestCardService_GetUserCardsPaginated_Success(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedResult := &repository.PaginatedResult[models.Card]{
		Items: []models.Card{
			{ID: uuid.New(), CardNumber: "1111", MerchantName: "Merchant 1"},
			{ID: uuid.New(), CardNumber: "2222", MerchantName: "Merchant 2"},
		},
		Total:      10,
		Page:       1,
		PerPage:    2,
		TotalPages: 5,
	}

	params := repository.PaginationParams{Page: 1, PerPage: 2}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(expectedResult, nil)

	result, err := service.GetUserCardsPaginated(ctx, userID, 1, 2)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(10), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.PerPage)
	assert.Equal(t, 5, result.TotalPages)
	mockRepo.AssertExpectations(t)
}

func TestCardService_GetUserCardsPaginated_EmptyResult(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedResult := &repository.PaginatedResult[models.Card]{
		Items:      []models.Card{},
		Total:      0,
		Page:       1,
		PerPage:    10,
		TotalPages: 0,
	}

	params := repository.PaginationParams{Page: 1, PerPage: 10}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(expectedResult, nil)

	result, err := service.GetUserCardsPaginated(ctx, userID, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
	mockRepo.AssertExpectations(t)
}

func TestCardService_GetUserCardsPaginated_RepositoryError(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := NewCardService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	params := repository.PaginationParams{Page: 1, PerPage: 10}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(nil, repoError)

	result, err := service.GetUserCardsPaginated(ctx, userID, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get paginated cards")
	mockRepo.AssertExpectations(t)
}
