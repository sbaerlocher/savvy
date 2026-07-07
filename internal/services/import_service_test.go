package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// ============================================================================
// MOCKS FOR IMPORT SERVICE
// ============================================================================

// mockImportCardService implements CardServiceInterface for import tests.
type mockImportCardService struct {
	mock.Mock
}

func (m *mockImportCardService) CreateCard(ctx context.Context, card *models.Card) error {
	return m.Called(ctx, card).Error(0)
}

func (m *mockImportCardService) GetCard(_ context.Context, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}

func (m *mockImportCardService) GetUserCards(_ context.Context, _ uuid.UUID) ([]models.Card, error) {
	return nil, nil
}

func (m *mockImportCardService) GetUserCardsPaginated(_ context.Context, _ uuid.UUID, _, _ int) (*repository.PaginatedResult[models.Card], error) {
	return nil, nil
}

func (m *mockImportCardService) UpdateCard(_ context.Context, _ *models.Card) error {
	return nil
}

func (m *mockImportCardService) DeleteCard(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockImportCardService) CountUserCards(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockImportCardService) CanUserAccessCard(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockImportCardService) CheckDuplicate(_ context.Context, _ string, _ uuid.UUID, _ *uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *mockImportCardService) CheckSharedDuplicate(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}

func (m *mockImportCardService) FindDeletedDuplicate(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}

func (m *mockImportCardService) RestoreCard(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}

var _ CardServiceInterface = (*mockImportCardService)(nil)

// mockImportVoucherService implements VoucherServiceInterface for import tests.
type mockImportVoucherService struct {
	mock.Mock
}

func (m *mockImportVoucherService) CreateVoucher(ctx context.Context, voucher *models.Voucher) error {
	return m.Called(ctx, voucher).Error(0)
}

func (m *mockImportVoucherService) GetVoucher(_ context.Context, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}

func (m *mockImportVoucherService) GetUserVouchers(_ context.Context, _ uuid.UUID) ([]models.Voucher, error) {
	return nil, nil
}

func (m *mockImportVoucherService) GetUserVouchersPaginated(_ context.Context, _ uuid.UUID, _, _ int) (*repository.PaginatedResult[models.Voucher], error) {
	return nil, nil
}

func (m *mockImportVoucherService) UpdateVoucher(_ context.Context, _ *models.Voucher) error {
	return nil
}

func (m *mockImportVoucherService) DeleteVoucher(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockImportVoucherService) CountUserVouchers(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockImportVoucherService) CheckDuplicate(_ context.Context, _ string, _ uuid.UUID, _ *uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}

func (m *mockImportVoucherService) FindDeletedDuplicate(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}

func (m *mockImportVoucherService) RestoreVoucher(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}

var _ VoucherServiceInterface = (*mockImportVoucherService)(nil)

// mockImportGiftCardService implements GiftCardServiceInterface for import tests.
type mockImportGiftCardService struct {
	mock.Mock
}

func (m *mockImportGiftCardService) CreateGiftCard(ctx context.Context, giftCard *models.GiftCard) error {
	return m.Called(ctx, giftCard).Error(0)
}

func (m *mockImportGiftCardService) GetGiftCard(_ context.Context, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}

func (m *mockImportGiftCardService) GetUserGiftCards(_ context.Context, _ uuid.UUID) ([]models.GiftCard, error) {
	return nil, nil
}

func (m *mockImportGiftCardService) GetUserGiftCardsPaginated(_ context.Context, _ uuid.UUID, _, _ int) (*repository.PaginatedResult[models.GiftCard], error) {
	return nil, nil
}

func (m *mockImportGiftCardService) UpdateGiftCard(_ context.Context, _ *models.GiftCard) error {
	return nil
}

func (m *mockImportGiftCardService) DeleteGiftCard(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockImportGiftCardService) CountUserGiftCards(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockImportGiftCardService) CheckDuplicate(_ context.Context, _ string, _ uuid.UUID, _ *uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}

func (m *mockImportGiftCardService) GetTotalBalance(_ context.Context, _ uuid.UUID) (float64, error) {
	return 0, nil
}

func (m *mockImportGiftCardService) GetCurrentBalance(_ context.Context, _ uuid.UUID) (float64, error) {
	return 0, nil
}

func (m *mockImportGiftCardService) CanUserAccessGiftCard(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockImportGiftCardService) CreateTransaction(_ context.Context, _ *models.GiftCardTransaction) error {
	return nil
}

func (m *mockImportGiftCardService) GetTransaction(_ context.Context, _, _ uuid.UUID) (*models.GiftCardTransaction, error) {
	return nil, nil
}

