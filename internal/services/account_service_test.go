package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"savvy/internal/models"
	"savvy/internal/repository"
)

// newAccountServiceForTest wires a real AccountService against the test DB.
// emailService is nil: DeleteAccount treats that as "skip confirmation mail",
// which keeps the test free of the async goroutine + SMTP.
func newAccountServiceForTest(db *gorm.DB) AccountServiceInterface {
	return NewAccountService(db, NewUserService(repository.NewUserRepository(db)), nil)
}

// seededIDs holds the row IDs created by seedUserWithData so assertions can
// target captured IDs rather than re-deriving them from rows that delete-cascade
// removes (which would make the assertion vacuous).
type seededIDs struct {
	owner, recipient  uuid.UUID
	cardID, voucherID uuid.UUID
	giftCardID        uuid.UUID
	giftCardTxID      uuid.UUID
}

// seedUserWithData creates a row in EVERY table DeleteAccount touches, owned by
// `owner`, plus a second `recipient` user that must survive. Outgoing shares
// (of the owner's resources) and incoming shares (to the owner) are both seeded.
func seedUserWithData(t *testing.T, db *gorm.DB) seededIDs {
	t.Helper()
	ids := seededIDs{owner: uuid.New(), recipient: uuid.New()}

	require.NoError(t, db.Create(&models.User{ID: ids.owner, Email: "owner-del@example.com", PasswordHash: "x", FirstName: "Own", LastName: "Er"}).Error)
	require.NoError(t, db.Create(&models.User{ID: ids.recipient, Email: "recipient-del@example.com", PasswordHash: "x", FirstName: "Rec", LastName: "Ip"}).Error)
	ownerRef := ids.owner

	// Owned resources.
	card := &models.Card{UserID: &ids.owner, CardNumber: "DEL-CARD", MerchantName: "M"}
	require.NoError(t, db.Create(card).Error)
	ids.cardID = card.ID

	voucher := &models.Voucher{UserID: &ids.owner, Code: "DEL-VOUCHER", Type: "fixed_amount", Value: 10, ValidFrom: time.Now(), ValidUntil: time.Now().Add(24 * time.Hour)}
	require.NoError(t, db.Create(voucher).Error)
	ids.voucherID = voucher.ID

	gc := &models.GiftCard{UserID: &ids.owner, CardNumber: "DEL-GC", MerchantName: "M", InitialBalance: 100, CurrentBalance: 100, Currency: "CHF"}
	require.NoError(t, db.Create(gc).Error)
	ids.giftCardID = gc.ID

	tx := &models.GiftCardTransaction{GiftCardID: gc.ID, Amount: -10, CreatedByUserID: &ownerRef}
	require.NoError(t, db.Create(tx).Error)
	ids.giftCardTxID = tx.ID

	// Outgoing shares (owner's resources shared TO recipient).
	require.NoError(t, db.Create(&models.CardShare{CardID: card.ID, SharedWithID: ids.recipient}).Error)
	require.NoError(t, db.Create(&models.VoucherShare{VoucherID: voucher.ID, SharedWithID: ids.recipient}).Error)
	require.NoError(t, db.Create(&models.GiftCardShare{GiftCardID: gc.ID, SharedWithID: ids.recipient}).Error)

	// Incoming shares (recipient's resources shared TO owner) — exercise the
	// "delete shares where shared_with_id = owner" stage.
	recipientCard := &models.Card{UserID: &ids.recipient, CardNumber: "REC-CARD", MerchantName: "M"}
	require.NoError(t, db.Create(recipientCard).Error)
	require.NoError(t, db.Create(&models.CardShare{CardID: recipientCard.ID, SharedWithID: ids.owner}).Error)

	// Preferences.
	require.NoError(t, db.Create(&models.UserFavorite{UserID: ids.owner, ResourceType: "card", ResourceID: card.ID}).Error)
	require.NoError(t, db.Create(&models.Notification{UserID: ids.owner, Type: models.NotificationTypeShareReceived, ResourceType: "card", ResourceID: card.ID}).Error)

	// Auth / device data.
	require.NoError(t, db.Create(&models.UserTOTP{UserID: ids.owner, Secret: "enc", Enabled: true}).Error)
	require.NoError(t, db.Create(&models.Session{UserID: &ownerRef, TokenHash: "session-hash", Data: []byte("{}")}).Error)
	require.NoError(t, db.Create(&models.EmailToken{UserID: ids.owner, TokenHash: "tok-hash", TokenType: "verify", ExpiresAt: time.Now().Add(time.Hour)}).Error)
	require.NoError(t, db.Create(&models.PushSubscription{UserID: ids.owner, Endpoint: "https://push.example/1", P256dhKey: "p", AuthKey: "a"}).Error)
	require.NoError(t, db.Create(&models.ExpiryReminderSent{UserID: ids.owner, ResourceType: "gift_card", ResourceID: gc.ID, DaysBefore: 7}).Error)

	return ids
}

