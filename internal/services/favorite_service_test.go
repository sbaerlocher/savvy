package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"savvy/internal/models"
)

// MockFavoriteRepository is a manual mock for FavoriteRepository
type MockFavoriteRepository struct {
	mock.Mock
}

func (m *MockFavoriteRepository) Create(ctx context.Context, favorite *models.UserFavorite) error {
	args := m.Called(ctx, favorite)
	return args.Error(0)
}

func (m *MockFavoriteRepository) GetByUserAndResource(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (*models.UserFavorite, error) {
	args := m.Called(ctx, userID, resourceType, resourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserFavorite), args.Error(1)
}

func (m *MockFavoriteRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.UserFavorite, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserFavorite), args.Error(1)
}

func (m *MockFavoriteRepository) Restore(ctx context.Context, favorite *models.UserFavorite) error {
	args := m.Called(ctx, favorite)
	return args.Error(0)
}

func (m *MockFavoriteRepository) Delete(ctx context.Context, favorite *models.UserFavorite) error {
	args := m.Called(ctx, favorite)
	return args.Error(0)
}

func (m *MockFavoriteRepository) IsFavorite(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, resourceType, resourceID)
	return args.Bool(0), args.Error(1)
}

// MockCardRepository is defined in card_service_test.go but we need it here too
type MockCardRepositoryFav struct {
	mock.Mock
}

func (m *MockCardRepositoryFav) Create(ctx context.Context, card *models.Card) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockCardRepositoryFav) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Card, error) {
	args := m.Called(ctx, id, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}

func (m *MockCardRepositoryFav) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Card), args.Error(1)
}

func (m *MockCardRepositoryFav) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Card), args.Error(1)
}

func (m *MockCardRepositoryFav) Update(ctx context.Context, card *models.Card) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockCardRepositoryFav) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCardRepositoryFav) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockVoucherRepositoryFav mock
type MockVoucherRepositoryFav struct {
	mock.Mock
}

func (m *MockVoucherRepositoryFav) Create(ctx context.Context, voucher *models.Voucher) error {
	args := m.Called(ctx, voucher)
	return args.Error(0)
}

func (m *MockVoucherRepositoryFav) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Voucher, error) {
	args := m.Called(ctx, id, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Voucher), args.Error(1)
}

func (m *MockVoucherRepositoryFav) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *MockVoucherRepositoryFav) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *MockVoucherRepositoryFav) Update(ctx context.Context, voucher *models.Voucher) error {
	args := m.Called(ctx, voucher)
	return args.Error(0)
}

func (m *MockVoucherRepositoryFav) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVoucherRepositoryFav) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockVoucherRepositoryFav) Redeem(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockGiftCardRepositoryFav mock
type MockGiftCardRepositoryFav struct {
	mock.Mock
}

func (m *MockGiftCardRepositoryFav) Create(ctx context.Context, giftCard *models.GiftCard) error {
	args := m.Called(ctx, giftCard)
	return args.Error(0)
}

func (m *MockGiftCardRepositoryFav) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.GiftCard, error) {
	args := m.Called(ctx, id, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCard), args.Error(1)
}

func (m *MockGiftCardRepositoryFav) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

func (m *MockGiftCardRepositoryFav) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

func (m *MockGiftCardRepositoryFav) Update(ctx context.Context, giftCard *models.GiftCard) error {
	args := m.Called(ctx, giftCard)
	return args.Error(0)
}

func (m *MockGiftCardRepositoryFav) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGiftCardRepositoryFav) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGiftCardRepositoryFav) GetTotalBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

// ============================================================================
// TESTS
// ============================================================================

func TestFavoriteService_ToggleFavorite_CreateNew(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	resourceID := uuid.New()

	// Favorite doesn't exist
	mockFavRepo.On("GetByUserAndResource", ctx, userID, "card", resourceID).Return(nil, assert.AnError)
	mockFavRepo.On("Create", ctx, mock.MatchedBy(func(f *models.UserFavorite) bool {
		return f.UserID == userID && f.ResourceType == "card" && f.ResourceID == resourceID
	})).Return(nil)

	err := service.ToggleFavorite(ctx, userID, "card", resourceID)

	assert.NoError(t, err)
	mockFavRepo.AssertExpectations(t)
}

func TestFavoriteService_ToggleFavorite_SoftDeleteActive(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	resourceID := uuid.New()

	existingFavorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   resourceID,
		DeletedAt:    gorm.DeletedAt{Valid: false}, // Active favorite
	}

	mockFavRepo.On("GetByUserAndResource", ctx, userID, "card", resourceID).Return(existingFavorite, nil)
	mockFavRepo.On("Delete", ctx, existingFavorite).Return(nil)

	err := service.ToggleFavorite(ctx, userID, "card", resourceID)

	assert.NoError(t, err)
	mockFavRepo.AssertExpectations(t)
}

