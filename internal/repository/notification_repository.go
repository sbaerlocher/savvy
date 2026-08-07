// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"savvy/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, userID, notificationID uuid.UUID) error
	// ArchiveOldRead archives read notifications older than cutoff and returns the count.
	ArchiveOldRead(ctx context.Context, cutoff time.Time) (int64, error)

	// ClaimPendingEmails atomically claims up to limit pending notifications for
	// email delivery, flipping them to 'sending'. Safe to call concurrently from
	// multiple replicas: a claimed row is never handed to a second caller.
	ClaimPendingEmails(ctx context.Context, limit int) ([]models.Notification, error)
	// MarkEmailResult records the outcome of a delivery attempt. A nil sendErr
	// marks the row sent; otherwise the row goes back to 'pending' for another
	// attempt, or to 'failed' once maxAttempts is reached.
	MarkEmailResult(ctx context.Context, id uuid.UUID, sendErr error, maxAttempts int) error
	// ResetStaleSendingEmails returns rows stuck in 'sending' since before cutoff
	// to 'pending' and reports how many were recovered.
	ResetStaleSendingEmails(ctx context.Context, cutoff time.Time) (int64, error)
	// CountPendingEmails returns the number of notifications awaiting delivery.
	CountPendingEmails(ctx context.Context) (int64, error)
}

// notificationRepository implements NotificationRepository
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// GetByID retrieves a notification by ID
func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetByUserID retrieves all notifications for a user with pagination
func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND archived_at IS NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// GetUnreadCount returns the number of unread notifications for a user
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = FALSE AND archived_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read (scoped to user for ownership check)
func (r *notificationRepository) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAllAsRead marks all unread notifications as read for a user
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = FALSE", userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

// Delete soft deletes a notification (scoped to user for ownership check)
func (r *notificationRepository) Delete(ctx context.Context, userID, notificationID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ArchiveOldRead archives notifications read before cutoff by stamping
// archived_at. Keying on read_at (not created_at) means the archive window
// counts from when the user read it, so an old notification just read stays
// visible for the full window. Archived rows drop out of the main list but
// stay in the table.
func (r *notificationRepository) ArchiveOldRead(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("is_read = TRUE AND archived_at IS NULL AND read_at IS NOT NULL AND read_at < ?", cutoff).
		Update("archived_at", gorm.Expr("CURRENT_TIMESTAMP"))
	return result.RowsAffected, result.Error
}

// ClaimPendingEmails claims pending notifications for delivery.
//
// FOR UPDATE SKIP LOCKED lets every replica claim a disjoint set of rows without
// a leader election: a row already locked by another transaction is passed over
// rather than waited on. The status flip to 'sending' commits before the caller
// touches SMTP, so a slow or hung send never holds a database lock.
//
// updated_at is written explicitly because stale-claim recovery ages rows by it.
func (r *notificationRepository) ClaimPendingEmails(ctx context.Context, limit int) ([]models.Notification, error) {
	var claimed []models.Notification

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var candidates []models.Notification
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("email_status = ?", models.EmailStatusPending).
			Order("created_at ASC").
			Limit(limit).
			Find(&candidates).Error; err != nil {
			return err
		}
		if len(candidates) == 0 {
			return nil
		}

		ids := make([]uuid.UUID, len(candidates))
		for i := range candidates {
			ids[i] = candidates[i].ID
		}

		if err := tx.Model(&models.Notification{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"email_status": models.EmailStatusSending,
				"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
			}).Error; err != nil {
			return err
		}

		for i := range candidates {
			candidates[i].EmailStatus = models.EmailStatusSending
		}
		claimed = candidates
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

// MarkEmailResult records the outcome of one delivery attempt.
//
// The pending-vs-failed decision is made in SQL against the stored counter
// rather than against a value read earlier in Go: two dispatcher runs could
// otherwise both read the same attempt count and neither would reach the limit.
func (r *notificationRepository) MarkEmailResult(ctx context.Context, id uuid.UUID, sendErr error, maxAttempts int) error {
	updates := map[string]interface{}{
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}

	if sendErr == nil {
		updates["email_status"] = models.EmailStatusSent
		updates["email_attempts"] = gorm.Expr("email_attempts + 1")
		updates["email_last_error"] = nil
	} else {
		updates["email_status"] = gorm.Expr(
			"CASE WHEN email_attempts + 1 >= ? THEN ? ELSE ? END",
			maxAttempts, string(models.EmailStatusFailed), string(models.EmailStatusPending),
		)
		updates["email_attempts"] = gorm.Expr("email_attempts + 1")
		updates["email_last_error"] = truncateError(sendErr.Error())
	}

	// Scoped to 'sending' so a late write from a pod whose row was already
	// stale-recovered and re-claimed elsewhere cannot overwrite the newer
	// attempt's state.
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND email_status = ?", id, models.EmailStatusSending).
		Updates(updates).Error
}

// maxEmailErrorLength caps what is stored in email_last_error. The column is
// TEXT, but an unbounded provider error string has no diagnostic value beyond
// its first lines and would bloat the row.
const maxEmailErrorLength = 500

// truncateError shortens a send error to a storable length.
//
// Slicing on a byte offset can split a multi-byte rune, and PostgreSQL rejects
// the resulting invalid UTF-8. That would fail the status write and strand the
// row in 'sending': the stale reset returns it to 'pending', the send fails the
// same way, and the attempt counter never advances — an endless resend loop.
// Non-ASCII reaches err.Error() routinely via provider responses and merchant
// names, so this is the common path, not an edge case.
func truncateError(msg string) string {
	if len(msg) > maxEmailErrorLength {
		return strings.ToValidUTF8(msg[:maxEmailErrorLength], "")
	}
	return msg
}

// ResetStaleSendingEmails recovers rows abandoned mid-send.
//
// A pod that dies between claiming a row and recording the outcome leaves it in
// 'sending' with nobody working on it. Returning such rows to 'pending' after a
// grace period is what makes delivery at-least-once; the cost is that a mail
// actually sent just before the crash can go out twice.
func (r *notificationRepository) ResetStaleSendingEmails(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("email_status = ? AND updated_at < ?", models.EmailStatusSending, cutoff).
		Updates(map[string]interface{}{
			"email_status": models.EmailStatusPending,
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		})
	return result.RowsAffected, result.Error
}

// CountPendingEmails returns how many notifications are waiting for delivery.
func (r *notificationRepository) CountPendingEmails(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("email_status = ?", models.EmailStatusPending).
		Count(&count).Error
	return count, err
}
