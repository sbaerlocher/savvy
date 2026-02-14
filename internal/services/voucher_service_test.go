package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// MockVoucherRepository is a manual mock for VoucherRepository
type MockVoucherRepository struct {
	mock.Mock
}

func (m *MockVoucherRepository) Create(ctx context.Context, voucher *models.Voucher) error {
	args := m.Called(ctx, voucher)
	return args.Error(0)
}

func (m *MockVoucherRepository) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*models.Voucher, error) {
	args := m.Called(ctx, id, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Voucher), args.Error(1)
}

func (m *MockVoucherRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *MockVoucherRepository) GetSharedWithUser(ctx context.Context, userID uuid.UUID) ([]models.Voucher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *MockVoucherRepository) Update(ctx context.Context, voucher *models.Voucher) error {
	args := m.Called(ctx, voucher)
	return args.Error(0)
}

func (m *MockVoucherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVoucherRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockVoucherRepository) CanRedeem(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockVoucherRepository) GetAllForUserPaginated(ctx context.Context, userID uuid.UUID, params repository.PaginationParams) (*repository.PaginatedResult[models.Voucher], error) {
	args := m.Called(ctx, userID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PaginatedResult[models.Voucher]), args.Error(1)
}

func (m *MockVoucherRepository) GetExpiringVouchers(_ context.Context, _ int) ([]models.Voucher, error) {
	return nil, nil
}

func (m *MockVoucherRepository) GetVouchersStartingTomorrow(_ context.Context) ([]models.Voucher, error) {
	return nil, nil
}

func (m *MockVoucherRepository) FindByVoucherCode(ctx context.Context, voucherCode string, userID uuid.UUID) (*models.Voucher, error) {
	args := m.Called(ctx, voucherCode, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Voucher), args.Error(1)
}

