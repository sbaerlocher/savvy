# Soft-Delete Duplicate-Check Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow re-creating a card, voucher, or gift card whose number matches a soft-deleted one the user owns, offering restore instead of a 500.

**Architecture:** Partial unique indexes exclude soft-deleted rows so the INSERT succeeds. The create handler additionally looks for a soft-deleted twin (Unscoped, user-scoped) and returns a `409` with `deleted:true` so the frontend can offer restore via a new `POST /:id/restore` endpoint.

**Tech Stack:** Go + Echo + GORM + gormigrate + PostgreSQL; SvelteKit + TypeScript.

## Global Constraints

- All development runs in Docker (`docker compose up`); do not start local dev servers. For Go builds/tests, `go build ./...` / `go test` on the host is acceptable (Go toolchain present).
- Duplicate-check and restore are ALWAYS per user: `user_id = ? AND deleted_at IS NOT NULL`. Never global, never cross-user.
- Restore = pure `deleted_at = NULL`. Old data returns unchanged; newly entered form values are discarded.
- Migrations run via gormigrate (`internal/setup/dependencies.go`), NOT `database.AutoMigrate()` (test-only). Keep model tags consistent with migration schema.
- Commit convention: `entwicklung:git-conventions` (English, repo is PUBLIC), `git commit -S --signoff`, no Claude trailers.
- Existing migration helpers: `createIndex(tx, sql)`, `dropIndex(tx, name)`, `addComment(tx, sql)` in `internal/migrations/migrations.go`.

---

### Task 1: Migration — partial unique indexes excluding soft-deleted rows

**Files:**
- Create: `internal/migrations/202607040031_partial_unique_indexes_exclude_soft_deleted.go`
- Modify: `internal/migrations/migrations.go` (append to `GetMigrations()` slice)

**Interfaces:**
- Consumes: `createIndex`, `dropIndex` helpers (already in `migrations.go`).
- Produces: migration func `partialUniqueIndexesExcludeSoftDeleted() *gormigrate.Migration`.

The three current indexes (from migration `202601250008`) are:
`idx_cards_user_card_number ON cards (user_id, card_number) WHERE user_id IS NOT NULL`,
`idx_vouchers_user_code ON vouchers (user_id, code) WHERE user_id IS NOT NULL`,
`idx_gift_cards_user_card_number ON gift_cards (user_id, card_number) WHERE user_id IS NOT NULL`.

- [ ] **Step 1: Write the migration file**

```go
package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// partialUniqueIndexesExcludeSoftDeleted rewrites the composite unique indexes
// on (user_id, card_number/code) to also exclude soft-deleted rows, so a user
// can re-create an entry whose number matches one they previously soft-deleted.
// Migration 000031 - 2026-07-04
func partialUniqueIndexesExcludeSoftDeleted() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607040031_partial_unique_indexes_exclude_soft_deleted",
		Migrate: func(tx *gorm.DB) error {
			steps := []struct{ drop, create string }{
				{"idx_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_user_card_number ON cards (user_id, card_number) WHERE user_id IS NOT NULL AND deleted_at IS NULL"},
				{"idx_vouchers_user_code",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_vouchers_user_code ON vouchers (user_id, code) WHERE user_id IS NOT NULL AND deleted_at IS NULL"},
				{"idx_gift_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_user_card_number ON gift_cards (user_id, card_number) WHERE user_id IS NOT NULL AND deleted_at IS NULL"},
			}
			for _, s := range steps {
				if err := dropIndex(tx, s.drop); err != nil {
					return err
				}
				if err := createIndex(tx, s.create); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			steps := []struct{ drop, create string }{
				{"idx_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_user_card_number ON cards (user_id, card_number) WHERE user_id IS NOT NULL"},
				{"idx_vouchers_user_code",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_vouchers_user_code ON vouchers (user_id, code) WHERE user_id IS NOT NULL"},
				{"idx_gift_cards_user_card_number",
					"CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_user_card_number ON gift_cards (user_id, card_number) WHERE user_id IS NOT NULL"},
			}
			for _, s := range steps {
				if err := dropIndex(tx, s.drop); err != nil {
					return err
				}
				if err := createIndex(tx, s.create); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
```