func (m *mockImportGiftCardService) DeleteTransaction(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockImportGiftCardService) FindDeletedDuplicate(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}

func (m *mockImportGiftCardService) RestoreGiftCard(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}

var _ GiftCardServiceInterface = (*mockImportGiftCardService)(nil)

// mockImportMerchantService implements MerchantServiceInterface for import tests.
type mockImportMerchantService struct {
	mock.Mock
}

func (m *mockImportMerchantService) GetMerchantByName(ctx context.Context, name string) (*models.Merchant, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Merchant), args.Error(1)
}

func (m *mockImportMerchantService) CreateMerchant(ctx context.Context, merchant *models.Merchant) error {
	return m.Called(ctx, merchant).Error(0)
}

func (m *mockImportMerchantService) GetMerchantByID(_ context.Context, _ uuid.UUID) (*models.Merchant, error) {
	return nil, nil
}

func (m *mockImportMerchantService) GetAllMerchants(_ context.Context) ([]models.Merchant, error) {
	return nil, nil
}

func (m *mockImportMerchantService) SearchMerchants(_ context.Context, _ string) ([]models.Merchant, error) {
	return nil, nil
}

func (m *mockImportMerchantService) UpdateMerchant(_ context.Context, _ *models.Merchant) error {
	return nil
}

func (m *mockImportMerchantService) DeleteMerchant(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockImportMerchantService) GetMerchantCount(_ context.Context) (int64, error) {
	return 0, nil
}

var _ MerchantServiceInterface = (*mockImportMerchantService)(nil)

// ============================================================================
// HELPER: create ImportService with mocks
// ============================================================================

type importTestSetup struct {
	service         ImportServiceInterface
	cardService     *mockImportCardService
	voucherService  *mockImportVoucherService
	giftCardService *mockImportGiftCardService
	merchantService *mockImportMerchantService
}

func newImportTestSetup() *importTestSetup {
	cs := new(mockImportCardService)
	vs := new(mockImportVoucherService)
	gs := new(mockImportGiftCardService)
	ms := new(mockImportMerchantService)
	return &importTestSetup{
		service:         NewImportService(cs, vs, gs, ms),
		cardService:     cs,
		voucherService:  vs,
		giftCardService: gs,
		merchantService: ms,
	}
}

// setupMerchantResolution configures the merchant mock to find an existing merchant or create a new one.
func (s *importTestSetup) setupMerchantFound(name string, merchantID uuid.UUID) {
	s.merchantService.On("GetMerchantByName", mock.Anything, name).
		Return(&models.Merchant{ID: merchantID, Name: name}, nil)
}

func (s *importTestSetup) setupMerchantNotFoundThenCreate(name string) {
	s.merchantService.On("GetMerchantByName", mock.Anything, name).
		Return(nil, errors.New("not found"))
	s.merchantService.On("CreateMerchant", mock.Anything, mock.MatchedBy(func(m *models.Merchant) bool {
		return m.Name == name
	})).Return(nil)
}

func (s *importTestSetup) setupMerchantCreateFails(name string) {
	s.merchantService.On("GetMerchantByName", mock.Anything, name).
		Return(nil, errors.New("not found"))
	s.merchantService.On("CreateMerchant", mock.Anything, mock.MatchedBy(func(m *models.Merchant) bool {
		return m.Name == name
	})).Return(errors.New("create merchant failed"))
}

// ============================================================================
// PreviewJSON TESTS
// ============================================================================

func TestImportService_PreviewJSON_Success(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()

	data := &ExportData{
		Cards: []ExportCard{
			{CardNumber: "111", MerchantName: "IKEA"},
			{CardNumber: "222", MerchantName: "Migros"},
		},
		Vouchers: []ExportVoucher{
			{Code: "SAVE10", MerchantName: "Coop"},
		},
		GiftCards: []ExportGiftCard{},
	}

	preview, err := s.service.PreviewJSON(ctx, data)

	assert.NoError(t, err)
	assert.NotNil(t, preview)
	assert.Equal(t, 2, preview.Cards)
	assert.Equal(t, 1, preview.Vouchers)
	assert.Equal(t, 0, preview.GiftCards)
}

func TestImportService_PreviewJSON_NilData(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()

	preview, err := s.service.PreviewJSON(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, preview)
	assert.Contains(t, err.Error(), "no data provided")
}

func TestImportService_PreviewJSON_EmptyData(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()

	data := &ExportData{}

	preview, err := s.service.PreviewJSON(ctx, data)

	assert.NoError(t, err)
	assert.NotNil(t, preview)
	assert.Equal(t, 0, preview.Cards)
	assert.Equal(t, 0, preview.Vouchers)
	assert.Equal(t, 0, preview.GiftCards)
}

