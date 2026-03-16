// Package services contains business logic.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"savvy/internal/models"
	"savvy/internal/repository"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
)

// PushPayload is the JSON payload sent to the browser.
type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

// PushServiceInterface defines push notification operations.
type PushServiceInterface interface {
	Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth, userAgent string) error
	Unsubscribe(ctx context.Context, endpoint string) error
	SendPushToUser(ctx context.Context, userID uuid.UUID, title, body, url string) error
	SendTestPush(ctx context.Context, userID uuid.UUID) error
	GetVAPIDPublicKey() string
	IsEnabled() bool
}

// PushService implements PushServiceInterface.
type PushService struct {
	repo           repository.PushSubscriptionRepository
	userRepo       repository.UserRepository
	vapidPublicKey string
	vapidPrivate   string
	vapidSubject   string
}

// NewPushService creates a new push service.
func NewPushService(repo repository.PushSubscriptionRepository, userRepo repository.UserRepository, vapidPublicKey, vapidPrivate, vapidSubject string) PushServiceInterface {
	// Strip "mailto:" prefix — webpush-go adds it automatically.
	// Passing "mailto:x@y" results in "mailto:mailto:x@y" in the JWT,
	// which Apple Push rejects with BadJwtToken.
	vapidSubject = strings.TrimPrefix(vapidSubject, "mailto:")

	return &PushService{
		repo:           repo,
		userRepo:       userRepo,
		vapidPublicKey: vapidPublicKey,
		vapidPrivate:   vapidPrivate,
		vapidSubject:   vapidSubject,
	}
}

// IsEnabled returns true if VAPID keys are configured.
func (s *PushService) IsEnabled() bool {
	return s.vapidPublicKey != "" && s.vapidPrivate != "" && s.vapidSubject != ""
}

// GetVAPIDPublicKey returns the VAPID public key for client subscription.
func (s *PushService) GetVAPIDPublicKey() string {
	return s.vapidPublicKey
}

// Subscribe registers a push subscription for a user.
// On the first subscription, push notification preferences are automatically enabled.
func (s *PushService) Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth, userAgent string) error {
	// Check if user already has existing subscriptions before creating new one
	existingSubs, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("check existing subscriptions: %w", err)
	}
	// isFirstSubscription is checked before Create — in a concurrent scenario two
	// goroutines could both pass this check and both enable preferences.
	// This is intentionally best-effort: the update is idempotent (bool → true).
	isFirstSubscription := len(existingSubs) == 0

	sub := &models.PushSubscription{
		UserID:    userID,
		Endpoint:  endpoint,
		P256dhKey: p256dh,
		AuthKey:   auth,
		UserAgent: userAgent,
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return fmt.Errorf("create push subscription: %w", err)
	}

	// Enable push notification preferences on first subscription
	if isFirstSubscription && s.userRepo != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			slog.WarnContext(ctx, "failed to enable push preferences for user", "user_id", userID, "error", err)
			return nil // subscription was created, preference update is best-effort
		}
		if !user.PushNotificationsEnabled {
			user.PushNotificationsEnabled = true
			user.PushRemindersEnabled = true
			user.PushSharingEnabled = true
			if err := s.userRepo.Update(ctx, user); err != nil {
				slog.WarnContext(ctx, "failed to update push preferences", "user_id", userID, "error", err)
			}
		}
	}

	return nil
}

// Unsubscribe removes a push subscription by endpoint.
func (s *PushService) Unsubscribe(ctx context.Context, endpoint string) error {
	return s.repo.DeleteByEndpoint(ctx, endpoint)
}

// SendTestPush sends a test push notification to all of the user's subscriptions.
// Unlike SendPushToUser, it returns an error if push is not enabled or no subscriptions exist,
// and reports delivery failures instead of silently continuing.
func (s *PushService) SendTestPush(ctx context.Context, userID uuid.UUID) error {
	if !s.IsEnabled() {
		return fmt.Errorf("push notifications not enabled")
	}

	subs, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get push subscriptions: %w", err)
	}

	if len(subs) == 0 {
		return fmt.Errorf("no push subscriptions found for this user — enable push notifications in your browser first")
	}

	payload := PushPayload{
		Title: "Test Push",
		Body:  "Push notifications are working!",
		URL:   "/admin/system-health",
		Icon:  "/favicon.png",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	var sent, failed int
	var lastErr string

	for _, sub := range subs {
		status, err := s.sendToSubscription(ctx, data, &sub)
		if err != nil {
			failed++
			lastErr = err.Error()
			continue
		}
		if status >= 400 {
			failed++
			lastErr = fmt.Sprintf("push service returned HTTP %d for endpoint %s", status, truncateEndpoint(sub.Endpoint))
			continue
		}
		sent++
	}

	if sent == 0 && failed > 0 {
		return fmt.Errorf("all %d push delivery attempts failed, last error: %s", failed, lastErr)
	}

	if failed > 0 {
		slog.WarnContext(ctx, "some test push deliveries failed", "sent", sent, "failed", failed)
	}

	return nil
}

// SendPushToUser sends a push notification to all of a user's subscriptions.
func (s *PushService) SendPushToUser(ctx context.Context, userID uuid.UUID, title, body, url string) error {
	if !s.IsEnabled() {
		return nil
	}

	subs, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get push subscriptions: %w", err)
	}

	if len(subs) == 0 {
		return nil
	}

	payload := PushPayload{
		Title: title,
		Body:  body,
		URL:   url,
		Icon:  "/favicon.png",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	for _, sub := range subs {
		_, _ = s.sendToSubscription(ctx, data, &sub) // best-effort delivery
	}

	return nil
}

// sendToSubscription sends a push notification to a single subscription and handles cleanup.
// Returns the HTTP status code and any error from the webpush library.
func (s *PushService) sendToSubscription(ctx context.Context, data []byte, sub *models.PushSubscription) (int, error) {
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: toBase64URL(sub.P256dhKey),
			Auth:   toBase64URL(sub.AuthKey),
		},
	}

	resp, err := webpush.SendNotification(data, subscription, &webpush.Options{
		Subscriber:      s.vapidSubject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivate,
		TTL:             86400,
	})
	if err != nil {
		slog.WarnContext(ctx, "push notification failed",
			"endpoint", truncateEndpoint(sub.Endpoint),
			"error", err,
		)
		return 0, err
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		slog.InfoContext(ctx, "removing expired push subscription", "endpoint", truncateEndpoint(sub.Endpoint))
		if delErr := s.repo.DeleteByEndpoint(ctx, sub.Endpoint); delErr != nil {
			slog.WarnContext(ctx, "failed to delete expired subscription", "error", delErr)
		}
	}

	if resp.StatusCode >= 400 {
		slog.WarnContext(ctx, "push service returned error",
			"endpoint", truncateEndpoint(sub.Endpoint),
			"status", resp.StatusCode,
			"body", string(body),
		)
	} else {
		slog.InfoContext(ctx, "push notification delivered",
			"endpoint", truncateEndpoint(sub.Endpoint),
			"status", resp.StatusCode,
		)
	}

	return resp.StatusCode, nil
}

// toBase64URL normalizes a base64 string to URL-safe base64 (RFC 4648 §5).
// This handles keys that were stored with standard base64 encoding (btoa).
func toBase64URL(s string) string {
	s = strings.ReplaceAll(s, "+", "-")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.TrimRight(s, "=")
	return s
}

// truncateEndpoint shortens a push endpoint URL for safe logging.
func truncateEndpoint(endpoint string) string {
	if len(endpoint) <= 60 {
		return endpoint
	}
	return endpoint[:30] + "..." + endpoint[len(endpoint)-20:]
}
