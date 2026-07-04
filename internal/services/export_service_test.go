package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// ============================================================================
// MOCKS FOR EXPORT SERVICE
// ============================================================================

// mockExportUserService implements UserServiceInterface for export tests.
type mockExportUserService struct {
	mock.Mock
}

func (m *mockExportUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockExportUserService) GetUserByEmail(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}

func (m *mockExportUserService) CreateUser(_ context.Context, _ *models.User) error { return nil }

func (m *mockExportUserService) UpdateUser(_ context.Context, _ *models.User) error { return nil }

func (m *mockExportUserService) DeleteUser(_ context.Context, _ uuid.UUID) error { return nil }

func (m *mockExportUserService) GetUserCount(_ context.Context) (int64, error) { return 0, nil }

func (m *mockExportUserService) GetAllUsers(_ context.Context) ([]models.User, error) {
	return nil, nil
}

func (m *mockExportUserService) SearchUsers(_ context.Context, _ string) ([]models.User, error) {
	return nil, nil
}

func (m *mockExportUserService) GetOAuthUserCount(_ context.Context) (int64, error) { return 0, nil }

func (m *mockExportUserService) GetLocalUserCount(_ context.Context) (int64, error) { return 0, nil }

var _ UserServiceInterface = (*mockExportUserService)(nil)

// mockExportCardRepo implements repository.CardRepository for export tests.
type mockExportCardRepo struct {
	mock.Mock
}

func (m *mockExportCardRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Card, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Card), args.Error(1)
}

func (m *mockExportCardRepo) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Card, error) {
	callArgs := []interface{}{ctx, id}
	for _, p := range preloads {
		callArgs = append(callArgs, p)
	}
	args := m.Called(callArgs...)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}

func (m *mockExportCardRepo) Create(_ context.Context, _ *models.Card) error { return nil }
func (m *mockExportCardRepo) GetSharedWithUser(_ context.Context, _ uuid.UUID) ([]models.Card, error) {
	return nil, nil
}
func (m *mockExportCardRepo) Update(_ context.Context, _ *models.Card) error { return nil }
func (m *mockExportCardRepo) Delete(_ context.Context, _ uuid.UUID) error    { return nil }
func (m *mockExportCardRepo) Count(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockExportCardRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.Card], error) {
	return nil, nil
}
func (m *mockExportCardRepo) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *mockExportCardRepo) FindDeletedByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *mockExportCardRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

var _ repository.CardRepository = (*mockExportCardRepo)(nil)

// mockExportVoucherRepo implements repository.VoucherRepository for export tests.
type mockExportVoucherRepo struct {
	mock.Mock
}

func (m *mockExportVoucherRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *mockExportVoucherRepo) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Voucher, error) {
	callArgs := []interface{}{ctx, id}
	for _, p := range preloads {
		callArgs = append(callArgs, p)
	}
	args := m.Called(callArgs...)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Voucher), args.Error(1)
}

func (m *mockExportVoucherRepo) Create(_ context.Context, _ *models.Voucher) error { return nil }
func (m *mockExportVoucherRepo) GetSharedWithUser(_ context.Context, _ uuid.UUID) ([]models.Voucher, error) {
	return nil, nil
}
func (m *mockExportVoucherRepo) Update(_ context.Context, _ *models.Voucher) error { return nil }
func (m *mockExportVoucherRepo) Delete(_ context.Context, _ uuid.UUID) error       { return nil }
func (m *mockExportVoucherRepo) Count(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockExportVoucherRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.Voucher], error) {
	return nil, nil
}
func (m *mockExportVoucherRepo) GetExpiringVouchers(_ context.Context, _ int) ([]models.Voucher, error) {
	return nil, nil
}
func (m *mockExportVoucherRepo) GetVouchersStartingTomorrow(_ context.Context) ([]models.Voucher, error) {
	return nil, nil
}
func (m *mockExportVoucherRepo) FindByVoucherCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *mockExportVoucherRepo) FindDeletedByCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}
func (m *mockExportVoucherRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

var _ repository.VoucherRepository = (*mockExportVoucherRepo)(nil)

// mockExportGiftCardRepo implements repository.GiftCardRepository for export tests.
type mockExportGiftCardRepo struct {
	mock.Mock
}

func (m *mockExportGiftCardRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.GiftCard, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GiftCard), args.Error(1)
}