func TestFavoriteService_ToggleFavorite_RestoreSoftDeleted(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	resourceID := uuid.New()

	existingFavorite := &models.UserFavorite{
		UserID:       userID,
		ResourceType: "card",
		ResourceID:   resourceID,
		DeletedAt:    gorm.DeletedAt{Time: time.Now(), Valid: true}, // Soft-deleted favorite
	}

	mockFavRepo.On("GetByUserAndResource", ctx, userID, "card", resourceID).Return(existingFavorite, nil)
	mockFavRepo.On("Restore", ctx, existingFavorite).Return(nil)

	err := service.ToggleFavorite(ctx, userID, "card", resourceID)

	assert.NoError(t, err)
	mockFavRepo.AssertExpectations(t)
}

func TestFavoriteService_GetUserFavorites_Success(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedFavorites := []models.UserFavorite{
		{UserID: userID, ResourceType: "card", ResourceID: uuid.New()},
		{UserID: userID, ResourceType: "voucher", ResourceID: uuid.New()},
	}

	mockFavRepo.On("GetByUser", ctx, userID).Return(expectedFavorites, nil)

	favorites, err := service.GetUserFavorites(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, favorites, 2)
	assert.Equal(t, expectedFavorites, favorites)
}

func TestFavoriteService_IsFavorite_True(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	resourceID := uuid.New()

	mockFavRepo.On("IsFavorite", ctx, userID, "card", resourceID).Return(true, nil)

	isFav, err := service.IsFavorite(ctx, userID, "card", resourceID)

	assert.NoError(t, err)
	assert.True(t, isFav)
}

func TestFavoriteService_IsFavorite_False(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	resourceID := uuid.New()

	mockFavRepo.On("IsFavorite", ctx, userID, "card", resourceID).Return(false, nil)

	isFav, err := service.IsFavorite(ctx, userID, "card", resourceID)

	assert.NoError(t, err)
	assert.False(t, isFav)
}

func TestFavoriteService_GetFavoriteCards_Success(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	cardID1 := uuid.New()
	cardID2 := uuid.New()

	favorites := []models.UserFavorite{
		{UserID: userID, ResourceType: "card", ResourceID: cardID1},
		{UserID: userID, ResourceType: "card", ResourceID: cardID2},
		{UserID: userID, ResourceType: "voucher", ResourceID: uuid.New()}, // Should be ignored
	}

	card1 := &models.Card{ID: cardID1, CardNumber: "1111"}
	card2 := &models.Card{ID: cardID2, CardNumber: "2222"}

	mockFavRepo.On("GetByUser", ctx, userID).Return(favorites, nil)
	mockCardRepo.On("GetByID", ctx, cardID1, []string{"Merchant", "User"}).Return(card1, nil)
	mockCardRepo.On("GetByID", ctx, cardID2, []string{"Merchant", "User"}).Return(card2, nil)

	cards, err := service.GetFavoriteCards(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, cards, 2)
	assert.Equal(t, "1111", cards[0].CardNumber)
	assert.Equal(t, "2222", cards[1].CardNumber)
}

func TestFavoriteService_GetFavoriteVouchers_Success(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	voucherID := uuid.New()

	favorites := []models.UserFavorite{
		{UserID: userID, ResourceType: "voucher", ResourceID: voucherID},
	}

	voucher := &models.Voucher{ID: voucherID, Code: "TEST123"}

	mockFavRepo.On("GetByUser", ctx, userID).Return(favorites, nil)
	mockVoucherRepo.On("GetByID", ctx, voucherID, []string{"Merchant", "User"}).Return(voucher, nil)

	vouchers, err := service.GetFavoriteVouchers(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, vouchers, 1)
	assert.Equal(t, "TEST123", vouchers[0].Code)
}

func TestFavoriteService_GetFavoriteGiftCards_Success(t *testing.T) {
	mockFavRepo := new(MockFavoriteRepository)
	mockCardRepo := new(MockCardRepositoryFav)
	mockVoucherRepo := new(MockVoucherRepositoryFav)
	mockGiftCardRepo := new(MockGiftCardRepositoryFav)

	service := NewFavoriteService(mockFavRepo, mockCardRepo, mockVoucherRepo, mockGiftCardRepo)
	ctx := context.Background()

	userID := uuid.New()
	giftCardID := uuid.New()

	favorites := []models.UserFavorite{
		{UserID: userID, ResourceType: "gift_card", ResourceID: giftCardID},
	}

	giftCard := &models.GiftCard{ID: giftCardID, CardNumber: "GIFT123"}

	mockFavRepo.On("GetByUser", ctx, userID).Return(favorites, nil)
	mockGiftCardRepo.On("GetByID", ctx, giftCardID, []string{"Merchant", "User"}).Return(giftCard, nil)

	giftCards, err := service.GetFavoriteGiftCards(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, giftCards, 1)
	assert.Equal(t, "GIFT123", giftCards[0].CardNumber)
}