func (m *MockVoucherRepository) Search(ctx context.Context, userID uuid.UUID, query string) ([]models.Voucher, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

var _ repository.VoucherRepository = (*MockVoucherRepository)(nil)

// ============================================================================
// TESTS
// ============================================================================

func TestVoucherService_CreateVoucher_Success(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	validFrom := time.Now()
	validUntil := validFrom.Add(30 * 24 * time.Hour)

	voucher := &models.Voucher{
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        20.0,
		ValidFrom:    validFrom,
		ValidUntil:   validUntil,
	}

	mockRepo.On("Create", ctx, voucher).Return(nil)

	err := service.CreateVoucher(ctx, voucher)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_CreateVoucher_MissingMerchantName(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	voucher := &models.Voucher{
		UserID: &userID,
		Code:   "SAVE20",
		Type:   "percentage",
		Value:  20.0,
	}

	err := service.CreateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestVoucherService_CreateVoucher_MissingCode(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	voucher := &models.Voucher{
		UserID:       &userID,
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        20.0,
	}

	err := service.CreateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voucher code is required")
}

func TestVoucherService_CreateVoucher_MissingType(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	voucher := &models.Voucher{
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Value:        20.0,
	}

	err := service.CreateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voucher type is required")
}

func TestVoucherService_CreateVoucher_InvalidValue(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	voucher := &models.Voucher{
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        0, // Invalid
	}

	err := service.CreateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voucher value must be positive")
}

func TestVoucherService_CreateVoucher_InvalidDateRange(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	validFrom := time.Now()
	validUntil := validFrom.Add(-1 * time.Hour) // Before validFrom!

	voucher := &models.Voucher{
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        20.0,
		ValidFrom:    validFrom,
		ValidUntil:   validUntil,
	}

	err := service.CreateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "valid_from must be before valid_until")
}

func TestVoucherService_GetVoucher_Success(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucherID := uuid.New()
	userID := uuid.New()
	expectedVoucher := &models.Voucher{
		ID:           voucherID,
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        20.0,
	}

	mockRepo.On("GetByID", ctx, voucherID, []string{"Merchant", "User"}).Return(expectedVoucher, nil)

	voucher, err := service.GetVoucher(ctx, voucherID)

	assert.NoError(t, err)
	assert.Equal(t, expectedVoucher, voucher)
}

func TestVoucherService_GetVoucher_NotFound(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucherID := uuid.New()

	mockRepo.On("GetByID", ctx, voucherID, []string{"Merchant", "User"}).Return(nil, gorm.ErrRecordNotFound)

	voucher, err := service.GetVoucher(ctx, voucherID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, voucher)
}

func TestVoucherService_GetUserVouchers_Success(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	ownedVouchers := []models.Voucher{
		{ID: uuid.New(), Code: "SAVE10", MerchantName: "Merchant 1"},
		{ID: uuid.New(), Code: "SAVE20", MerchantName: "Merchant 2"},
	}

	sharedVouchers := []models.Voucher{
		{ID: uuid.New(), Code: "SAVE30", MerchantName: "Merchant 3"},
	}

	mockRepo.On("GetByUserID", ctx, userID).Return(ownedVouchers, nil)
	mockRepo.On("GetSharedWithUser", ctx, userID).Return(sharedVouchers, nil)

	vouchers, err := service.GetUserVouchers(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, vouchers, 3)
}

func TestVoucherService_UpdateVoucher_Success(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucherID := uuid.New()
	userID := uuid.New()
	voucher := &models.Voucher{
		ID:           voucherID,
		UserID:       &userID,
		Code:         "UPDATED20",
		MerchantName: "Updated Merchant",
		Type:         "percentage",
		Value:        25.0,
	}

	mockRepo.On("Update", ctx, voucher).Return(nil)

	err := service.UpdateVoucher(ctx, voucher)

	assert.NoError(t, err)
}

func TestVoucherService_UpdateVoucher_ValidationError(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	voucher := &models.Voucher{
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        -10, // Invalid
	}

	err := service.UpdateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voucher value must be positive")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestVoucherService_DeleteVoucher_Success(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucherID := uuid.New()

	mockRepo.On("Delete", ctx, voucherID).Return(nil)

	err := service.DeleteVoucher(ctx, voucherID)

	assert.NoError(t, err)
}

func TestVoucherService_CountUserVouchers_Success(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedCount := int64(15)

	mockRepo.On("Count", ctx, userID).Return(expectedCount, nil)

	count, err := service.CountUserVouchers(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
}

// ============================================================================
// ADDITIONAL TESTS FOR ERROR PATH COVERAGE
// ============================================================================

func TestVoucherService_GetUserVouchers_OwnedError(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, repoError)

	vouchers, err := service.GetUserVouchers(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, vouchers)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_GetUserVouchers_SharedError(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	ownedVouchers := []models.Voucher{
		{ID: uuid.New(), Code: "TEST1"},
	}
	repoError := errors.New("shared query failed")

	mockRepo.On("GetByUserID", ctx, userID).Return(ownedVouchers, nil)
	mockRepo.On("GetSharedWithUser", ctx, userID).Return(nil, repoError)

	vouchers, err := service.GetUserVouchers(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, vouchers)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_UpdateVoucher_ValidationMissingMerchant(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucher := &models.Voucher{
		ID:           uuid.New(),
		Code:         "TEST",
		MerchantName: "", // Missing merchant
		Type:         "discount",
		Value:        10.0,
	}

	err := service.UpdateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestVoucherService_UpdateVoucher_ValidationMissingCode(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucher := &models.Voucher{
		ID:           uuid.New(),
		Code:         "", // Missing code
		MerchantName: "Test",
		Type:         "discount",
		Value:        10.0,
	}

	err := service.UpdateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voucher code is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestVoucherService_UpdateVoucher_ValidationMissingType(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucher := &models.Voucher{
		ID:           uuid.New(),
		Code:         "TEST",
		MerchantName: "Test",
		Type:         "", // Missing type
		Value:        10.0,
	}

	err := service.UpdateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voucher type is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestVoucherService_UpdateVoucher_ValidationNegativeValue(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucher := &models.Voucher{
		ID:           uuid.New(),
		Code:         "TEST",
		MerchantName: "Test",
		Type:         "discount",
		Value:        -5.0, // Negative value
	}

	err := service.UpdateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voucher value must be positive")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestVoucherService_UpdateVoucher_ValidationInvalidDateRange(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	voucher := &models.Voucher{
		ID:           uuid.New(),
		Code:         "TEST",
		MerchantName: "Test",
		Type:         "discount",
		Value:        10.0,
		ValidFrom:    time.Now().Add(24 * time.Hour),
		ValidUntil:   time.Now(), // ValidUntil before ValidFrom
	}

	err := service.UpdateVoucher(ctx, voucher)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "valid_from must be before valid_until")
	mockRepo.AssertNotCalled(t, "Update")
}

// ============================================================================
// CHECK DUPLICATE TESTS
// ============================================================================

func TestVoucherService_CheckDuplicate_DuplicateFound(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	existingVoucher := &models.Voucher{
		ID:           uuid.New(),
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        20.0,
	}

	mockRepo.On("FindByVoucherCode", ctx, "SAVE20", userID).Return(existingVoucher, nil)

	result, err := service.CheckDuplicate(ctx, "SAVE20", userID, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingVoucher.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_CheckDuplicate_NoDuplicate(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("FindByVoucherCode", ctx, "NONEXISTENT", userID).Return(nil, nil)

	result, err := service.CheckDuplicate(ctx, "NONEXISTENT", userID, nil)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_CheckDuplicate_ExcludeSameVoucher(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	voucherID := uuid.New()
	existingVoucher := &models.Voucher{
		ID:           voucherID,
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        20.0,
	}

	mockRepo.On("FindByVoucherCode", ctx, "SAVE20", userID).Return(existingVoucher, nil)

	// Exclude the same voucher (update case) - should not be considered a duplicate
	result, err := service.CheckDuplicate(ctx, "SAVE20", userID, &voucherID)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_CheckDuplicate_ExcludeDifferentVoucher(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	existingVoucherID := uuid.New()
	differentVoucherID := uuid.New()
	existingVoucher := &models.Voucher{
		ID:           existingVoucherID,
		UserID:       &userID,
		Code:         "SAVE20",
		MerchantName: "Test Merchant",
		Type:         "percentage",
		Value:        20.0,
	}

	mockRepo.On("FindByVoucherCode", ctx, "SAVE20", userID).Return(existingVoucher, nil)

	// Exclude a different voucher - should still be considered a duplicate
	result, err := service.CheckDuplicate(ctx, "SAVE20", userID, &differentVoucherID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingVoucherID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_CheckDuplicate_RepositoryError(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	mockRepo.On("FindByVoucherCode", ctx, "SAVE20", userID).Return(nil, repoError)

	result, err := service.CheckDuplicate(ctx, "SAVE20", userID, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "check duplicate")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// GET USER VOUCHERS PAGINATED TESTS
// ============================================================================

func TestVoucherService_GetUserVouchersPaginated_Success(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedResult := &repository.PaginatedResult[models.Voucher]{
		Items: []models.Voucher{
			{ID: uuid.New(), Code: "SAVE10", MerchantName: "Merchant 1", Type: "percentage", Value: 10.0},
			{ID: uuid.New(), Code: "SAVE20", MerchantName: "Merchant 2", Type: "fixed", Value: 20.0},
		},
		Total:      8,
		Page:       1,
		PerPage:    2,
		TotalPages: 4,
	}

	params := repository.PaginationParams{Page: 1, PerPage: 2}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(expectedResult, nil)

	result, err := service.GetUserVouchersPaginated(ctx, userID, 1, 2)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(8), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.PerPage)
	assert.Equal(t, 4, result.TotalPages)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_GetUserVouchersPaginated_EmptyResult(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedResult := &repository.PaginatedResult[models.Voucher]{
		Items:      []models.Voucher{},
		Total:      0,
		Page:       1,
		PerPage:    10,
		TotalPages: 0,
	}

	params := repository.PaginationParams{Page: 1, PerPage: 10}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(expectedResult, nil)

	result, err := service.GetUserVouchersPaginated(ctx, userID, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
	mockRepo.AssertExpectations(t)
}

func TestVoucherService_GetUserVouchersPaginated_RepositoryError(t *testing.T) {
	mockRepo := new(MockVoucherRepository)
	service := NewVoucherService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	repoError := errors.New("database error")

	params := repository.PaginationParams{Page: 1, PerPage: 10}
	mockRepo.On("GetAllForUserPaginated", ctx, userID, params).Return(nil, repoError)

	result, err := service.GetUserVouchersPaginated(ctx, userID, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get paginated vouchers")
	mockRepo.AssertExpectations(t)
}