// ============================================================================
// ImportJSON TESTS
// ============================================================================

func TestImportService_ImportJSON_Success(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("IKEA", merchantID)
	s.setupMerchantFound("Coop", merchantID)
	s.setupMerchantFound("Manor", merchantID)

	s.cardService.On("CreateCard", ctx, mock.AnythingOfType("*models.Card")).Return(nil)
	s.voucherService.On("CreateVoucher", ctx, mock.AnythingOfType("*models.Voucher")).Return(nil)
	s.giftCardService.On("CreateGiftCard", ctx, mock.AnythingOfType("*models.GiftCard")).Return(nil)

	data := &ExportData{
		Cards:     []ExportCard{{CardNumber: "111", MerchantName: "IKEA", Status: "active"}},
		Vouchers:  []ExportVoucher{{Code: "SAVE10", MerchantName: "Coop", Type: "percentage", Value: 10}},
		GiftCards: []ExportGiftCard{{CardNumber: "GC-001", MerchantName: "Manor", InitialBalance: 50, Currency: "CHF"}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.CardsImported)
	assert.Equal(t, 1, result.VouchersImported)
	assert.Equal(t, 1, result.GiftCardsImported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
	s.cardService.AssertExpectations(t)
	s.voucherService.AssertExpectations(t)
	s.giftCardService.AssertExpectations(t)
}

func TestImportService_ImportJSON_NilData(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	result, err := s.service.ImportJSON(ctx, userID, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no data provided")
}

func TestImportService_ImportJSON_EmptyData(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	data := &ExportData{
		Cards:     []ExportCard{},
		Vouchers:  []ExportVoucher{},
		GiftCards: []ExportGiftCard{},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.CardsImported)
	assert.Equal(t, 0, result.VouchersImported)
	assert.Equal(t, 0, result.GiftCardsImported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImportService_ImportJSON_MerchantResolution_ExistingMerchant(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	// Merchant already exists - should be found via GetMerchantByName
	s.setupMerchantFound("IKEA", merchantID)
	s.cardService.On("CreateCard", ctx, mock.MatchedBy(func(c *models.Card) bool {
		return c.MerchantID != nil && *c.MerchantID == merchantID && c.MerchantName == "IKEA"
	})).Return(nil)

	data := &ExportData{
		Cards: []ExportCard{{CardNumber: "111", MerchantName: "IKEA"}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.CardsImported)
	s.merchantService.AssertCalled(t, "GetMerchantByName", mock.Anything, "IKEA")
	s.merchantService.AssertNotCalled(t, "CreateMerchant", mock.Anything, mock.Anything)
}

func TestImportService_ImportJSON_MerchantResolution_NewMerchant(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	// Merchant not found, should be created
	s.setupMerchantNotFoundThenCreate("NewShop")
	s.cardService.On("CreateCard", ctx, mock.AnythingOfType("*models.Card")).Return(nil)

	data := &ExportData{
		Cards: []ExportCard{{CardNumber: "999", MerchantName: "NewShop"}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.CardsImported)
	s.merchantService.AssertCalled(t, "GetMerchantByName", mock.Anything, "NewShop")
	s.merchantService.AssertCalled(t, "CreateMerchant", mock.Anything, mock.MatchedBy(func(m *models.Merchant) bool {
		return m.Name == "NewShop"
	}))
}

func TestImportService_ImportJSON_MerchantResolution_EmptyName(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	// Empty merchant name should result in nil MerchantID (no lookup)
	s.cardService.On("CreateCard", ctx, mock.MatchedBy(func(c *models.Card) bool {
		return c.MerchantID == nil && c.MerchantName == ""
	})).Return(nil)

	data := &ExportData{
		Cards: []ExportCard{{CardNumber: "555", MerchantName: ""}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.CardsImported)
	s.merchantService.AssertNotCalled(t, "GetMerchantByName", mock.Anything, mock.Anything)
}

func TestImportService_ImportJSON_MerchantCreateFails(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	s.setupMerchantCreateFails("BadMerchant")

	data := &ExportData{
		Cards: []ExportCard{{CardNumber: "111", MerchantName: "BadMerchant"}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.CardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "merchant_name", result.Errors[0].Field)
}

func TestImportService_ImportJSON_CardCreateFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("IKEA", merchantID)
	s.cardService.On("CreateCard", ctx, mock.AnythingOfType("*models.Card")).
		Return(errors.New("duplicate card number"))

	data := &ExportData{
		Cards: []ExportCard{{CardNumber: "111", MerchantName: "IKEA"}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err) // Import itself succeeds, individual items may fail
	assert.Equal(t, 0, result.CardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "duplicate card number")
}

func TestImportService_ImportJSON_VoucherCreateFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("Coop", merchantID)
	s.voucherService.On("CreateVoucher", ctx, mock.AnythingOfType("*models.Voucher")).
		Return(errors.New("voucher code exists"))

	data := &ExportData{
		Vouchers: []ExportVoucher{{Code: "DUP-CODE", MerchantName: "Coop", Value: 10}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.VouchersImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "voucher code exists")
}

func TestImportService_ImportJSON_GiftCardCreateFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.AnythingOfType("*models.GiftCard")).
		Return(errors.New("gift card error"))

	data := &ExportData{
		GiftCards: []ExportGiftCard{{CardNumber: "GC-FAIL", MerchantName: "Manor", InitialBalance: 100}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.GiftCardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "gift card error")
}

func TestImportService_ImportJSON_PartialSuccess(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("IKEA", merchantID)
	s.setupMerchantFound("Migros", merchantID)

	// First card succeeds, second fails
	s.cardService.On("CreateCard", ctx, mock.MatchedBy(func(c *models.Card) bool {
		return c.CardNumber == "111"
	})).Return(nil)
	s.cardService.On("CreateCard", ctx, mock.MatchedBy(func(c *models.Card) bool {
		return c.CardNumber == "222"
	})).Return(errors.New("constraint violation"))

	data := &ExportData{
		Cards: []ExportCard{
			{CardNumber: "111", MerchantName: "IKEA"},
			{CardNumber: "222", MerchantName: "Migros"},
		},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.CardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, 2, result.Errors[0].Row)
}

func TestImportService_ImportJSON_DefaultValues(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("IKEA", merchantID)
	s.setupMerchantFound("Coop", merchantID)
	s.setupMerchantFound("Manor", merchantID)

	// Verify default values are applied when fields are empty
	s.cardService.On("CreateCard", ctx, mock.MatchedBy(func(c *models.Card) bool {
		return c.Status == "active" && c.BarcodeType == "CODE128" // defaults when empty
	})).Return(nil)

	s.voucherService.On("CreateVoucher", ctx, mock.MatchedBy(func(v *models.Voucher) bool {
		return v.Type == "percentage" && v.UsageLimitType == "single_use" && v.BarcodeType == "CODE128"
	})).Return(nil)

	s.giftCardService.On("CreateGiftCard", ctx, mock.MatchedBy(func(g *models.GiftCard) bool {
		return g.Currency == "CHF" && g.BarcodeType == "CODE128" // defaults when empty
	})).Return(nil)

	data := &ExportData{
		Cards:     []ExportCard{{CardNumber: "111", MerchantName: "IKEA", Status: ""}},
		Vouchers:  []ExportVoucher{{Code: "V1", MerchantName: "Coop", Type: "", UsageLimitType: "", BarcodeType: "", Value: 10}},
		GiftCards: []ExportGiftCard{{CardNumber: "GC-1", MerchantName: "Manor", Currency: ""}},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.CardsImported)
	assert.Equal(t, 1, result.VouchersImported)
	assert.Equal(t, 1, result.GiftCardsImported)
	s.cardService.AssertExpectations(t)
	s.voucherService.AssertExpectations(t)
	s.giftCardService.AssertExpectations(t)
}

func TestImportService_ImportJSON_GiftCardWithExpiresAt(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.MatchedBy(func(g *models.GiftCard) bool {
		return g.ExpiresAt != nil && g.ExpiresAt.Year() == 2027
	})).Return(nil)

	data := &ExportData{
		GiftCards: []ExportGiftCard{
			{CardNumber: "GC-EXP", MerchantName: "Manor", ExpiresAt: "2027-06-15T00:00:00Z", InitialBalance: 100},
		},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.GiftCardsImported)
	s.giftCardService.AssertExpectations(t)
}

func TestImportService_ImportJSON_GiftCardWithInvalidExpiresAt(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("Manor", merchantID)

	data := &ExportData{
		GiftCards: []ExportGiftCard{
			{CardNumber: "GC-BAD", MerchantName: "Manor", ExpiresAt: "not-a-date", InitialBalance: 50},
		},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.GiftCardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "expires_at", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "invalid expires_at date format")
}

func TestImportService_ImportJSON_GiftCardWithoutExpiresAt(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.MatchedBy(func(g *models.GiftCard) bool {
		return g.ExpiresAt == nil
	})).Return(nil)

	data := &ExportData{
		GiftCards: []ExportGiftCard{
			{CardNumber: "GC-NOEXP", MerchantName: "Manor", ExpiresAt: "", InitialBalance: 25},
		},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.GiftCardsImported)
	s.giftCardService.AssertExpectations(t)
}

func TestImportService_ImportJSON_RowNumbersInErrors(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	s.setupMerchantFound("IKEA", merchantID)
	s.setupMerchantFound("Migros", merchantID)
	s.setupMerchantFound("Coop", merchantID)

	// All three cards fail
	s.cardService.On("CreateCard", ctx, mock.AnythingOfType("*models.Card")).
		Return(errors.New("fail"))

	data := &ExportData{
		Cards: []ExportCard{
			{CardNumber: "AAA", MerchantName: "IKEA"},
			{CardNumber: "BBB", MerchantName: "Migros"},
			{CardNumber: "CCC", MerchantName: "Coop"},
		},
	}

	result, err := s.service.ImportJSON(ctx, userID, data)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.CardsImported)
	assert.Equal(t, 3, result.Skipped)
	assert.Len(t, result.Errors, 3)
	// Row numbers should be 1-based
	assert.Equal(t, 1, result.Errors[0].Row)
	assert.Equal(t, 2, result.Errors[1].Row)
	assert.Equal(t, 3, result.Errors[2].Row)
}

// ============================================================================
// ImportCardsCSV TESTS
// ============================================================================

func TestImportService_ImportCardsCSV_Success(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,program,card_number,barcode_type,status,notes\nIKEA,Family,123456,CODE128,active,My card\nMigros,Cumulus,789012,EAN13,active,Cumulus card"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("IKEA", merchantID)
	s.setupMerchantFound("Migros", merchantID)
	s.cardService.On("CreateCard", ctx, mock.AnythingOfType("*models.Card")).Return(nil)

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.CardsImported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImportService_ImportCardsCSV_EmptyCSV(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	// Only header row, no data rows
	csvData := "merchant_name,program,card_number,barcode_type,status,notes\n"
	reader := strings.NewReader(csvData)

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.CardsImported)
}

func TestImportService_ImportCardsCSV_HeaderOnly(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	// Single row (header) - less than 2 records
	csvData := "merchant_name,program,card_number,barcode_type,status,notes"
	reader := strings.NewReader(csvData)

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.CardsImported)
}

func TestImportService_ImportCardsCSV_MissingCardNumber(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,program,card_number,barcode_type,status,notes\nIKEA,Family,,CODE128,active,My card"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("IKEA", merchantID)

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.CardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "card_number", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "card_number is required")
	assert.Equal(t, 2, result.Errors[0].Row) // Row 2 (1-indexed, after header)
}

func TestImportService_ImportCardsCSV_CreateFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,program,card_number,barcode_type,status,notes\nIKEA,Family,123456,CODE128,active,Test"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("IKEA", merchantID)
	s.cardService.On("CreateCard", ctx, mock.AnythingOfType("*models.Card")).
		Return(errors.New("database error"))

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.CardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "database error")
}

func TestImportService_ImportCardsCSV_MerchantResolutionFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	csvData := "merchant_name,program,card_number,barcode_type,status,notes\nBadMerchant,Loyalty,123456,CODE128,active,Note"
	reader := strings.NewReader(csvData)

	s.setupMerchantCreateFails("BadMerchant")

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.CardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "merchant_name", result.Errors[0].Field)
}

