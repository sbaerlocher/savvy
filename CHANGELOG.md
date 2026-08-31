# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.7.0] - 2026-08-31

### Added

- **Platform-native design system** - Every screen was rebuilt from the design
  mockups per platform: Material 3 on Android, Liquid Glass on iOS, solid card
  chrome on desktop — dashboard (#336-#338), wallet (#339-#341), auth
  (#378-#380), password reset (#374-#376), notifications (#381, #382, #384),
  resource details (#386-#388), merchants (#389-#391), import dialog
  (#395-#397), settings (#398-#400), admin (#401-#403) and the batch flows
  (#405-#407). The look is carried by new token sets in `tokens.css` — M3
  shape/tonal-surface/elevation tokens, a liquid-glass palette with a
  no-backdrop-filter fallback, and dedicated transfer and gift-card accents
  (#317, #320, #327, #328, #329, #333, #335).
- **Unified page structure (PageShell) (#412)** - Every page renders through a
  shared `PageShell`: one content container, one title row carrying the back
  affordance and the right-hand actions, with the page content in a colocated
  `Section` per route. Replaces five competing per-page container dialects
  that double-padded content on some screens. The layer model is documented in
  `client/COMPONENTS.md`.
- **Structural baseline suite (#412)** - `client/tests/structure/` guards the
  shared skeleton per route × platform: structure rules (one container, one
  `h1`, identical nav), 60 committed screenshot baselines and recorded axe
  violation counts, plus a route-coverage guard (`structure:routes`) so no new
  page escapes the baseline. Recording is gated behind an explicit confirm.
- **Admin on mobile (#402, #403)** - iOS and Android get an admin entry in the
  profile hub (users, merchants, audit log, system health) that previously
  existed only in the desktop nav; desktop admin moved onto elevated panels
  (#401).
- **Merged settings screens (#398-#400)** - Desktop combines profile, security
  and notification preferences into one tabbed settings page; Android and iOS
  each merge them into a single native settings screen including sessions, the
  service-worker recovery entry and a logout row.
- **OpenAPI specification (#297)** - `just openapi` generates
  `docs/openapi/` from handler annotations (swag v2); the `cards` slice is
  fully annotated as the reference, Swagger UI is served at `/api/v1/docs/` in
  development builds only, and `cmd/openapi-fix` collapses swag's
  unsatisfiable `oneOf` body wrappers.
- **Shared UI primitives (#290, #412)** - `Button`, `EmptyState` and
  `Skeleton` replace per-call-site duplication, and a shared `StateScreen`
  renders the offline/404/error pages.
- **Resource names in share/transfer notifications (#295)** - Notifications
  and their emails name the merchant and description instead of a generic
  message; monetary values and free-form notes are deliberately kept out of
  push bodies so they never appear on a lockscreen.
- **Notification email dispatcher** - Background job (1-minute tick) that claims
  pending notification emails with `FOR UPDATE SKIP LOCKED`, so the multi-replica
  deployment needs no leader election. Rows stranded by a dying pod return to the
  queue after 10 minutes; delivery is at-least-once, meaning a crash between SMTP
  success and the status write can resend one mail.
- **Delivery metrics** - `notification_emails_sent_total`,
  `notification_emails_failed_total` and the `notification_emails_pending` gauge.
  The gauge is the load-bearing one: without it a stalled dispatcher looks exactly
  like an idle one.
- **Firefox and Mobile Chrome E2E jobs (#350)** - Both Playwright projects now
  run on every pull request (reporting-only, not yet required checks).

### Changed

- **`/settings` redirects to `/profile` (#412)** - The account destination is
  `/profile`; `?tab=security` maps to `/security` and `?tab=notifications` to
  `/notifications/settings`.
- **Native detail actions behind the more menu (#386-#388, #412)** - On
  Android and iOS, share, transfer and delete sit behind the ••• menu on the
  detail title row (M3 bottom sheet / glass context menu); Android edits
  through a FAB instead of a header button, and the bottom navigation stays
  visible on detail routes.
- **Wallet chrome (#339-#341)** - Search moved into the platform chrome
  (desktop nav bar, Android header, iOS bottom-nav pill); desktop gains a
  title-row toolbar with filter, select mode, a persisted barcode toggle and
  import. Select mode rearranges the native bottom chrome (contextual top app
  bar on Android, floating batch bar in the nav slot on iOS).
- **Transfer accent (#333)** - Transfer UI moved from purple to a dedicated
  warm `--color-transfer-*` scale (WCAG-AA checked); purple is now the admin
  accent.
- **Share/transfer email gating** - The recipient lookup now happens before the
  notification row is created (the same query the push gate already made, so no
  extra round trip). A recipient that cannot be loaded yields a `skipped` email
  state instead of an attempted send.
- **make → just (#311)** - The Makefile was replaced by a `justfile`; docs and
  CI templates updated accordingly.
- **Toolchain and dependency maintenance** - Go moved to 1.27 across
  `go.mod`, CI and the `golang:1.27-alpine` build stages (#352, #394, #404);
  the Node.js runtime moved to 24.20 (#314, #334, #418),
  `@testing-library/jest-dom` to v7 (#298), `bwip-js` to 4.11 (#377), and
  TypeScript is capped below 6.1 because `@typescript-eslint/*` declares
  `typescript <6.1.0` as a required peer (#304). The remaining Renovate range
  is routine non-major updates, action pins and Compose image digests.

### Fixed

- **Failed notification emails were lost permanently** - Email delivery ran inline
  while the notification row was created and a send error was only logged, after
  which the reminder was marked as sent unconditionally. Because
  `expiry_reminder_sents` then reported the reminder as delivered, it was never
  retried: one SMTP hiccup dropped the mail for good. Delivery is now decoupled
  from row creation — the notification row carries `email_status`,
  `email_attempts` and `email_last_error`, and a dispatcher retries a failed send
  for about three hours — long enough to ride out a typical hosted-SMTP incident
  — before parking it as `failed`. Share and transfer emails,
  which previously had no retry at all and lost queued mail on restart, go
  through the same path.
- **Push notifications opened a blank page (#296)** - The service worker
  posted a navigation message no client listened for, and gift-card links
  pointed at `/gift_cards` instead of `/gift-cards`; the corrected route
  mapping is shared with the share/transfer emails, which carried the same
  broken link.
- **Barcode scanner always uses the WASM polyfill (#289)** - Chrome on Android
  can construct a native `BarcodeDetector` that never detects anything when
  the Play Services barcode module is missing; the polyfill is now the single
  path on all platforms.
- **Dashboard favorites no longer capped at five per type (#285)** - Favorites
  are an explicit user selection and load in full; the recent-items fallback
  keeps its limit of 5.
- **Idempotent init-schema foreign keys (#291)** - All 13 FK constraints guard
  against `pg_constraint`, so a half-applied init migration can replay without
  failing on SQLSTATE 42710.
- **Concurrent map write in the readiness check (#371)** - The
  `not_configured` writes for SMTP, OAuth, VAPID and TOTP ran unlocked while
  check goroutines were in flight — a fatal concurrent map write; they now
  take the mutex.
- **Platform detection order (#386)** - An Android user agent on a MacIntel
  platform was misclassified as iOS because the iPadOS touch heuristic ran
  first; Android is checked first now.
- **BottomSheet max-height (#322)** - Applied via inline style instead of an
  interpolated Tailwind class the JIT never emitted.
- **BottomSheet was outside the accessibility tree (#350)** - The scrim
  carried `aria-hidden="true"`, which propagates to the whole subtree, so the
  `role="dialog"` sheet nested inside it was absent for assistive technology;
  the scrim is `role="presentation"` now, which drops only the scrim itself
  from the accessibility tree and leaves the nested dialog exposed.
- **iOS polish (#372, #373)** - The type filter drops its redundant "all"
  segment (tapping the active segment clears the filter on every platform) and
  the resource tile matches the design component.
- **Seeded notifications match production metadata (#385)** - Seeds carry
  `merchant_name` and `days_left` so dev notifications render like real ones.

## [1.6.1] - 2026-07-23

### Fixed

- **Dashboard favorites showed non-favorited cards and gift cards (#284)** - The
  favorites-vs-recent fallback was decided per resource type, so a user with only
  voucher favorites got recent, non-favorited cards and gift cards mixed into the
  favorites section. The decision is now global: as soon as any favorite exists,
  every type returns only its favorites; the recent-items fallback remains for
  users without any favorites.
- **Wallet default view showed expired cards (#284)** - Cards carry a manual status
  (active/inactive/expired/lost/blocked), but the default "active" filter only
  excluded `inactive`, letting expired, lost and blocked cards leak through. The
  status match now lives in `wallet/filter.ts` ('active' matches exactly active,
  'inactive' groups every non-active status), and the dashboard checkout section
  applies the same rule via a new `isCardActive` helper.

## [1.6.0] - 2026-07-22

### Added

- **Admin merchant management (#246)** - New `/admin/merchants` table (mirroring
  `/admin/users`) for listing, searching, editing and deleting merchants, linked from
  the desktop admin menu. Creating a merchant is also an admin-only entry in the
  global "new" dialog, so it works on mobile where the admin menu is not available.
- **Cross-category batch selection (#243)** - Entering select mode no longer forces
  the wallet to a single type; items can be selected across cards, vouchers and gift
  cards. The selection is grouped by type at action time and dispatched per type
  (delete, share, transfer, export) via `Promise.allSettled`, aggregating partial
  success so a failed group never hides a committed one.
- **Service worker re-registration button (#268)** - An installed PWA shortcut can get
  stuck on a zombie service worker; the existing update check only called
  `registration.update()`. A new profile button unregisters all workers and registers
  a fresh one, leaving caches and stored data intact, and reports whether registration
  actually succeeded.
- **Post-merge revalidation workflow (#274)** - `.github/workflows/merge.yml` reruns Go CI,
  Client CI and Helm test on `push: main` (plus `workflow_dispatch`). Without a
  merge queue, two individually-green PRs can break once combined; this catches it
  on merge instead of on the next PR's red base. E2E and the Claude review are
  intentionally excluded (E2E dominates wall-clock and already gates every PR;
  the review has no post-merge PR to comment on).
- **Frontend vitest via dde plugin (#259)** - `dde project:client:test` runs the Vitest
  suite inside the running client container, so developers no longer need a matching
  local Node toolchain or the `--root` workaround.

### Changed

- **Design tokens consolidated onto a single Savvy scale (#250, #252, #258, #262)** -
  `tokens.css` reduced to one design direction and made the single source for status
  colors: dead Redesign tokens pruned, numbered danger/success/warning/purple scales
  added (WCAG-AA warm-neutral ramps), and the raw Tailwind red/green/yellow/amber/
  orange/emerald/purple utilities across 44 components migrated onto them. Per-type
  resource colors dropped (`category-colors.ts` removed); merchant fallbacks unified
  on one `--color-merchant-default` token. Off-scale `text-[Npx]`/`h-[42px]`/raw
  shadow values replaced with scale tokens and shared `.control` utility.
- **Shared ResourceDetail for the three detail pages (#253)** - The card, voucher and
  gift-card detail pages were the same screen three times over (~2700 lines). Collapsed
  into one kind-parameterized `ui/ResourceDetail.svelte` plus a `GiftCardLedger` slot;
  per-type variance lives in a kind config table. Also adds the missing not-found
  branch to card and voucher, closing a latent white screen.
- **Shared WalletView for wallet and merchant detail (#245)** - `/wallet` and
  `/merchants/[id]` were ~90% copy-paste, including two drifting copies of the batch
  logic. The shared list surface (toolbar, filters, tile grid, batch flow, modals) is
  now one `ui/WalletView.svelte`; merchant detail gains the active-default status
  filter and cross-type batch handling for free.
- **Shared client lib modules and UI primitives (#263, #267)** - Duplicated logic
  extracted into `lib/` (barcode helpers, date/user formatting, wallet filter/sort,
  detail-route mapping, service-worker registration, portal action) with unit tests;
  repeated UI extracted into shared components (Modal shell for all dialogs, icon
  constants, BarcodeFields, ShareSection, DesktopNav/AppFooter shell splits).
- **Dashboard stat tiles aligned with platform mockups (#279)** - iOS keeps bordered
  cards, Android uses borderless M3 surfaces with larger radii, desktop adds a card
  shadow. Stat values render in the mono face; the favorites empty-state title gains
  proper emphasis.

### Performance

- **bwip-js loaded on demand (#244)** - The static import pulled its ~930 KB chunk
  into the first-paint bundle of every route mounting a ResourceTile, even though
  barcodes there are collapsed by default. It is now dynamically imported inside the
  draw path and cached, shipping only when a barcode actually renders.

### Security

- **Log injection hardening (#251, #280)** - CodeQL flagged user-controlled strings
  (emails, merchant names, voucher codes) written to slog without sanitization,
  allowing CRLF-based log forging. New `internal/logsafe` helper strips control
  characters at every flagged call site, plus a `SanitizingHandler` wrapping the slog
  chain as defense-in-depth so any future log call is covered.

### Fixed

- **Wallet defaults to active items (#243)** - The wallet opened with status filter
  "all", so expired vouchers, expired gift cards and inactive loyalty cards showed by
  default. It now defaults to active/valid (also on `?type=` deep links), and cards
  are status-filtered for the first time.
- **Dashboard favorites hide unusable items (#254)** - Expired vouchers, inactive
  vouchers and depleted/expired gift cards no longer surface in the favorites
  quick-access; the favorite itself stays visible on wallet and detail pages.
- **Mobile safe-area fixes (#247, #248, #249)** - The floating action button no longer
  overlaps the home indicator (missing `.mobile-nav-fab` rule added); the offline
  banner and fullscreen barcode overlay respect `safe-area-inset-*`; confirm dialogs
  present as bottom sheets on mobile (above the bottom nav) instead of centered
  dialogs with padding hacks.
- **dde stack bootstrapping (#260, #261)** - The app database is created in a pre-up
  hook (reading the resolved `DATABASE_URL`, so git worktrees get their per-branch
  database) instead of crash-looping until the post-up seed; the client dev server and
  tooling run as the dde user.

### Tests

- **E2E stabilization (#265, #266)** - Host `node_modules` installed before Playwright
  runs; the forgot-password submit gates on config load so the CSRF cookie exists
  before the mutation, and the logout IndexedDB assertion scopes to the app-owned
  database instead of Workbox internals.

### Dependencies

- Updated all non-major dependencies (#257, #264, #270), `@sveltejs/kit` 2.70.0 (#273),
  svelte-check 4.7.3 (#269), happy-dom 20.11.0 (#277), `google.golang.org/grpc`
  1.82.1 \[SECURITY\] (#278).
- Updated Docker digests: distroless static-debian12 (#256, #272), Grafana 13.0 (#275).
- Updated `sbaerlocher/.github` action to v2026-07-19 (#271, #276) and
  v2026-07-21 (#281).

## [1.5.0] - 2026-07-13

### Added

- **Unified wallet overview (#239)** - Cards, vouchers and gift cards merged into a
  single `/wallet` page with type filtering. The standalone vouchers list page is
  removed; `/vouchers` now 307-redirects to `/wallet?type=vouchers`.
- **Unified ResourceTile (#238)** - One `ResourceTile` component (fed by a `TileModel`
  adapter that normalises all three DTOs) replaces the duplicated tile markup across
  the dashboard favorites, cards/vouchers/gift-card lists and merchant detail. Share
  status renders as a compact icon row (lock / people + count / people + owner name);
  the cards list gains a per-list, localStorage-persisted barcode toggle.
- **Global new dialog (#240)** - A single global "new" dialog to create cards, vouchers
  and gift cards from anywhere.

### Changed

- **Dashboard redesigned as favorites quick-access (#237)** - The dashboard now leads
  with favorites for fast at-the-till access.
- **App shell reduced to three places (#240)** - Navigation collapsed to three primary
  destinations plus the global new dialog.
- **Platform-native mobile redesign on design tokens (#241)** - Mobile UI rebuilt on a
  shared design-token system (`client/src/tokens.css`) for a platform-native feel.
- **Database migrate and seed via dde plugins (#233)** - Migrations and seeding run
  through dde plugins instead of Makefile targets, matching the existing e2e and
  observability plugins. Dead Makefile targets removed.

## [1.4.0] - 2026-07-11

### Added

- **Free voucher type (#225)** - New `free` voucher type whose value stays 0. The
  value field is hidden and no longer required; detail, list, merchant and favorite
  views render it as "Free" instead of a bare 0. Expiry reminder emails resolve the
  localized "Free" label instead of a blank value (#231).
- **Inline favorite barcodes on the dashboard (#221)** - Favorite cards, vouchers and
  gift cards render their barcode directly in the dashboard tile, scannable at the till
  without opening the detail page. Tapping still opens the enlarge modal.
- **Multi-recipient sharing (#208)** - The share action on card, voucher and gift-card
  detail pages now shares with several recipients in one call. New `emails[]` endpoint
  returns a partial-success response (`success_count`, `failed[]`, `shares`); unknown
  email, self-share and already-shared become `failed[]` entries instead of a 4xx/5xx.
  422 when all recipients fail, 201 on partial success. `EmailAutocomplete` gains a
  chip-based multiple mode. Limit 50 recipients.
- **Revoke all shares per resource (#212)** - "Revoke all shares" action for cards,
  vouchers and gift cards (one `DELETE /:id/shares` endpoint each) with a confirmation
  dialog on the detail pages. Reuses existing repo queries; no schema change. i18n
  DE/EN/FR.
- **Auto-archive read notifications (#209)** - Read in-app notifications are archived
  out of the main list after a configurable period (`NOTIFICATION_ARCHIVE_AFTER_DAYS`,
  default 30, 0 disables) without data loss (`archived_at` stamped, rows kept). A
  dedicated 24h background goroutine runs independently of expiry reminders. Migration
  0032 adds the column and index; the archive window counts from `read_at`.
- **Restore soft-deleted resources on re-create (#204)** - Re-creating a card, voucher
  or gift card whose number matches a soft-deleted one the user owns now returns a 409
  offering restore instead of a 500. Backed by per-user partial unique indexes that
  exclude soft-deleted rows and a user-scoped restore endpoint. `DuplicateWarningBanner`
  offers the restore action. i18n DE/EN/FR.
- **Symbology-content warning (#190)** - A reactive warning under the barcode-type select
  in the card, voucher and gift-card forms flags content unsuitable for the chosen
  symbology (e.g. non-numeric content with EAN-13) before saving.

### Changed

- **Barcodes hidden on mobile list views (#230)** - The barcode block is hidden below
  the `sm` breakpoint on the cards, vouchers, gift cards and merchant detail overview
  lists, saving screen space and render time; the barcode stays on each item's detail
  page.
- **Dashboard page split into components (#211)** - The 897-line dashboard page moved
  into five focused components under `client/src/lib/components/dashboard/` (header,
  favorites, quick actions, tips, barcode modal); favorite tiles restyled to match the
  list pages. No behavior or API change.
- **Oversized API handlers split (#192)** - `gift_cards.go`, `auth.go` and `admin.go`
  each had their most self-contained endpoint cluster extracted into a new file in the
  same package (transactions, auth tokens, admin diagnostics); all now under 700 lines.
  Pure relocation, no behavior change.
- **Frontend quality gates run locally in worktrees** - `lefthook.yml` `pre-push`
  now mirrors the CI `client-ci` job (`format:check`, `lint`, `typecheck`) instead
  of typecheck alone, so prettier/eslint failures surface before the push rather
  than after a CI round-trip. Documented the git-worktree caveat in `DEVELOPMENT.md`:
  `client/node_modules` is git-ignored and not shared across worktrees, so a fresh
  worktree needs `cd client && npm ci` once, else the frontend hooks fail the
  push with a missing-binary error and the check only runs in CI. DB/E2E gates
  stay CI-only.

### Fixed

- **Code-scanning findings in SW and E2E helper (#232)** - The service worker message
  handler now rejects cross-origin senders before acting on `SKIP_WAITING`, so an
  embedded third-party frame cannot force an early activation (CodeQL
  `js/missing-origin-check`). Dropped a no-op identity `.replace()` in the resource
  list page object (CodeQL `js/identity-replacement`).
- **White screen on Android homescreen shortcut launch (#224)** - Added
  `launch_handler` `navigate-existing` to both manifest sources so Android reuses an
  existing client, and warm the app shell (`/`) into the navigation cache on install so
  a cold first launch finds the shell instead of rendering nothing.
- **Shared cards detected as duplicates (#215)** - Creating a card with a number already
  shared with the user now attaches an advisory duplicate warning naming the sharer,
  instead of silently creating an indistinguishable copy. Creation still proceeds
  (family cards are intentionally allowed); matching is scoped to the merchant.
- **Audit resource id on conditional deletes (#205)** - The delete audit hook read the
  id from `Statement.Dest`, which was zero for `WHERE`-clause deletes (gift cards,
  merchants, transactions, all `*_share`), writing `uuid.Nil` and making those entries
  unrestorable. The hook now re-selects the targeted rows (mirroring the DELETE's own
  scoping) and writes one entry per row with its real id and legible `resource_data`.
  Also fixes bulk deletes producing a single nil entry.
- **Gift-card `barcode_type` dropped (#203)** - The barcode type was missing from the
  create request, response DTO and its mapper, so every gift card fell back to CODE128
  regardless of input. Wired through all four paths and validated on create like cards.

### Tests

- **Full-stack cards handler integration tests (#191)** - New tests drive the cards
  handler through real services, repositories and PostgreSQL (via `testutil.NewTestDB`)
  to verify end-to-end wiring (list round-trip, ownership/forbidden authorization,
  create persistence with auto-created merchant); skip when no test DB is reachable.

### Dependencies

- Updated all non-major dependencies (#189, #193, #195, #198, #199, #206, #214, #219,
  #228), `@sveltejs/kit` 2.69.1 (#210), `cookie` v2 (#202), Prettier 3.9.1 (#201),
  Node.js 24.18.0 (#200).
- Updated Docker images (#188, #196): Prometheus v3.13.1 (#229, #207), golang:1.26-alpine
  digests (#216, #223), distroless static-debian12 (#222-range), otel-collector-contrib
  (#218-range).
- Updated `sbaerlocher/.github` reusables to 2026-07-10 (#227) and action to v2026-06-26
  (#197).

### Docs

- **Self-hosted scope documented (#194)** - Added a Project Scope section to the README
  (self-hosted OSS only, PWA-only client, core-vs-optional feature table driven by
  `ENABLE_*` toggles) and recorded the maintainability-audit decision in GOVERNANCE:
  keep the full feature set, manage scope via configuration.

## [1.3.2] - 2026-06-21

### Security

- **Existing 2FA enforced regardless of `ENABLE_2FA` (#164)** - The TOTP service was
  only instantiated when `ENABLE_2FA=true`, so with the flag off the login handler
  skipped the 2FA check entirely and any user who had set up 2FA could log in with just
  a password. Enforcement now keys off `TOTP_ENCRYPTION_KEY` being present, so existing
  second factors stay active no matter the flag; `ENABLE_2FA` now gates only new
  enrollment and UI advertisement. Existing-2FA management (disable, backup codes,
  status) stays reachable so users are never locked out.
- **Gift-card balance race on concurrent transactions (#166)** - `CreateTransaction`
  did a bare insert; the balance-guard trigger reads `SUM(amount)`, so two concurrent
  transactions in separate DB transactions both read the pre-insert balance, both pass
  the overdraw check, and both commit (TOCTOU). The insert now runs inside a DB
  transaction that takes `SELECT … FOR UPDATE` on the gift-card row, serializing
  concurrent inserts so the trigger always sees the committed sum.

### Changed

- **Migrations split one-per-file** - `internal/migrations/migrations.go` (~2000 LoC,
  30 migrations inline) split into one `<migration-id>.go` file per migration, leaving
  `migrations.go` as the shared helpers plus the `GetMigrations()` registry. Each new
  migration adds a file instead of growing the monolith, which removes the guaranteed
  merge conflict on parallel branches. Migration bodies and registry order are
  unchanged (no behavior change).
- **Repo aligned with the sbaerlocher standard (#177, #178)** - Added REVIEW.md,
  lefthook, and ESLint (407-finding first lint pass); migrated `.github/workflows/` to
  the standard layout (`pull-request.yml`, `tag.yml`, `weekly-security.yml`; no
  `merge.yml` since deploys run via Fleet GitOps, not CI);
  removed a stray tracked `.ansible/` cache. No runtime behavior change.

### Fixed

- **Long-content QR/barcode scannability (#122)** - QR codes with long payloads
  (URLs, long tokens) now use error-correction level M and a larger module scale so
  they stay scannable from a phone. The barcode in the detail view is tappable to open
  a fullscreen view (previously only landscape on touch devices), and 2D codes grow to
  a large square in fullscreen instead of the short 1D-barcode height.
- **White screen on missing gift card (#121)** - The gift-card detail page had no
  `{:else}` branch, so a finished load with no gift card (e.g. right after an ownership
  transfer removed access, or a load that errored before redirecting) rendered nothing.
  Now shows a "not found" message and a back-to-list button. New i18n key
  `giftCards.backToList` (DE/EN/FR).
- **False audit entry on failed transfer (#168)** - `logTransferAudit` ran before the
  repository transfer in all three transfer methods, so a failed ownership transfer
  still left a "transfer" audit entry. The audit write now runs only after the
  repository call succeeds for cards, vouchers, and gift cards.
- **Duplicate `component` label in Helm manifests (#183)** - The common `savvy.labels`
  helper hardcoded `app.kubernetes.io/component: backend`; templates that set their own
  component (configmap, externalsecret) appended a second key, producing invalid YAML
  (duplicate map key). Dropped `component` from the common labels and set it per resource
  (`backend` on the Deployment, `config`/`secrets` where they belong). Surfaced by the
  rendered-manifest kubeconform validation in ci-gitops.

### Tests

- **Backend i18n drift guard** - `TestLocaleKeysInSync` asserts every supported language
  (`internal/assets/locales/{de,en,fr}.json`) defines the identical set of message IDs.
  Catches a missing/extra translation key in CI; the typed `TranslationKeys` interface
  already covers the frontend TS locales.
- **Data-export completeness E2E** - New Playwright test asserts `GET /api/v1/export`
  actually returns every user-data category (cards, vouchers, gift cards with nested
  transactions, favorites) with real content, not just that a download is triggered.
  Guards the advertised GDPR data-portability promise against a category silently
  dropping out of the export.
- **GDPR account-deletion coverage (#167)** - `account_service.go` (9-stage hard delete,
  the most destructive service — cascades across cards, vouchers, gift cards, shares,
  transactions, TOTP, sessions, audit refs) had no test. New `account_service_test.go`
  asserts all user data is removed, a second recipient survives, the audit trail is kept
  with the user reference nulled, and that delete on an unknown/already-deleted account
  errors cleanly instead of partially succeeding.
- **Gift-card transfer post-action guard (#121)** - The gift-card transfer E2E only
  asserted the HTTP response, so a white screen after the transfer (no redirect / empty
  page) would have gone unnoticed. The test now asserts the user lands back on the
  `/gift-cards` list, the list actually renders, and the transferred card is gone.
- **Stable toast matching in E2E (#182)** - `expectToast` asserted the _first_
  `role="status"` node, but toasts stack and linger; a lingering service-worker warning
  toast could occupy `.first()` and fail the share-success assertion intermittently. Now
  matches the toast that _contains_ the expected message via `filter({ hasText })`,
  fixing the flake across every spec — no app change.

## [1.3.1] - 2026-03-31

### Fixed

- **Voucher min purchase currency** - Replaced static currency text with selectable currency
  dropdown (CHF/EUR/USD/GBP), synced with the value field currency
- **Points multiplier display** - Fixed duplicate `x` in detail view (`5x x Punkte` → `5x Punkte`)

## [1.3.0] - 2026-03-31

### Added

- **`bonus_points` voucher type** - New voucher type with `+{value} Points` display, full i18n
  support (DE/EN/FR), and barcode display rendering
- **`min_purchase_amount` field on vouchers** - Optional minimum purchase amount exposed in API
  DTOs and frontend form (2-column grid layout alongside usage limit)
- **Duplicate barcode blocking (409)** - Card/voucher/gift card creation now returns HTTP 409
  with `DuplicateErrorResponse` instead of a non-blocking warning; TOCTOU race handled at DB layer
- **`DuplicateWarningBanner` component** - Frontend banner displayed on create pages when a
  duplicate barcode/code is detected, with navigation to the existing resource
- **Custom OpenTelemetry Echo v5 middleware** - In-house OTel tracing middleware replacing
  third-party `otelecho` (Echo v4 only); W3C trace context propagation, HTTP semantic conventions,
  span naming (`METHOD /route`), error status recording

### Changed

- **Echo v4 → v5 migration** - Full backend migration to `github.com/labstack/echo/v5`;
  handler signatures changed to `*echo.Context`, path params via `SetPathValues`, body limit
  as int64, graceful shutdown via `echo.StartConfig` + `signal.NotifyContext()`
- **Metrics middleware** - Uses `echo.ResolveResponseStatus()` for accurate status code
  classification instead of `c.Response().Status` (which could be 0)
- **OTelLogger middleware** - Replaced removed Echo v4 `c.Logger().SetPrefix()` with
  `slog.Default().With()` stored in context
- Updated `github.com/labstack/echo/v4` → `v5 v5.0.4`; removed `labstack/gommon`, `otelecho`
- Updated Svelte 5.53→5.55, Tailwind CSS 4.2.1→4.2.2, Vitest 4.1.0→4.1.2
- Updated Docker images: Node 24.14.1, otel-collector 0.148.0, Loki 3.7.1

### Fixed

- **DST-safe reminder calculation** - Fixed timezone bug where `calculateDaysLeft` could
  produce non-24h differences during DST transitions; now compares calendar dates in UTC
- **Hardcoded German in voucher display** - Dashboard and detail page `formatVoucherValue`
  now use localized i18n keys instead of hardcoded `{value}x Punkte`
- **`bonus_points` in expiry reminders** - Added missing `bonus_points` case to
  `formatVoucherValue()` in reminder service
- **Password manager autofill on share email** - EmailAutocomplete input was detected as
  login field by password managers; fixed with `autocomplete="new-password"`,
  `name="share-recipient"`, `data-1p-ignore`, and `data-lpignore`

## [1.2.0] - 2026-03-19

### Added

- **2FA brute-force lockout** - Per-session failed attempt counter (`SessionKey2FAFailedAttempts`);
  pending session destroyed after 5 failed challenges to prevent enumeration
- **Admin-flag in impersonation session** - `SessionKeyOriginalUserIsAdmin` propagates the
  original user's admin status into the impersonation session, removing the extra DB lookup
  on each request and closing a session-store-bug escalation vector
- **`EmailAutocomplete` component** - Reusable email input with debounced user search and
  autocomplete dropdown (300ms debounce, min 2 chars); ARIA combobox pattern
  (`role="combobox"`, `aria-expanded`, `aria-controls`, `aria-activedescendant`);
  ArrowUp/Down/Enter/Escape keyboard navigation; `required` prop controls label asterisk
- **`MerchantSelect` component** - Accessible combobox for merchant selection with
  keyboard navigation (Arrow keys, Enter, Escape), live filtering, and clear button
- **`ResourceHeader` component** - Unified header for resource detail pages combining
  favorite toggle and edit button with offline-aware disabled states
- **`ShareListItem` component** - Reusable share list item with view/edit modes,
  permission badges, and configurable delete/save actions
- **`SharePermissions` component** - Permission checkbox group for sharing
  (`can_edit`, `can_delete`, optional `can_edit_transactions`)
- **`SharedInfoBox` component** - Info box for displaying permission summaries on shared items
- **`TransferBox` component** - Transfer ownership UI with warning box, email autocomplete,
  and expandable form

### Changed

- **`cards/new` and `vouchers/new`** - Inline email autocomplete replaced with the shared
  `EmailAutocomplete` component (~50 lines of duplicate code removed per page)
- **`TransferBox`** - Transfer button disabled when email is empty (`!email.trim()`)
- **Resource detail pages** - Cards, Vouchers, Gift Cards refactored to use new shared
  components; significant LOC reduction (`[id]/+page.svelte`: cards -650, gift cards -800,
  vouchers -730 lines)
- **New/edit forms** - `CardForm`, `VoucherForm`, `GiftCardForm` migrated to
  `MerchantSelect` component

### Fixed

- **Voucher Transfer modal regression** - `oncancel` handler was missing; clicking Cancel
  left the modal open permanently and required a page reload
- **`ShareListItem` offline button** - Added `aria-label` using `common.offlineEditDisabled`
  so screen readers announce the reason instead of silence
- **E2E form-validation test** - Gift card initial balance selector corrected from
  `input#original_balance` to `input#initialBalance` (test was silently skipped before)
- **OTel SSRF protection** - `OTEL_EXPORTER_OTLP_ENDPOINT` now rejects raw RFC-1918 and
  link-local IP literals; hostname-based endpoints remain permitted
- **DDL injection in migrations** - `createTrigger`/`dropTrigger`/`dropFunction` validate
  identifier characters (`[a-z0-9_]`) and trigger timing/event keywords before interpolation;
  compound events (`INSERT OR UPDATE`) are supported
- **Weak session secret warning** - `slog.Warn` emitted when the known-weak default
  `SESSION_SECRET` is used outside the `development` environment
- **`SaveSession` silent discard** - `destroy2FAPendingSession` and the 2FA attempt counter
  increment now log a `slog.WarnContext` on persistence failure instead of discarding the error
- **`MerchantSelect` keyboard handling** - Corrected ArrowDown/ArrowUp navigation and
  `required` validation via `setCustomValidity`
- **TypeScript null narrowing in Svelte 5 snippet blocks** - Resolved TS errors when
  narrowing nullable types inside `{#snippet}` / `{@render}` contexts

## [1.1.4] - 2026-03-17

### Fixed

- **Barcode Scanner WASM route** - Add `/*.wasm` static asset route so the Go server serves
  `zxing_reader.wasm` correctly; without it, requests fell through to the SPA fallback and
  returned `index.html` instead of the WASM binary, causing "both async and sync fetching
  of the wasm failed" runtime errors

## [1.1.3] - 2026-03-17

### Fixed

- **Barcode Scanner WASM loading** - Serve zxing-wasm binary from same origin instead of
  jsDelivr CDN; fixes silent polyfill failure on iOS Safari and Firefox where CSP
  `connect-src 'self'` blocked the CDN fetch

## [1.1.2] - 2026-03-17

### Fixed

- **Barcode Scanner error handling** - Comprehensive error messages for all camera failure
  scenarios (no camera, permission denied, HTTPS required, camera in use, unsupported browser,
  security policy blocked, overconstrained config, initialization timeout)
- **Barcode Scanner retry loop** - Fixed `$effect` re-triggering `startScanning()` endlessly
  after camera error by introducing `hasError` guard state
- **Session cookie Secure flag** - Removed hardcoded `Secure=true` from PGStore cookie helper
  and default options; now set dynamically by `SaveSession` based on TLS/X-Forwarded-Proto
  (fixes CSRF failures in HTTP development environments)

### Added

- **i18n scanner keys** - 6 new scanner error translation keys (DE/EN/FR): camera not supported,
  constraints error, security blocked, timeout, HTTPS required, camera not available
- **Camera initialization timeout** - 10-second timeout prevents scanner from hanging indefinitely
  when camera metadata never loads

## [1.1.1] - 2026-03-16

### Fixed

- **Barcode Scanner iOS** - Use offscreen canvas frame capture for barcode detection
  on iOS Safari (`createImageBitmap` from video elements unsupported in WebKit)
- **CI test stability** - Prevent deadlocks and cross-package data conflicts in
  repository tests by using targeted DELETEs in a PL/pgSQL transaction block,
  connection pool limits (`MaxOpenConns=5`), and proper FK-dependent table cleanup

## [1.1.0] - 2026-03-16

### Added

- **Splash Screen** - Animated logo splash screen during SvelteKit hydration with dark mode support
- **Cross-browser select styling** - Normalized `<select>` appearance across all browsers

### Changed

- **Barcode Scanner** - Migrated from zbar-wasm to barcode-detector polyfill
  (spec-compliant BarcodeDetector API, smaller bundle)
- **Notification preference defaults** - Changed from `true` to `false`;
  email prefs auto-enabled on verification, push prefs on first subscription
- **Cookie Secure attribute** - Set statically to `true` instead of dynamic
  per-request detection (app always runs behind HTTPS/Traefik)

### Fixed

- **Email header injection** - Added `sanitizeHeader()` to strip CR/LF from email headers (CodeQL `go/email-injection`)
- **Open redirect prevention** - Block protocol-relative URLs (`//evil.com`) in i18n redirect validation
- **Sliding sessions** - Cookie MaxAge now refreshed on every session update, preventing premature cookie expiry
- **Session regeneration** - Fixed stale cookie overwrite after `RegenerateSession` by stashing raw token in memory
- **Structured logging** - Complete migration from `c.Logger()` to `slog`
  across all handlers (prevents log injection)
- **Admin OAuth roles** - Made OAuth user roles readonly in admin panel (cannot be changed server-side)

### Security

- Resolved 13 CodeQL code scanning alerts (email injection, cookie secure, open redirect, log injection)
- Enforced `Secure=true` on all session cookies
- Added email header sanitization against CRLF injection

### CI/CD

- Added Claude code review workflow for PRs
- Updated reusable workflows to `2026-03-16`
- Added gitleaks allowlist for test file false positives
- Fixed security workflow for Go 1.26 and GHCR access

### Dependencies

- Migrated `@nickvdyck/barcode-scanner` / `zbar-wasm` to `barcode-detector` polyfill
- Updated Node.js to v24.14.0
- Updated Go dependencies and Docker base images
- Updated `cookie` to >=1.1.1

## [1.0.0] - 2026-03-04

### Added

#### Frontend

- **SvelteKit SPA** with TypeScript and JSON API backend (`/api/v1/`)
- **Progressive Web App (PWA)** with offline-first functionality
  - NetworkFirst caching with automatic warmup cache on install
  - Offline viewing of cards, vouchers, and gift cards
  - Installable app experience with PWA manifest
- **Global Search** across cards, vouchers, and gift cards with filters and debounced input

#### Resource Management

- **Cards** - Digital customer cards with barcode support and merchant association
- **Vouchers** - Discount vouchers with expiration tracking and active/inactive/expired lifecycle
- **Gift Cards** - Gift cards with balance tracking, transaction history, and granular sharing permissions
- **Favorites** - Pin cards, vouchers, and gift cards to dashboard
- **Merchant Overview** - Browse data by merchant with aggregated counts and balances

#### Sharing & Transfer

- **Granular sharing permissions** per resource type (view, edit, delete, manage transactions)
- **Email-based sharing** with recipient management and permission editing
- **Ownership transfer** with audit logging and automatic share cleanup
- **Batch operations** - Bulk delete, share, transfer, and export (max 50 items per request)

#### User Features

- **Dashboard** with real-time statistics and favorites integration
- **Internationalization** - German, English, French with language switcher
- **In-app notifications** for shares, permission changes, and transfers
- **Data import** - JSON and CSV import with preview and partial success reporting
- **Data export** - Full JSON export of all user data
- **Profile editing**, password change, and GDPR-compliant account deletion
- **Settings pages** - Dedicated `/profile`, `/security`, and `/notifications` pages

#### Security & Authentication

- **Server-side sessions** - PostgreSQL-backed with SHA-256 hashed tokens and automatic expired session cleanup
- **Active session management** - View, revoke, and revoke all other sessions
- **Stale session invalidation** - Sessions automatically invalidated after password change
- **OAuth/OIDC support** with automatic user creation on first login
- **Two-Factor Authentication (TOTP)** with backup codes and AES-256-GCM encrypted secrets
- **Email verification** and secure password reset with expiring tokens
- **Production secrets validation** to prevent deployment with default credentials
- **Centralized authorization** with consistent permission checks across all resources
- **CSRF protection**, Content Security Policy, rate limiting, and security headers

#### Notifications

- **Web Push Notifications** - VAPID-based multi-device push via service worker
- **Expiry Reminders** - Automatic multi-channel reminders (in-app, push, email) before voucher/gift card expiry
- **Per-channel notification preferences** - 6 granular toggles for push/email channels and reminder/sharing categories
- **SMTP email service** with HTML templates and multi-language support (DE, EN, FR)

#### Barcode Features

- **Server-side generation** (bwip-js) - CODE128, EAN13, EAN8, QR codes
- **Browser scanning** (html5-qrcode) via camera

#### Observability

- **Audit logging** for all deletions with JSONB resource snapshots
- **Health checks** (`/health`, `/ready`) with comprehensive service validation and two-tier readiness model
- **Admin System Health Dashboard** with real-time monitoring and test email functionality
- **Structured logging** (slog, JSON) with correlation IDs
- **Prometheus metrics** for application, HTTP, and notification monitoring

#### Admin

- **User management** with search, sort, filter, and role toggling
- **Merchant management** and audit log viewer

#### Database & Deployment

- **PostgreSQL** with UUID primary keys, embedded migrations, and database triggers
- **Docker** multi-stage build with distroless production image and Watch mode for hot reload
- **Feature toggles** for cards, vouchers, gift cards, login, registration, and 2FA
- **CORS** configurable via `CORS_ALLOWED_ORIGINS` environment variable
- **OpenTelemetry** support (optional)

#### Quality & Community

- Comprehensive test suite with race detection and E2E tests (Playwright)
- Community standards: CODE_OF_CONDUCT.md, SECURITY.md, MIT license
- Mobile-optimized navigation with iPhone Safe Area support

### Changed

- Dashboard and gift card balance queries optimized for performance
- Update operations changed from POST to PATCH for RESTful compliance

### Fixed

- SMTP email sending corrected to use RFC 4409 compliant STARTTLS protocol
- Email templates updated with correct Savvy branding and dynamic logo URL
- Service Worker path and registration issues resolved
- PWA update banner i18n translations corrected

[Unreleased]: https://github.com/sbaerlocher/savvy/compare/v1.7.0...HEAD
[1.7.0]: https://github.com/sbaerlocher/savvy/compare/v1.6.1...v1.7.0
[1.6.1]: https://github.com/sbaerlocher/savvy/compare/v1.6.0...v1.6.1
[1.6.0]: https://github.com/sbaerlocher/savvy/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/sbaerlocher/savvy/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/sbaerlocher/savvy/compare/v1.3.2...v1.4.0
[1.3.2]: https://github.com/sbaerlocher/savvy/releases/tag/v1.3.2
[1.3.1]: https://github.com/sbaerlocher/savvy/releases/tag/v1.3.1
[1.3.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.3.0
[1.2.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.2.0
[1.1.4]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.4
[1.1.3]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.3
[1.1.2]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.2
[1.1.1]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.1
[1.1.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.0
[1.0.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.0.0
