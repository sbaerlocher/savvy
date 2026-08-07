package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"savvy/internal/email"
	"savvy/internal/metrics"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

const (
	// defaultEmailBatchSize caps how many notifications one dispatcher run claims.
	//
	// The batch is delivered serially, so its worst-case duration must stay well
	// inside defaultStaleSendingAfter: once a run outlives that window, another
	// replica treats the rows still being sent as abandoned and delivers them a
	// second time. At a pessimistic 12s per SMTP round trip, 20 rows take about
	// four minutes against a ten-minute window.
	defaultEmailBatchSize = 20
	// defaultMaxEmailAttempts is how often a single email is retried before the
	// row is parked as failed.
	//
	// 'failed' is terminal — nothing requeues it — so the budget has to outlast
	// a realistic provider incident rather than a brief blip. At the one-minute
	// dispatch interval 180 attempts span roughly three hours, which covers the
	// 15-60 minute outages hosted SMTP providers routinely have. A five-minute
	// budget would park the entire queue permanently and reproduce the very
	// data loss this outbox exists to prevent, just recorded in a column
	// instead of a log line.
	//
	// The cost of a high limit is bounded: each attempt is one SMTP dial, and a
	// genuinely rejected message (bad address) burns them against a provider
	// that answers immediately.
	defaultMaxEmailAttempts = 180
	// defaultStaleSendingAfter is how long a row may sit in 'sending' before it
	// is assumed abandoned by a dead pod and returned to the queue.
	defaultStaleSendingAfter = 10 * time.Minute
)

// EmailDispatchServiceInterface defines email outbox dispatching.
type EmailDispatchServiceInterface interface {
	// DispatchPending delivers the emails of pending notifications and returns
	// how many were sent successfully.
	DispatchPending(ctx context.Context) (int, error)
	// PendingCount reports how many notifications await delivery.
	PendingCount(ctx context.Context) (int64, error)
}

// EmailDispatchService delivers the emails queued on notification rows.
//
// Notification creation only records that an email is due; this service is what
// actually sends it. Splitting the two is the point: a failing send no longer
// disappears into a log line, it stays on the row and is retried.
type EmailDispatchService struct {
	notifRepo         repository.NotificationRepository
	userRepo          repository.UserRepository
	emailService      email.ServiceInterface
	emailTokenService EmailTokenServiceInterface
	frontendURL       string

	batchSize         int
	maxAttempts       int
	staleSendingAfter time.Duration
}

// NewEmailDispatchService creates a dispatcher for the notification email outbox.
func NewEmailDispatchService(
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	emailService email.ServiceInterface,
	emailTokenService EmailTokenServiceInterface,
	frontendURL string,
) *EmailDispatchService {
	return &EmailDispatchService{
		notifRepo:         notifRepo,
		userRepo:          userRepo,
		emailService:      emailService,
		emailTokenService: emailTokenService,
		frontendURL:       frontendURL,
		batchSize:         defaultEmailBatchSize,
		maxAttempts:       defaultMaxEmailAttempts,
		staleSendingAfter: defaultStaleSendingAfter,
	}
}

// PendingCount reports how many notifications are waiting for delivery.
func (s *EmailDispatchService) PendingCount(ctx context.Context) (int64, error) {
	count, err := s.notifRepo.CountPendingEmails(ctx)
	if err != nil {
		return 0, fmt.Errorf("count pending emails: %w", err)
	}
	return count, nil
}