func TestImportService_ImportCardsCSV_DefaultBarcodeAndStatus(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	// barcode_type and status are empty => should default to CODE128 and active
	csvData := "merchant_name,program,card_number,barcode_type,status,notes\nIKEA,Family,123456,,,Test"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("IKEA", merchantID)
	s.cardService.On("CreateCard", ctx, mock.MatchedBy(func(c *models.Card) bool {
		return c.BarcodeType == "CODE128" && c.Status == "active"
	})).Return(nil)

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.CardsImported)
	s.cardService.AssertExpectations(t)
}

func TestImportService_ImportCardsCSV_InvalidCSV(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	// Malformed CSV: inconsistent number of fields (not lazy-quote safe)
	csvData := "merchant_name,card_number\nIKEA,123,extra_field"
	reader := strings.NewReader(csvData)

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	// csv.ReadAll returns an error for inconsistent field counts
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse CSV")
}

func TestImportService_ImportCardsCSV_WhitespaceInHeaders(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	// Headers with extra whitespace should be normalized
	csvData := " Merchant_Name , Program , Card_Number ,barcode_type,status,notes\nIKEA,Family,123456,CODE128,active,Test"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("IKEA", merchantID)
	s.cardService.On("CreateCard", ctx, mock.MatchedBy(func(c *models.Card) bool {
		return c.CardNumber == "123456" && c.MerchantName == "IKEA"
	})).Return(nil)

	result, err := s.service.ImportCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.CardsImported)
}

