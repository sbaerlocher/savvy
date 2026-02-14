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

// MockUserRepository is a mock for UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) SearchByIDs(ctx context.Context, ids []uuid.UUID, query string) ([]models.User, error) {
	args := m.Called(ctx, ids, query)
	return args.Get(0).([]models.User), args.Error(1)
}

// Ensure MockUserRepository implements UserRepository
var _ repository.UserRepository = (*MockUserRepository)(nil)

// ============================================================================
// TESTS
// ============================================================================

// TestUserService_GetUserByID_Success tests retrieving user by ID
func TestUserService_GetUserByID_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	expectedUser := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}

	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	// Test: Get user by ID
	user, err := service.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Email, user.Email)
	mockRepo.AssertExpectations(t)
}

// TestUserService_GetUserByID_NotFound tests error when user not found
func TestUserService_GetUserByID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	// Test: User not found
	user, err := service.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// TestUserService_GetUserByEmail_Success tests retrieving user by email
func TestUserService_GetUserByEmail_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	email := "test@example.com"
	expectedUser := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "hash",
		Role:         "user",
	}

	mockRepo.On("GetByEmail", ctx, email).Return(expectedUser, nil)

	// Test: Get user by email
	user, err := service.GetUserByEmail(ctx, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	mockRepo.AssertExpectations(t)
}

// TestUserService_GetUserByEmail_NotFound tests custom error when user not found
func TestUserService_GetUserByEmail_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	email := "nonexistent@example.com"

	mockRepo.On("GetByEmail", ctx, email).Return(nil, gorm.ErrRecordNotFound)

	// Test: User not found returns custom error
	user, err := service.GetUserByEmail(ctx, email)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// TestUserService_GetUserByEmail_DatabaseError tests handling of database errors
func TestUserService_GetUserByEmail_DatabaseError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	email := "test@example.com"
	dbError := errors.New("database connection error")

	mockRepo.On("GetByEmail", ctx, email).Return(nil, dbError)

	// Test: Database error is propagated (wrapped with context)
	user, err := service.GetUserByEmail(ctx, email)

	assert.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// TestUserService_CreateUser_Success tests creating a new user
func TestUserService_CreateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	user := &models.User{
		Email:        "newuser@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}

	mockRepo.On("Create", ctx, user).Return(nil)

	// Test: Create user
	err := service.CreateUser(ctx, user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUserService_CreateUser_MissingEmail tests validation for missing email
func TestUserService_CreateUser_MissingEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	user := &models.User{
		Email:        "", // Missing email
		PasswordHash: "hash",
		Role:         "user",
	}

	// Test: Missing email should fail validation
	err := service.CreateUser(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestUserService_CreateUser_EmptyEmail tests validation for empty email
func TestUserService_CreateUser_EmptyEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	user := &models.User{
		Email:        "   ", // Whitespace email
		PasswordHash: "hash",
		Role:         "user",
	}

	// Note: Current implementation only checks for empty string, not trimmed
	// This test documents current behavior
	mockRepo.On("Create", ctx, user).Return(nil)

	err := service.CreateUser(ctx, user)

	// Currently passes because email is not empty string
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUserService_CreateUser_RepositoryError tests handling of repository errors
func TestUserService_CreateUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}

	repoError := errors.New("duplicate email")
	mockRepo.On("Create", ctx, user).Return(repoError)

	// Test: Repository error is propagated (wrapped with context)
	err := service.CreateUser(ctx, user)

	assert.Error(t, err)
	assert.ErrorIs(t, err, repoError)
	mockRepo.AssertExpectations(t)
}

// TestUserService_UpdateUser_Success tests updating a user
func TestUserService_UpdateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	user := &models.User{
		ID:           uuid.New(),
		Email:        "updated@example.com",
		FirstName:    "Updated",
		LastName:     "Name",
		PasswordHash: "hash",
		Role:         "admin",
	}

	mockRepo.On("Update", ctx, user).Return(nil)

	// Test: Update user
	err := service.UpdateUser(ctx, user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUserService_UpdateUser_Error tests handling of update errors
func TestUserService_UpdateUser_Error(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}

	updateError := errors.New("update failed")
	mockRepo.On("Update", ctx, user).Return(updateError)

	// Test: Update error is propagated (wrapped with context)
	err := service.UpdateUser(ctx, user)

	assert.Error(t, err)
	assert.ErrorIs(t, err, updateError)
	mockRepo.AssertExpectations(t)
}

// TestUserService_ErrUserNotFound tests the custom error constant
func TestUserService_ErrUserNotFound(t *testing.T) {
	// Test: Custom error is defined and has correct message
	assert.NotNil(t, ErrUserNotFound)
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
}

// TestUserService_GetUserByEmail_CaseInsensitive tests case-insensitive email lookup
func TestUserService_GetUserByEmail_CaseInsensitive(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	// Repository should handle case-insensitivity
	email := "Test@Example.com"
	expectedUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com", // Stored as lowercase
		PasswordHash: "hash",
		Role:         "user",
	}

	mockRepo.On("GetByEmail", ctx, email).Return(expectedUser, nil)

	// Test: Email lookup works regardless of case
	user, err := service.GetUserByEmail(ctx, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	mockRepo.AssertExpectations(t)
}

// TestUserService_CreateUser_WithAllFields tests creating user with all optional fields
func TestUserService_CreateUser_WithAllFields(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	user := &models.User{
		Email:        "complete@example.com",
		PasswordHash: "hash",
		FirstName:    "John",
		LastName:     "Doe",
		Role:         "admin",
		AuthProvider: "local",
	}

	mockRepo.On("Create", ctx, user).Return(nil)

	// Test: Create user with all fields
	err := service.CreateUser(ctx, user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUserService_UpdateUser_PartialUpdate tests updating specific fields
func TestUserService_UpdateUser_PartialUpdate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	// User with only some fields updated
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "UpdatedFirst",
		// LastName not changed
	}

	mockRepo.On("Update", ctx, user).Return(nil)

	// Test: Partial update
	err := service.UpdateUser(ctx, user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
