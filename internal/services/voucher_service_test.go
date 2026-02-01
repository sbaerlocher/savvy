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