// ============================================================================
// ImportVouchersCSV TESTS
// ============================================================================

func TestImportService_ImportVouchersCSV_Success(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,code,type,value,description,valid_from,valid_until,usage_limit_type,barcode_type\nCoop,SAVE10,percentage,10,10% off,,,,CODE128"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Coop", merchantID)
	s.voucherService.On("CreateVoucher", ctx, mock.AnythingOfType("*models.Voucher")).Return(nil)

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.VouchersImported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImportService_ImportVouchersCSV_MissingCode(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,code,type,value,description,valid_from,valid_until,usage_limit_type,barcode_type\nCoop,,percentage,10,Discount,,,,CODE128"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Coop", merchantID)

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.VouchersImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "code", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "code is required")
}

func TestImportService_ImportVouchersCSV_CreateFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,code,type,value,description,valid_from,valid_until,usage_limit_type,barcode_type\nCoop,DUP,percentage,10,10% off,,,,CODE128"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Coop", merchantID)
	s.voucherService.On("CreateVoucher", ctx, mock.AnythingOfType("*models.Voucher")).
		Return(errors.New("duplicate code"))

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.VouchersImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Contains(t, result.Errors[0].Message, "duplicate code")
}

func TestImportService_ImportVouchersCSV_HeaderOnly(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	csvData := "merchant_name,code,type,value"
	reader := strings.NewReader(csvData)

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.VouchersImported)
}

