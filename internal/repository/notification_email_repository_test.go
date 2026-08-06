package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"savvy/internal/models"
	"savvy/internal/testutil"
)

// seedNotification inserts a notification with an explicit email delivery state.
func seedNotification(t *testing.T, db *gorm.DB, userID uuid.UUID, status models.EmailStatus) *models.Notification {
	t.Helper()
	n := &models.Notification{
		ID:           uuid.New(),
		UserID:       userID,
		Type:         models.NotificationTypeShareReceived,
		ResourceType: "card",
		ResourceID:   uuid.New(),
		Metadata:     models.NotificationMetadata{"from_user_name": "Tester"},
		EmailStatus:  status,
	}
	require.NoError(t, db.Create(n).Error)
	return n
}

func TestClaimPendingEmails_OnlyClaimsPending(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	pending := seedNotification(t, db, userID, models.EmailStatusPending)
	seedNotification(t, db, userID, models.EmailStatusSkipped)
	seedNotification(t, db, userID, models.EmailStatusSent)
	seedNotification(t, db, userID, models.EmailStatusFailed)

	claimed, err := repo.ClaimPendingEmails(ctx, 10)
	require.NoError(t, err)

	require.Len(t, claimed, 1, "only the pending row may be claimed")
	assert.Equal(t, pending.ID, claimed[0].ID)

	// The claim must be durable, otherwise a second dispatcher would send twice.
	var stored models.Notification
	require.NoError(t, db.First(&stored, "id = ?", pending.ID).Error)
	assert.Equal(t, models.EmailStatusSending, stored.EmailStatus)
}

func TestClaimPendingEmails_RespectsLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	for i := 0; i < 5; i++ {
		seedNotification(t, db, userID, models.EmailStatusPending)
	}

	claimed, err := repo.ClaimPendingEmails(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, claimed, 2)
}

func TestClaimPendingEmails_SecondClaimIsEmpty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	seedNotification(t, db, userID, models.EmailStatusPending)

	first, err := repo.ClaimPendingEmails(ctx, 10)
	require.NoError(t, err)
	require.Len(t, first, 1)

	second, err := repo.ClaimPendingEmails(ctx, 10)
	require.NoError(t, err)
	assert.Empty(t, second, "a claimed row must not be handed out twice")
}

func TestMarkEmailResult_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	n := seedNotification(t, db, userID, models.EmailStatusSending)

	require.NoError(t, repo.MarkEmailResult(ctx, n.ID, nil, 5))

	var stored models.Notification
	require.NoError(t, db.First(&stored, "id = ?", n.ID).Error)
	assert.Equal(t, models.EmailStatusSent, stored.EmailStatus)
	assert.Equal(t, 1, stored.EmailAttempts)
	assert.Nil(t, stored.EmailLastError)
}

func TestMarkEmailResult_FailureBelowLimitRetries(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	n := seedNotification(t, db, userID, models.EmailStatusSending)

	require.NoError(t, repo.MarkEmailResult(ctx, n.ID, errors.New("smtp timeout"), 5))

	var stored models.Notification
	require.NoError(t, db.First(&stored, "id = ?", n.ID).Error)
	// Back to pending is the whole point: the old code marked this sent and the
	// mail was never retried.
	assert.Equal(t, models.EmailStatusPending, stored.EmailStatus)
	assert.Equal(t, 1, stored.EmailAttempts)
	require.NotNil(t, stored.EmailLastError)
	assert.Contains(t, *stored.EmailLastError, "smtp timeout")
}

func TestMarkEmailResult_FailureAtLimitFails(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	n := seedNotification(t, db, userID, models.EmailStatusSending)
	// One attempt short of the limit.
	require.NoError(t, db.Model(&models.Notification{}).Where("id = ?", n.ID).
		Update("email_attempts", 4).Error)

	require.NoError(t, repo.MarkEmailResult(ctx, n.ID, errors.New("smtp rejected"), 5))

	var stored models.Notification
	require.NoError(t, db.First(&stored, "id = ?", n.ID).Error)
	assert.Equal(t, models.EmailStatusFailed, stored.EmailStatus)
	assert.Equal(t, 5, stored.EmailAttempts)
}

func TestResetStaleSendingEmails(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	stale := seedNotification(t, db, userID, models.EmailStatusSending)
	fresh := seedNotification(t, db, userID, models.EmailStatusSending)

	// Age one row past the recovery threshold.
	staleTime := time.Now().Add(-30 * time.Minute)
	require.NoError(t, db.Model(&models.Notification{}).Where("id = ?", stale.ID).
		UpdateColumn("updated_at", staleTime).Error)

	recovered, err := repo.ResetStaleSendingEmails(ctx, time.Now().Add(-10*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, int64(1), recovered)

	var staleStored, freshStored models.Notification
	require.NoError(t, db.First(&staleStored, "id = ?", stale.ID).Error)
	require.NoError(t, db.First(&freshStored, "id = ?", fresh.ID).Error)
	assert.Equal(t, models.EmailStatusPending, staleStored.EmailStatus, "abandoned row must return to the queue")
	assert.Equal(t, models.EmailStatusSending, freshStored.EmailStatus, "a row still being worked on must be left alone")
}

func TestCountPendingEmails(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, db)

	seedNotification(t, db, userID, models.EmailStatusPending)
	seedNotification(t, db, userID, models.EmailStatusPending)
	seedNotification(t, db, userID, models.EmailStatusSent)

	count, err := repo.CountPendingEmails(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// TestClaimPendingEmails_ConcurrentWorkersClaimDisjointSets is the reason
// FOR UPDATE SKIP LOCKED is there at all: production runs several replicas, and
// without it two of them would claim the same row and mail the user twice.
//
// This test needs NewTestDBDirect — the transaction-isolated helper cannot model
// concurrent sessions, since the competing workers would all sit inside one
// transaction and never actually contend for row locks.
func TestClaimPendingEmails_ConcurrentWorkersClaimDisjointSets(t *testing.T) {
	db := testutil.NewTestDBDirect(t)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        fmt.Sprintf("concurrent-%s@example.com", userID.String()[:8]),
		PasswordHash: "hashed",
		FirstName:    "Test",
		LastName:     "User",
	}
	require.NoError(t, db.Create(user).Error)

	const totalRows = 20
	for i := 0; i < totalRows; i++ {
		seedNotification(t, db, userID, models.EmailStatusPending)
	}

	const workers = 4
	var (
		mu       sync.Mutex
		claimed  = make(map[uuid.UUID]int)
		wg       sync.WaitGroup
		firstErr error
	)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repo := NewNotificationRepository(db)
			for {
				batch, err := repo.ClaimPendingEmails(ctx, 3)
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				if len(batch) == 0 {
					return
				}
				mu.Lock()
				for i := range batch {
					claimed[batch[i].ID]++
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	require.NoError(t, firstErr)
	assert.Len(t, claimed, totalRows, "every row must be claimed exactly once across all workers")
	for id, times := range claimed {
		assert.Equal(t, 1, times, "row %s was claimed %d times — that is a duplicate email", id, times)
	}

	var leftover int64
	require.NoError(t, db.Model(&models.Notification{}).
		Where("email_status = ?", models.EmailStatusPending).Count(&leftover).Error)
	assert.Zero(t, leftover, "no pending row may be left behind")
}
