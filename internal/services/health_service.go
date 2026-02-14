// Package services contains business logic for health checks.
package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"savvy/internal/config"
	"savvy/internal/email"
	"sync"
	"time"

	"gorm.io/gorm"
)

// HealthCheckServiceInterface defines the contract for health checking.
type HealthCheckServiceInterface interface {
	CheckReadiness(ctx context.Context) (*ReadinessReport, error)
}

// HealthCheckService orchestrates health checks for all services.
type HealthCheckService struct {
	db           *gorm.DB
	emailService email.ServiceInterface
	config       *config.Config
	httpClient   *http.Client
}

// ReadinessReport contains the overall readiness status and individual check results.
type ReadinessReport struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of a single health check.
type CheckResult struct {
	Status    string `json:"status"`               // "healthy", "unhealthy", "not_configured"
	Enabled   bool   `json:"enabled"`              // Is the service enabled?
	LatencyMs *int64 `json:"latency_ms,omitempty"` // Only for database
	Error     string `json:"error,omitempty"`      // Error details (sanitized in production)
}

// NewHealthCheckService creates a new health check service.
func NewHealthCheckService(
	db *gorm.DB,
	emailService email.ServiceInterface,
	cfg *config.Config,
) *HealthCheckService {
	return &HealthCheckService{
		db:           db,
		emailService: emailService,
		config:       cfg,
		httpClient: &http.Client{
			Timeout: 2 * time.Second, // Global HTTP timeout for OAuth checks
		},
	}
}

// CheckReadiness performs all health checks in parallel and returns a readiness report.
func (s *HealthCheckService) CheckReadiness(ctx context.Context) (*ReadinessReport, error) {
	// Global timeout: 5 seconds
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	var mu sync.Mutex // Protect map writes
	var wg sync.WaitGroup

	// 1. Database check (CRITICAL)
	wg.Add(1)
	go func() {
		defer wg.Done()
		result := s.checkDatabase(ctx)
		mu.Lock()
		checks["database"] = result
		mu.Unlock()
	}()

	// 2. SMTP check (OPTIONAL) - only if enabled
	if s.config.IsSMTPEnabled() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := s.checkSMTP(ctx)
			mu.Lock()
			checks["smtp"] = result
			mu.Unlock()
		}()
	} else {
		checks["smtp"] = CheckResult{Status: "not_configured", Enabled: false}
	}

	// 3. OAuth check (OPTIONAL) - only if enabled
	if s.config.IsOAuthEnabled() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := s.checkOAuth(ctx)
			mu.Lock()
			checks["oauth"] = result
			mu.Unlock()
		}()
	} else {
		checks["oauth"] = CheckResult{Status: "not_configured", Enabled: false}
	}

	// 4. VAPID check (OPTIONAL) - only if enabled
	if s.config.IsPushEnabled() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := s.checkVAPID(ctx)
			mu.Lock()
			checks["vapid"] = result
			mu.Unlock()
		}()
	} else {
		checks["vapid"] = CheckResult{Status: "not_configured", Enabled: false}
	}

	// 5. TOTP encryption check (OPTIONAL) - only if enabled
	if s.config.Is2FAEnabled() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := s.checkTOTPEncryption(ctx)
			mu.Lock()
			checks["totp_encryption"] = result
			mu.Unlock()
		}()
	} else {
		checks["totp_encryption"] = CheckResult{Status: "not_configured", Enabled: false}
	}

	wg.Wait()

	// Determine overall status
	status := s.determineStatus(checks)

	return &ReadinessReport{
		Status:    status,
		Timestamp: time.Now(),
		Checks:    checks,
	}, nil
}