- [ ] **Step 2: Register the migration**

In `internal/migrations/migrations.go`, add as the LAST entry in the `GetMigrations()` slice, after `fixNotificationPreferenceDefaults(),`:

```go
		fixNotificationPreferenceDefaults(),
		partialUniqueIndexesExcludeSoftDeleted(),
	}
}
```

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: success, no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/migrations/202607040031_partial_unique_indexes_exclude_soft_deleted.go internal/migrations/migrations.go
git commit -S --signoff -m "fix(migrations): exclude soft-deleted rows from unique indexes"
```

---

### Task 2: Align model tags with the partial composite index

**Files:**
- Modify: `internal/models/card.go:20` (`CardNumber`)
- Modify: `internal/models/voucher.go:19` (`Code`)
- Modify: `internal/models/gift_card.go:20` (`CardNumber`)

**Interfaces:**
- Consumes: nothing.
- Produces: model tags that make test-time `database.AutoMigrate()` build the same partial composite index as Task 1's migration.

Current tags use a bare global `uniqueIndex` which only affects test AutoMigrate (prod uses gormigrate) and builds a global index prod does not have.

- [ ] **Step 1: Update Card tag**

`internal/models/card.go` line 20, replace:

```go
	CardNumber   string         `gorm:"uniqueIndex;not null" json:"card_number"`
```

with:

```go
	CardNumber   string         `gorm:"uniqueIndex:idx_cards_user_card_number,where:user_id IS NOT NULL AND deleted_at IS NULL,composite:user_card;not null" json:"card_number"`
```

Note: GORM composite unique indexes need both columns tagged. Also tag `UserID` on line 14. Replace:

```go
	UserID       *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
```

with:

```go
	UserID       *uuid.UUID     `gorm:"type:uuid;index;uniqueIndex:idx_cards_user_card_number,where:user_id IS NOT NULL AND deleted_at IS NULL,composite:user_card,priority:1" json:"user_id"`
```

And set `priority:2` on `CardNumber`:

```go
	CardNumber   string         `gorm:"uniqueIndex:idx_cards_user_card_number,where:user_id IS NOT NULL AND deleted_at IS NULL,composite:user_card,priority:2;not null" json:"card_number"`
```

- [ ] **Step 2: Update Voucher tag**

`internal/models/voucher.go`: apply the same pattern with index name `idx_vouchers_user_code`, tagging `UserID` (priority:1) and `Code` (priority:2, line 19). Read the file first to get `UserID`'s exact current tag, then add the `uniqueIndex:idx_vouchers_user_code,where:user_id IS NOT NULL AND deleted_at IS NULL,composite:user_code,priority:1` clause to `UserID` and `uniqueIndex:idx_vouchers_user_code,...,priority:2` to `Code`, removing the bare `uniqueIndex` from `Code`.

- [ ] **Step 3: Update GiftCard tag**

`internal/models/gift_card.go`: same pattern with index name `idx_gift_cards_user_card_number`, tagging `UserID` (priority:1) and `CardNumber` (priority:2, line 20).

- [ ] **Step 4: Build + run the AutoMigrate test**

Run: `go build ./... && go test ./internal/database/ -run TestAutoMigrate -count=1`
Expected: PASS. If GORM's composite tag syntax fails to parse, fall back to a bare named `uniqueIndex:idx_..._user_...` on both columns without `where:` (AutoMigrate is test-only; the migration in Task 1 is the source of truth for prod). Document the fallback with a `// ponytail:` comment.

- [ ] **Step 5: Commit**

```bash
git add internal/models/card.go internal/models/voucher.go internal/models/gift_card.go
git commit -S --signoff -m "fix(models): align unique index tags with per-user partial index"
```

---

### Task 3: Repository — find soft-deleted twin + restore (Cards)

**Files:**
- Modify: `internal/repository/card_repository.go` (interface `CardRepository`)
- Modify: `internal/repository/card_repository_impl.go` (`GormCardRepository`)
- Test: `internal/repository/card_repository_test.go`

