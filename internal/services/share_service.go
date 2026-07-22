// Package services contains business logic.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"savvy/internal/audit"
	"savvy/internal/logsafe"
	"savvy/internal/models"
	"savvy/internal/repository"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotOwner is returned when caller is not the resource owner.
var ErrNotOwner = errors.New("caller is not the resource owner")

// ErrAlreadyShared is returned when a resource is already shared with the target user.
var ErrAlreadyShared = errors.New("already shared with this user")

// isDuplicateShareError checks if an error is a unique constraint violation for shares.
func isDuplicateShareError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unique_active")
}

// ShareServiceInterface defines the interface for share business logic.
type ShareServiceInterface interface {
	CreateCardShare(ctx context.Context, callerUserID, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error
	CreateVoucherShare(ctx context.Context, callerUserID, voucherID, sharedWithID uuid.UUID) error
	CreateGiftCardShare(ctx context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error
	GetCardShares(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error)
	GetVoucherShares(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error)
	GetGiftCardShares(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error)
	GetCardShareCounts(ctx context.Context, cardIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetVoucherShareCounts(ctx context.Context, voucherIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetGiftCardShareCounts(ctx context.Context, giftCardIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	UpdateCardShare(ctx context.Context, callerUserID, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error
	UpdateGiftCardShare(ctx context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error
	DeleteCardShare(ctx context.Context, callerUserID, cardID, sharedWithID uuid.UUID) error
	DeleteVoucherShare(ctx context.Context, callerUserID, voucherID, sharedWithID uuid.UUID) error
	DeleteGiftCardShare(ctx context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID) error
	DeleteAllCardShares(ctx context.Context, callerUserID, cardID uuid.UUID) error
	DeleteAllVoucherShares(ctx context.Context, callerUserID, voucherID uuid.UUID) error
	DeleteAllGiftCardShares(ctx context.Context, callerUserID, giftCardID uuid.UUID) error
	GetSharedUsers(ctx context.Context, userID uuid.UUID, searchQuery string) ([]models.User, error)
}

// ShareService implements ShareServiceInterface.
type ShareService struct {
	db                  *gorm.DB // for atomic ownership-verified share creation (defense-in-depth)
	cardRepo            repository.CardRepository
	voucherRepo         repository.VoucherRepository
	giftCardRepo        repository.GiftCardRepository
	cardShareRepo       repository.CardShareRepository
	voucherShareRepo    repository.VoucherShareRepository
	giftCardShareRepo   repository.GiftCardShareRepository
	userRepo            repository.UserRepository
	auditLogRepo        repository.AuditLogRepository
	notificationService NotificationServiceInterface
}

// NewShareService creates a new share service.
// db is used for atomic ownership-verified share creation (defense-in-depth against TOCTOU).
// Pass nil for db in unit tests that use mock repositories.
func NewShareService(
	db *gorm.DB,
	cardRepo repository.CardRepository,
	voucherRepo repository.VoucherRepository,
	giftCardRepo repository.GiftCardRepository,
	cardShareRepo repository.CardShareRepository,
	voucherShareRepo repository.VoucherShareRepository,
	giftCardShareRepo repository.GiftCardShareRepository,
	userRepo repository.UserRepository,
	auditLogRepo repository.AuditLogRepository,
	notificationService NotificationServiceInterface,
) ShareServiceInterface {
	return &ShareService{
		db:                  db,
		cardRepo:            cardRepo,
		voucherRepo:         voucherRepo,
		giftCardRepo:        giftCardRepo,
		cardShareRepo:       cardShareRepo,
		voucherShareRepo:    voucherShareRepo,
		giftCardShareRepo:   giftCardShareRepo,
		userRepo:            userRepo,
		auditLogRepo:        auditLogRepo,
		notificationService: notificationService,
	}
}

// CreateCardShare creates a new card share. callerUserID must be the resource owner.
func (s *ShareService) CreateCardShare(ctx context.Context, callerUserID, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	if cardID == uuid.Nil {
		return errors.New("card ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("card not found")
		}
		return err
	}

	// Defense-in-depth: verify caller is the resource owner
	if card.UserID == nil || *card.UserID != callerUserID {
		return ErrNotOwner
	}

	if card.UserID != nil && *card.UserID == sharedWithID {
		return errors.New("cannot share card with its owner")
	}

	share := models.CardShare{
		CardID:       cardID,
		SharedWithID: sharedWithID,
		CanEdit:      canEdit,
		CanDelete:    canDelete,
	}

	// Atomic share creation with ownership re-verification (defense-in-depth against TOCTOU).
	// If ownership changed between the check above and this INSERT, RowsAffected == 0.
	if s.db != nil {
		result := s.db.WithContext(ctx).Exec(`
			INSERT INTO card_shares (id, card_id, shared_with_id, can_edit, can_delete, created_at, updated_at)
			SELECT gen_random_uuid(), ?, ?, ?, ?, NOW(), NOW()
			WHERE EXISTS (SELECT 1 FROM cards WHERE id = ? AND user_id = ? AND deleted_at IS NULL)`,
			cardID, sharedWithID, canEdit, canDelete, cardID, callerUserID)
		if result.Error != nil {
			if isDuplicateShareError(result.Error) {
				return ErrAlreadyShared
			}
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotOwner
		}
	} else {
		if err := s.cardShareRepo.Create(ctx, &share); err != nil {
			if isDuplicateShareError(err) {
				return ErrAlreadyShared
			}
			return err
		}
	}

	// Create notification for the shared user
	slog.Info("Attempting to create share notification",
		"card_id", cardID,
		"shared_with_id", sharedWithID,
		"has_owner", card.UserID != nil)

	if card.UserID != nil {
		ownerUser, err := s.userRepo.GetByID(ctx, *card.UserID)
		if err == nil {
			slog.Info("Owner user found, creating notification",
				"owner_id", *card.UserID,
				"owner_name", ownerUser.DisplayName())

			if err := s.notificationService.CreateShareNotification(
				ctx, sharedWithID, *card.UserID, ownerUser.DisplayName(),
				"card", cardID,
				map[string]bool{"can_edit": canEdit, "can_delete": canDelete},
			); err != nil {
				slog.Warn("Failed to create share notification for card",
					"card_id", cardID, "shared_with_id", sharedWithID, "error", err)
			} else {
				slog.Info("Share notification created successfully",
					"card_id", cardID, "shared_with_id", sharedWithID)
			}
		} else {
			slog.Warn("Owner user not found", "owner_id", *card.UserID, "error", err)
		}
	} else {
		slog.Warn("Card has no owner, skipping notification", "card_id", cardID)
	}

	return nil
}

// CreateVoucherShare creates a new voucher share. callerUserID must be the resource owner.
func (s *ShareService) CreateVoucherShare(ctx context.Context, callerUserID, voucherID, sharedWithID uuid.UUID) error {
	if voucherID == uuid.Nil {
		return errors.New("voucher ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("voucher not found")
		}
		return err
	}

	// Defense-in-depth: verify caller is the resource owner
	if voucher.UserID == nil || *voucher.UserID != callerUserID {
		return ErrNotOwner
	}

	if voucher.UserID != nil && *voucher.UserID == sharedWithID {
		return errors.New("cannot share voucher with its owner")
	}

	share := models.VoucherShare{
		VoucherID:    voucherID,
		SharedWithID: sharedWithID,
	}

	// Atomic share creation with ownership re-verification (defense-in-depth against TOCTOU).
	if s.db != nil {
		result := s.db.WithContext(ctx).Exec(`
			INSERT INTO voucher_shares (id, voucher_id, shared_with_id, created_at, updated_at)
			SELECT gen_random_uuid(), ?, ?, NOW(), NOW()
			WHERE EXISTS (SELECT 1 FROM vouchers WHERE id = ? AND user_id = ? AND deleted_at IS NULL)`,
			voucherID, sharedWithID, voucherID, callerUserID)
		if result.Error != nil {
			if isDuplicateShareError(result.Error) {
				return ErrAlreadyShared
			}
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotOwner
		}
	} else {
		if err := s.voucherShareRepo.Create(ctx, &share); err != nil {
			if isDuplicateShareError(err) {
				return ErrAlreadyShared
			}
			return err
		}
	}

	if voucher.UserID != nil {
		ownerUser, err := s.userRepo.GetByID(ctx, *voucher.UserID)
		if err == nil {
			if err := s.notificationService.CreateShareNotification(
				ctx, sharedWithID, *voucher.UserID, ownerUser.DisplayName(),
				"voucher", voucherID, nil,
			); err != nil {
				slog.Warn("Failed to create share notification for voucher",
					"voucher_id", voucherID, "shared_with_id", sharedWithID, "error", err)
			}
		}
	}

	return nil
}

// CreateGiftCardShare creates a new gift card share. callerUserID must be the resource owner.
func (s *ShareService) CreateGiftCardShare(ctx context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	if giftCardID == uuid.Nil {
		return errors.New("gift card ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("gift card not found")
		}
		return err
	}

	// Defense-in-depth: verify caller is the resource owner
	if giftCard.UserID == nil || *giftCard.UserID != callerUserID {
		return ErrNotOwner
	}

	if giftCard.UserID != nil && *giftCard.UserID == sharedWithID {
		return errors.New("cannot share gift card with its owner")
	}

	share := models.GiftCardShare{
		GiftCardID:          giftCardID,
		SharedWithID:        sharedWithID,
		CanEdit:             canEdit,
		CanDelete:           canDelete,
		CanEditTransactions: canEditTransactions,
	}

	// Atomic share creation with ownership re-verification (defense-in-depth against TOCTOU).
	if s.db != nil {
		result := s.db.WithContext(ctx).Exec(`
			INSERT INTO gift_card_shares (id, gift_card_id, shared_with_id, can_edit, can_delete, can_edit_transactions, created_at, updated_at)
			SELECT gen_random_uuid(), ?, ?, ?, ?, ?, NOW(), NOW()
			WHERE EXISTS (SELECT 1 FROM gift_cards WHERE id = ? AND user_id = ? AND deleted_at IS NULL)`,
			giftCardID, sharedWithID, canEdit, canDelete, canEditTransactions, giftCardID, callerUserID)
		if result.Error != nil {
			if isDuplicateShareError(result.Error) {
				return ErrAlreadyShared
			}
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotOwner
		}
	} else {
		if err := s.giftCardShareRepo.Create(ctx, &share); err != nil {
			if isDuplicateShareError(err) {
				return ErrAlreadyShared
			}
			return err
		}
	}

	if giftCard.UserID != nil {
		ownerUser, err := s.userRepo.GetByID(ctx, *giftCard.UserID)
		if err == nil {
			if err := s.notificationService.CreateShareNotification(
				ctx, sharedWithID, *giftCard.UserID, ownerUser.DisplayName(),
				"gift_card", giftCardID,
				map[string]bool{
					"can_edit": canEdit, "can_delete": canDelete,
					"can_edit_transactions": canEditTransactions,
				},
			); err != nil {
				slog.Warn("Failed to create share notification for gift card",
					"gift_card_id", giftCardID, "shared_with_id", sharedWithID, "error", err)
			}
		}
	}

	return nil
}

// GetCardShares retrieves all active (non-deleted) shares for a card.
func (s *ShareService) GetCardShares(ctx context.Context, cardID uuid.UUID) ([]models.CardShare, error) {
	return s.cardShareRepo.GetByCardID(ctx, cardID)
}

// GetVoucherShares retrieves all active (non-deleted) shares for a voucher.
func (s *ShareService) GetVoucherShares(ctx context.Context, voucherID uuid.UUID) ([]models.VoucherShare, error) {
	return s.voucherShareRepo.GetByVoucherID(ctx, voucherID)
}

// GetGiftCardShares retrieves all active (non-deleted) shares for a gift card.
func (s *ShareService) GetGiftCardShares(ctx context.Context, giftCardID uuid.UUID) ([]models.GiftCardShare, error) {
	return s.giftCardShareRepo.GetByGiftCardID(ctx, giftCardID)
}

// GetCardShareCounts returns the number of shares per card for the given card IDs.
func (s *ShareService) GetCardShareCounts(ctx context.Context, cardIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return s.cardShareRepo.CountByCardIDs(ctx, cardIDs)
}

// GetVoucherShareCounts returns the number of shares per voucher for the given voucher IDs.
func (s *ShareService) GetVoucherShareCounts(ctx context.Context, voucherIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return s.voucherShareRepo.CountByVoucherIDs(ctx, voucherIDs)
}

// GetGiftCardShareCounts returns the number of shares per gift card for the given gift card IDs.
func (s *ShareService) GetGiftCardShareCounts(ctx context.Context, giftCardIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return s.giftCardShareRepo.CountByGiftCardIDs(ctx, giftCardIDs)
}

// GetSharedUsers retrieves all unique users that the given user has shared resources with.
func (s *ShareService) GetSharedUsers(ctx context.Context, userID uuid.UUID, searchQuery string) ([]models.User, error) {
	cardUserIDs, _ := s.cardShareRepo.GetSharedUserIDs(ctx, userID)
	voucherUserIDs, _ := s.voucherShareRepo.GetSharedUserIDs(ctx, userID)
	giftCardUserIDs, _ := s.giftCardShareRepo.GetSharedUserIDs(ctx, userID)

	uniqueMap := make(map[uuid.UUID]bool)
	var allIDs []uuid.UUID
	for _, ids := range [][]uuid.UUID{cardUserIDs, voucherUserIDs, giftCardUserIDs} {
		for _, id := range ids {
			if !uniqueMap[id] {
				uniqueMap[id] = true
				allIDs = append(allIDs, id)
			}
		}
	}

	if len(allIDs) == 0 {
		return []models.User{}, nil
	}

	return s.userRepo.SearchByIDs(ctx, allIDs, searchQuery)
}

// DeleteCardShare removes a card share by cardID and sharedWithID. callerUserID must be the resource owner.
func (s *ShareService) DeleteCardShare(ctx context.Context, callerUserID, cardID, sharedWithID uuid.UUID) error {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("card not found")
		}
		return err
	}
	if card.UserID == nil || *card.UserID != callerUserID {
		return ErrNotOwner
	}
	return s.cardShareRepo.DeleteByCardAndUser(ctx, cardID, sharedWithID)
}

// DeleteVoucherShare removes a voucher share by voucherID and sharedWithID. callerUserID must be the resource owner.
func (s *ShareService) DeleteVoucherShare(ctx context.Context, callerUserID, voucherID, sharedWithID uuid.UUID) error {
	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("voucher not found")
		}
		return err
	}
	if voucher.UserID == nil || *voucher.UserID != callerUserID {
		return ErrNotOwner
	}
	return s.voucherShareRepo.DeleteByVoucherAndUser(ctx, voucherID, sharedWithID)
}

// DeleteGiftCardShare removes a gift card share by giftCardID and sharedWithID. callerUserID must be the resource owner.
func (s *ShareService) DeleteGiftCardShare(ctx context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID) error {
	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("gift card not found")
		}
		return err
	}
	if giftCard.UserID == nil || *giftCard.UserID != callerUserID {
		return ErrNotOwner
	}
	return s.giftCardShareRepo.DeleteByGiftCardAndUser(ctx, giftCardID, sharedWithID)
}

// DeleteAllCardShares removes all shares for a card. callerUserID must be the resource owner.
func (s *ShareService) DeleteAllCardShares(ctx context.Context, callerUserID, cardID uuid.UUID) error {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("card not found")
		}
		return err
	}
	if card.UserID == nil || *card.UserID != callerUserID {
		return ErrNotOwner
	}
	return s.cardShareRepo.DeleteByCardID(ctx, cardID)
}

