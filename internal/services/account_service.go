// Package services contains business logic.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"savvy/internal/email"
	"savvy/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountServiceInterface defines account management operations.
type AccountServiceInterface interface {
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
}

// AccountService implements AccountServiceInterface.
type AccountService struct {
	db           *gorm.DB
	userService  UserServiceInterface
	emailService email.ServiceInterface
}

// NewAccountService creates a new account service.
func NewAccountService(db *gorm.DB, userService UserServiceInterface, emailService email.ServiceInterface) AccountServiceInterface {
	return &AccountService{
		db:           db,
		userService:  userService,
		emailService: emailService,
	}
}

// DeleteAccount permanently deletes a user account and all associated data.
// This is a GDPR-compliant hard delete operation.
func (s *AccountService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	// Get user before deletion for confirmation email
	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	userEmail := user.Email
	userName := user.FirstName
	userLanguage := user.Language

	// Run all deletions in a single transaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Stage 1: Delete incoming shares (shared WITH this user)
		if err := tx.Unscoped().Where("shared_with_id = ?", userID).Delete(&models.CardShare{}).Error; err != nil {
			return fmt.Errorf("delete card shares (incoming): %w", err)
		}
		if err := tx.Unscoped().Where("shared_with_id = ?", userID).Delete(&models.VoucherShare{}).Error; err != nil {
			return fmt.Errorf("delete voucher shares (incoming): %w", err)
		}
		if err := tx.Unscoped().Where("shared_with_id = ?", userID).Delete(&models.GiftCardShare{}).Error; err != nil {
			return fmt.Errorf("delete gift card shares (incoming): %w", err)
		}

		// Stage 2: Delete user preferences
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.UserFavorite{}).Error; err != nil {
			return fmt.Errorf("delete favorites: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.Notification{}).Error; err != nil {
			return fmt.Errorf("delete notifications: %w", err)
		}

		// Stage 3: Delete owned resource shares (shares OF user's resources)
		// Get IDs of user's cards/vouchers/gift cards
		var cardIDs []uuid.UUID
		tx.Model(&models.Card{}).Unscoped().Where("user_id = ?", userID).Pluck("id", &cardIDs)
		if len(cardIDs) > 0 {
			if err := tx.Unscoped().Where("card_id IN ?", cardIDs).Delete(&models.CardShare{}).Error; err != nil {
				return fmt.Errorf("delete card shares (outgoing): %w", err)
			}
		}

		var voucherIDs []uuid.UUID
		tx.Model(&models.Voucher{}).Unscoped().Where("user_id = ?", userID).Pluck("id", &voucherIDs)
		if len(voucherIDs) > 0 {
			if err := tx.Unscoped().Where("voucher_id IN ?", voucherIDs).Delete(&models.VoucherShare{}).Error; err != nil {
				return fmt.Errorf("delete voucher shares (outgoing): %w", err)
			}
		}

		var giftCardIDs []uuid.UUID
		tx.Model(&models.GiftCard{}).Unscoped().Where("user_id = ?", userID).Pluck("id", &giftCardIDs)
		if len(giftCardIDs) > 0 {
			if err := tx.Unscoped().Where("gift_card_id IN ?", giftCardIDs).Delete(&models.GiftCardShare{}).Error; err != nil {
				return fmt.Errorf("delete gift card shares (outgoing): %w", err)
			}
			// Delete transactions of user's gift cards
			if err := tx.Unscoped().Where("gift_card_id IN ?", giftCardIDs).Delete(&models.GiftCardTransaction{}).Error; err != nil {
				return fmt.Errorf("delete gift card transactions: %w", err)
			}
		}

		// Stage 4: Delete owned resources
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.Card{}).Error; err != nil {
			return fmt.Errorf("delete cards: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.Voucher{}).Error; err != nil {
			return fmt.Errorf("delete vouchers: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.GiftCard{}).Error; err != nil {
			return fmt.Errorf("delete gift cards: %w", err)
		}

		// Stage 5: Delete email tokens
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.EmailToken{}).Error; err != nil {
			return fmt.Errorf("delete email tokens: %w", err)
		}

		// Stage 5a: Delete push subscriptions
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.PushSubscription{}).Error; err != nil {
			return fmt.Errorf("delete push subscriptions: %w", err)
		}

		// Stage 5b: Delete TOTP data (encrypted secrets, backup codes)
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.UserTOTP{}).Error; err != nil {
			return fmt.Errorf("delete TOTP data: %w", err)
		}

		// Stage 5c: Delete server-side sessions
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.Session{}).Error; err != nil {
			return fmt.Errorf("delete sessions: %w", err)
		}

		// Stage 5d: Delete expiry reminder tracking
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.ExpiryReminderSent{}).Error; err != nil {
			return fmt.Errorf("delete expiry reminders: %w", err)
		}

		// Stage 6: Create audit log entry before deleting user
		auditLog := &models.AuditLog{
			UserID:       &userID,
			Action:       "account_deleted",
			ResourceType: "users",
			ResourceID:   userID,
		}
		if err := tx.Create(auditLog).Error; err != nil {
			return fmt.Errorf("create audit log: %w", err)
		}

		// Stage 7: Nullify user references in audit logs (preserve audit trail)
		if err := tx.Model(&models.AuditLog{}).Where("user_id = ?", userID).Update("user_id", nil).Error; err != nil {
			return fmt.Errorf("nullify audit log user references: %w", err)
		}

		// Stage 8: Nullify created_by_user_id in remaining gift card transactions (from other users' gift cards)
		if err := tx.Model(&models.GiftCardTransaction{}).Where("created_by_user_id = ?", userID).Update("created_by_user_id", nil).Error; err != nil {
			return fmt.Errorf("nullify transaction user references: %w", err)
		}

		// Stage 9: Hard delete the user (GDPR)
		if err := tx.Unscoped().Where("id = ?", userID).Delete(&models.User{}).Error; err != nil {
			return fmt.Errorf("delete user: %w", err)
		}

		return nil
	})

	if err != nil {
		slog.Error("Account deletion failed", "user_id", userID, "error", err)
		return err
	}

	slog.Info("Account deleted successfully", "user_id", userID)

	// Send confirmation email (async, non-blocking)
	if s.emailService != nil {
		go func() {
			if sendErr := s.emailService.SendAccountDeletionConfirmation(context.Background(), userEmail, userName, userLanguage); sendErr != nil {
				slog.Error("Failed to send account deletion confirmation", "email", userEmail, "error", sendErr)
			}
		}()
	}

	return nil
}
