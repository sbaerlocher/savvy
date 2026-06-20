# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Tests

- **Gift-card transfer post-action guard (#121)** - The gift-card transfer E2E only
  asserted the HTTP response, so a white screen after the transfer (no redirect / empty
  page) would have gone unnoticed. The test now asserts the user lands back on the
  `/gift-cards` list, the list actually renders, and the transferred card is gone.

### Changed

- **Migrations split one-per-file** - `internal/migrations/migrations.go` (~2000 LoC,
  30 migrations inline) split into one `<migration-id>.go` file per migration, leaving
  `migrations.go` as the shared helpers plus the `GetMigrations()` registry. Each new
  migration adds a file instead of growing the monolith, which removes the guaranteed
  merge conflict on parallel branches. Migration bodies and registry order are
  unchanged (no behavior change).

### Fixed

- **Long-content QR/barcode scannability (#122)** - QR codes with long payloads
  (URLs, long tokens) now use error-correction level M and a larger module scale so
  they stay scannable from a phone. The barcode in the detail view is tappable to open
  a fullscreen view (previously only landscape on touch devices), and 2D codes grow to
  a large square in fullscreen instead of the short 1D-barcode height.

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

[Unreleased]: https://github.com/sbaerlocher/savvy/compare/v1.3.1...HEAD
[1.3.1]: https://github.com/sbaerlocher/savvy/releases/tag/v1.3.1
[1.3.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.3.0
[1.2.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.2.0
[1.1.4]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.4
[1.1.3]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.3
[1.1.2]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.2
[1.1.1]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.1
[1.1.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.1.0
[1.0.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.0.0