// DeleteAllVoucherShares removes all shares for a voucher. callerUserID must be the resource owner.
func (s *ShareService) DeleteAllVoucherShares(ctx context.Context, callerUserID, voucherID uuid.UUID) error {
	voucher, err := s.voucherRepo.GetByID(ctx, voucherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("voucher not found")
		}
		return err
	}
	if voucher.UserID == nil || *voucher.UserID != callerUserID {
		return ErrNotOwner
	}
	return s.voucherShareRepo.DeleteByVoucherID(ctx, voucherID)
}

// DeleteAllGiftCardShares removes all shares for a gift card. callerUserID must be the resource owner.
func (s *ShareService) DeleteAllGiftCardShares(ctx context.Context, callerUserID, giftCardID uuid.UUID) error {
	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("gift card not found")
		}
		return err
	}
	if giftCard.UserID == nil || *giftCard.UserID != callerUserID {
		return ErrNotOwner
	}
	return s.giftCardShareRepo.DeleteByGiftCardID(ctx, giftCardID)
}

// UpdateCardShare updates card share permissions. callerUserID must be the resource owner.
func (s *ShareService) UpdateCardShare(ctx context.Context, callerUserID, cardID, sharedWithID uuid.UUID, canEdit, canDelete bool) error {
	if cardID == uuid.Nil {
		return errors.New("card ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	share, err := s.cardShareRepo.GetByCardAndUser(ctx, cardID, sharedWithID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("share not found")
		}
		return err
	}

	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}

	// Defense-in-depth: verify caller is the resource owner
	if card.UserID == nil || *card.UserID != callerUserID {
		return ErrNotOwner
	}

	if card.UserID != nil {
		ipAddress, userAgent := audit.ExtractAuditInfo(ctx)
		dataJSON, err := json.Marshal(share)
		if err != nil {
			slog.Warn("Failed to marshal share data for audit log", "share_id", share.ID, "error", err)
			dataJSON = []byte("{}")
		}
		if err := s.auditLogRepo.Create(ctx, &models.AuditLog{
			UserID: card.UserID, Action: "update", ResourceType: "card_shares",
			ResourceID: share.ID, ResourceData: string(dataJSON),
			IPAddress: ipAddress, UserAgent: userAgent,
		}); err != nil {
			slog.Warn("Failed to create audit log for card share update",
				"card_id", cardID, "share_id", share.ID, "shared_with_id", sharedWithID, "error", err)
		}
	}

	share.CanEdit = canEdit
	share.CanDelete = canDelete

	if err := s.cardShareRepo.Update(ctx, share); err != nil {
		return err
	}

	slog.Info("Card share permissions updated",
		"card_id", logsafe.UUID(cardID), "shared_with_id", logsafe.UUID(sharedWithID),
		"can_edit", canEdit, "can_delete", canDelete)

	return nil
}

