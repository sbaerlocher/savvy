// Package services contains business logic.
package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// ErrTOTPAlreadyEnabled is returned when TOTP is already enabled for the user
	ErrTOTPAlreadyEnabled = errors.New("TOTP is already enabled")
	// ErrTOTPNotEnabled is returned when TOTP is not enabled for the user
	ErrTOTPNotEnabled = errors.New("TOTP is not enabled")
	// ErrTOTPInvalidCode is returned when the TOTP code is invalid
	ErrTOTPInvalidCode = errors.New("invalid TOTP code")
	// ErrTOTPNotSetup is returned when TOTP setup has not been initiated
	ErrTOTPNotSetup = errors.New("TOTP setup not initiated")
)

// TOTPSetupResponse contains the data needed for TOTP setup
type TOTPSetupResponse struct {
	Secret      string   `json:"secret"` // #nosec G117 -- struct field name, not a hardcoded secret
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// TOTPServiceInterface defines the TOTP service interface.
type TOTPServiceInterface interface {
	// GenerateSetup creates a new TOTP secret and backup codes for a user
	GenerateSetup(ctx context.Context, userID uuid.UUID, email string) (*TOTPSetupResponse, error)

	// VerifyAndEnable verifies a TOTP code and enables 2FA for the user
	VerifyAndEnable(ctx context.Context, userID uuid.UUID, code string) error

	// Verify verifies a TOTP code for an already-enabled user
	Verify(ctx context.Context, userID uuid.UUID, code string) (bool, error)

	// VerifyBackupCode verifies and consumes a backup code
	VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)

	// Disable disables 2FA for a user (requires valid TOTP code)
	Disable(ctx context.Context, userID uuid.UUID, code string) error

	// IsEnabled checks if TOTP is enabled for a user
	IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error)

	// RegenerateBackupCodes generates new backup codes for a user
	RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error)
}

// TOTPService implements TOTPServiceInterface.
type TOTPService struct {
	totpRepo      repository.TOTPRepository
	encryptionKey []byte // 32 bytes for AES-256
	issuer        string
}

// NewTOTPService creates a new TOTP service.
// The encryptionKey must be exactly 32 bytes for AES-256-GCM.
func NewTOTPService(totpRepo repository.TOTPRepository, encryptionKey string, issuer string) (*TOTPService, error) {
	keyBytes := []byte(encryptionKey)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("TOTP_ENCRYPTION_KEY must be exactly 32 bytes, got %d", len(keyBytes))
	}
	return &TOTPService{
		totpRepo:      totpRepo,
		encryptionKey: keyBytes,
		issuer:        issuer,
	}, nil
}

// GenerateSetup creates a new TOTP secret and backup codes for a user.
func (s *TOTPService) GenerateSetup(ctx context.Context, userID uuid.UUID, email string) (*TOTPSetupResponse, error) {
	// Check if TOTP is already enabled
	existing, err := s.totpRepo.GetByUserID(ctx, userID)
	if err == nil && existing != nil && existing.Enabled {
		return nil, ErrTOTPAlreadyEnabled
	}

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate 10 backup codes
	backupCodes, err := generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Hash backup codes with bcrypt
	hashedCodes, err := hashBackupCodes(backupCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to hash backup codes: %w", err)
	}

	// Encrypt the TOTP secret
	encryptedSecret, err := s.encrypt(key.Secret())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}

	// Serialize hashed codes to JSON
	hashedCodesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize backup codes: %w", err)
	}

	// Delete any existing unverified setup
	if existing != nil && !existing.Enabled {
		if err := s.totpRepo.Delete(ctx, userID); err != nil {
			slog.Warn("Failed to delete existing TOTP setup", "user_id", userID, "error", err)
		}
	}

	// Save to database (not yet enabled/verified)
	userTOTP := &models.UserTOTP{
		UserID:      userID,
		Secret:      encryptedSecret,
		BackupCodes: string(hashedCodesJSON),
		Enabled:     false,
		Verified:    false,
	}

	if err := s.totpRepo.Create(ctx, userTOTP); err != nil {
		return nil, fmt.Errorf("failed to save TOTP setup: %w", err)
	}

	return &TOTPSetupResponse{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyAndEnable verifies a TOTP code and enables 2FA for the user.
func (s *TOTPService) VerifyAndEnable(ctx context.Context, userID uuid.UUID, code string) error {
	userTOTP, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTOTPNotSetup
		}
		return fmt.Errorf("failed to get TOTP config: %w", err)
	}

	if userTOTP.Enabled {
		return ErrTOTPAlreadyEnabled
	}

	// Decrypt the secret
	secret, err := s.decrypt(userTOTP.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	// Validate the code
	valid := totp.Validate(code, secret)
	if !valid {
		return ErrTOTPInvalidCode
	}

	// Enable TOTP
	now := time.Now()
	userTOTP.Enabled = true
	userTOTP.Verified = true
	userTOTP.EnabledAt = &now

	if err := s.totpRepo.Update(ctx, userTOTP); err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	slog.Info("TOTP enabled for user", "user_id", userID)
	return nil
}