func TestImportService_ImportVouchersCSV_DefaultValues(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	// Empty type, usage_limit_type, barcode_type should get defaults
	csvData := "merchant_name,code,type,value,description,valid_from,valid_until,usage_limit_type,barcode_type\nCoop,SAVE20,,20,Discount,,,,"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Coop", merchantID)
	s.voucherService.On("CreateVoucher", ctx, mock.MatchedBy(func(v *models.Voucher) bool {
		return v.Type == "percentage" && v.UsageLimitType == "single_use" && v.BarcodeType == "CODE128"
	})).Return(nil)

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.VouchersImported)
	s.voucherService.AssertExpectations(t)
}

func TestImportService_ImportVouchersCSV_WithValidDates(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,code,type,value,description,valid_from,valid_until,usage_limit_type,barcode_type\nCoop,DATED,percentage,15,15% off,2026-01-01T00:00:00Z,2026-12-31T23:59:59Z,single_use,CODE128"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Coop", merchantID)
	s.voucherService.On("CreateVoucher", ctx, mock.MatchedBy(func(v *models.Voucher) bool {
		return v.ValidFrom.Year() == 2026 && v.ValidUntil.Year() == 2026
	})).Return(nil)

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.VouchersImported)
	s.voucherService.AssertExpectations(t)
}

func TestImportService_ImportVouchersCSV_MerchantResolutionFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	csvData := "merchant_name,code,type,value\nBadShop,CODE1,percentage,10"
	reader := strings.NewReader(csvData)

	s.setupMerchantCreateFails("BadShop")

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.VouchersImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Equal(t, "merchant_name", result.Errors[0].Field)
}

func TestImportService_ImportVouchersCSV_MultipleRows(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,code,type,value\nCoop,V1,percentage,10\nCoop,V2,fixed_amount,5\nCoop,V3,percentage,20"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Coop", merchantID)
	s.voucherService.On("CreateVoucher", ctx, mock.AnythingOfType("*models.Voucher")).Return(nil)

	result, err := s.service.ImportVouchersCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 3, result.VouchersImported)
	assert.Equal(t, 0, result.Skipped)
}

// ============================================================================
// ImportGiftCardsCSV TESTS
// ============================================================================

func TestImportService_ImportGiftCardsCSV_Success(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency,pin,expires_at,notes\nManor,GC-001,100.00,CHF,1234,,Birthday gift"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.AnythingOfType("*models.GiftCard")).Return(nil)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.GiftCardsImported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImportService_ImportGiftCardsCSV_MissingCardNumber(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency,pin,expires_at,notes\nManor,,100.00,CHF,1234,,Gift"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.GiftCardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "card_number", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "card_number is required")
}

