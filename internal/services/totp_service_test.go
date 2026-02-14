package services

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"savvy/internal/models"
	"savvy/internal/repository"
)

// ============================================================================
// MOCK
// ============================================================================

// MockTOTPRepository is a manual mock for TOTPRepository.
type MockTOTPRepository struct {
	mock.Mock
}

func (m *MockTOTPRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserTOTP, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserTOTP), args.Error(1)
}

func (m *MockTOTPRepository) Create(ctx context.Context, userTOTP *models.UserTOTP) error {
	args := m.Called(ctx, userTOTP)
	return args.Error(0)
}

func (m *MockTOTPRepository) Update(ctx context.Context, userTOTP *models.UserTOTP) error {
	args := m.Called(ctx, userTOTP)
	return args.Error(0)
}

func (m *MockTOTPRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Ensure MockTOTPRepository implements TOTPRepository
var _ repository.TOTPRepository = (*MockTOTPRepository)(nil)

// ============================================================================
// HELPERS
// ============================================================================

const testEncryptionKey = "test-encryption-key-32-bytes-ok!"
const testIssuer = "SavvyTest"

// newTestTOTPService creates a TOTPService with a test encryption key and mock repository.
func newTestTOTPService(mockRepo *MockTOTPRepository) *TOTPService {
	svc, err := NewTOTPService(mockRepo, testEncryptionKey, testIssuer)
	if err != nil {
		panic("newTestTOTPService: " + err.Error())
	}
	return svc
}

// generateTestTOTPKey creates a real TOTP key and encrypts the secret using the service.
// Returns the raw secret and a UserTOTP model with encrypted secret and hashed backup codes.
func generateTestTOTPKey(t *testing.T, service *TOTPService, userID uuid.UUID, enabled bool) (string, *models.UserTOTP) {
	t.Helper()

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      testIssuer,
		AccountName: "test@example.com",
	})
	assert.NoError(t, err)

	secret := key.Secret()
	encryptedSecret, err := service.encrypt(secret)
	assert.NoError(t, err)

	// Create hashed backup codes
	backupCodes := []string{"ABCD-EFGH", "JKLM-NPQR"}
	hashedCodes, err := hashBackupCodes(backupCodes)
	assert.NoError(t, err)
	hashedCodesJSON, err := json.Marshal(hashedCodes)
	assert.NoError(t, err)

	now := time.Now()
	var enabledAt *time.Time
	if enabled {
		enabledAt = &now
	}

	userTOTP := &models.UserTOTP{
		ID:          uuid.New(),
		UserID:      userID,
		Secret:      encryptedSecret,
		BackupCodes: string(hashedCodesJSON),
		Enabled:     enabled,
		Verified:    enabled,
		EnabledAt:   enabledAt,
	}

	return secret, userTOTP
}

// generateValidCode creates a valid TOTP code for the given secret at the current time.
func generateValidCode(t *testing.T, secret string) string {
	t.Helper()
	code, err := totp.GenerateCode(secret, time.Now())
	assert.NoError(t, err)
	return code
}

// ============================================================================
// TESTS: GenerateSetup
// ============================================================================