// Verify verifies a TOTP code for an already-enabled user.
func (s *TOTPService) Verify(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	userTOTP, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrTOTPNotEnabled
		}
		return false, fmt.Errorf("failed to get TOTP config: %w", err)
	}

	if !userTOTP.Enabled {
		return false, ErrTOTPNotEnabled
	}

	// Decrypt the secret
	secret, err := s.decrypt(userTOTP.Secret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	return totp.Validate(code, secret), nil
}

// VerifyBackupCode verifies and consumes a backup code.
func (s *TOTPService) VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	userTOTP, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrTOTPNotEnabled
		}
		return false, fmt.Errorf("failed to get TOTP config: %w", err)
	}

	if !userTOTP.Enabled {
		return false, ErrTOTPNotEnabled
	}

	// Deserialize hashed backup codes
	var hashedCodes []string
	if err := json.Unmarshal([]byte(userTOTP.BackupCodes), &hashedCodes); err != nil {
		return false, fmt.Errorf("failed to deserialize backup codes: %w", err)
	}

	// Try each hashed code
	for i, hashedCode := range hashedCodes {
		if bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(code)) == nil {
			// Remove the used code
			hashedCodes = append(hashedCodes[:i], hashedCodes[i+1:]...)
			updatedJSON, err := json.Marshal(hashedCodes)
			if err != nil {
				return false, fmt.Errorf("failed to serialize updated backup codes: %w", err)
			}
			userTOTP.BackupCodes = string(updatedJSON)
			if err := s.totpRepo.Update(ctx, userTOTP); err != nil {
				return false, fmt.Errorf("failed to update backup codes: %w", err)
			}
			slog.Info("Backup code used", "user_id", userID, "remaining", len(hashedCodes))
			return true, nil
		}
	}

	return false, nil
}

// Disable disables 2FA for a user after verifying the TOTP code.
func (s *TOTPService) Disable(ctx context.Context, userID uuid.UUID, code string) error {
	userTOTP, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTOTPNotEnabled
		}
		return fmt.Errorf("failed to get TOTP config: %w", err)
	}

	if !userTOTP.Enabled {
		return ErrTOTPNotEnabled
	}

	// Verify the code before disabling
	secret, err := s.decrypt(userTOTP.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	if !totp.Validate(code, secret) {
		return ErrTOTPInvalidCode
	}

	// Delete the TOTP config
	if err := s.totpRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete TOTP config: %w", err)
	}

	slog.Info("TOTP disabled for user", "user_id", userID)
	return nil
}

// IsEnabled checks if TOTP is enabled for a user.
func (s *TOTPService) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	userTOTP, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get TOTP config: %w", err)
	}
	return userTOTP.Enabled, nil
}

// RegenerateBackupCodes generates new backup codes for a user after verifying the TOTP code.
func (s *TOTPService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	userTOTP, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTOTPNotEnabled
		}
		return nil, fmt.Errorf("failed to get TOTP config: %w", err)
	}

	if !userTOTP.Enabled {
		return nil, ErrTOTPNotEnabled
	}

	// Verify the TOTP code
	secret, err := s.decrypt(userTOTP.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	if !totp.Validate(code, secret) {
		return nil, ErrTOTPInvalidCode
	}

	// Generate new backup codes
	backupCodes, err := generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	hashedCodes, err := hashBackupCodes(backupCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to hash backup codes: %w", err)
	}

	hashedCodesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize backup codes: %w", err)
	}

	userTOTP.BackupCodes = string(hashedCodesJSON)
	if err := s.totpRepo.Update(ctx, userTOTP); err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	slog.Info("Backup codes regenerated", "user_id", userID)
	return backupCodes, nil
}

// encrypt encrypts plaintext using AES-256-GCM
func (s *TOTPService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts ciphertext using AES-256-GCM
func (s *TOTPService) decrypt(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// generateBackupCodes generates n random 8-character alphanumeric codes
func generateBackupCodes(n int) ([]string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude ambiguous chars (0,O,1,I)
	codes := make([]string, n)

	for i := 0; i < n; i++ {
		code := make([]byte, 8)
		for j := range code {
			b := make([]byte, 1)
			if _, err := rand.Read(b); err != nil {
				return nil, err
			}
			code[j] = charset[int(b[0])%len(charset)]
		}
		// Format as XXXX-XXXX
		codes[i] = string(code[:4]) + "-" + string(code[4:])
	}

	return codes, nil
}

// hashBackupCodes hashes backup codes with bcrypt
func hashBackupCodes(codes []string) ([]string, error) {
	hashed := make([]string, len(codes))
	for i, code := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashed[i] = string(hash)
	}
	return hashed, nil
}