// DispatchPending claims pending notifications and sends their emails.
//
// Stale recovery runs first so rows stranded by a previous crash rejoin this
// batch instead of waiting for the next one.
func (s *EmailDispatchService) DispatchPending(ctx context.Context) (int, error) {
	if s.emailService == nil {
		return 0, nil
	}

	cutoff := time.Now().Add(-s.staleSendingAfter)
	if recovered, err := s.notifRepo.ResetStaleSendingEmails(ctx, cutoff); err != nil {
		slog.WarnContext(ctx, "failed to reset stale sending notifications", "error", err)
	} else if recovered > 0 {
		slog.WarnContext(ctx, "recovered stale notification emails", "count", recovered)
	}

	claimed, err := s.notifRepo.ClaimPendingEmails(ctx, s.batchSize)
	if err != nil {
		return 0, fmt.Errorf("claim pending emails: %w", err)
	}

	sent, failed := 0, 0
	for i := range claimed {
		n := &claimed[i]
		sendErr := s.deliver(ctx, n)

		// A permanent error will fail identically on every retry, so it parks
		// on the first attempt instead of occupying the retry budget. Passing
		// an effective limit of 1 makes the SQL CASE resolve straight to
		// 'failed' — retrying an unroutable type only delays the same outcome.
		attemptLimit := s.maxAttempts
		if isPermanentSendError(sendErr) {
			attemptLimit = 1
		}

		if err := s.notifRepo.MarkEmailResult(ctx, n.ID, sendErr, attemptLimit); err != nil {
			slog.ErrorContext(ctx, "failed to record email delivery result", "notification_id", n.ID, "error", err)
			continue
		}
		if sendErr == nil {
			sent++
		} else {
			failed++
			slog.WarnContext(ctx, "notification email delivery failed", "notification_id", n.ID, "type", n.Type, "error", sendErr)
		}
	}

	metrics.RecordNotificationEmailResult(sent, failed)

	return sent, nil
}

// errPermanentSend marks a delivery error that no retry can resolve.
//
// MarkEmailResult cannot tell a transient SMTP failure from a permanent one by
// looking at the error, so the distinction has to be made here.
var errPermanentSend = errors.New("permanent delivery failure")

// isPermanentSendError reports whether retrying would produce the same result.
func isPermanentSendError(err error) bool {
	return err != nil && errors.Is(err, errPermanentSend)
}

// deliver sends the email for a single claimed notification.
func (s *EmailDispatchService) deliver(ctx context.Context, n *models.Notification) error {
	recipient, err := s.userRepo.GetByID(ctx, n.UserID)
	if err != nil {
		return fmt.Errorf("get recipient %s: %w", n.UserID, err)
	}

	// Reminder and sharing emails unsubscribe from different categories, so they
	// carry different token types. Using one for both would let a user opting out
	// of reminders silently switch off their share notifications instead.
	switch n.Type {
	case models.NotificationTypeExpiryReminder:
		return s.sendExpiryReminder(ctx, recipient, n, s.reminderUnsubscribeURL(ctx, n.UserID))
	case models.NotificationTypeValidityStart:
		return s.sendValidityStart(ctx, recipient, n, s.reminderUnsubscribeURL(ctx, n.UserID))
	case models.NotificationTypeShareReceived:
		return s.sendShare(ctx, recipient, n, s.sharingUnsubscribeURL(ctx, n.UserID))
	case models.NotificationTypeTransferReceived:
		return s.sendTransfer(ctx, recipient, n, s.sharingUnsubscribeURL(ctx, n.UserID))
	default:
		// An unroutable type must fail rather than stay pending: a row nobody
		// knows how to send would otherwise be re-claimed on every single run.
		// Wrapped as permanent so it parks now instead of retrying an outcome
		// that cannot change.
		return fmt.Errorf("%w: unknown notification type %q", errPermanentSend, n.Type)
	}
}

func (s *EmailDispatchService) sendExpiryReminder(ctx context.Context, recipient *models.User, n *models.Notification, unsubscribeURL string) error {
	data := email.ExpiryReminderData{
		MerchantName: metadataString(n.Metadata, "merchant_name"),
		ResourceType: n.ResourceType,
		DaysLeft:     metadataInt(n.Metadata, "days_left"),
		ExpiresAt:    metadataString(n.Metadata, "expires_at"),
		Code:         metadataString(n.Metadata, "code"),
		Value:        metadataString(n.Metadata, "value"),
		ResourceURL:  s.resourceURL(n),
	}
	return s.emailService.SendExpiryReminder(ctx, recipient.Email, recipient.DisplayName(), data, unsubscribeURL, recipient.Language)
}