**Interfaces:**
- Consumes: `models.Card`, `GormCardRepository` (receiver `r`, field `r.db`).
- Produces:
  - `FindDeletedByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error)` — Unscoped, `deleted_at IS NOT NULL`, user-scoped; returns nil if none.
  - `RestoreByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error` — Unscoped update `deleted_at = NULL` where `id` and `user_id` match and row is soft-deleted.

- [ ] **Step 1: Write failing test**

Add to `internal/repository/card_repository_test.go` (follow the existing test setup in that file for DB + repo construction):

```go
func TestCardRepository_FindDeletedByCardNumber(t *testing.T) {
	db := setupTestDB(t) // use whatever helper the existing tests use
	repo := NewGormCardRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	card := &models.Card{UserID: &userID, Program: "P", CardNumber: "DEL-1"}
	require.NoError(t, repo.Create(ctx, card))
	require.NoError(t, repo.Delete(ctx, card.ID)) // soft-delete

	// Active lookup does not see it
	active, err := repo.FindByCardNumber(ctx, "DEL-1", userID)
	require.NoError(t, err)
	require.Nil(t, active)

	// Deleted lookup finds it
	found, err := repo.FindDeletedByCardNumber(ctx, "DEL-1", userID)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, card.ID, found.ID)

	// Restore brings it back
	require.NoError(t, repo.RestoreByID(ctx, card.ID, userID))
	active2, err := repo.FindByCardNumber(ctx, "DEL-1", userID)
	require.NoError(t, err)
	require.NotNil(t, active2)
}
```

Match import style and `setupTestDB`/`NewGormCardRepository` names to the existing test file — read it first.

- [ ] **Step 2: Run test, verify it fails**

Run: `go test ./internal/repository/ -run TestCardRepository_FindDeletedByCardNumber -count=1`
Expected: FAIL — `FindDeletedByCardNumber`/`RestoreByID` undefined.

- [ ] **Step 3: Add interface methods**

In `internal/repository/card_repository.go`, inside `CardRepository` interface, after `FindByCardNumber`:

```go
	// FindDeletedByCardNumber finds a soft-deleted card by card number for a specific user
	FindDeletedByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error)

	// RestoreByID clears deleted_at for a soft-deleted card owned by the user
	RestoreByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
```

- [ ] **Step 4: Implement methods**

In `internal/repository/card_repository_impl.go`, after the existing `FindByCardNumber` implementation:

```go
// FindDeletedByCardNumber finds a soft-deleted card by card number for a specific user.
func (r *GormCardRepository) FindDeletedByCardNumber(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error) {
	var card models.Card
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("card_number = ? AND user_id = ? AND deleted_at IS NOT NULL", cardNumber, userID).
		First(&card).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &card, nil
}

// RestoreByID clears deleted_at for a soft-deleted card owned by the user.
func (r *GormCardRepository) RestoreByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&models.Card{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil).Error
}
```

- [ ] **Step 5: Run test, verify pass**

Run: `go test ./internal/repository/ -run TestCardRepository_FindDeletedByCardNumber -count=1`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/repository/card_repository.go internal/repository/card_repository_impl.go internal/repository/card_repository_test.go
git commit -S --signoff -m "feat(repository): add deleted-twin lookup and restore for cards"
```

---

### Task 4: Repository — find soft-deleted twin + restore (Vouchers + Gift Cards)

**Files:**
- Modify: `internal/repository/voucher_repository.go`, `internal/repository/voucher_repository_impl.go`
- Modify: `internal/repository/gift_card_repository.go`, `internal/repository/gift_card_repository_impl.go`
- Test: `internal/repository/voucher_repository_test.go`, `internal/repository/gift_card_repository_test.go`

**Interfaces:**
- Produces (Voucher, receiver `GormVoucherRepository`):
  - `FindDeletedByCode(ctx, code string, userID uuid.UUID) (*models.Voucher, error)`
  - `RestoreByID(ctx, id uuid.UUID, userID uuid.UUID) error`
- Produces (GiftCard, receiver `GormGiftCardRepository`):
  - `FindDeletedByCardNumber(ctx, cardNumber string, userID uuid.UUID) (*models.GiftCard, error)`
  - `RestoreByID(ctx, id uuid.UUID, userID uuid.UUID) error`

- [ ] **Step 1: Write failing tests**

Mirror Task 3's test for each: voucher uses field `Code` and lookup `FindDeletedByCode`; gift card uses `CardNumber` and `FindDeletedByCardNumber`. Build minimal valid models (voucher needs `ValidFrom`/`ValidUntil`/`Currency`; read the model to get required fields). Follow each existing `_test.go` setup.

- [ ] **Step 2: Run tests, verify they fail**

Run: `go test ./internal/repository/ -run 'FindDeletedByCode|GiftCard.*FindDeleted' -count=1`
Expected: FAIL — methods undefined.

- [ ] **Step 3: Add interface + impl methods**

Voucher impl (`voucher_repository_impl.go`):

```go
func (r *GormVoucherRepository) FindDeletedByCode(ctx context.Context, code string, userID uuid.UUID) (*models.Voucher, error) {
	var voucher models.Voucher
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("code = ? AND user_id = ? AND deleted_at IS NOT NULL", code, userID).
		First(&voucher).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &voucher, nil
}

