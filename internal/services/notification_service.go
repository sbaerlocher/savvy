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

// NotificationServiceInterface defines the interface for notification business logic
type NotificationServiceInterface interface {
	CreateShareNotification(ctx context.Context, recipientID, fromUserID uuid.UUID, fromUserName, resourceType string, resourceID uuid.UUID, permissions map[string]bool) error
	CreateTransferNotification(ctx context.Context, recipientID, fromUserID uuid.UUID, fromUserName, resourceType string, resourceID uuid.UUID) error
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
func (s *NotificationService) CreateShareNotification(
	ctx context.Context,
	recipientID, fromUserID uuid.UUID,
	fromUserName, resourceType string,
	resourceID uuid.UUID,
	permissions map[string]bool,
) error {
	// Build metadata
	metadata := models.NotificationMetadata{
		"from_user_id":   fromUserID.String(),
		"from_user_name": fromUserName,
	}

	// Add permissions if provided
	if permissions != nil {
		metadata["permissions"] = permissions
	}

	notification := &models.Notification{
		UserID:       recipientID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		IsRead:       false,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("create share notification: %w", err)
	}

	slog.Info("Share notification created", "recipient_id", recipientID, "resource_type", resourceType, "resource_id", resourceID)

	resourceURL := "/" + resourceType + "s"

	// Get recipient for preference checks
	recipient := s.getRecipient(ctx, recipientID)
	lang := ""
	if recipient != nil {
		lang = recipient.Language
	}

	// Send push notification (gated by channel + category preferences)
	if recipient == nil || (recipient.PushNotificationsEnabled && recipient.PushSharingEnabled) {
		lctx := i18nCtx(lang)
		resource := i18n.T(lctx, "push.resource."+resourceType)
		article := pushArticle(resourceType, lang)
		title := i18n.T(lctx, "push.share.title")
		body := i18n.T(lctx, "push.share.body", map[string]any{"User": fromUserName, "Resource": resource, "Article": article})
		s.sendPush(ctx, recipientID, title, body, resourceURL)
	}

	// Send email notification (gated by channel + category preferences)
	s.sendShareEmail(ctx, recipientID, fromUserName, resourceType, resourceURL)

	return nil
}

// CreateTransferNotification creates a notification when resource ownership is transferred to a user
func (s *NotificationService) CreateTransferNotification(
	ctx context.Context,
	recipientID, fromUserID uuid.UUID,
	fromUserName, resourceType string,
	resourceID uuid.UUID,
) error {
	notification := &models.Notification{
		UserID:       recipientID,
		Type:         models.NotificationTypeTransferReceived,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata: models.NotificationMetadata{
			"from_user_id":   fromUserID.String(),
			"from_user_name": fromUserName,
		},
		IsRead: false,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("create transfer notification: %w", err)
	}

	slog.Info("Transfer notification created", "recipient_id", recipientID, "resource_type", resourceType, "resource_id", resourceID)

	resourceURL := "/" + resourceType + "s"

	// Get recipient for preference checks
	recipient := s.getRecipient(ctx, recipientID)
	lang := ""
	if recipient != nil {
		lang = recipient.Language
	}

	// Send push notification (gated by channel + category preferences)
	if recipient == nil || (recipient.PushNotificationsEnabled && recipient.PushSharingEnabled) {
		lctx := i18nCtx(lang)
		resource := i18n.T(lctx, "push.resource."+resourceType)
		article := pushArticle(resourceType, lang)
		title := i18n.T(lctx, "push.transfer.title")
		body := i18n.T(lctx, "push.transfer.body", map[string]any{"User": fromUserName, "Resource": resource, "Article": article})
		s.sendPush(ctx, recipientID, title, body, resourceURL)
	}

	// Send email notification (gated by channel + category preferences)
	s.sendTransferEmail(ctx, recipientID, fromUserName, resourceType, resourceURL)

	return nil
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

// sendShareEmail sends a share notification email (best-effort).
// Gated by both EmailSharingEnabled (category) and EmailNotificationsEnabled (channel).
func (s *NotificationService) sendShareEmail(ctx context.Context, recipientID uuid.UUID, fromUserName, resourceType, resourcePath string) {
	if s.emailService == nil || s.userRepo == nil {
		return
	}

	recipient, err := s.userRepo.GetByID(ctx, recipientID)
	if err != nil {
		slog.WarnContext(ctx, "failed to get recipient for share email", "user_id", recipientID, "error", err)
		return
	}

	if !recipient.EmailSharingEnabled || !recipient.EmailNotificationsEnabled {
		return
	}

	resourceURL := s.frontendURL + resourcePath
	unsubscribeURL := s.generateUnsubscribeURL(ctx, recipientID)
	if err := s.emailService.SendShareNotification(ctx, recipient.Email, recipient.DisplayName(), fromUserName, resourceType, resourceURL, unsubscribeURL, recipient.Language); err != nil {
		slog.WarnContext(ctx, "failed to send share notification email", "user_id", recipientID, "error", err)
	}
}

// sendTransferEmail sends a transfer notification email (best-effort).
// Gated by both EmailSharingEnabled (category) and EmailNotificationsEnabled (channel).
func (s *NotificationService) sendTransferEmail(ctx context.Context, recipientID uuid.UUID, fromUserName, resourceType, resourcePath string) {
	if s.emailService == nil || s.userRepo == nil {
		return
	}

	recipient, err := s.userRepo.GetByID(ctx, recipientID)
	if err != nil {
		slog.WarnContext(ctx, "failed to get recipient for transfer email", "user_id", recipientID, "error", err)
		return
	}

	if !recipient.EmailSharingEnabled || !recipient.EmailNotificationsEnabled {
		return
	}

	resourceURL := s.frontendURL + resourcePath
	unsubscribeURL := s.generateUnsubscribeURL(ctx, recipientID)
	if err := s.emailService.SendTransferNotification(ctx, recipient.Email, recipient.DisplayName(), fromUserName, resourceType, resourceURL, unsubscribeURL, recipient.Language); err != nil {
		slog.WarnContext(ctx, "failed to send transfer notification email", "user_id", recipientID, "error", err)
	}
}

// generateUnsubscribeURL creates a one-click unsubscribe URL with a token.
func (s *NotificationService) generateUnsubscribeURL(ctx context.Context, userID uuid.UUID) string {
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