func TestTOTPService_GenerateSetup_Success(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// No existing TOTP record
	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.UserTOTP")).Return(nil)

	resp, err := service.GenerateSetup(ctx, userID, "user@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Secret)
	assert.NotEmpty(t, resp.QRCodeURL)
	assert.Len(t, resp.BackupCodes, 10)

	// Verify backup codes match XXXX-XXXX format
	codePattern := regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}$`)
	for _, code := range resp.BackupCodes {
		assert.Regexp(t, codePattern, code)
	}

	// Verify QR code URL contains the issuer
	assert.Contains(t, resp.QRCodeURL, "SavvyTest")
	assert.Contains(t, resp.QRCodeURL, "user@example.com")

	// Verify the created model was stored correctly
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(m *models.UserTOTP) bool {
		return m.UserID == userID && !m.Enabled && !m.Verified && m.Secret != "" && m.BackupCodes != ""
	}))
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_GenerateSetup_AlreadyEnabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	existing := &models.UserTOTP{
		UserID:  userID,
		Enabled: true,
	}
	mockRepo.On("GetByUserID", ctx, userID).Return(existing, nil)

	resp, err := service.GenerateSetup(ctx, userID, "user@example.com")

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTOTPAlreadyEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_GenerateSetup_ReplacesPendingSetup(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Existing unverified/not-enabled TOTP setup
	existing := &models.UserTOTP{
		UserID:  userID,
		Enabled: false,
	}
	mockRepo.On("GetByUserID", ctx, userID).Return(existing, nil)
	mockRepo.On("Delete", ctx, userID).Return(nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.UserTOTP")).Return(nil)

	resp, err := service.GenerateSetup(ctx, userID, "user@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.BackupCodes, 10)

	// Verify the old setup was deleted
	mockRepo.AssertCalled(t, "Delete", ctx, userID)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_GenerateSetup_CreateFails(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.UserTOTP")).Return(gorm.ErrInvalidDB)

	resp, err := service.GenerateSetup(ctx, userID, "user@example.com")

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save TOTP setup")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: VerifyAndEnable
// ============================================================================

func TestTOTPService_VerifyAndEnable_Success(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Create an unenabled TOTP record with a real secret
	secret, userTOTP := generateTestTOTPKey(t, service, userID, false)
	validCode := generateValidCode(t, secret)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(m *models.UserTOTP) bool {
		return m.Enabled && m.Verified && m.EnabledAt != nil
	})).Return(nil)

	err := service.VerifyAndEnable(ctx, userID, validCode)

	assert.NoError(t, err)
	assert.True(t, userTOTP.Enabled)
	assert.True(t, userTOTP.Verified)
	assert.NotNil(t, userTOTP.EnabledAt)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyAndEnable_NotSetup(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	err := service.VerifyAndEnable(ctx, userID, "123456")

	assert.ErrorIs(t, err, ErrTOTPNotSetup)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyAndEnable_AlreadyEnabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	err := service.VerifyAndEnable(ctx, userID, "123456")

	assert.ErrorIs(t, err, ErrTOTPAlreadyEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyAndEnable_InvalidCode(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, false)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	err := service.VerifyAndEnable(ctx, userID, "000000")

	assert.ErrorIs(t, err, ErrTOTPInvalidCode)
	assert.False(t, userTOTP.Enabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyAndEnable_UpdateFails(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	secret, userTOTP := generateTestTOTPKey(t, service, userID, false)
	validCode := generateValidCode(t, secret)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(gorm.ErrInvalidDB)

	err := service.VerifyAndEnable(ctx, userID, validCode)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to enable TOTP")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: Verify
// ============================================================================

func TestTOTPService_Verify_Success(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	secret, userTOTP := generateTestTOTPKey(t, service, userID, true)
	validCode := generateValidCode(t, secret)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	valid, err := service.Verify(ctx, userID, validCode)

	assert.NoError(t, err)
	assert.True(t, valid)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Verify_NotEnabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Case 1: No TOTP record at all
	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	valid, err := service.Verify(ctx, userID, "123456")

	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Verify_NotEnabled_ExistsButDisabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// TOTP record exists but is not enabled
	_, userTOTP := generateTestTOTPKey(t, service, userID, false)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	valid, err := service.Verify(ctx, userID, "123456")

	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Verify_InvalidCode(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	valid, err := service.Verify(ctx, userID, "000000")

	assert.NoError(t, err)
	assert.False(t, valid)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Verify_RepositoryError(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrInvalidDB)

	valid, err := service.Verify(ctx, userID, "123456")

	assert.False(t, valid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get TOTP config")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: VerifyBackupCode
// ============================================================================

func TestTOTPService_VerifyBackupCode_Success(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Create an enabled TOTP record with known backup codes
	backupCode := "ABCD-EFGH"
	hashedCodes, err := hashBackupCodes([]string{backupCode, "JKLM-NPQR"})
	assert.NoError(t, err)
	hashedCodesJSON, err := json.Marshal(hashedCodes)
	assert.NoError(t, err)

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)
	userTOTP.BackupCodes = string(hashedCodesJSON)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(m *models.UserTOTP) bool {
		// After consuming one backup code, only 1 should remain
		var remaining []string
		if err := json.Unmarshal([]byte(m.BackupCodes), &remaining); err != nil {
			return false
		}
		return len(remaining) == 1
	})).Return(nil)

	valid, err := service.VerifyBackupCode(ctx, userID, backupCode)

	assert.NoError(t, err)
	assert.True(t, valid)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyBackupCode_InvalidCode(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	valid, err := service.VerifyBackupCode(ctx, userID, "ZZZZ-ZZZZ")

	assert.NoError(t, err)
	assert.False(t, valid)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyBackupCode_NotEnabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	valid, err := service.VerifyBackupCode(ctx, userID, "ABCD-EFGH")

	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyBackupCode_NotEnabled_ExistsButDisabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, false)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	valid, err := service.VerifyBackupCode(ctx, userID, "ABCD-EFGH")

	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyBackupCode_UpdateFails(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	backupCode := "WXYZ-5678"
	hashedCodes, err := hashBackupCodes([]string{backupCode})
	assert.NoError(t, err)
	hashedCodesJSON, err := json.Marshal(hashedCodes)
	assert.NoError(t, err)

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)
	userTOTP.BackupCodes = string(hashedCodesJSON)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(gorm.ErrInvalidDB)

	valid, err := service.VerifyBackupCode(ctx, userID, backupCode)

	assert.False(t, valid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update backup codes")
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_VerifyBackupCode_ConsumesCode(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Create 3 backup codes
	codes := []string{"AAAA-BBBB", "CCCC-DDDD", "EEEE-FFFF"}
	hashedCodes, err := hashBackupCodes(codes)
	assert.NoError(t, err)
	hashedCodesJSON, err := json.Marshal(hashedCodes)
	assert.NoError(t, err)

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)
	userTOTP.BackupCodes = string(hashedCodesJSON)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	// Capture the updated backup codes to verify the middle code was consumed
	var updatedBackupCodes string
	mockRepo.On("Update", ctx, mock.MatchedBy(func(m *models.UserTOTP) bool {
		updatedBackupCodes = m.BackupCodes
		return true
	})).Return(nil)

	// Consume the second code
	valid, err := service.VerifyBackupCode(ctx, userID, "CCCC-DDDD")

	assert.NoError(t, err)
	assert.True(t, valid)

	// Verify only 2 codes remain
	var remaining []string
	err = json.Unmarshal([]byte(updatedBackupCodes), &remaining)
	assert.NoError(t, err)
	assert.Len(t, remaining, 2)

	// Verify the first and third hashed codes still exist (second was consumed)
	assert.Equal(t, hashedCodes[0], remaining[0])
	assert.Equal(t, hashedCodes[2], remaining[1])
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: Disable
// ============================================================================

func TestTOTPService_Disable_Success(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	secret, userTOTP := generateTestTOTPKey(t, service, userID, true)
	validCode := generateValidCode(t, secret)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Delete", ctx, userID).Return(nil)

	err := service.Disable(ctx, userID, validCode)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Delete", ctx, userID)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Disable_NotEnabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	err := service.Disable(ctx, userID, "123456")

	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Disable_NotEnabled_ExistsButDisabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, false)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	err := service.Disable(ctx, userID, "123456")

	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Disable_InvalidCode(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	err := service.Disable(ctx, userID, "000000")

	assert.ErrorIs(t, err, ErrTOTPInvalidCode)
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_Disable_DeleteFails(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	secret, userTOTP := generateTestTOTPKey(t, service, userID, true)
	validCode := generateValidCode(t, secret)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Delete", ctx, userID).Return(gorm.ErrInvalidDB)

	err := service.Disable(ctx, userID, validCode)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete TOTP config")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: IsEnabled
// ============================================================================

func TestTOTPService_IsEnabled_True(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	enabled, err := service.IsEnabled(ctx, userID)

	assert.NoError(t, err)
	assert.True(t, enabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_IsEnabled_False_NoRecord(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	enabled, err := service.IsEnabled(ctx, userID)

	assert.NoError(t, err)
	assert.False(t, enabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_IsEnabled_False_ExistsButDisabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, false)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	enabled, err := service.IsEnabled(ctx, userID)

	assert.NoError(t, err)
	assert.False(t, enabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_IsEnabled_RepositoryError(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrInvalidDB)

	enabled, err := service.IsEnabled(ctx, userID)

	assert.False(t, enabled)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get TOTP config")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: RegenerateBackupCodes
// ============================================================================

func TestTOTPService_RegenerateBackupCodes_Success(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	secret, userTOTP := generateTestTOTPKey(t, service, userID, true)
	validCode := generateValidCode(t, secret)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(m *models.UserTOTP) bool {
		// Verify backup codes were updated (should be a valid JSON array)
		var hashed []string
		if err := json.Unmarshal([]byte(m.BackupCodes), &hashed); err != nil {
			return false
		}
		return len(hashed) == 10
	})).Return(nil)

	codes, err := service.RegenerateBackupCodes(ctx, userID, validCode)

	assert.NoError(t, err)
	assert.Len(t, codes, 10)

	// Verify format XXXX-XXXX
	codePattern := regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}$`)
	for _, code := range codes {
		assert.Regexp(t, codePattern, code)
	}

	// Verify all codes are unique
	uniqueCodes := make(map[string]struct{})
	for _, code := range codes {
		uniqueCodes[code] = struct{}{}
	}
	assert.Len(t, uniqueCodes, 10)

	mockRepo.AssertExpectations(t)
}

