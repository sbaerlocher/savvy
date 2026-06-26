package api

// Integration tests for the cards handler that exercise the FULL stack —
// handler → real services → real repositories → real PostgreSQL — instead of
// mocks. They close the gap noted in the intake audit ("handler tests are
// mock-only, no cross-service/DB integration"): the mock tests verify each
// handler in isolation, these verify the wiring actually holds end to end
// (authorization against real rows, persistence, DTO mapping).
//
// They use testutil.NewTestDB, which runs each test in a transaction rolled
// back on cleanup, and skips automatically when no test database is reachable.

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/labstack/echo/v5"

	"savvy/internal/models"
	"savvy/internal/repository"
	"savvy/internal/services"
	"savvy/internal/testutil"
)

// newCardsHandlerWithRealStack wires a CardsHandler over real services and
// repositories backed by the given test DB — the same composition the
// production container builds, minus push/email side-channels not used by the
// card endpoints under test.
func newCardsHandlerWithRealStack(db *gorm.DB) *CardsHandler {
	cardRepo := repository.NewCardRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	giftCardRepo := repository.NewGiftCardRepository(db)
	merchantRepo := repository.NewMerchantRepository(db)
	userRepo := repository.NewUserRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	cardShareRepo := repository.NewCardShareRepository(db)
	voucherShareRepo := repository.NewVoucherShareRepository(db)
	giftCardShareRepo := repository.NewGiftCardShareRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	transferRepo := repository.NewTransferRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	notificationService := services.NewNotificationService(notificationRepo, userRepo)

	return NewCardsHandler(
		services.NewCardService(cardRepo),
		services.NewAuthzService(
			cardRepo, voucherRepo, giftCardRepo,
			cardShareRepo, voucherShareRepo, giftCardShareRepo,
		),
		services.NewMerchantService(merchantRepo),
		services.NewUserService(userRepo),
		services.NewFavoriteService(favoriteRepo, cardRepo, voucherRepo, giftCardRepo),
		services.NewShareService(
			db, cardRepo, voucherRepo, giftCardRepo,
			cardShareRepo, voucherShareRepo, giftCardShareRepo,
			userRepo, auditLogRepo, notificationService,
		),
		services.NewTransferService(
			cardRepo, voucherRepo, giftCardRepo,
			userRepo, transferRepo, auditLogRepo, notificationService,
		),
		services.NewAdminService(userRepo, auditLogRepo),
	)
}

// seedUser inserts a user and returns it.
func seedUser(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()
	u := &models.User{
		ID:        uuid.New(),
		Email:     email,
		FirstName: "Int",
		LastName:  "Test",
		Role:      "user",
	}
	require.NoError(t, db.Create(u).Error)
	return u
}

// seedCard inserts a card owned by ownerID and returns it.
func seedCard(t *testing.T, db *gorm.DB, ownerID uuid.UUID, number string) *models.Card {
	t.Helper()
	card := &models.Card{
		ID:           uuid.New(),
		UserID:       &ownerID,
		MerchantName: "Int Merchant",
		CardNumber:   number,
		BarcodeType:  "CODE128",
		Status:       "active",
	}
	require.NoError(t, db.Create(card).Error)
	return card
}

func TestCardsHandler_Integration_List(t *testing.T) {
	db := testutil.NewTestDB(t)
	handler := newCardsHandlerWithRealStack(db)

	user := seedUser(t, db, "int-list@example.com")
	seedCard(t, db, user.ID, "INT-LIST-1")
	seedCard(t, db, user.ID, "INT-LIST-2")

	c, rec := createTestContext(http.MethodGet, "/api/v1/cards", "")
	c.Set("current_user", user)

	require.NoError(t, handler.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp CardListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Cards, 2, "both owned cards round-trip through service+repo+DB")
}

func TestCardsHandler_Integration_Show_Owner(t *testing.T) {
	db := testutil.NewTestDB(t)
	handler := newCardsHandlerWithRealStack(db)

	user := seedUser(t, db, "int-show-owner@example.com")
	card := seedCard(t, db, user.ID, "INT-SHOW-1")

	c, rec := createTestContext(http.MethodGet, "/api/v1/cards/:id", "")
	c.Set("current_user", user)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: card.ID.String()}})

	require.NoError(t, handler.Show(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp CardDetailResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, card.ID.String(), resp.Card.ID)
	assert.True(t, resp.Permissions.IsOwner, "real AuthzService resolves ownership from the DB row")
}

func TestCardsHandler_Integration_Show_Forbidden(t *testing.T) {
	db := testutil.NewTestDB(t)
	handler := newCardsHandlerWithRealStack(db)

	owner := seedUser(t, db, "int-owner@example.com")
	stranger := seedUser(t, db, "int-stranger@example.com")
	card := seedCard(t, db, owner.ID, "INT-FORBIDDEN-1")

	c, rec := createTestContext(http.MethodGet, "/api/v1/cards/:id", "")
	c.Set("current_user", stranger) // not the owner, no share
	c.SetPathValues(echo.PathValues{{Name: "id", Value: card.ID.String()}})

	require.NoError(t, handler.Show(c))
	assert.Equal(t, http.StatusForbidden, rec.Code,
		"AuthzService denies a user with neither ownership nor a share")
}

func TestCardsHandler_Integration_Create_Persists(t *testing.T) {
	db := testutil.NewTestDB(t)
	handler := newCardsHandlerWithRealStack(db)

	user := seedUser(t, db, "int-create@example.com")

	body := `{"card_number":"INT-CREATE-1","new_merchant_name":"Created Merchant","barcode_type":"CODE128"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/cards", body)
	c.Set("current_user", user)

	require.NoError(t, handler.Create(c))
	assert.Equal(t, http.StatusCreated, rec.Code, "create returns 201; body: %s", rec.Body.String())

	// The card and its auto-created merchant must actually be in the DB.
	var count int64
	require.NoError(t, db.Model(&models.Card{}).
		Where("card_number = ? AND user_id = ?", "INT-CREATE-1", user.ID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count, "created card is persisted by the real repository")

	var merchantCount int64
	require.NoError(t, db.Model(&models.Merchant{}).
		Where("name = ?", "Created Merchant").Count(&merchantCount).Error)
	assert.Equal(t, int64(1), merchantCount, "new merchant is persisted in the same flow")
}