func (m *mockExportGiftCardRepo) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.GiftCard, error) {
	callArgs := []interface{}{ctx, id}
	for _, p := range preloads {
		callArgs = append(callArgs, p)
	}
	args := m.Called(callArgs...)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GiftCard), args.Error(1)
}

func (m *mockExportGiftCardRepo) Create(_ context.Context, _ *models.GiftCard) error { return nil }
func (m *mockExportGiftCardRepo) GetSharedWithUser(_ context.Context, _ uuid.UUID) ([]models.GiftCard, error) {
	return nil, nil
}
func (m *mockExportGiftCardRepo) Update(_ context.Context, _ *models.GiftCard) error { return nil }
func (m *mockExportGiftCardRepo) Delete(_ context.Context, _ uuid.UUID) error        { return nil }
func (m *mockExportGiftCardRepo) Count(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockExportGiftCardRepo) GetTotalBalance(_ context.Context, _ uuid.UUID) (float64, error) {
	return 0, nil
}
func (m *mockExportGiftCardRepo) CreateTransaction(_ context.Context, _ *models.GiftCardTransaction) error {
	return nil
}
func (m *mockExportGiftCardRepo) GetTransaction(_ context.Context, _, _ uuid.UUID) (*models.GiftCardTransaction, error) {
	return nil, nil
}
func (m *mockExportGiftCardRepo) DeleteTransaction(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockExportGiftCardRepo) GetAllForUserPaginated(_ context.Context, _ uuid.UUID, _ repository.PaginationParams) (*repository.PaginatedResult[models.GiftCard], error) {
	return nil, nil
}
func (m *mockExportGiftCardRepo) GetExpiringGiftCards(_ context.Context, _ int) ([]models.GiftCard, error) {
	return nil, nil
}
func (m *mockExportGiftCardRepo) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *mockExportGiftCardRepo) FindDeletedByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
func (m *mockExportGiftCardRepo) RestoreByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

var _ repository.GiftCardRepository = (*mockExportGiftCardRepo)(nil)

// mockExportFavoriteRepo implements repository.FavoriteRepository for export tests.
type mockExportFavoriteRepo struct {
	mock.Mock
}

func (m *mockExportFavoriteRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.UserFavorite, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserFavorite), args.Error(1)
}

func (m *mockExportFavoriteRepo) Create(_ context.Context, _ *models.UserFavorite) error {
	return nil
}
func (m *mockExportFavoriteRepo) GetByUserAndResource(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) (*models.UserFavorite, error) {
	return nil, nil
}
func (m *mockExportFavoriteRepo) Delete(_ context.Context, _ *models.UserFavorite) error {
	return nil
}
func (m *mockExportFavoriteRepo) Restore(_ context.Context, _ *models.UserFavorite) error {
	return nil
}
func (m *mockExportFavoriteRepo) IsFavorite(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}

var _ repository.FavoriteRepository = (*mockExportFavoriteRepo)(nil)

// ============================================================================
// TEST SETUP
// ============================================================================

type exportTestSetup struct {
	service      ExportServiceInterface
	userService  *mockExportUserService
	cardRepo     *mockExportCardRepo
	voucherRepo  *mockExportVoucherRepo
	giftCardRepo *mockExportGiftCardRepo
	favoriteRepo *mockExportFavoriteRepo
}

func newExportTestSetup() *exportTestSetup {
	us := new(mockExportUserService)
	cr := new(mockExportCardRepo)
	vr := new(mockExportVoucherRepo)
	gr := new(mockExportGiftCardRepo)
	fr := new(mockExportFavoriteRepo)
	return &exportTestSetup{
		service:      NewExportService(us, cr, vr, gr, fr),
		userService:  us,
		cardRepo:     cr,
		voucherRepo:  vr,
		giftCardRepo: gr,
		favoriteRepo: fr,
	}
}

// ============================================================================
// ExportCardsByIDs - Ownership Enforcement Tests
// ============================================================================

