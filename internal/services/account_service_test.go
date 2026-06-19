package services

import (
	"context"
	"testing"

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

// seedUserWithData creates a user owning one card (shared to a second user),
// one gift card with a transaction, a TOTP row, a session and a notification —
// i.e. at least one row in every table DeleteAccount touches. Returns the
// owner's ID and the recipient's ID.
func seedUserWithData(t *testing.T, db *gorm.DB) (owner, recipient uuid.UUID) {
	t.Helper()
	owner = uuid.New()
	recipient = uuid.New()

	require.NoError(t, db.Create(&models.User{ID: owner, Email: "owner-del@example.com", PasswordHash: "x", FirstName: "Own", LastName: "Er"}).Error)
	require.NoError(t, db.Create(&models.User{ID: recipient, Email: "recipient-del@example.com", PasswordHash: "x", FirstName: "Rec", LastName: "Ip"}).Error)

	card := &models.Card{UserID: &owner, CardNumber: "DEL-CARD", MerchantName: "M"}
	require.NoError(t, db.Create(card).Error)
	require.NoError(t, db.Create(&models.CardShare{CardID: card.ID, SharedWithID: recipient}).Error)

	gc := &models.GiftCard{UserID: &owner, CardNumber: "DEL-GC", MerchantName: "M", InitialBalance: 100, CurrentBalance: 100, Currency: "CHF"}
	require.NoError(t, db.Create(gc).Error)
	ownerRef := owner
	require.NoError(t, db.Create(&models.GiftCardTransaction{GiftCardID: gc.ID, Amount: -10, CreatedByUserID: &ownerRef}).Error)

	require.NoError(t, db.Create(&models.UserFavorite{UserID: owner, ResourceType: "card", ResourceID: card.ID}).Error)
	require.NoError(t, db.Create(&models.Notification{UserID: owner, Type: models.NotificationTypeShareReceived, ResourceType: "card", ResourceID: card.ID}).Error)
	require.NoError(t, db.Create(&models.UserTOTP{UserID: owner, Secret: "enc", Enabled: true}).Error)
	require.NoError(t, db.Create(&models.Session{UserID: &ownerRef, TokenHash: "h"}).Error)
	return owner, recipient
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
	owner, recipient := seedUserWithData(t, db)

	require.NoError(t, svc.DeleteAccount(context.Background(), owner))

	// User is hard-deleted.
	assert.Zero(t, countWhere(t, db, &models.User{}, "id = ?", owner), "user row")
	// Owned resources + their children gone.
	assert.Zero(t, countWhere(t, db, &models.Card{}, "user_id = ?", owner), "cards")
	assert.Zero(t, countWhere(t, db, &models.GiftCard{}, "user_id = ?", owner), "gift cards")
	assert.Zero(t, countWhere(t, db, &models.GiftCardTransaction{}, "gift_card_id IN (?)", db.Model(&models.GiftCard{}).Unscoped().Select("id").Where("user_id = ?", owner)), "gift card transactions")
	// Preferences gone.
	assert.Zero(t, countWhere(t, db, &models.UserFavorite{}, "user_id = ?", owner), "favorites")
	assert.Zero(t, countWhere(t, db, &models.Notification{}, "user_id = ?", owner), "notifications")
	// Auth data gone.
	assert.Zero(t, countWhere(t, db, &models.UserTOTP{}, "user_id = ?", owner), "totp")
	assert.Zero(t, countWhere(t, db, &models.Session{}, "user_id = ?", owner), "sessions")
	// Incoming share (owner's card shared to recipient) gone.
	assert.Zero(t, countWhere(t, db, &models.CardShare{}, "shared_with_id = ?", recipient), "incoming share of deleted card")

	// Recipient is untouched.
	assert.Equal(t, int64(1), countWhere(t, db, &models.User{}, "id = ?", recipient), "recipient survives")

	// Audit trail preserved but user reference nulled: the account_deleted
	// entry exists with a NULL user_id, not removed.
	var auditWithUser int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("user_id = ?", owner).Count(&auditWithUser).Error)
	assert.Zero(t, auditWithUser, "no audit log still references the deleted user")
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
	owner, _ := seedUserWithData(t, db)

	require.NoError(t, svc.DeleteAccount(context.Background(), owner))
	// Second run: user already gone → fails cleanly at the get-user stage,
	// no panic, no partial side effects.
	err := svc.DeleteAccount(context.Background(), owner)
	assert.Error(t, err, "re-running delete on an already-deleted account must error, not succeed")
}
