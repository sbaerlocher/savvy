// Package security provides cryptographic utilities for secure token generation and validation
package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidToken indicates the token is malformed or tampered with
	ErrInvalidToken = errors.New("invalid or tampered token")
	// ErrTokenExpired indicates the token has expired
	ErrTokenExpired = errors.New("token has expired")
	// ErrInvalidClaims indicates the token claims are invalid
	ErrInvalidClaims = errors.New("invalid token claims")
)

// BarcodeTokenClaims represents the data embedded in a barcode token
type BarcodeTokenClaims struct {
	ResourceID   uuid.UUID `json:"rid"`   // Resource ID (card, voucher, or gift card)
	ResourceType string    `json:"rtype"` // "card", "voucher", or "gift_card"
	UserID       uuid.UUID `json:"uid"`   // User ID (for access control)
	ExpiresAt    int64     `json:"exp"`   // Unix timestamp
}

// tokenSecret holds the HMAC secret key (set via Init)
var tokenSecret []byte

// Init initializes the security package with the session secret
func Init(secret string) {
	tokenSecret = []byte(secret)
}

// GenerateBarcodeToken creates a cryptographically signed token for barcode access
// The token is valid for the specified duration (e.g., 60 seconds)
func GenerateBarcodeToken(resourceID uuid.UUID, resourceType string, userID uuid.UUID, validDuration time.Duration) (string, error) {
	claims := BarcodeTokenClaims{
		ResourceID:   resourceID,
		ResourceType: resourceType,
		UserID:       userID,
		ExpiresAt:    time.Now().Add(validDuration).Unix(),
	}

	// Marshal claims to JSON
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// Encode claims as base64
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Generate HMAC signature
	signature := generateHMAC(claimsB64, tokenSecret)
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	// Token format: {claims}.{signature}
	token := claimsB64 + "." + signatureB64

	return token, nil
}

// ValidateBarcodeToken verifies the token signature and expiration
// Returns the claims if valid, or an error if invalid/expired
func ValidateBarcodeToken(token string) (*BarcodeTokenClaims, error) {
	// Split token into claims and signature
	var claimsB64, signatureB64 string
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			claimsB64 = token[:i]
			signatureB64 = token[i+1:]
			break
		}
	}

	if claimsB64 == "" || signatureB64 == "" {
		return nil, ErrInvalidToken
	}

	// Verify signature
	expectedSignature := generateHMAC(claimsB64, tokenSecret)
	providedSignature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !hmac.Equal(expectedSignature, providedSignature) {
		return nil, ErrInvalidToken
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims BarcodeTokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Validate claims
	if claims.ResourceID == uuid.Nil || claims.UserID == uuid.Nil {
		return nil, ErrInvalidClaims
	}

	if claims.ResourceType != "card" && claims.ResourceType != "voucher" && claims.ResourceType != "gift_card" {
		return nil, ErrInvalidClaims
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// generateHMAC creates an HMAC-SHA256 signature
func generateHMAC(data string, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return h.Sum(nil)
}
