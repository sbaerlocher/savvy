# Soft-Delete bei Duplikat-Check berücksichtigen

**Datum:** 2026-07-04
**Typ:** Bugfix (P1)
**Ressourcen:** Cards, Vouchers, Gift Cards
**Notion:** https://app.notion.com/p/399f17b8e6a7488a8e7cfb488d91b0ea

## Problem

Wird eine Karte, ein Gutschein oder eine Geschenkkarte soft-gelöscht und
anschliessend ein neuer Eintrag mit derselben Nummer/demselben Code angelegt,
schlägt die Neuanlage mit einem generischen 500er fehl.

Zwei Defekte, plus die gewählte UX-Erweiterung:

1. **DB-Index (Root Cause).** Die Composite-Unique-Indizes aus Migration
   `202601250008` — `idx_cards_user_card_number (user_id, card_number)`,
   `idx_vouchers_user_code (user_id, code)`,
   `idx_gift_cards_user_card_number (user_id, card_number)` — haben nur
   `WHERE user_id IS NOT NULL`, kein `AND deleted_at IS NULL`. Ein
   soft-gelöschter Datensatz behält Nummer und `user_id`, blockiert also
   den Re-Insert.

2. **Handler-500 (Symptom).** Der `Create`-Pre-Check `CheckDuplicate` läuft
   über den GORM-Soft-Delete-Scope, findet den gelöschten Datensatz nicht,
   lässt den INSERT zu. Der INSERT verletzt den Index → `ErrDuplicatedKey`.
   Der Catch-Zweig ruft erneut `CheckDuplicate` (scoped) → nil → 500.
   Nach dem Index-Fix gelingt der INSERT, der 500-Pfad entfällt für den
   Deleted-Twin-Fall.

3. **Model-Tag-Drift.** Die Struct-Tags in `card.go`/`voucher.go`/
   `gift_card.go` deklarieren einen **globalen** `uniqueIndex` (ohne Namen).
   Produktion läuft über `gormigrate` (`internal/setup/dependencies.go`),
   nicht über `database.AutoMigrate()` — Letzteres läuft nur in Tests
   (`database_test.go`). Der globale Tag ist Prod-tot, baut aber im
   Test-Schema einen globalen Index, den Prod nicht hat → Tests validieren
   gegen ein falsches Schema.

## Geklärte Entscheidungen

- **Scope:** Restore und Duplicate-Check sind **immer pro User**
  (`user_id = ? AND deleted_at IS NOT NULL`). Nie global. Ein fremder
  Datensatz mit gleicher Nummer bleibt unsichtbar und irrelevant — der
  Partial-Index erlaubt ihn ohnehin.
- **UX:** Wiederherstellen anbieten (kein harter Ausschluss).
- **Restore-Semantik:** Reines `deleted_at = NULL`. Alte Daten kommen
  unverändert zurück (inkl. Shares, Transaktionen, Favoriten-Historie).
  Neu eingegebene Formularwerte werden verworfen.
- **Model-Tags:** Auf Composite + Partial angleichen, damit Test-Schema ==
  Prod-Schema.
- **Frontend:** Bestehendes Duplicate-Pattern erweitern
  (`DuplicateWarning` + `DuplicateWarningBanner`), keine neue Komponente.

## Komponenten

### Migration `202607040031_partial_unique_indexes_exclude_soft_deleted.go`

- Droppt die 3 Composite-Indizes und legt sie neu an mit
  `WHERE user_id IS NOT NULL AND deleted_at IS NULL`.
- Rollback stellt die aktuelle Form (`WHERE user_id IS NOT NULL`) wieder her.

### Model-Tags

`Card.CardNumber`, `Voucher.Code`, `GiftCard.CardNumber`: globalen
`uniqueIndex` ersetzen durch benannten Composite-Partial-Index, der das
Migrations-Schema spiegelt (`uniqueIndex:idx_..._user_...,where:user_id IS NOT
NULL AND deleted_at IS NULL`). Damit erzeugt das Test-AutoMigrate dasselbe
Schema wie Prod.

### Repository (pro Ressource: Card / Voucher / GiftCard)

```go
// Unscoped, deleted_at IS NOT NULL, user-scoped
FindDeletedByCardNumber(ctx, number, userID) (*Card, error)

// Unscoped Update deleted_at = NULL, nur Owner
Restore(ctx, id, userID) error
```

### Service (dünner Passthrough)

```go
FindDeletedDuplicate(ctx, number, userID) (*Card, error)
// Owner-Check via Unscoped-Fetch (UserID == userID, DeletedAt != nil), dann repo.Restore
RestoreCard(ctx, id, userID) error
```

### Handler

- `Create`: nach dem bestehenden `CheckDuplicate` zusätzlich
  `FindDeletedDuplicate`. Twin gefunden → `409` mit
  `DuplicateWarning{Deleted: true, ExistingID}`.
- Neuer `Restore`-Handler: `POST /api/v1/{cards,vouchers,gift-cards}/:id/restore`
  → gibt die restaurierte DTO zurück (200). 404 wenn kein gelöschter Twin
  des Users, 403 wenn nicht Owner.

### DTO

`DuplicateWarning` um `Deleted bool json:"deleted"` erweitern.

## Data-Flow

**Create:**
```
POST /cards {card_number:"12345"}
  → CheckDuplicate (scoped, aktiv)            → aktiver Twin?  409 duplicate_barcode (bestehend)
  → FindDeletedDuplicate (unscoped, gelöscht) → gelöschter Twin? 409 {deleted:true, existing_id}
  → INSERT                                     → gelingt (Partial-Index erlaubt jetzt)
```

**Restore:**
```
POST /cards/:id/restore
  → Service: Unscoped-Fetch, prüft UserID == user.ID && DeletedAt != nil
  → repo.Restore (deleted_at = NULL)
  → 200 CardDTO
```

## Frontend (cards/vouchers/gift-cards `new/+page.svelte`)

- `DuplicateWarning`-Type um `deleted?: boolean` erweitern.
- `DuplicateWarningBanner`: wenn `deleted === true`, statt reinem Warntext
  einen "Wiederherstellen"-Button zeigen.
- Button → `restoreApi(id)` → bei Erfolg zur wiederhergestellten Karte
  navigieren. Verworfen → Nutzer bleibt im Formular.
- Neue API-Methode `restore(id)` je Ressourcen-Client
  (`cards.ts`/`vouchers.ts`/`giftCards.ts`).
- i18n-Keys (DE/EN/FR) für Dialog-Text + Button.

## Tests

Backend (pro Ressource):
1. Karte anlegen → soft-deleten → Neuanlage mit gleicher Nummer → 409
   `deleted:true`.
2. Restore-Endpoint → gelöschte Karte wird aktiv, alte Daten intakt.
3. Restore fremder Karte → 403/404.
4. Neuanlage mit Nummer eines aktiven Twins → weiterhin 409 `duplicate_barcode`
   (kein Restore-Angebot).
5. Partial-Index: zwei User, gleiche Nummer → beide erlaubt.

Frontend E2E: Restore-Dialog erscheint, Restore navigiert korrekt,
Verwerfen bleibt im Formular.

## Nicht im Scope

- Kein Merge neu eingegebener Formularwerte in die restaurierte Karte.
- Kein globaler (cross-user) Restore.
- Keine Änderung am bestehenden aktiven-Duplikat-Verhalten.