func (r *GormVoucherRepository) RestoreByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&models.Voucher{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil).Error
}
```

Gift card impl (`gift_card_repository_impl.go`): identical shape with `models.GiftCard` and column `card_number`, method `FindDeletedByCardNumber`.

Add the matching method signatures to the two interface files (`voucher_repository.go`, `gift_card_repository.go`).

- [ ] **Step 4: Run tests, verify pass**

Run: `go test ./internal/repository/ -run 'FindDeletedByCode|GiftCard.*FindDeleted' -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/repository/voucher_repository*.go internal/repository/gift_card_repository*.go
git commit -S --signoff -m "feat(repository): add deleted-twin lookup and restore for vouchers and gift cards"
```

---

### Task 5: Service — deleted-duplicate check + restore (all three)

**Files:**
- Modify: `internal/services/card_service.go`, `internal/services/voucher_service.go`, `internal/services/gift_card_service.go`
- Test: `internal/services/card_service_test.go`

**Interfaces:**
- Consumes: repo methods from Tasks 3–4.
- Produces (per service interface + impl):
  - Card: `FindDeletedDuplicate(ctx, cardNumber string, userID uuid.UUID) (*models.Card, error)`, `RestoreCard(ctx, id uuid.UUID, userID uuid.UUID) (*models.Card, error)`
  - Voucher: `FindDeletedDuplicate(ctx, code string, userID uuid.UUID) (*models.Voucher, error)`, `RestoreVoucher(ctx, id, userID) (*models.Voucher, error)`
  - GiftCard: `FindDeletedDuplicate(ctx, cardNumber string, userID uuid.UUID) (*models.GiftCard, error)`, `RestoreGiftCard(ctx, id, userID) (*models.GiftCard, error)`

`RestoreX` returns the restored, freshly-loaded model so the handler can map it to a DTO.

- [ ] **Step 1: Write failing test (Card)**

Add to `internal/services/card_service_test.go` (use existing mock repo pattern in that file). Test that `FindDeletedDuplicate` returns the repo's deleted card, and `RestoreCard` calls `RestoreByID` then re-fetches via `GetByID`.

```go
func TestCardService_RestoreCard(t *testing.T) {
	// arrange mock repo returning a deleted card for FindDeletedByCardNumber,
	// expecting RestoreByID(id, userID) then GetByID(id) -> active card
	// assert returned card is non-nil, err nil
}
```

Fill in with the concrete mock the existing tests use (read the file for the mock type name and expectation style).

- [ ] **Step 2: Run test, verify fail**

Run: `go test ./internal/services/ -run TestCardService_RestoreCard -count=1`
Expected: FAIL — method undefined.

- [ ] **Step 3: Implement (Card service)**

Add to `CardServiceInterface` in `card_service.go`:

```go
	FindDeletedDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error)
	RestoreCard(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Card, error)
```

Impl:

```go
// FindDeletedDuplicate returns a soft-deleted card with the same number owned by the user, or nil.
func (s *CardService) FindDeletedDuplicate(ctx context.Context, cardNumber string, userID uuid.UUID) (*models.Card, error) {
	return s.repo.FindDeletedByCardNumber(ctx, cardNumber, userID)
}