func TestExportCardsByIDs_OnlyExportsOwnedCards(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	ownedCardID := uuid.New()
	foreignCardID := uuid.New()

	ownedCard := models.Card{
		ID:           ownedCardID,
		UserID:       &userID,
		MerchantName: "IKEA",
		Program:      "Family",
		CardNumber:   "CARD-001",
		BarcodeType:  "CODE128",
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	// GetByUserID returns only the user's own cards
	s.cardRepo.On("GetByUserID", mock.Anything, userID).
		Return([]models.Card{ownedCard}, nil)

	// GetByID should only be called for the owned card
	s.cardRepo.On("GetByID", mock.Anything, ownedCardID, "Merchant").
		Return(&ownedCard, nil)

	// Request export for both owned and foreign card
	result, err := s.service.ExportCardsByIDs(ctx, userID, []uuid.UUID{ownedCardID, foreignCardID})

	require.NoError(t, err)
	assert.Len(t, result.Cards, 1)
	assert.Equal(t, ownedCardID.String(), result.Cards[0].ID)

	// GetByID should NOT have been called for the foreign card
	s.cardRepo.AssertNotCalled(t, "GetByID", mock.Anything, foreignCardID, "Merchant")
}

func TestExportCardsByIDs_EmptyWhenNoOwnedCards(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	foreignCardID := uuid.New()

	// User owns no cards
	s.cardRepo.On("GetByUserID", mock.Anything, userID).
		Return([]models.Card{}, nil)

	result, err := s.service.ExportCardsByIDs(ctx, userID, []uuid.UUID{foreignCardID})

	require.NoError(t, err)
	assert.Empty(t, result.Cards)
}

func TestExportCardsByIDs_HandlesGetByUserIDError(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	// GetByUserID fails → treated as "no owned cards"
	s.cardRepo.On("GetByUserID", mock.Anything, userID).
		Return(nil, errors.New("db error"))

	result, err := s.service.ExportCardsByIDs(ctx, userID, []uuid.UUID{uuid.New()})

	require.NoError(t, err)
	assert.Empty(t, result.Cards)
}

// ============================================================================
// ExportVouchersByIDs - Ownership Enforcement Tests
// ============================================================================

func TestExportVouchersByIDs_OnlyExportsOwnedVouchers(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	ownedVoucherID := uuid.New()
	foreignVoucherID := uuid.New()

	ownedVoucher := models.Voucher{
		ID:           ownedVoucherID,
		UserID:       &userID,
		MerchantName: "Zalando",
		Code:         "SAVE20",
		Type:         "percentage",
		Value:        20,
		ValidFrom:    time.Now(),
		ValidUntil:   time.Now().Add(30 * 24 * time.Hour),
		BarcodeType:  "CODE128",
		CreatedAt:    time.Now(),
	}

	s.voucherRepo.On("GetByUserID", mock.Anything, userID).
		Return([]models.Voucher{ownedVoucher}, nil)

	s.voucherRepo.On("GetByID", mock.Anything, ownedVoucherID, "Merchant").
		Return(&ownedVoucher, nil)

	result, err := s.service.ExportVouchersByIDs(ctx, userID, []uuid.UUID{ownedVoucherID, foreignVoucherID})

	require.NoError(t, err)
	assert.Len(t, result.Vouchers, 1)
	assert.Equal(t, ownedVoucherID.String(), result.Vouchers[0].ID)

	s.voucherRepo.AssertNotCalled(t, "GetByID", mock.Anything, foreignVoucherID, "Merchant")
}

func TestExportVouchersByIDs_EmptyWhenNoOwnedVouchers(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	s.voucherRepo.On("GetByUserID", mock.Anything, userID).
		Return([]models.Voucher{}, nil)

	result, err := s.service.ExportVouchersByIDs(ctx, userID, []uuid.UUID{uuid.New()})

	require.NoError(t, err)
	assert.Empty(t, result.Vouchers)
}

// ============================================================================
// ExportGiftCardsByIDs - Ownership Enforcement Tests
// ============================================================================

func TestExportGiftCardsByIDs_OnlyExportsOwnedGiftCards(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	ownedGCID := uuid.New()
	foreignGCID := uuid.New()

	ownedGC := models.GiftCard{
		ID:             ownedGCID,
		UserID:         &userID,
		MerchantName:   "Amazon",
		CardNumber:     "GC-001",
		InitialBalance: 100,
		CurrentBalance: 75,
		Currency:       "CHF",
		BarcodeType:    "CODE128",
		CreatedAt:      time.Now(),
	}

	s.giftCardRepo.On("GetByUserID", mock.Anything, userID).
		Return([]models.GiftCard{ownedGC}, nil)

	s.giftCardRepo.On("GetByID", mock.Anything, ownedGCID, "Merchant", "Transactions").
		Return(&ownedGC, nil)

	result, err := s.service.ExportGiftCardsByIDs(ctx, userID, []uuid.UUID{ownedGCID, foreignGCID})

	require.NoError(t, err)
	assert.Len(t, result.GiftCards, 1)
	assert.Equal(t, ownedGCID.String(), result.GiftCards[0].ID)
	assert.Equal(t, float64(75), result.GiftCards[0].CurrentBalance)

	s.giftCardRepo.AssertNotCalled(t, "GetByID", mock.Anything, foreignGCID, "Merchant", "Transactions")
}

func TestExportGiftCardsByIDs_EmptyWhenNoOwnedGiftCards(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	s.giftCardRepo.On("GetByUserID", mock.Anything, userID).
		Return([]models.GiftCard{}, nil)

	result, err := s.service.ExportGiftCardsByIDs(ctx, userID, []uuid.UUID{uuid.New()})

	require.NoError(t, err)
	assert.Empty(t, result.GiftCards)
}

func TestExportGiftCardsByIDs_UsesResolvedMerchantName(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	gcID := uuid.New()

	merchantName := "Amazon DE"
	gc := models.GiftCard{
		ID:             gcID,
		UserID:         &userID,
		MerchantName:   "fallback",
		Merchant:       &models.Merchant{Name: merchantName},
		CardNumber:     "GC-002",
		InitialBalance: 50,
		CurrentBalance: 50,
		Currency:       "EUR",
		BarcodeType:    "QR_CODE",
		CreatedAt:      time.Now(),
	}

	s.giftCardRepo.On("GetByUserID", mock.Anything, userID).
		Return([]models.GiftCard{gc}, nil)
	s.giftCardRepo.On("GetByID", mock.Anything, gcID, "Merchant", "Transactions").
		Return(&gc, nil)

	result, err := s.service.ExportGiftCardsByIDs(ctx, userID, []uuid.UUID{gcID})

	require.NoError(t, err)
	require.Len(t, result.GiftCards, 1)
	assert.Equal(t, merchantName, result.GiftCards[0].MerchantName)
}

// ============================================================================
// ExportUserData Tests
// ============================================================================

func TestExportService_ExportUserData_Success(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	user := &models.User{
		ID:            userID,
		Email:         "test@example.com",
		FirstName:     "John",
		LastName:      "Doe",
		AuthProvider:  "local",
		EmailVerified: true,
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	cardID := uuid.New()
	cards := []models.Card{
		{
			ID:           cardID,
			UserID:       &userID,
			MerchantName: "IKEA",
			Merchant:     &models.Merchant{Name: "IKEA Store"},
			Program:      "Family",
			CardNumber:   "CARD-001",
			BarcodeType:  "CODE128",
			Status:       "active",
			Notes:        "My IKEA card",
			CreatedAt:    time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	voucherID := uuid.New()
	vouchers := []models.Voucher{
		{
			ID:                voucherID,
			UserID:            &userID,
			MerchantName:      "Zalando",
			Code:              "SAVE20",
			Type:              "percentage",
			Value:             20,
			Description:       "20% off",
			MinPurchaseAmount: 50,
			ValidFrom:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			ValidUntil:        time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			UsageLimitType:    "single",
			BarcodeType:       "QR_CODE",
			CreatedAt:         time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	gcID := uuid.New()
	txID := uuid.New()
	giftCards := []models.GiftCard{
		{
			ID:             gcID,
			UserID:         &userID,
			MerchantName:   "Amazon",
			CardNumber:     "GC-001",
			InitialBalance: 100,
			CurrentBalance: 75,
			Currency:       "CHF",
			PIN:            "1234",
			BarcodeType:    "CODE128",
			Notes:          "Birthday gift",
			CreatedAt:      time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC),
			Transactions: []models.GiftCardTransaction{
				{
					ID:              txID,
					GiftCardID:      gcID,
					Amount:          -25,
					Description:     "Purchase",
					TransactionDate: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					CreatedAt:       time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	favID := uuid.New()
	favorites := []models.UserFavorite{
		{
			UserID:       userID,
			ResourceType: "card",
			ResourceID:   favID,
		},
	}

	s.userService.On("GetUserByID", ctx, userID).Return(user, nil)
	s.cardRepo.On("GetByUserID", ctx, userID).Return(cards, nil)
	s.voucherRepo.On("GetByUserID", ctx, userID).Return(vouchers, nil)
	s.giftCardRepo.On("GetByUserID", ctx, userID).Return(giftCards, nil)
	s.favoriteRepo.On("GetByUser", ctx, userID).Return(favorites, nil)

	result, err := s.service.ExportUserData(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify user
	assert.Equal(t, userID.String(), result.User.ID)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Equal(t, "John", result.User.FirstName)
	assert.Equal(t, "Doe", result.User.LastName)
	assert.Equal(t, "local", result.User.AuthProvider)
	assert.True(t, result.User.EmailVerified)

	// Verify cards (uses Merchant.Name when available)
	require.Len(t, result.Cards, 1)
	assert.Equal(t, cardID.String(), result.Cards[0].ID)
	assert.Equal(t, "IKEA Store", result.Cards[0].MerchantName)
	assert.Equal(t, "Family", result.Cards[0].Program)
	assert.Equal(t, "CARD-001", result.Cards[0].CardNumber)

	// Verify vouchers
	require.Len(t, result.Vouchers, 1)
	assert.Equal(t, voucherID.String(), result.Vouchers[0].ID)
	assert.Equal(t, "Zalando", result.Vouchers[0].MerchantName)
	assert.Equal(t, "SAVE20", result.Vouchers[0].Code)
	assert.Equal(t, float64(20), result.Vouchers[0].Value)

	// Verify gift cards
	require.Len(t, result.GiftCards, 1)
	assert.Equal(t, gcID.String(), result.GiftCards[0].ID)
	assert.Equal(t, "Amazon", result.GiftCards[0].MerchantName)
	assert.Equal(t, float64(75), result.GiftCards[0].CurrentBalance)
	assert.Equal(t, "1234", result.GiftCards[0].PIN)

	// Verify transactions
	require.Len(t, result.GiftCards[0].Transactions, 1)
	assert.Equal(t, txID.String(), result.GiftCards[0].Transactions[0].ID)
	assert.Equal(t, float64(-25), result.GiftCards[0].Transactions[0].Amount)

	// Verify favorites
	require.Len(t, result.Favorites, 1)
	assert.Equal(t, "card", result.Favorites[0].ResourceType)
	assert.Equal(t, favID.String(), result.Favorites[0].ResourceID)

	// Verify exported_at is set
	assert.NotEmpty(t, result.ExportedAt)
}

func TestExportService_ExportUserData_EmptyData(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	user := &models.User{
		ID:        userID,
		Email:     "empty@example.com",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	s.userService.On("GetUserByID", ctx, userID).Return(user, nil)
	s.cardRepo.On("GetByUserID", ctx, userID).Return([]models.Card{}, nil)
	s.voucherRepo.On("GetByUserID", ctx, userID).Return([]models.Voucher{}, nil)
	s.giftCardRepo.On("GetByUserID", ctx, userID).Return([]models.GiftCard{}, nil)
	s.favoriteRepo.On("GetByUser", ctx, userID).Return([]models.UserFavorite{}, nil)

	result, err := s.service.ExportUserData(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID.String(), result.User.ID)
	assert.Empty(t, result.Cards)
	assert.Empty(t, result.Vouchers)
	assert.Empty(t, result.GiftCards)
	assert.Empty(t, result.Favorites)
}

func TestExportService_ExportUserData_CardRepoError(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	s.userService.On("GetUserByID", ctx, userID).Return(user, nil)
	s.cardRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("card db error"))

	result, err := s.service.ExportUserData(ctx, userID)

	assert.Nil(t, result)
	assert.EqualError(t, err, "card db error")
}

func TestExportService_ExportUserData_VoucherRepoError(t *testing.T) {
	s := newExportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	s.userService.On("GetUserByID", ctx, userID).Return(user, nil)
	s.cardRepo.On("GetByUserID", ctx, userID).Return([]models.Card{}, nil)
	s.voucherRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("voucher db error"))

	result, err := s.service.ExportUserData(ctx, userID)

	assert.Nil(t, result)
	assert.EqualError(t, err, "voucher db error")
}
