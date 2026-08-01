// Package services contains business logic.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"savvy/internal/audit"
	"savvy/internal/models"
	"savvy/internal/repository"

	"github.com/google/uuid"
)

// TransferServiceInterface defines the interface for ownership transfer business logic.
type TransferServiceInterface interface {
	TransferCardOwnership(ctx context.Context, cardID, newOwnerID, currentOwnerID uuid.UUID) error
	TransferVoucherOwnership(ctx context.Context, voucherID, newOwnerID, currentOwnerID uuid.UUID) error
	TransferGiftCardOwnership(ctx context.Context, giftCardID, newOwnerID, currentOwnerID uuid.UUID) error
}

// TransferService implements TransferServiceInterface.
type TransferService struct {
	cardRepo            repository.CardRepository
	voucherRepo         repository.VoucherRepository
	giftCardRepo        repository.GiftCardRepository
	userRepo            repository.UserRepository
	transferRepo        repository.TransferRepository
	auditLogRepo        repository.AuditLogRepository
	notificationService NotificationServiceInterface
}

// transferAuditData wraps resource data with transfer-specific info for audit logging
type transferAuditData struct {
	Resource      any       `json:"resource"`
	NewOwnerID    uuid.UUID `json:"new_owner_id"`
	NewOwnerEmail string    `json:"new_owner_email"`
}

// NewTransferService creates a new transfer service.
func NewTransferService(
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	userRepo repository.UserRepository,
	transferRepo repository.TransferRepository,
	auditLogRepo repository.AuditLogRepository,
	notificationService NotificationServiceInterface,
) TransferServiceInterface {
	return &TransferService{
		cardRepo:            cardRepo,
		voucherRepo:         voucherRepo,
		giftCardRepo:        giftCardRepo,
		userRepo:            userRepo,
		transferRepo:        transferRepo,
		auditLogRepo:        auditLogRepo,
		notificationService: notificationService,
	}
}

// validateNewOwner validates that the new owner exists and is different from current owner.
func (s *TransferService) validateNewOwner(ctx context.Context, newOwnerID, currentOwnerID uuid.UUID) (*models.User, error) {
	newOwner, err := s.userRepo.GetByID(ctx, newOwnerID)
	if err != nil {
		return nil, errors.New("new owner not found")
	}

	if newOwnerID == currentOwnerID {
		return nil, errors.New("cannot transfer to yourself")
	}

	return newOwner, nil
}

// sendTransferNotification sends a notification to the new owner (best effort).
// merchantName, description and value describe the transferred resource so the
// notification can name it; value is nil for resources without a value (Card).
func (s *TransferService) sendTransferNotification(ctx context.Context, resourceType string, resourceID, newOwnerID, currentOwnerID uuid.UUID, merchantName, description string, value *NotificationValue) {
	currentUser, err := s.userRepo.GetByID(ctx, currentOwnerID)
	if err == nil {
		if err := s.notificationService.CreateTransferNotification(ctx, TransferNotificationInput{
			RecipientID:  newOwnerID,
			FromUserID:   currentOwnerID,
			FromUserName: currentUser.DisplayName(),
			ResourceType: resourceType,
			ResourceID:   resourceID,
			MerchantName: merchantName,
			Description:  description,
			Value:        value,
		}); err != nil {
			slog.Warn("Failed to create transfer notification",
				"resource_type", resourceType, "resource_id", resourceID,
				"new_owner_id", newOwnerID, "error", err)
		}
	}
}

// logTransferAudit creates an audit log entry for a transfer operation
func (s *TransferService) logTransferAudit(ctx context.Context, currentOwnerID uuid.UUID, resourceType string, resourceID uuid.UUID, auditData transferAuditData) {
	ipAddress, userAgent := audit.ExtractAuditInfo(ctx)
	dataJSON, err := json.Marshal(auditData)
	if err != nil {
		slog.Warn("Failed to marshal transfer data for audit log", "resource_type", resourceType, "resource_id", resourceID, "error", err)
		dataJSON = []byte("{}")
	}
	if err := s.auditLogRepo.Create(ctx, &models.AuditLog{
		UserID: &currentOwnerID, Action: "transfer", ResourceType: resourceType,
		ResourceID: resourceID, ResourceData: string(dataJSON),
		IPAddress: ipAddress, UserAgent: userAgent,
	}); err != nil {
		slog.Warn("Failed to create audit log for transfer",
			"resource_type", resourceType, "resource_id", resourceID,
			"current_owner_id", currentOwnerID, "error", err)
	}
}