// RestoreCard clears deleted_at for the user's soft-deleted card and returns the restored card.
func (s *CardService) RestoreCard(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Card, error) {
	if err := s.repo.RestoreByID(ctx, id, userID); err != nil {
		return nil, fmt.Errorf("restore card: %w", err)
	}
	restored, err := s.repo.GetByID(ctx, id, "Merchant")
	if err != nil {
		return nil, fmt.Errorf("load restored card: %w", err)
	}
	return restored, nil
}
```

Mirror for Voucher (`FindDeletedByCode`, `RestoreVoucher`) and GiftCard (`FindDeletedByCardNumber`, `RestoreGiftCard`) — same shape, correct model + preload args (match the preload strings each service's `GetByID`/`GetX` already uses).

- [ ] **Step 4: Run test + build, verify pass**

Run: `go test ./internal/services/ -run TestCardService_RestoreCard -count=1 && go build ./...`
Expected: PASS + build success.

- [ ] **Step 5: Commit**

```bash
git add internal/services/card_service.go internal/services/voucher_service.go internal/services/gift_card_service.go internal/services/card_service_test.go
git commit -S --signoff -m "feat(services): add deleted-duplicate check and restore"
```

---

### Task 6: DTO — add Deleted flag to DuplicateWarning

**Files:**
- Modify: `internal/handlers/api/dto.go:43-48`

**Interfaces:**
- Produces: `DuplicateWarning.Deleted bool json:"deleted"`.

- [ ] **Step 1: Add field**

In `internal/handlers/api/dto.go`, `DuplicateWarning` struct, add after `ExistingID`:

```go
type DuplicateWarning struct {
	HasDuplicate   bool   `json:"has_duplicate"`
	MerchantName   string `json:"merchant_name,omitempty"`
	ResourceNumber string `json:"resource_number,omitempty"` // Card number, voucher code, etc.
	ExistingID     string `json:"existing_id,omitempty"`
	Deleted        bool   `json:"deleted"` // true = existing_id refers to a soft-deleted resource that can be restored
}
```

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/api/dto.go
git commit -S --signoff -m "feat(api): add deleted flag to duplicate warning dto"
```

---

### Task 7: Handler — deleted-duplicate 409 in Create + Restore endpoint (Cards)

**Files:**
- Modify: `internal/handlers/api/cards.go` (`Create`, add `Restore`)
- Modify: `internal/setup/routes.go` (register `cardsAPI.POST("/:id/restore", ...)`)
- Test: `internal/handlers/api/cards_test.go`

**Interfaces:**
- Consumes: `CardServiceInterface.FindDeletedDuplicate`, `RestoreCard`; `DuplicateWarning.Deleted`.
- Produces: `POST /api/v1/cards/:id/restore` → 200 `CardDetailResponse`-shaped body with restored `CardDTO`; 404 if no deleted twin owned by user.

- [ ] **Step 1: Write failing test**

In `internal/handlers/api/cards_test.go`, add a test that: creates a card, deletes it, POSTs create with the same number → expects `409` with body `duplicate.deleted == true` and `duplicate.existing_id == <deleted card id>`. Follow the existing handler-test harness (mock service or real DB — match what `cards_test.go` already does).

- [ ] **Step 2: Run test, verify fail**

Run: `go test ./internal/handlers/api/ -run TestCards.*Restore -count=1` (name the test accordingly)
Expected: FAIL.

- [ ] **Step 3: Add deleted-duplicate branch in Create**

In `internal/handlers/api/cards.go`, in `Create`, right after the existing active-duplicate block (the `if duplicate != nil { ... 409 duplicate_barcode ... }`), add:

```go
	// Soft-deleted twin owned by this user → offer restore instead of a hard failure
	deletedDup, err := h.cardService.FindDeletedDuplicate(c.Request().Context(), req.CardNumber, user.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to check deleted duplicate", "error", err)
	}
	if deletedDup != nil {
		return c.JSON(http.StatusConflict, DuplicateErrorResponse{
			Error:   "duplicate_barcode",
			Message: "A soft-deleted card with this number exists and can be restored",
			Duplicate: &DuplicateWarning{
				HasDuplicate:   true,
				MerchantName:   deletedDup.MerchantName,
				ResourceNumber: deletedDup.CardNumber,
				ExistingID:     deletedDup.ID.String(),
				Deleted:        true,
			},
		})
	}
```