func TestImportService_ImportGiftCardsCSV_WithExpiresAt(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency,pin,expires_at,notes\nManor,GC-EXP,200.00,EUR,5678,2027-12-31T23:59:59Z,Expiring card"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.MatchedBy(func(g *models.GiftCard) bool {
		return g.ExpiresAt != nil && g.ExpiresAt.Year() == 2027
	})).Return(nil)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.GiftCardsImported)
	s.giftCardService.AssertExpectations(t)
}

func TestImportService_ImportGiftCardsCSV_InvalidExpiresAt(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency,pin,expires_at,notes\nManor,GC-BAD,50,CHF,,bad-date,Note"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.GiftCardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "expires_at", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "invalid expires_at date format")
}

func TestImportService_ImportGiftCardsCSV_CreateFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency,pin,expires_at,notes\nManor,GC-FAIL,100,CHF,,,Note"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.AnythingOfType("*models.GiftCard")).
		Return(errors.New("db constraint"))

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.GiftCardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Contains(t, result.Errors[0].Message, "db constraint")
}

func TestImportService_ImportGiftCardsCSV_HeaderOnly(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency,pin,expires_at,notes"
	reader := strings.NewReader(csvData)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.GiftCardsImported)
}

func TestImportService_ImportGiftCardsCSV_DefaultCurrency(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	// Empty currency should default to CHF
	csvData := "merchant_name,card_number,initial_balance,currency,pin,expires_at,notes\nManor,GC-DEF,75,,,,A gift"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.MatchedBy(func(g *models.GiftCard) bool {
		return g.Currency == "CHF"
	})).Return(nil)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.GiftCardsImported)
	s.giftCardService.AssertExpectations(t)
}

func TestImportService_ImportGiftCardsCSV_MerchantResolutionFailure(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency\nBadShop,GC-1,100,CHF"
	reader := strings.NewReader(csvData)

	s.setupMerchantCreateFails("BadShop")

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.GiftCardsImported)
	assert.Equal(t, 1, result.Skipped)
	assert.Equal(t, "merchant_name", result.Errors[0].Field)
}

func TestImportService_ImportGiftCardsCSV_MultipleRows(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency\nManor,GC-1,50,CHF\nManor,GC-2,100,EUR\nManor,GC-3,200,CHF"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.AnythingOfType("*models.GiftCard")).Return(nil)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 3, result.GiftCardsImported)
	assert.Equal(t, 0, result.Skipped)
}

func TestImportService_ImportGiftCardsCSV_ParsesInitialBalance(t *testing.T) {
	s := newImportTestSetup()
	ctx := context.Background()
	userID := uuid.New()
	merchantID := uuid.New()

	csvData := "merchant_name,card_number,initial_balance,currency\nManor,GC-BAL,99.95,CHF"
	reader := strings.NewReader(csvData)

	s.setupMerchantFound("Manor", merchantID)
	s.giftCardService.On("CreateGiftCard", ctx, mock.MatchedBy(func(g *models.GiftCard) bool {
		return g.InitialBalance == 99.95
	})).Return(nil)

	result, err := s.service.ImportGiftCardsCSV(ctx, userID, reader)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.GiftCardsImported)
	s.giftCardService.AssertExpectations(t)
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

func TestDefaultIfEmpty_EmptyReturnsDefault(t *testing.T) {
	result := defaultIfEmpty("", "fallback")
	assert.Equal(t, "fallback", result)
}

func TestDefaultIfEmpty_NonEmptyReturnsValue(t *testing.T) {
	result := defaultIfEmpty("custom", "fallback")
	assert.Equal(t, "custom", result)
}

func TestNormalizeCSVHeader_TrimsAndLowercases(t *testing.T) {
	header := []string{" Merchant_Name ", " CARD_NUMBER", "BarcodeType "}
	normalized := normalizeCSVHeader(header)

	assert.Equal(t, "merchant_name", normalized[0])
	assert.Equal(t, "card_number", normalized[1])
	assert.Equal(t, "barcodetype", normalized[2])
}

func TestNormalizeCSVHeader_Empty(t *testing.T) {
	header := []string{}
	normalized := normalizeCSVHeader(header)

	assert.Empty(t, normalized)
}

func TestNormalizeCSVHeader_AlreadyNormalized(t *testing.T) {
	header := []string{"merchant_name", "card_number", "status"}
	normalized := normalizeCSVHeader(header)

	assert.Equal(t, header, normalized)
}

func TestMapCSVRow_MapsHeaderToRecord(t *testing.T) {
	header := []string{"name", "value", "notes"}
	record := []string{"IKEA", "123", "Test note"}

	m := mapCSVRow(header, record)

	assert.Equal(t, "IKEA", m["name"])
	assert.Equal(t, "123", m["value"])
	assert.Equal(t, "Test note", m["notes"])
}

