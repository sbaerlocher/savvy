// Package services contains business logic.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"savvy/internal/email"
	"savvy/internal/i18n"
	"savvy/internal/models"
	"savvy/internal/repository"
	"time"

	"github.com/google/uuid"
)

// NotificationValue holds a monetary value for a notification. It is shown in
// the email and stored in the notification metadata, but never in the push body
// (lockscreen privacy). A nil *NotificationValue means the resource has no value
// (e.g. Card).
type NotificationValue struct {
	Amount   float64
	Currency string
}

// ShareNotificationInput carries the data for a share notification. Using an
// options struct instead of positional parameters keeps the signature stable as
// fields are added and preserves compile-time checks at call sites.
type ShareNotificationInput struct {
	RecipientID  uuid.UUID
	FromUserID   uuid.UUID
	FromUserName string
	ResourceType string
	ResourceID   uuid.UUID
	Permissions  map[string]bool
	MerchantName string
	Description  string
	Value        *NotificationValue // nil for resources without a value (Card)
}

// TransferNotificationInput carries the data for a transfer notification.
// Same rationale as ShareNotificationInput; transfers have no permissions.
type TransferNotificationInput struct {
	RecipientID  uuid.UUID
	FromUserID   uuid.UUID
	FromUserName string
	ResourceType string
	ResourceID   uuid.UUID
	MerchantName string
	Description  string
	Value        *NotificationValue // nil for resources without a value (Card)
}

// NotificationServiceInterface defines the interface for notification business logic
type NotificationServiceInterface interface {
	CreateShareNotification(ctx context.Context, in ShareNotificationInput) error
	CreateTransferNotification(ctx context.Context, in TransferNotificationInput) error
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, userID, notificationID uuid.UUID) error
	ArchiveOldRead(ctx context.Context, olderThanDays int) (int64, error)
	SetPushService(pushService PushServiceInterface)
	SetEmailService(emailService email.ServiceInterface, emailTokenService EmailTokenServiceInterface, frontendURL string)
}

// NotificationService implements NotificationServiceInterface
type NotificationService struct {
	repo              repository.NotificationRepository
	userRepo          repository.UserRepository
	pushService       PushServiceInterface
	emailService      email.ServiceInterface
	emailTokenService EmailTokenServiceInterface
	frontendURL       string
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo repository.NotificationRepository, userRepo repository.UserRepository) NotificationServiceInterface {
	return &NotificationService{repo: repo, userRepo: userRepo}
}

// SetPushService sets the push service for sending push notifications.
func (s *NotificationService) SetPushService(pushService PushServiceInterface) {
	s.pushService = pushService
}

// SetEmailService sets the email service for sending email notifications.
func (s *NotificationService) SetEmailService(emailService email.ServiceInterface, emailTokenService EmailTokenServiceInterface, frontendURL string) {
	s.emailService = emailService
	s.emailTokenService = emailTokenService
	s.frontendURL = frontendURL
}

// CreateShareNotification creates a notification when a resource is shared with a user
func (s *NotificationService) CreateShareNotification(ctx context.Context, in ShareNotificationInput) error {
	// Build metadata. The in-app renderer reads merchant_name; description and
	// value are stored for future use. value is never put in the push body.
	metadata := models.NotificationMetadata{
		"from_user_id":   in.FromUserID.String(),
		"from_user_name": in.FromUserName,
	}
	if in.Permissions != nil {
		metadata["permissions"] = in.Permissions
	}
	addResourceMetadata(metadata, in.MerchantName, in.Description, in.Value)

	resourceURL := resourceListPath(in.ResourceType)

	// Look the recipient up before creating the row: the same lookup the push
	// gate already needed now also decides the row's email state, so this costs
	// no extra query.
	recipient := s.getRecipient(ctx, in.RecipientID)

	notification := &models.Notification{
		UserID:       in.RecipientID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: in.ResourceType,
		ResourceID:   in.ResourceID,
		Metadata:     metadata,
		IsRead:       false,
		EmailStatus:  s.emailStatusForSharing(recipient),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("create share notification: %w", err)
	}

	slog.Info("Share notification created", "recipient_id", in.RecipientID, "resource_type", in.ResourceType, "resource_id", in.ResourceID)

	lang := ""
	if recipient != nil {
		lang = recipient.Language
	}

	// Send push notification (gated by channel + category preferences).
	// Push body carries merchant + description but never the value (lockscreen privacy).
	// Push stays inline and best-effort; only email moved to the outbox.
	if recipient == nil || (recipient.PushNotificationsEnabled && recipient.PushSharingEnabled) {
		title := i18n.T(i18nCtx(lang), "push.share.title")
		body := sharePushBody(lang, in.FromUserName, in.ResourceType, in.MerchantName, in.Description)
		s.sendPush(ctx, in.RecipientID, title, body, resourceURL)
	}

	return nil
}