- [ ] **Step 4: Add Restore handler**

Add to `internal/handlers/api/cards.go`:

```go
// Restore restores a soft-deleted card owned by the current user
// POST /api/v1/cards/:id/restore
func (h *CardsHandler) Restore(c *echo.Context) error {
	user := c.Get("current_user").(*models.User)

	cardID, err := parseResourceID(c, "card")
	if err != nil {
		return err
	}

	restored, err := h.cardService.RestoreCard(c.Request().Context(), cardID, user.ID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "failed to restore card", "card_id", cardID, "error", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to restore card",
		})
	}
	if restored == nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "No restorable card found",
		})
	}

	isFavorite, _ := h.favoriteService.IsFavorite(c.Request().Context(), user.ID, "card", cardID)
	dto := ToCardDTO(restored, isFavorite)
	return c.JSON(http.StatusOK, CardResponse{Card: dto})
}
```

Note: use whatever single-card response wrapper `cards.go` already uses for create (check the `Create` success return — if it returns `CardResponse{Card: dto}` or `CardDetailResponse`, match it). If `RestoreByID` updated zero rows, `GetByID` still returns the row (now active) — to return 404 for "nothing to restore", have `RestoreCard` check rows-affected; simplest: if `FindDeletedByCardNumber`-style guard is desired, the handler's `restored == nil` covers the GetByID-not-found case. Keep it simple: 404 only when `GetByID` errors with not-found (adjust `RestoreCard` to return `(nil, nil)` on not-found rather than a wrapped error).

- [ ] **Step 5: Register route**

In `internal/setup/routes.go`, after `cardsAPI.DELETE("/:id", cardsAPIHandler.Delete)` (around line 413):

```go
	cardsAPI.POST("/:id/restore", cardsAPIHandler.Restore)
```

- [ ] **Step 6: Run test + build, verify pass**

Run: `go test ./internal/handlers/api/ -run TestCards -count=1 && go build ./...`
Expected: PASS + build success.

- [ ] **Step 7: Commit**

```bash
git add internal/handlers/api/cards.go internal/setup/routes.go internal/handlers/api/cards_test.go
git commit -S --signoff -m "feat(api): offer restore for soft-deleted card duplicates"
```

---

### Task 8: Handler — Create branch + Restore endpoint (Vouchers + Gift Cards)

**Files:**
- Modify: `internal/handlers/api/vouchers.go`, `internal/handlers/api/gift_cards.go`
- Modify: `internal/setup/routes.go` (two more `POST /:id/restore` routes)
- Test: `internal/handlers/api/vouchers_test.go`, `internal/handlers/api/gift_cards_test.go`

**Interfaces:**
- Produces: `POST /api/v1/vouchers/:id/restore`, `POST /api/v1/gift-cards/:id/restore`.

- [ ] **Step 1: Write failing tests**

Mirror Task 7's create-collision test for vouchers (field `Code`) and gift cards (`CardNumber`), asserting `409` + `duplicate.deleted == true`.

- [ ] **Step 2: Run, verify fail**

Run: `go test ./internal/handlers/api/ -run 'TestVoucher.*Restore|TestGiftCard.*Restore' -count=1`
Expected: FAIL.

- [ ] **Step 3: Implement**

In `vouchers.go` `Create`: after the active-duplicate block, add the deleted-twin branch using `h.voucherService.FindDeletedDuplicate(ctx, req.Code, user.ID)` and `ResourceNumber: deletedDup.Code`. Add a `Restore` handler calling `h.voucherService.RestoreVoucher`, mapping via `ToVoucherDTO`.

In `gift_cards.go` `Create`: same, using `req.CardNumber` and `h.giftCardService.FindDeletedDuplicate`. Add `Restore` calling `RestoreGiftCard`, mapping via `ToGiftCardDTO`.

Match each file's existing single-resource response wrapper.

- [ ] **Step 4: Register routes**

In `internal/setup/routes.go`, next to the voucher and gift-card DELETE routes, add:

