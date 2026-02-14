package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"savvy/internal/config"
)

// ============================================================================
// determineStatus Tests
// ============================================================================

func TestDetermineStatus_AllHealthy(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{}}
	checks := map[string]CheckResult{
		"database": {Status: "healthy", Enabled: true},
		"smtp":     {Status: "healthy", Enabled: true},
		"oauth":    {Status: "healthy", Enabled: true},
	}
	assert.Equal(t, "ready", svc.determineStatus(checks))
}

func TestDetermineStatus_DatabaseUnhealthy(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{}}
	checks := map[string]CheckResult{
		"database": {Status: "unhealthy", Enabled: true},
		"smtp":     {Status: "healthy", Enabled: true},
	}
	assert.Equal(t, "not_ready", svc.determineStatus(checks))
}

func TestDetermineStatus_DatabaseMissing(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{}}
	checks := map[string]CheckResult{
		"smtp": {Status: "healthy", Enabled: true},
	}
	assert.Equal(t, "not_ready", svc.determineStatus(checks))
}

func TestDetermineStatus_OptionalUnhealthy(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{}}
	checks := map[string]CheckResult{
		"database": {Status: "healthy", Enabled: true},
		"smtp":     {Status: "unhealthy", Enabled: true},
	}
	assert.Equal(t, "degraded", svc.determineStatus(checks))
}

func TestDetermineStatus_OptionalNotConfigured(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{}}
	checks := map[string]CheckResult{
		"database": {Status: "healthy", Enabled: true},
		"smtp":     {Status: "not_configured", Enabled: false},
	}
	// Not enabled + not healthy should not trigger degraded
	assert.Equal(t, "ready", svc.determineStatus(checks))
}

// ============================================================================
// sanitizeError Tests
// ============================================================================

func TestSanitizeError_Production(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{Environment: "production"}}
	result := svc.sanitizeError(assert.AnError)
	assert.Equal(t, "check failed", result)
}

func TestSanitizeError_Development(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{Environment: "development"}}
	result := svc.sanitizeError(assert.AnError)
	assert.Equal(t, assert.AnError.Error(), result)
}

// ============================================================================
// checkVAPID Tests
// ============================================================================

func TestCheckVAPID_Healthy(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{
		VAPIDPublicKey:  "BDd3_hVL9fZi9Ybo2UUzA284WG5FZR30_95YeZJsiApwXKpNcF1rRPF3foIiBHXRdJI2Qhumhf6_LFTeZaNndIo",
		VAPIDPrivateKey: "xV3GGQE9h3qyVGOZsVqoaN4wR6rUHxjCNEWi5S3pPKI",
		VAPIDSubject:    "mailto:test@example.com",
	}}
	result := svc.checkVAPID(context.TODO())
	assert.Equal(t, "healthy", result.Status)
	assert.True(t, result.Enabled)
}

func TestCheckVAPID_MissingKeys(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{
		VAPIDPublicKey:  "",
		VAPIDPrivateKey: "",
	}}
	result := svc.checkVAPID(context.TODO())
	assert.Equal(t, "unhealthy", result.Status)
}

func TestCheckVAPID_InvalidPublicKeyFormat(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{
		VAPIDPublicKey:  "!!!invalid-base64!!!",
		VAPIDPrivateKey: "xV3GGQE9h3qyVGOZsVqoaN4wR6rUHxjCNEWi5S3pPKI",
		VAPIDSubject:    "mailto:test@example.com",
	}}
	result := svc.checkVAPID(context.TODO())
	assert.Equal(t, "unhealthy", result.Status)
}

func TestCheckVAPID_InvalidPrivateKeyFormat(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{
		VAPIDPublicKey:  "BDd3_hVL9fZi9Ybo2UUzA284WG5FZR30_95YeZJsiApwXKpNcF1rRPF3foIiBHXRdJI2Qhumhf6_LFTeZaNndIo",
		VAPIDPrivateKey: "!!!invalid!!!",
		VAPIDSubject:    "mailto:test@example.com",
	}}
	result := svc.checkVAPID(context.TODO())
	assert.Equal(t, "unhealthy", result.Status)
}

func TestCheckVAPID_MissingSubject(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{
		VAPIDPublicKey:  "BDd3_hVL9fZi9Ybo2UUzA284WG5FZR30_95YeZJsiApwXKpNcF1rRPF3foIiBHXRdJI2Qhumhf6_LFTeZaNndIo",
		VAPIDPrivateKey: "xV3GGQE9h3qyVGOZsVqoaN4wR6rUHxjCNEWi5S3pPKI",
		VAPIDSubject:    "",
	}}
	result := svc.checkVAPID(context.TODO())
	assert.Equal(t, "unhealthy", result.Status)
}

// ============================================================================
// checkTOTPEncryption Tests
// ============================================================================

func TestCheckTOTPEncryption_Healthy(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{
		TOTPEncryptionKey: "12345678901234567890123456789012", // 32 bytes
	}}
	result := svc.checkTOTPEncryption(context.TODO())
	assert.Equal(t, "healthy", result.Status)
	assert.True(t, result.Enabled)
}

func TestCheckTOTPEncryption_WrongLength(t *testing.T) {
	svc := &HealthCheckService{config: &config.Config{
		TOTPEncryptionKey: "too-short",
	}}
	result := svc.checkTOTPEncryption(context.TODO())
	assert.Equal(t, "unhealthy", result.Status)
}

// ============================================================================
// NewHealthCheckService Tests
// ============================================================================

func TestNewHealthCheckService(t *testing.T) {
	cfg := &config.Config{}
	svc := NewHealthCheckService(nil, nil, cfg)
	assert.NotNil(t, svc)
	assert.Equal(t, cfg, svc.config)
	assert.NotNil(t, svc.httpClient)
}