// UpdateGiftCardShare updates gift card share permissions. callerUserID must be the resource owner.
func (s *ShareService) UpdateGiftCardShare(ctx context.Context, callerUserID, giftCardID, sharedWithID uuid.UUID, canEdit, canDelete, canEditTransactions bool) error {
	if giftCardID == uuid.Nil {
		return errors.New("gift card ID is required")
	}
	if sharedWithID == uuid.Nil {
		return errors.New("shared with user ID is required")
	}

	share, err := s.giftCardShareRepo.GetByGiftCardAndUser(ctx, giftCardID, sharedWithID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("share not found")
		}
		return err
	}

	giftCard, err := s.giftCardRepo.GetByID(ctx, giftCardID)
	if err != nil {
		return err
	}

	// Defense-in-depth: verify caller is the resource owner
	if giftCard.UserID == nil || *giftCard.UserID != callerUserID {
		return ErrNotOwner
	}

	if giftCard.UserID != nil {
		ipAddress, userAgent := audit.ExtractAuditInfo(ctx)
		dataJSON, err := json.Marshal(share)
		if err != nil {
			slog.Warn("Failed to marshal share data for audit log", "share_id", share.ID, "error", err)
			dataJSON = []byte("{}")
		}
		if err := s.auditLogRepo.Create(ctx, &models.AuditLog{
			UserID: giftCard.UserID, Action: "update", ResourceType: "gift_card_shares",
			ResourceID: share.ID, ResourceData: string(dataJSON),
			IPAddress: ipAddress, UserAgent: userAgent,
		}); err != nil {
			slog.Warn("Failed to create audit log for gift card share update",
				"gift_card_id", giftCardID, "share_id", share.ID, "shared_with_id", sharedWithID, "error", err)
		}
	}

	share.CanEdit = canEdit
	share.CanDelete = canDelete
	share.CanEditTransactions = canEditTransactions

	if err := s.giftCardShareRepo.Update(ctx, share); err != nil {
		return err
	}

	slog.Info("Gift card share permissions updated",
		"gift_card_id", logsafe.UUID(giftCardID), "shared_with_id", logsafe.UUID(sharedWithID),
		"can_edit", canEdit, "can_delete", canDelete, "can_edit_transactions", canEditTransactions)

	return nil
}