```go
	vouchersAPI.POST("/:id/restore", vouchersAPIHandler.Restore)
	giftCardsAPI.POST("/:id/restore", giftCardsAPIHandler.Restore)
```

(Use the exact group variable names used for voucher/gift-card routes — grep `vouchersAPI.` and `giftCardsAPI.` in routes.go first.)

- [ ] **Step 5: Run tests + build, verify pass**

Run: `go test ./internal/handlers/api/ -count=1 && go build ./...`
Expected: PASS + build success.

- [ ] **Step 6: Commit**

```bash
git add internal/handlers/api/vouchers.go internal/handlers/api/gift_cards.go internal/setup/routes.go internal/handlers/api/vouchers_test.go internal/handlers/api/gift_cards_test.go
git commit -S --signoff -m "feat(api): offer restore for soft-deleted voucher and gift card duplicates"
```

---

### Task 9: Frontend — restore API + type flag

**Files:**
- Modify: `client/src/lib/types/api.ts:38-43` (`DuplicateWarning`)
- Modify: `client/src/lib/api/cards.ts`, `client/src/lib/api/vouchers.ts`, `client/src/lib/api/giftCards.ts` (add `restore`)

**Interfaces:**
- Produces: `DuplicateWarning.deleted?: boolean`; `xApi.restore(id): Promise<...>` per resource.

- [ ] **Step 1: Add type flag**

`client/src/lib/types/api.ts`, `DuplicateWarning`:

```ts
export interface DuplicateWarning {
	has_duplicate: boolean;
	merchant_name?: string;
	resource_number?: string;
	existing_id?: string;
	deleted?: boolean;
}
```

- [ ] **Step 2: Add restore method to each API client**