func (s *EmailDispatchService) sendValidityStart(ctx context.Context, recipient *models.User, n *models.Notification, unsubscribeURL string) error {
	data := email.ValidityStartData{
		MerchantName: metadataString(n.Metadata, "merchant_name"),
		ValidFrom:    metadataString(n.Metadata, "valid_from"),
		Code:         metadataString(n.Metadata, "code"),
		Value:        metadataString(n.Metadata, "value"),
		ResourceURL:  s.resourceURL(n),
	}
	return s.emailService.SendValidityStart(ctx, recipient.Email, recipient.DisplayName(), data, unsubscribeURL, recipient.Language)
}

func (s *EmailDispatchService) sendShare(ctx context.Context, recipient *models.User, n *models.Notification, unsubscribeURL string) error {
	amount, currency := metadataValue(n.Metadata)
	return s.emailService.SendShareNotification(
		ctx, recipient.Email, recipient.DisplayName(),
		metadataString(n.Metadata, "from_user_name"),
		n.ResourceType,
		metadataString(n.Metadata, "merchant_name"),
		metadataString(n.Metadata, "description"),
		amount, currency,
		s.resourceURL(n), unsubscribeURL, recipient.Language,
	)
}

func (s *EmailDispatchService) sendTransfer(ctx context.Context, recipient *models.User, n *models.Notification, unsubscribeURL string) error {
	amount, currency := metadataValue(n.Metadata)
	return s.emailService.SendTransferNotification(
		ctx, recipient.Email, recipient.DisplayName(),
		metadataString(n.Metadata, "from_user_name"),
		n.ResourceType,
		metadataString(n.Metadata, "merchant_name"),
		metadataString(n.Metadata, "description"),
		amount, currency,
		s.resourceURL(n), unsubscribeURL, recipient.Language,
	)
}

// resourceURL resolves the link the email points at. A stored resource_url is
// preferred because reminders deep-link to the individual item; falling back to
// the list path would silently downgrade those links to a listing page.
func (s *EmailDispatchService) resourceURL(n *models.Notification) string {
	if stored := metadataString(n.Metadata, "resource_url"); stored != "" {
		return stored
	}
	return s.frontendURL + resourceListPath(n.ResourceType)
}

// reminderUnsubscribeURL builds the opt-out link for expiry and validity emails.
func (s *EmailDispatchService) reminderUnsubscribeURL(ctx context.Context, userID uuid.UUID) string {
	if s.emailTokenService == nil {
		return ""
	}
	token, err := s.emailTokenService.CreateUnsubscribeReminderToken(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "failed to create unsubscribe reminder token", "user_id", userID, "error", err)
		return ""
	}
	return s.frontendURL + "/unsubscribe?token=" + token + "&type=reminders"
}

// sharingUnsubscribeURL builds the opt-out link for share and transfer emails.
func (s *EmailDispatchService) sharingUnsubscribeURL(ctx context.Context, userID uuid.UUID) string {
	if s.emailTokenService == nil {
		return ""
	}
	token, err := s.emailTokenService.CreateUnsubscribeToken(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "failed to create unsubscribe token", "user_id", userID, "error", err)
		return ""
	}
	return s.frontendURL + "/unsubscribe?token=" + token + "&type=notifications"
}

// metadataString reads a string field from notification metadata.
func metadataString(m models.NotificationMetadata, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// metadataInt reads an integer field from notification metadata.
//
// Metadata round-trips through JSONB, and encoding/json decodes every number
// into a float64 — asserting straight to int would silently yield 0 for a value
// that is present and correct. Both forms are accepted because a value written
// in-process this run has not been through the database yet.
func metadataInt(m models.NotificationMetadata, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// metadataValue reads the stored monetary value as (amount, currency).
func metadataValue(m models.NotificationMetadata) (float64, string) {
	raw, ok := m["value"].(map[string]interface{})
	if !ok {
		return 0, ""
	}

	var amount float64
	switch v := raw["amount"].(type) {
	case float64:
		amount = v
	case int:
		amount = float64(v)
	case int64:
		amount = float64(v)
	}

	currency, _ := raw["currency"].(string)
	return amount, currency
}
