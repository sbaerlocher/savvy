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

func (m *MockMerchantRepository) GetByName(ctx context.Context, name string) (*models.Merchant, error) {
	args := m.Called(ctx, name)
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

	// Mock GetByName to return nil (merchant doesn't exist)
	mockRepo.On("GetByName", ctx, "Test Merchant").Return(nil, nil)
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
	mockRepo.AssertNotCalled(t, "GetByName")
}

func TestMerchantService_CreateMerchant_DuplicateName(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	existingMerchant := &models.Merchant{
		ID:    uuid.New(),
		Name:  "Existing Merchant",
		Color: "#FF0000",
	}

	newMerchant := &models.Merchant{
		Name:  "Existing Merchant",
		Color: "#00FF00",
	}

	// Mock GetByName to return existing merchant
	mockRepo.On("GetByName", ctx, "Existing Merchant").Return(existingMerchant, nil)

	err := service.CreateMerchant(ctx, newMerchant)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant with this name already exists")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestMerchantService_CreateMerchant_DefaultColor(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchant := &models.Merchant{
		Name: "Test Merchant",
	}

	// Mock GetByName to return nil (merchant doesn't exist)
	mockRepo.On("GetByName", ctx, "Test Merchant").Return(nil, nil)
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

	// Mock GetByName to return nil (no other merchant with same name exists)
	mockRepo.On("GetByName", ctx, "Updated Merchant").Return(nil, nil)
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
	mockRepo.AssertNotCalled(t, "GetByName")
}

func TestMerchantService_UpdateMerchant_DuplicateName(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	merchantID := uuid.New()
	otherMerchantID := uuid.New()

	existingMerchant := &models.Merchant{
		ID:    otherMerchantID,
		Name:  "Existing Merchant",
		Color: "#FF0000",
	}

	merchantToUpdate := &models.Merchant{
		ID:    merchantID,
		Name:  "Existing Merchant", // Same name as another merchant
		Color: "#00FF00",
	}

	// Mock GetByName to return the other merchant with the same name
	mockRepo.On("GetByName", ctx, "Existing Merchant").Return(existingMerchant, nil)

	err := service.UpdateMerchant(ctx, merchantToUpdate)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "merchant with this name already exists")
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

func TestMerchantService_GetMerchantByName_Success(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	expectedMerchant := &models.Merchant{
		ID:   uuid.New(),
		Name: "Test Merchant",
	}

	mockRepo.On("GetByName", ctx, "Test Merchant").Return(expectedMerchant, nil)

	merchant, err := service.GetMerchantByName(ctx, "Test Merchant")

	assert.NoError(t, err)
	assert.NotNil(t, merchant)
	assert.Equal(t, "Test Merchant", merchant.Name)
	mockRepo.AssertExpectations(t)
}

func TestMerchantService_GetMerchantByName_NotFound(t *testing.T) {
	mockRepo := new(MockMerchantRepository)
	service := NewMerchantService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByName", ctx, "Nonexistent").Return((*models.Merchant)(nil), assert.AnError)

	merchant, err := service.GetMerchantByName(ctx, "Nonexistent")

	assert.Error(t, err)
	assert.Nil(t, merchant)
	mockRepo.AssertExpectations(t)
}