// TransferCardOwnership transfers ownership of a card to a new owner.
func (s *TransferService) TransferCardOwnership(ctx context.Context, cardID, newOwnerID, currentOwnerID uuid.UUID) error {
	newOwner, err := s.validateNewOwner(ctx, newOwnerID, currentOwnerID)
	if err != nil {
		return err
	}

	card, err := s.cardRepo.GetByID(ctx, cardID, "Merchant")
	if err != nil {
		return err
	}
	if card.UserID == nil || *card.UserID != currentOwnerID {
		return errors.New("only owner can transfer")
	}

	if err := s.transferRepo.TransferCardOwnership(ctx, card, newOwnerID); err != nil {
		return fmt.Errorf("transfer card ownership: %w", err)
	}

	s.logTransferAudit(ctx, currentOwnerID, "cards", cardID,
		transferAuditData{Resource: card, NewOwnerID: newOwnerID, NewOwnerEmail: newOwner.Email})

	// Description stays empty: Card.Notes is free-form and the only field where
	// a PIN or door code can live, so it must not reach the push body.
	s.sendTransferNotification(ctx, "card", cardID, newOwnerID, currentOwnerID, cardMerchantName(card), "", nil)
	return nil
}

// TransferVoucherOwnership transfers ownership of a voucher to a new owner.
func (s *TransferService) TransferVoucherOwnership(ctx context.Context, voucherID, newOwnerID, currentOwnerID uuid.UUID) error {
	newOwner, err := s.validateNewOwner(ctx, newOwnerID, currentOwnerID)
	if err != nil {
		return err
	}

	voucher, err := s.voucherRepo.GetByID(ctx, voucherID, "Merchant")
	if err != nil {
		return err
	}
	if voucher.UserID == nil || *voucher.UserID != currentOwnerID {
		return errors.New("only owner can transfer")
	}

	if err := s.transferRepo.TransferVoucherOwnership(ctx, voucher, newOwnerID); err != nil {
		return fmt.Errorf("transfer voucher ownership: %w", err)
	}

	s.logTransferAudit(ctx, currentOwnerID, "vouchers", voucherID,
		transferAuditData{Resource: voucher, NewOwnerID: newOwnerID, NewOwnerEmail: newOwner.Email})

	s.sendTransferNotification(ctx, "voucher", voucherID, newOwnerID, currentOwnerID, voucherMerchantName(voucher), voucher.Description, voucherNotificationValue(voucher))
	return nil
}

// TransferGiftCardOwnership transfers ownership of a gift card to a new owner.
func (s *TransferService) TransferGiftCardOwnership(ctx context.Context, giftCardID, newOwnerID, currentOwnerID uuid.UUID) error {
	newOwner, err := s.validateNewOwner(ctx, newOwnerID, currentOwnerID)
	if err != nil {
		return err
	}

	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID, "Merchant")
	if err != nil {
		return err
	}
	if giftCard.UserID == nil || *giftCard.UserID != currentOwnerID {
		return errors.New("only owner can transfer")
	}

	if err := s.transferRepo.TransferGiftCardOwnership(ctx, giftCard, newOwnerID); err != nil {
		return fmt.Errorf("transfer gift card ownership: %w", err)
	}

	s.logTransferAudit(ctx, currentOwnerID, "gift_cards", giftCardID,
		transferAuditData{Resource: giftCard, NewOwnerID: newOwnerID, NewOwnerEmail: newOwner.Email})

	// Same as Card: GiftCard.Notes sits next to the PIN field, so it stays out
	// of the push body.
	s.sendTransferNotification(ctx, "gift_card", giftCardID, newOwnerID, currentOwnerID, giftCardMerchantName(giftCard), "", giftCardNotificationValue(giftCard))
	return nil
}
