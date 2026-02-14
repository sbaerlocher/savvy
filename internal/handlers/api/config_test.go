package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"savvy/internal/config"
)

func TestConfigHandler_GetConfig_OAuthEnabled(t *testing.T) {
	cfg := &config.Config{
		OAuthClientID:      "test-client-id",
		OAuthClientSecret:  "test-client-secret",
		OAuthIssuer:        "https://auth.example.com",
		EnableCards:        true,
		EnableVouchers:     true,
		EnableGiftCards:    true,
		EnableLocalLogin:   false,
		EnableRegistration: false,
		LogLevel:           "DEBUG",
	}

	handler := NewConfigHandler(cfg)
	c, rec := createTestContext(http.MethodGet, "/api/v1/config", "")

	err := handler.GetConfig(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ConfigResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)

	assert.True(t, response.OAuth.Enabled)
	assert.Equal(t, "/auth/oauth/login", response.OAuth.LoginURL)
	assert.True(t, response.Features.Cards)
	assert.True(t, response.Features.Vouchers)
	assert.True(t, response.Features.GiftCards)
	assert.False(t, response.LocalLogin)
	assert.False(t, response.Registration)
}

func TestConfigHandler_GetConfig_OAuthDisabled(t *testing.T) {
	cfg := &config.Config{
		EnableCards:        true,
		EnableVouchers:     false,
		EnableGiftCards:    true,
		EnableLocalLogin:   true,
		EnableRegistration: true,
		LogLevel:           "INFO",
	}

	handler := NewConfigHandler(cfg)
	c, rec := createTestContext(http.MethodGet, "/api/v1/config", "")

	err := handler.GetConfig(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ConfigResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)

	assert.False(t, response.OAuth.Enabled)
	assert.Empty(t, response.OAuth.LoginURL)
	assert.True(t, response.Features.Cards)
	assert.False(t, response.Features.Vouchers)
	assert.True(t, response.Features.GiftCards)
	assert.True(t, response.LocalLogin)
	assert.True(t, response.Registration)
}

func TestConfigHandler_GetConfig_AllFeaturesDisabled(t *testing.T) {
	cfg := &config.Config{
		EnableCards:        false,
		EnableVouchers:     false,
		EnableGiftCards:    false,
		EnableLocalLogin:   false,
		EnableRegistration: false,
		LogLevel:           "WARN",
	}

	handler := NewConfigHandler(cfg)
	c, rec := createTestContext(http.MethodGet, "/api/v1/config", "")

	err := handler.GetConfig(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ConfigResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)

	assert.False(t, response.Features.Cards)
	assert.False(t, response.Features.Vouchers)
	assert.False(t, response.Features.GiftCards)
	assert.False(t, response.LocalLogin)
	assert.False(t, response.Registration)
}