func TestMapCSVRow_RecordShorterThanHeader(t *testing.T) {
	header := []string{"name", "value", "notes"}
	record := []string{"IKEA"} // Only one field

	m := mapCSVRow(header, record)

	assert.Equal(t, "IKEA", m["name"])
	assert.Equal(t, "", m["value"]) // Not set
	assert.Equal(t, "", m["notes"]) // Not set
}

func TestMapCSVRow_EmptyRecord(t *testing.T) {
	header := []string{"name", "value"}
	record := []string{}

	m := mapCSVRow(header, record)

	assert.Equal(t, "", m["name"])
	assert.Equal(t, "", m["value"])
}

func TestMapCSVRow_TrimsWhitespace(t *testing.T) {
	header := []string{"name", "value"}
	record := []string{" IKEA ", " 123 "}

	m := mapCSVRow(header, record)

	assert.Equal(t, "IKEA", m["name"])
	assert.Equal(t, "123", m["value"])
}

func TestParseTimeOrNow_ValidRFC3339(t *testing.T) {
	result := parseTimeOrNow("2026-06-15T10:30:00Z")

	assert.Equal(t, 2026, result.Year())
	assert.Equal(t, 6, int(result.Month()))
	assert.Equal(t, 15, result.Day())
}

func TestParseTimeOrNow_ValidDateOnly(t *testing.T) {
	result := parseTimeOrNow("2026-06-15")

	assert.Equal(t, 2026, result.Year())
	assert.Equal(t, 6, int(result.Month()))
	assert.Equal(t, 15, result.Day())
}

func TestParseTimeOrNow_EmptyString(t *testing.T) {
	result := parseTimeOrNow("")

	// Should return approximately now
	assert.InDelta(t, float64(result.Unix()), float64(result.Unix()), 2)
}

func TestParseTimeOrNow_InvalidFormat(t *testing.T) {
	result := parseTimeOrNow("not-a-date")

	// Should return approximately now (fallback)
	assert.InDelta(t, float64(result.Unix()), float64(result.Unix()), 2)
}

func TestParseTimeOrDefault_ValidRFC3339(t *testing.T) {
	def := parseTimeOrNow("2020-01-01T00:00:00Z")
	result := parseTimeOrDefault("2026-12-25T00:00:00Z", def)

	assert.Equal(t, 2026, result.Year())
	assert.Equal(t, 12, int(result.Month()))
	assert.Equal(t, 25, result.Day())
}

func TestParseTimeOrDefault_ValidDateOnly(t *testing.T) {
	def := parseTimeOrNow("2020-01-01T00:00:00Z")
	result := parseTimeOrDefault("2026-12-25", def)

	assert.Equal(t, 2026, result.Year())
	assert.Equal(t, 12, int(result.Month()))
	assert.Equal(t, 25, result.Day())
}

func TestParseTimeOrDefault_EmptyReturnsDefault(t *testing.T) {
	def := parseTimeOrNow("2020-01-01T00:00:00Z")
	result := parseTimeOrDefault("", def)

	assert.Equal(t, def.Year(), result.Year())
}

func TestParseTimeOrDefault_InvalidReturnsDefault(t *testing.T) {
	def := parseTimeOrNow("2020-01-01T00:00:00Z")
	result := parseTimeOrDefault("garbage", def)

	assert.Equal(t, def.Year(), result.Year())
}

// ============================================================================
// CONSTRUCTOR TEST
// ============================================================================

func TestNewImportService_ReturnsNonNil(t *testing.T) {
	cs := new(mockImportCardService)
	vs := new(mockImportVoucherService)
	gs := new(mockImportGiftCardService)
	ms := new(mockImportMerchantService)

	service := NewImportService(cs, vs, gs, ms)

	assert.NotNil(t, service)
}

// FindByCardNumber stub for mockImportCardService
func (m *mockImportCardService) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}
func (m *mockImportCardService) FindSharedByCardNumber(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) (*models.Card, error) {
	return nil, nil
}

// FindByVoucherCode stub for mockImportVoucherService
func (m *mockImportVoucherService) FindByVoucherCode(_ context.Context, _ string, _ uuid.UUID) (*models.Voucher, error) {
	return nil, nil
}

// FindByCardNumber stub for mockImportGiftCardService
func (m *mockImportGiftCardService) FindByCardNumber(_ context.Context, _ string, _ uuid.UUID) (*models.GiftCard, error) {
	return nil, nil
}