func TestTOTPService_RegenerateBackupCodes_InvalidCode(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, true)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	codes, err := service.RegenerateBackupCodes(ctx, userID, "000000")

	assert.Nil(t, codes)
	assert.ErrorIs(t, err, ErrTOTPInvalidCode)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_RegenerateBackupCodes_NotEnabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	codes, err := service.RegenerateBackupCodes(ctx, userID, "123456")

	assert.Nil(t, codes)
	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_RegenerateBackupCodes_NotEnabled_ExistsButDisabled(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	_, userTOTP := generateTestTOTPKey(t, service, userID, false)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)

	codes, err := service.RegenerateBackupCodes(ctx, userID, "123456")

	assert.Nil(t, codes)
	assert.ErrorIs(t, err, ErrTOTPNotEnabled)
	mockRepo.AssertExpectations(t)
}

func TestTOTPService_RegenerateBackupCodes_UpdateFails(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	secret, userTOTP := generateTestTOTPKey(t, service, userID, true)
	validCode := generateValidCode(t, secret)

	mockRepo.On("GetByUserID", ctx, userID).Return(userTOTP, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(gorm.ErrInvalidDB)

	codes, err := service.RegenerateBackupCodes(ctx, userID, validCode)

	assert.Nil(t, codes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update backup codes")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TESTS: Encrypt / Decrypt
// ============================================================================

func TestTOTPService_EncryptDecrypt(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)

	plaintext := "JBSWY3DPEHPK3PXP"

	encrypted, err := service.encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := service.decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestTOTPService_EncryptDecrypt_DifferentCiphertexts(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)

	plaintext := "JBSWY3DPEHPK3PXP"

	encrypted1, err := service.encrypt(plaintext)
	assert.NoError(t, err)

	encrypted2, err := service.encrypt(plaintext)
	assert.NoError(t, err)

	// Same plaintext should produce different ciphertexts due to random nonce
	assert.NotEqual(t, encrypted1, encrypted2)

	// Both should decrypt to the same value
	decrypted1, err := service.decrypt(encrypted1)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := service.decrypt(encrypted2)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestTOTPService_Decrypt_InvalidBase64(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)

	_, err := service.decrypt("not-valid-base64!!!")
	assert.Error(t, err)
}

func TestTOTPService_Decrypt_TamperedCiphertext(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)

	encrypted, err := service.encrypt("test-secret")
	assert.NoError(t, err)

	// Tamper with the ciphertext (flip a character)
	tampered := []byte(encrypted)
	if tampered[len(tampered)-1] == 'A' {
		tampered[len(tampered)-1] = 'B'
	} else {
		tampered[len(tampered)-1] = 'A'
	}

	_, err = service.decrypt(string(tampered))
	assert.Error(t, err)
}

func TestTOTPService_EncryptDecrypt_EmptyString(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)

	encrypted, err := service.encrypt("")
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := service.decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

func TestTOTPService_EncryptDecrypt_LongString(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service := newTestTOTPService(mockRepo)

	// Test with a long string
	longStr := ""
	for i := 0; i < 1000; i++ {
		longStr += "A"
	}

	encrypted, err := service.encrypt(longStr)
	assert.NoError(t, err)

	decrypted, err := service.decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, longStr, decrypted)
}