In `client/src/lib/api/cards.ts`, inside `cardsApi`, add (mirror the existing `delete(id)` method's request/return style):

```ts
	async restore(id: string): Promise<{ card: Card }> {
		return apiClient.post(`/cards/${id}/restore`);
	},
```

Use the actual HTTP helper the file already imports (grep `delete(id)` in the file for the exact call, e.g. `apiFetch`/`apiClient`). Repeat in `vouchers.ts` (`/vouchers/${id}/restore` → `{ voucher: Voucher }`) and `giftCards.ts` (`/gift-cards/${id}/restore` → `{ gift_card: GiftCard }`).

- [ ] **Step 3: Typecheck**

Run: `docker compose exec -T client npm run check` (or the project's svelte-check command; if client container not up, `cd client && npm run check`).
Expected: no new type errors.

- [ ] **Step 4: Commit**

```bash
git add client/src/lib/types/api.ts client/src/lib/api/cards.ts client/src/lib/api/vouchers.ts client/src/lib/api/giftCards.ts
git commit -S --signoff -m "feat(client): add restore api and deleted duplicate flag"
```

---

### Task 10: Frontend — restore button in DuplicateWarningBanner + wire into forms

**Files:**
- Modify: `client/src/lib/components/DuplicateWarningBanner.svelte`
- Modify: `client/src/routes/cards/new/+page.svelte`, `client/src/routes/vouchers/new/+page.svelte`, `client/src/routes/gift-cards/new/+page.svelte`
- Modify: `client/src/lib/i18n/` (DE/EN/FR keys)

**Interfaces:**
- Consumes: `DuplicateWarning.deleted`, `existing_id`; `xApi.restore`.

- [ ] **Step 1: Add restore affordance to the banner**

In `DuplicateWarningBanner.svelte`, accept an `onrestore?: () => void` prop and, when `warning.deleted === true`, render a restore button (label from i18n key `duplicate.restore`) instead of / in addition to the existing warning text. Read the component first to match its prop and styling conventions.

- [ ] **Step 2: Wire into each new-form page**

In `cards/new/+page.svelte`, add a handler:

```ts
	async function handleRestore() {
		if (!duplicateWarning?.existing_id) return;
		await cardsApi.restore(duplicateWarning.existing_id);
		await goto(`/cards/${duplicateWarning.existing_id}`);
	}
```

Pass `onrestore={handleRestore}` to `<DuplicateWarningBanner>`. Import `goto` from `$app/navigation` if not already. Mirror in vouchers/gift-cards new pages with the right api + route prefix.

- [ ] **Step 3: Add i18n keys**

Add `duplicate.restore` and `duplicate.deletedMessage` to DE/EN/FR locale files (match existing i18n structure — grep an existing `duplicate.` key first). Suggested copy:
- DE: `"Gelöschten Eintrag wiederherstellen"` / `"Ein gelöschter Eintrag mit dieser Nummer existiert."`
- EN: `"Restore deleted entry"` / `"A deleted entry with this number exists."`
- FR: `"Restaurer l'entrée supprimée"` / `"Une entrée supprimée avec ce numéro existe."`

- [ ] **Step 4: Typecheck**

Run: `cd client && npm run check`
Expected: no new type errors.

- [ ] **Step 5: Commit**

```bash
git add client/src/lib/components/DuplicateWarningBanner.svelte client/src/routes/cards/new/+page.svelte client/src/routes/vouchers/new/+page.svelte client/src/routes/gift-cards/new/+page.svelte client/src/lib/i18n/
git commit -S --signoff -m "feat(client): offer restore in duplicate warning banner"
```

---

### Task 11: E2E test — full restore flow

**Files:**
- Create or modify: `client/tests/e2e/soft-delete-restore.spec.ts` (or extend `cards.spec.ts`)

**Interfaces:**
- Consumes: running e2e stack via `dde project:e2e:start`.

- [ ] **Step 1: Write E2E test**

Test flow for cards: create card with number `E2E-RESTORE-1` → delete it → create again with same number → assert restore banner appears → click restore → assert navigation to the card detail and card is active. Follow the existing `cards.spec.ts` selectors/fixtures.

- [ ] **Step 2: Run E2E**

Run:
```bash
dde project:e2e:start
dde project:e2e:test -- tests/e2e/soft-delete-restore.spec.ts
```
Expected: PASS.

- [ ] **Step 3: Tear down + commit**

```bash
dde project:e2e:down
git add client/tests/e2e/soft-delete-restore.spec.ts
git commit -S --signoff -m "test(e2e): cover soft-delete duplicate restore flow"
```

---

### Task 12: Full verification

- [ ] **Step 1: Backend build + full test with race**

Run: `go build ./... && go test ./... -race -count=1`
Expected: all PASS.

- [ ] **Step 2: Frontend typecheck**

Run: `cd client && npm run check`
Expected: no errors.

- [ ] **Step 3: Manual smoke via docker (optional)**

Run: `docker compose up -d`, create → delete → recreate a card, confirm restore banner + restore works, and confirm two different users can hold the same number.

No commit — verification only.

## Self-Review

**Spec coverage:**
- Partial indexes exclude soft-deleted → Task 1. ✓
- Model-tag drift → Task 2. ✓
- Repo deleted-lookup + restore (3 resources) → Tasks 3–4. ✓
- Service layer → Task 5. ✓
- DTO `Deleted` flag → Task 6. ✓
- Handler 409 + restore endpoint (3 resources) → Tasks 7–8. ✓
- Per-user scope, never global → enforced in every repo query (`user_id = ?`). ✓
- Restore = pure `deleted_at = NULL`, form values discarded → Task 5 (`Update("deleted_at", nil)`, no field merge). ✓
- Frontend restore via existing banner → Tasks 9–10. ✓
- Tests (backend per resource + E2E) → Tasks 3–5, 7–8, 11. ✓
- Not-in-scope items (no field merge, no global restore, active-duplicate behavior unchanged) → respected; active-duplicate block untouched, deleted branch added after it. ✓

**Type consistency:** `FindDeletedByCardNumber`/`FindDeletedByCode`/`RestoreByID` (repo) → `FindDeletedDuplicate`/`RestoreCard`/`RestoreVoucher`/`RestoreGiftCard` (service) → handler `Restore`. `DuplicateWarning.Deleted` (Go) ↔ `deleted` (TS). Consistent across tasks.

**Placeholder scan:** Test bodies for Tasks 4/5/8 reference "follow existing harness / mock pattern" rather than full code because the exact mock/DB-setup helper names live in files the implementer must read first; the assertions and method calls under test are fully specified. Acceptable — the behavior is concrete, only the fixture boilerplate is deferred to the file's own convention.
