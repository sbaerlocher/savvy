package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"savvy/internal/models"
)

// MockMerchantRepository is a manual mock for MerchantRepository
type MockMerchantRepository struct {
	mock.Mock
}

func (m *MockMerchantRepository) Create(ctx context.Context, merchant *models.Merchant) error {
	args := m.Called(ctx, merchant)
	return args.Error(0)
}

func (m *MockMerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Merchant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Merchant), args.Error(1)
}

func (m *MockMerchantRepository) GetAll(ctx context.Context) ([]models.Merchant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Merchant), args.Error(1)
}

func (m *MockMerchantRepository) Search(ctx context.Context, query string) ([]models.Merchant, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Merchant), args.Error(1)
}

func (m *MockMerchantRepository) Update(ctx context.Context, merchant *models.Merchant) error {
	args := m.Called(ctx, merchant)
	return args.Error(0)
}

func (m *MockMerchantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMerchantRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// ============================================================================
// TESTS
// ============================================================================

func TestMerchantService_CreateMerchant_Success(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchant := &models.Merchant{
		Name:  "Test Merchant",
		Color: "#FF0000",
	}

	mockRepo.On("Create", ctx, merchant).Return(nil)

	err := service.CreateMerchant(ctx, merchant)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMerchantService_CreateMerchant_ValidationError_MissingName(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchant := &models.Merchant{
		Color: "#FF0000",
	}

	err := service.CreateMerchant(ctx, merchant)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestMerchantService_CreateMerchant_DefaultColor(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchant := &models.Merchant{
		Name: "Test Merchant",
	}

	mockRepo.On("Create", ctx, merchant).Return(nil)

	err := service.CreateMerchant(ctx, merchant)

	assert.NoError(t, err)
	assert.Equal(t, "#3B82F6", merchant.Color) // Default blue color
	mockRepo.AssertExpectations(t)
}

func TestMerchantService_GetMerchantByID_Success(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchantID := uuid.New()
	expectedMerchant := &models.Merchant{
		ID:    merchantID,
		Name:  "Test Merchant",
		Color: "#FF0000",
	}

	mockRepo.On("GetByID", ctx, merchantID).Return(expectedMerchant, nil)

	merchant, err := service.GetMerchantByID(ctx, merchantID)

	assert.NoError(t, err)
	assert.Equal(t, expectedMerchant, merchant)
}

func TestMerchantService_GetAllMerchants_Success(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	expectedMerchants := []models.Merchant{
		{ID: uuid.New(), Name: "Merchant 1", Color: "#FF0000"},
		{ID: uuid.New(), Name: "Merchant 2", Color: "#00FF00"},
	}

	mockRepo.On("GetAll", ctx).Return(expectedMerchants, nil)

	merchants, err := service.GetAllMerchants(ctx)

	assert.NoError(t, err)
	assert.Len(t, merchants, 2)
	assert.Equal(t, expectedMerchants, merchants)
}

func TestMerchantService_SearchMerchants_WithQuery(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	query := "Test"
	expectedMerchants := []models.Merchant{
		{ID: uuid.New(), Name: "Test Merchant", Color: "#FF0000"},
	}

	mockRepo.On("Search", ctx, query).Return(expectedMerchants, nil)

	merchants, err := service.SearchMerchants(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, merchants, 1)
	assert.Equal(t, expectedMerchants, merchants)
}

func TestMerchantService_SearchMerchants_EmptyQuery(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	expectedMerchants := []models.Merchant{
		{ID: uuid.New(), Name: "Merchant 1", Color: "#FF0000"},
		{ID: uuid.New(), Name: "Merchant 2", Color: "#00FF00"},
	}

	mockRepo.On("GetAll", ctx).Return(expectedMerchants, nil)

	merchants, err := service.SearchMerchants(ctx, "")

	assert.NoError(t, err)
	assert.Len(t, merchants, 2)
	mockRepo.AssertNotCalled(t, "Search")
}

func TestMerchantService_UpdateMerchant_Success(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchant := &models.Merchant{
		ID:    uuid.New(),
		Name:  "Updated Merchant",
		Color: "#0000FF",
	}

	mockRepo.On("Update", ctx, merchant).Return(nil)

	err := service.UpdateMerchant(ctx, merchant)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMerchantService_UpdateMerchant_ValidationError_MissingName(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchant := &models.Merchant{
		ID:    uuid.New(),
		Color: "#0000FF",
	}

	err := service.UpdateMerchant(ctx, merchant)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant name is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestMerchantService_UpdateMerchant_ValidationError_MissingColor(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchant := &models.Merchant{
		ID:   uuid.New(),
		Name: "Updated Merchant",
	}

	err := service.UpdateMerchant(ctx, merchant)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant color is required")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestMerchantService_DeleteMerchant_Success(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchantID := uuid.New()

	mockRepo.On("Delete", ctx, merchantID).Return(nil)

	err := service.DeleteMerchant(ctx, merchantID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMerchantService_GetMerchantCount_Success(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	expectedCount := int64(42)

	mockRepo.On("Count", ctx).Return(expectedCount, nil)

	count, err := service.GetMerchantCount(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockRepo.AssertExpectations(t)
}