// ============================================================================
// TESTS: Encryption Key Padding
// ============================================================================

func TestNewTOTPService_ShortKey(t *testing.T) {
	mockRepo := new(MockTOTPRepository)

	// Short key must be rejected
	service, err := NewTOTPService(mockRepo, "short", testIssuer)
	assert.Nil(t, service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be exactly 32 bytes")
	assert.Contains(t, err.Error(), "got 5")
}

func TestNewTOTPService_LongKey(t *testing.T) {
	mockRepo := new(MockTOTPRepository)

	// Long key must be rejected
	longKey := "this-is-a-very-long-encryption-key-that-exceeds-32-bytes"
	service, err := NewTOTPService(mockRepo, longKey, testIssuer)
	assert.Nil(t, service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be exactly 32 bytes")
}

func TestNewTOTPService_ExactKey(t *testing.T) {
	mockRepo := new(MockTOTPRepository)

	// Exactly 32-byte key should succeed
	exactKey := "12345678901234567890123456789012" // 32 chars
	service, err := NewTOTPService(mockRepo, exactKey, testIssuer)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Len(t, service.encryptionKey, 32)
	assert.Equal(t, []byte(exactKey), service.encryptionKey)
}

func TestNewTOTPService_EmptyKey(t *testing.T) {
	mockRepo := new(MockTOTPRepository)

	// Empty key must be rejected
	service, err := NewTOTPService(mockRepo, "", testIssuer)
	assert.Nil(t, service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be exactly 32 bytes")
	assert.Contains(t, err.Error(), "got 0")
}

// ============================================================================
// TESTS: generateBackupCodes
// ============================================================================

func TestGenerateBackupCodes(t *testing.T) {
	codes, err := generateBackupCodes(10)

	assert.NoError(t, err)
	assert.Len(t, codes, 10)

	codePattern := regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{4}$`)
	for _, code := range codes {
		assert.Regexp(t, codePattern, code, "backup code should match XXXX-XXXX format with valid charset")
		assert.Len(t, code, 9) // 4 + 1 (dash) + 4
	}
}

func TestGenerateBackupCodes_Uniqueness(t *testing.T) {
	codes, err := generateBackupCodes(100)
	assert.NoError(t, err)

	uniqueCodes := make(map[string]struct{})
	for _, code := range codes {
		uniqueCodes[code] = struct{}{}
	}

	// With 30 possible characters and 8 positions, collisions should be virtually impossible
	assert.Len(t, uniqueCodes, 100, "all backup codes should be unique")
}

func TestGenerateBackupCodes_ZeroCodes(t *testing.T) {
	codes, err := generateBackupCodes(0)
	assert.NoError(t, err)
	assert.Empty(t, codes)
}

func TestGenerateBackupCodes_OneCodes(t *testing.T) {
	codes, err := generateBackupCodes(1)
	assert.NoError(t, err)
	assert.Len(t, codes, 1)
}

func TestGenerateBackupCodes_NoAmbiguousChars(t *testing.T) {
	// Generate many codes and verify no ambiguous characters (0, O, 1, I)
	codes, err := generateBackupCodes(50)
	assert.NoError(t, err)

	for _, code := range codes {
		assert.NotContains(t, code, "0", "should not contain 0")
		assert.NotContains(t, code, "O", "should not contain O")
		assert.NotContains(t, code, "1", "should not contain 1")
		assert.NotContains(t, code, "I", "should not contain I")
	}
}

// ============================================================================
// TESTS: hashBackupCodes
// ============================================================================

func TestHashBackupCodes(t *testing.T) {
	codes := []string{"ABCD-EFGH", "JKLM-NPQR", "STUV-WX23"}
	hashed, err := hashBackupCodes(codes)

	assert.NoError(t, err)
	assert.Len(t, hashed, 3)

	// Verify each hash is a valid bcrypt hash and matches the original code
	for i, hash := range hashed {
		assert.NotEqual(t, codes[i], hash, "hash should differ from plaintext")
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(codes[i]))
		assert.NoError(t, err, "bcrypt hash should match original code")
	}
}

func TestHashBackupCodes_Empty(t *testing.T) {
	hashed, err := hashBackupCodes([]string{})
	assert.NoError(t, err)
	assert.Empty(t, hashed)
}

func TestHashBackupCodes_DifferentHashesForSameInput(t *testing.T) {
	codes := []string{"ABCD-EFGH"}

	hashed1, err := hashBackupCodes(codes)
	assert.NoError(t, err)

	hashed2, err := hashBackupCodes(codes)
	assert.NoError(t, err)

	// bcrypt generates different hashes for the same input due to salt
	assert.NotEqual(t, hashed1[0], hashed2[0])

	// Both should still validate against the original code
	err = bcrypt.CompareHashAndPassword([]byte(hashed1[0]), []byte(codes[0]))
	assert.NoError(t, err)
	err = bcrypt.CompareHashAndPassword([]byte(hashed2[0]), []byte(codes[0]))
	assert.NoError(t, err)
}

// ============================================================================
// TESTS: Interface Compliance
// ============================================================================

func TestTOTPService_ImplementsInterface(t *testing.T) {
	mockRepo := new(MockTOTPRepository)
	service, err := NewTOTPService(mockRepo, testEncryptionKey, testIssuer)
	assert.NoError(t, err)

	// Verify TOTPService implements TOTPServiceInterface
	var _ TOTPServiceInterface = service
}
