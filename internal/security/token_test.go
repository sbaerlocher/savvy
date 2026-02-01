package security

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetValidityWindow(t *testing.T) {
	now := time.Now()
	expectedMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	t.Run("Returns midnight + duration", func(t *testing.T) {
		duration := 7 * 24 * time.Hour
		result := getValidityWindow(duration)

		expected := expectedMidnight.Add(duration)
		assert.Equal(t, expected.Unix(), result.Unix())
	})

	t.Run("Returns same time for multiple calls in same day", func(t *testing.T) {
		duration := 7 * 24 * time.Hour

		result1 := getValidityWindow(duration)
		time.Sleep(10 * time.Millisecond) // Small delay
		result2 := getValidityWindow(duration)

		// Both should return the exact same timestamp (deterministic)
		assert.Equal(t, result1.Unix(), result2.Unix())
	})
}

func TestDeterministicTokenGeneration(t *testing.T) {
	// Initialize security package
	Init("test-secret-key-for-deterministic-tokens")

	resourceID := uuid.New()
	userID := uuid.New()
	resourceType := "card"
	duration := 7 * 24 * time.Hour

	t.Run("Same parameters generate same token on same day", func(t *testing.T) {
		token1, err1 := GenerateBarcodeToken(resourceID, resourceType, userID, duration)
		assert.NoError(t, err1)

		// Small delay
		time.Sleep(10 * time.Millisecond)

		token2, err2 := GenerateBarcodeToken(resourceID, resourceType, userID, duration)
		assert.NoError(t, err2)

		// Tokens should be identical (deterministic)
		assert.Equal(t, token1, token2, "Tokens generated on same day should be identical")
	})

	t.Run("Different resources generate different tokens", func(t *testing.T) {
		resourceID1 := uuid.New()
		resourceID2 := uuid.New()

		token1, err1 := GenerateBarcodeToken(resourceID1, resourceType, userID, duration)
		assert.NoError(t, err1)

		token2, err2 := GenerateBarcodeToken(resourceID2, resourceType, userID, duration)
		assert.NoError(t, err2)

		assert.NotEqual(t, token1, token2, "Different resource IDs should generate different tokens")
	})

	t.Run("Different users generate different tokens", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()

		token1, err1 := GenerateBarcodeToken(resourceID, resourceType, userID1, duration)
		assert.NoError(t, err1)

		token2, err2 := GenerateBarcodeToken(resourceID, resourceType, userID2, duration)
		assert.NoError(t, err2)

		assert.NotEqual(t, token1, token2, "Different user IDs should generate different tokens")
	})

	t.Run("Generated token validates successfully", func(t *testing.T) {
		token, err := GenerateBarcodeToken(resourceID, resourceType, userID, duration)
		assert.NoError(t, err)

		claims, err := ValidateBarcodeToken(token)
		assert.NoError(t, err)
		assert.Equal(t, resourceID, claims.ResourceID)
		assert.Equal(t, resourceType, claims.ResourceType)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("Token expiration is set to midnight + duration", func(t *testing.T) {
		token, err := GenerateBarcodeToken(resourceID, resourceType, userID, duration)
		assert.NoError(t, err)

		claims, err := ValidateBarcodeToken(token)
		assert.NoError(t, err)

		now := time.Now()
		expectedMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		expectedExpiry := expectedMidnight.Add(duration)

		assert.Equal(t, expectedExpiry.Unix(), claims.ExpiresAt)
	})
}

func TestBarcodeTokenValidation(t *testing.T) {
	Init("test-secret-key-for-validation")

	resourceID := uuid.New()
	userID := uuid.New()

	t.Run("Valid token passes validation", func(t *testing.T) {
		token, err := GenerateBarcodeToken(resourceID, "card", userID, 7*24*time.Hour)
		assert.NoError(t, err)

		claims, err := ValidateBarcodeToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
	})

	t.Run("Expired token returns error", func(t *testing.T) {
		// Generate token with negative duration (already expired)
		token, err := GenerateBarcodeToken(resourceID, "card", userID, -1*time.Hour)
		assert.NoError(t, err)

		_, err = ValidateBarcodeToken(token)
		assert.ErrorIs(t, err, ErrTokenExpired)
	})

	t.Run("Invalid token format returns error", func(t *testing.T) {
		_, err := ValidateBarcodeToken("invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("Tampered token returns error", func(t *testing.T) {
		token, err := GenerateBarcodeToken(resourceID, "card", userID, 7*24*time.Hour)
		assert.NoError(t, err)

		// Tamper with token
		tamperedToken := token[:len(token)-5] + "XXXXX"

		_, err = ValidateBarcodeToken(tamperedToken)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