func countByID(t *testing.T, db *gorm.DB, model interface{}, id uuid.UUID) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Unscoped().Model(model).Where("id = ?", id).Count(&n).Error)
	return n
}

func countWhere(t *testing.T, db *gorm.DB, model interface{}, query string, arg interface{}) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Unscoped().Model(model).Where(query, arg).Count(&n).Error)
	return n
}

func TestAccountService_DeleteAccount_RemovesAllUserData(t *testing.T) {
	db := setupDirectTestDB(t) // direct DB: DeleteAccount runs its own transaction + a goroutine guard
	svc := newAccountServiceForTest(db)
	ids := seedUserWithData(t, db)

	require.NoError(t, svc.DeleteAccount(context.Background(), ids.owner))

	// User hard-deleted.
	assert.Zero(t, countByID(t, db, &models.User{}, ids.owner), "user row")
	// Owned resources gone (by captured ID — survives cascade ordering).
	assert.Zero(t, countByID(t, db, &models.Card{}, ids.cardID), "card")
	assert.Zero(t, countByID(t, db, &models.Voucher{}, ids.voucherID), "voucher")
	assert.Zero(t, countByID(t, db, &models.GiftCard{}, ids.giftCardID), "gift card")
	// Gift card transaction gone — asserted by its OWN captured ID, not via a
	// subquery on the already-deleted gift_cards table (which would be vacuous).
	assert.Zero(t, countByID(t, db, &models.GiftCardTransaction{}, ids.giftCardTxID), "gift card transaction")
	// Preferences gone.
	assert.Zero(t, countWhere(t, db, &models.UserFavorite{}, "user_id = ?", ids.owner), "favorites")
	assert.Zero(t, countWhere(t, db, &models.Notification{}, "user_id = ?", ids.owner), "notifications")
	// Auth / device data gone.
	assert.Zero(t, countWhere(t, db, &models.UserTOTP{}, "user_id = ?", ids.owner), "totp")
	assert.Zero(t, countWhere(t, db, &models.Session{}, "user_id = ?", ids.owner), "sessions")
	assert.Zero(t, countWhere(t, db, &models.EmailToken{}, "user_id = ?", ids.owner), "email tokens")
	assert.Zero(t, countWhere(t, db, &models.PushSubscription{}, "user_id = ?", ids.owner), "push subscriptions")
	assert.Zero(t, countWhere(t, db, &models.ExpiryReminderSent{}, "user_id = ?", ids.owner), "expiry reminders")
	// Outgoing shares gone.
	assert.Zero(t, countWhere(t, db, &models.VoucherShare{}, "voucher_id = ?", ids.voucherID), "outgoing voucher share")
	assert.Zero(t, countWhere(t, db, &models.GiftCardShare{}, "gift_card_id = ?", ids.giftCardID), "outgoing gift card share")
	// Incoming share (recipient's card shared to owner) gone.
	assert.Zero(t, countWhere(t, db, &models.CardShare{}, "shared_with_id = ?", ids.owner), "incoming share to deleted user")

	// Recipient untouched.
	assert.Equal(t, int64(1), countByID(t, db, &models.User{}, ids.recipient), "recipient survives")

	// Audit trail: the account_deleted entry must SURVIVE (preserved trail)…
	var deletedEntries int64
	require.NoError(t, db.Model(&models.AuditLog{}).
		Where("action = ? AND resource_id = ?", "account_deleted", ids.owner).
		Count(&deletedEntries).Error)
	assert.Equal(t, int64(1), deletedEntries, "account_deleted audit entry is preserved")
	// …with its user_id nulled (no audit row references the deleted user).
	assert.Zero(t, countWhere(t, db, &models.AuditLog{}, "user_id = ?", ids.owner), "no audit row references the deleted user")
}

func TestAccountService_DeleteAccount_UnknownUser(t *testing.T) {
	db := setupDirectTestDB(t)
	svc := newAccountServiceForTest(db)

	err := svc.DeleteAccount(context.Background(), uuid.New())
	assert.Error(t, err, "deleting a non-existent user must error (get user stage), not silently succeed")
}

func TestAccountService_DeleteAccount_Idempotent(t *testing.T) {
	db := setupDirectTestDB(t)
	svc := newAccountServiceForTest(db)
	ids := seedUserWithData(t, db)

	require.NoError(t, svc.DeleteAccount(context.Background(), ids.owner))
	// Second run: user already gone → fails cleanly at the get-user stage,
	// no panic, no partial side effects.
	err := svc.DeleteAccount(context.Background(), ids.owner)
	assert.Error(t, err, "re-running delete on an already-deleted account must error, not succeed")
}
