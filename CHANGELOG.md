# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.0.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.0.0