// checkDatabase verifies database connectivity with a 2-second timeout.
func (s *HealthCheckService) checkDatabase(ctx context.Context) CheckResult {
	start := time.Now()

	// Timeout for DB check: 2 seconds
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Get underlying *sql.DB
	sqlDB, err := s.db.DB()
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("failed to get DB: %w", err)),
		}
	}

	// Ping database
	if err := sqlDB.PingContext(checkCtx); err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("database ping failed: %w", err)),
		}
	}

	latency := time.Since(start).Milliseconds()
	return CheckResult{
		Status:    "healthy",
		Enabled:   true,
		LatencyMs: &latency,
	}
}

// checkSMTP verifies SMTP server connectivity with a 3-second timeout.
func (s *HealthCheckService) checkSMTP(ctx context.Context) CheckResult {
	// Timeout: 3 seconds
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Use EmailService interface method
	if err := s.emailService.CheckConnection(checkCtx); err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(err),
		}
	}

	return CheckResult{Status: "healthy", Enabled: true}
}

// checkOAuth verifies OAuth issuer is reachable via .well-known/openid-configuration.
func (s *HealthCheckService) checkOAuth(ctx context.Context) CheckResult {
	// Timeout: 2 seconds (inherited from s.httpClient)
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Construct .well-known URL
	wellKnownURL := s.config.OAuthIssuer + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, wellKnownURL, nil)
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("failed to create request: %w", err)),
		}
	}

	resp, err := s.httpClient.Do(req) // #nosec G704 -- URL from trusted config (OAuthIssuer), not user input
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("failed to reach OAuth issuer: %w", err)),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("OAuth issuer returned status %d", resp.StatusCode)),
		}
	}

	return CheckResult{Status: "healthy", Enabled: true}
}

// checkVAPID validates VAPID keys are set and have valid format.
func (s *HealthCheckService) checkVAPID(_ context.Context) CheckResult {
	// Fast local check (no network I/O)
	if s.config.VAPIDPublicKey == "" || s.config.VAPIDPrivateKey == "" {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("VAPID keys not configured")),
		}
	}

	// Validate Base64 format (VAPID keys are unpadded Base64 URL-encoded per RFC 4648)
	if _, err := base64.RawURLEncoding.DecodeString(s.config.VAPIDPublicKey); err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("invalid VAPID public key format: %w", err)),
		}
	}

	if _, err := base64.RawURLEncoding.DecodeString(s.config.VAPIDPrivateKey); err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("invalid VAPID private key format: %w", err)),
		}
	}

	// Validate subject (must be mailto: or https:)
	if s.config.VAPIDSubject == "" {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("VAPID subject not configured")),
		}
	}

	return CheckResult{Status: "healthy", Enabled: true}
}

// checkTOTPEncryption validates TOTP encryption key is set and has correct length.
func (s *HealthCheckService) checkTOTPEncryption(_ context.Context) CheckResult {
	// Fast local check (no network I/O)
	if len(s.config.TOTPEncryptionKey) != 32 {
		return CheckResult{
			Status:  "unhealthy",
			Enabled: true,
			Error:   s.sanitizeError(fmt.Errorf("TOTP encryption key must be 32 bytes, got %d", len(s.config.TOTPEncryptionKey))),
		}
	}

	return CheckResult{Status: "healthy", Enabled: true}
}

// determineStatus aggregates individual check results into overall status.
func (s *HealthCheckService) determineStatus(checks map[string]CheckResult) string {
	// Critical check: database
	dbCheck, exists := checks["database"]
	if !exists || dbCheck.Status != "healthy" {
		return "not_ready" // Critical failure
	}

	// Optional checks: if any fail, status is "degraded"
	for name, check := range checks {
		if name == "database" {
			continue // Already checked
		}
		if check.Enabled && check.Status != "healthy" {
			return "degraded" // Non-critical failure
		}
	}

	return "ready" // All checks passed
}

// sanitizeError returns a sanitized error message.
// In production: generic error message.
// In development: detailed error message.
func (s *HealthCheckService) sanitizeError(err error) string {
	if s.config.IsProduction() {
		return "check failed" // Generic message in production
	}
	return err.Error() // Detailed message in development
}