// CreateTransferNotification creates a notification when resource ownership is transferred to a user
func (s *NotificationService) CreateTransferNotification(ctx context.Context, in TransferNotificationInput) error {
	metadata := models.NotificationMetadata{
		"from_user_id":   in.FromUserID.String(),
		"from_user_name": in.FromUserName,
	}
	addResourceMetadata(metadata, in.MerchantName, in.Description, in.Value)

	resourceURL := resourceListPath(in.ResourceType)

	// Recipient lookup moved ahead of the row for the same reason as in
	// CreateShareNotification: it decides the row's email state.
	recipient := s.getRecipient(ctx, in.RecipientID)

	notification := &models.Notification{
		UserID:       in.RecipientID,
		Type:         models.NotificationTypeTransferReceived,
		ResourceType: in.ResourceType,
		ResourceID:   in.ResourceID,
		Metadata:     metadata,
		IsRead:       false,
		EmailStatus:  s.emailStatusForSharing(recipient),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("create transfer notification: %w", err)
	}

	slog.Info("Transfer notification created", "recipient_id", in.RecipientID, "resource_type", in.ResourceType, "resource_id", in.ResourceID)

	lang := ""
	if recipient != nil {
		lang = recipient.Language
	}

	// Send push notification (gated by channel + category preferences).
	// Push body carries merchant + description but never the value (lockscreen privacy).
	// Push stays inline and best-effort; only email moved to the outbox.
	if recipient == nil || (recipient.PushNotificationsEnabled && recipient.PushSharingEnabled) {
		title := i18n.T(i18nCtx(lang), "push.transfer.title")
		body := transferPushBody(lang, in.FromUserName, in.ResourceType, in.MerchantName, in.Description)
		s.sendPush(ctx, in.RecipientID, title, body, resourceURL)
	}

	return nil
}

// emailStatusForSharing decides whether a share/transfer row queues an email.
//
// A recipient that cannot be loaded yields 'skipped'. That matches the previous
// behaviour, where the same failed lookup ended the send: queueing instead would
// park a mail with no address to deliver it to.
func (s *NotificationService) emailStatusForSharing(recipient *models.User) models.EmailStatus {
	if s.emailService == nil || recipient == nil {
		return models.EmailStatusSkipped
	}
	return emailStatusFor(recipient.EmailSharingEnabled && recipient.EmailNotificationsEnabled)
}

// addResourceMetadata adds merchant_name, description and value to notification
// metadata when present. The in-app renderer reads merchant_name; description
// and value are stored for future use, and value is deliberately kept out of
// the push body (see *PushBody helpers).
func addResourceMetadata(metadata models.NotificationMetadata, merchantName, description string, value *NotificationValue) {
	if merchantName != "" {
		metadata["merchant_name"] = merchantName
	}
	if description != "" {
		metadata["description"] = description
	}
	if value != nil {
		metadata["value"] = map[string]any{
			"amount":   value.Amount,
			"currency": value.Currency,
		}
	}
}

// GetUserNotifications retrieves all notifications for a user with pagination
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	notifications, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get user notifications: %w", err)
	}
	return notifications, nil
}

// GetUnreadCount returns the number of unread notifications for a user
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get unread notification count: %w", err)
	}
	return count, nil
}

// MarkAsRead marks a notification as read (scoped to user for ownership check)
func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	if err := s.repo.MarkAsRead(ctx, userID, notificationID); err != nil {
		return fmt.Errorf("mark notification as read %s: %w", notificationID, err)
	}
	return nil
}

// MarkAllAsRead marks all unread notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}

// DeleteNotification deletes a notification (scoped to user for ownership check)
func (s *NotificationService) DeleteNotification(ctx context.Context, userID, notificationID uuid.UUID) error {
	if err := s.repo.Delete(ctx, userID, notificationID); err != nil {
		return fmt.Errorf("delete notification %s: %w", notificationID, err)
	}
	return nil
}

// ArchiveOldRead archives read notifications last read more than olderThanDays
// ago and returns how many were archived.
func (s *NotificationService) ArchiveOldRead(ctx context.Context, olderThanDays int) (int64, error) {
	cutoff := time.Now().Add(-time.Duration(olderThanDays) * 24 * time.Hour)
	count, err := s.repo.ArchiveOldRead(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("archive old notifications: %w", err)
	}
	return count, nil
}

// sendPush sends a push notification if the push service is configured.
func (s *NotificationService) sendPush(ctx context.Context, userID uuid.UUID, title, body, url string) {
	if s.pushService == nil {
		return
	}
	if err := s.pushService.SendPushToUser(ctx, userID, title, body, url); err != nil {
		slog.WarnContext(ctx, "failed to send push notification", "user_id", userID, "error", err)
	}
}

// getRecipient looks up the recipient user for preference checks and language.
func (s *NotificationService) getRecipient(ctx context.Context, userID uuid.UUID) *models.User {
	if s.userRepo == nil {
		return nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil
	}
	return user
}
