<p align="center">
  <img src="client/static/logo.png" alt="Savvy Logo" width="128" />
</p>

# Savvy

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://golang.org/dl/)
[![Node Version](https://img.shields.io/badge/Node-18%2B-green.svg)](https://nodejs.org/)
[![Pull Request](https://github.com/sbaerlocher/savvy/workflows/Pull%20Request/badge.svg)](https://github.com/sbaerlocher/savvy/actions/workflows/pull-request.yml)
[![Weekly Security Scan](https://github.com/sbaerlocher/savvy/workflows/Weekly%20Security%20Scan/badge.svg)](https://github.com/sbaerlocher/savvy/actions/workflows/weekly-security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sbaerlocher/savvy)](https://goreportcard.com/report/github.com/sbaerlocher/savvy)
[![codecov](https://codecov.io/gh/sbaerlocher/savvy/branch/main/graph/badge.svg)](https://codecov.io/gh/sbaerlocher/savvy)

> Digital management system for loyalty cards, vouchers, and gift cards with sharing functionality

A modern web-based system for managing loyalty cards, discount vouchers, and prepaid gift cards. Features integrated
barcode scanner, transaction history, and flexible sharing with other users.

## Project Scope

- **Hosting model**: **Self-hosted, open source (MIT)**. There is no hosted SaaS
  or paid tier. The stack (Go binary + PostgreSQL + Helm chart) is built to run
  on your own infrastructure. Run it with Docker Compose or the bundled Helm
  chart — see [Deployment](#deployment).
- **Client model**: **PWA only**. The installable Progressive Web App covers
  iOS/Android/Desktop; there is intentionally no native (Capacitor/Tauri)
  build. Apple Wallet pass integration would require a native shell and is
  deliberately out of scope.

### Core vs. optional features

Every feature ships in the binary; scope is controlled at runtime via
`ENABLE_*` environment variables, not by deleting code. This keeps a small
deployment small without forking.

| Group           | Feature                         | Toggle                    | Default |
| --------------- | ------------------------------- | ------------------------- | ------- |
| **Core**        | Loyalty cards + barcode + share | `ENABLE_CARDS`            | on      |
| Optional        | Discount vouchers               | `ENABLE_VOUCHERS`         | on      |
| Optional        | Prepaid gift cards              | `ENABLE_GIFT_CARDS`       | on      |
| Optional (auth) | Email/password login            | `ENABLE_LOCAL_LOGIN`      | on      |
| Optional (auth) | Self-registration               | `ENABLE_REGISTRATION`     | on      |
| Optional (auth) | Two-factor (TOTP)               | `ENABLE_2FA`              | off     |
| Optional        | Expiry reminders                | `ENABLE_EXPIRY_REMINDERS` | on      |

> Defaults reflect `internal/config/config.go`. At least one auth method
> (`ENABLE_LOCAL_LOGIN` or OAuth) must stay enabled — the config validates this.

A minimal single-user deployment can run on loyalty cards alone
(`ENABLE_VOUCHERS=false ENABLE_GIFT_CARDS=false`); the broader feature set is
opt-in per instance. See [Feature Toggles in AGENTS.md](AGENTS.md) and
[Deployment](#deployment) for the full variable reference.

## Features

### Loyalty Cards

- Digital storage of loyalty cards and membership cards
- Barcode support (CODE128, QR, EAN13, EAN8)
- Barcode scanning via smartphone/webcam (barcode-detector polyfill)
- Server-side barcode generation (bwip-js)
- Status tracking (Active, Inactive)
- Merchant management with colors and logos
- Merchant filter for quick merchant-based filtering
- Share with other users (with edit permissions)

### Vouchers

- Discount vouchers (percentage, fixed amount, points multiplier)
- Multiple usage models:
  - Single-Use (one-time)
  - One-per-Customer (once per customer)
  - Multiple-Use (multiple times with/without card tracking)
  - Unlimited
- Validity period and minimum order value
- Barcode scanning for quick capture
- Merchant filter for targeted searching
- Sharing (always read-only for shared users)

### Gift Cards

- Prepaid balance with automatic calculation
- Transaction history (expenses and recharges)
- Optional PIN protection
- Barcode scanning for card numbers
- Merchant filter for organized overview
- Expiration date management
- Share with granular permissions:
  - Edit (modify details)
  - Delete (remove card)
  - Manage transactions (record expenses)

### Sharing & Permissions

- All three resource types can be shared
- Flexible permissions per share
- Edit/Delete/View permissions for Cards
- Transaction management for Gift Cards
- Overview of shared items in dashboard
- User-specific favorites (shared items can be favorited individually)
- Owner display for shared items ("from [Name]")

### Ownership Transfer

- **Complete ownership transfer** for Cards, Vouchers & Gift Cards
- Email-based recipient selection with autocomplete
- Only owner can transfer (Authorization via AuthzService)
- Clean slate approach: All shares are deleted on transfer
- Audit logging for all ownership transfers
- Inline form with warnings before transfer
- Reactive SvelteKit UI without page reload

### Dashboard

- Statistics (number of Cards/Vouchers/Gift Cards)
- Total balance of all Gift Cards
- Favorites system (pinning) - Quick access to frequently used items
- Recently added items (when no favorites available)
- Quick access to create new items
- Mobile-optimized view (favorites before statistics)

### Merchant Overview

- **Merchant-centric browsing** at `/merchants`
- Aggregated counts (cards, vouchers, gift cards per merchant) and total gift card balances
- Sort by: name A-Z, name Z-A, most items, highest balance
- Filter by resource type with persistent filter state
- Merchant detail page with all associated resources
- Responsive: desktop side-panel + mobile bottom sheet for filters

### Search & Filter

- Full-text search by merchant/code
- Filter by owner (Mine / All)
- Filter by status (Active / Expired)
- Sort by merchant or date
- Client-side filtering (Svelte reactive statements) for fast results

### Progressive Web App (PWA)

- **Installable**: Can be installed as app on iOS/Android/Desktop
- **Offline-first**: Automatic warmup cache for instant offline support
- **Custom Service Worker**: NetworkFirst strategy with Workbox Recipes
- **Offline detection**: Visual feedback for network issues
- **Automatic updates**: Service worker updates transparently in background
- **Zero 500 errors**: Cache fallback prevents offline errors

**Offline features**:

- View cards/vouchers/gift cards (automatically cached on Service Worker install)
- Display barcode details
- Filter & sort (client-side)
- Browse favorites
- Dashboard statistics (automatically cached)
- Create/edit/delete items (online only)
- Share management (online only)

### Batch Operations

- **Bulk delete, share, transfer, and export** for all resource types
- Max 50 items per batch request
- Delete: All-or-nothing (checks all permissions first)
- Share/Transfer: Partial success (continues on error, reports failures)
- Export: JSON file download with permission-checked items
- Duplicate and permission validation

### Two-Factor Authentication (2FA)

- **TOTP-based** with authenticator app support (Google Authenticator, Authy, etc.)
- QR code setup with manual key fallback
- **10 backup codes** per user (format: `XXXX-XXXX`)
- AES-256-GCM encryption for TOTP secrets
- Bcrypt-hashed backup codes
- Rate-limited verification endpoints
- Enable/disable/regenerate with code verification

### Data Import & Export

- **JSON Import**: Full import with cards, vouchers, and gift cards
- **CSV Import**: Resource-specific with column mapping
- **Preview mode**: See counts before executing import
- Partial success with detailed error reporting (row numbers)
- Merchant auto-resolution/creation
- **Full JSON Export**: Download all user data

### Email Verification & Password Reset

- Token-based email verification flow
- Secure password reset with expiring tokens
- SHA-256 hashed tokens in database
- Multi-language email templates (DE, EN, FR)
- SMTP configuration with TLS support

### Web Push Notifications

- VAPID-based Web Push via service worker
- Multi-device support (all subscriptions per user)
- Auto-cleanup of expired subscriptions
- Configurable in user settings

### Expiry Reminders

- Automatic reminders before voucher/gift card expiry
- **Multi-channel**: In-app notifications, push, and email
- **Per-channel preferences**: Independent category toggles per channel (push/email)
- Configurable reminder windows (default: 7, 3, 1 days before)
- Duplicate prevention (no repeated reminders)
- Localized notifications (DE, EN, FR)

### Account Management

- **Dedicated profile page** (`/profile`) for name editing, account info, language, data export
- **Dedicated security page** (`/security`) for password change, 2FA, active sessions
- **Dedicated notifications page** (`/notifications`) with full list and notification preferences
- Password change with client-side strength indicator (12+ chars, 3/4 complexity groups)
- **GDPR-compliant account deletion** (complete data removal in 9 stages)
- Confirmation email on deletion

### Notifications

- In-app notification system for shares, transfers, and expiry reminders
- Real-time notifications when items are shared with you
- Permission change alerts for shared resources
- Ownership transfer notifications
- Expiry reminder alerts (multi-channel)
- **Per-channel notification preferences** (6 toggles):
  - Push/Email channel toggles (global on/off per channel)
  - Push/Email category toggles (reminders, sharing — per channel)
- Mark as read/unread functionality
- Delete individual or all notifications
- Automatic notification generation via service layer

### Internationalization

- **Multi-language support**: German (DE), English (EN), French (FR)
- Language switcher in navigation
- Fully translated UI including forms, buttons, and messages
- SvelteKit i18n store with reactive language switching
- Persistent language preference (localStorage)
- Date and number localization per language

### Authentication & Authorization

- **OAuth/OIDC**: Provider-agnostic authentication (Google, Microsoft, etc.)
- **Local Authentication**: Email/Password with bcrypt hashing
- **Two-Factor Authentication**: Optional TOTP with backup codes
- **Server-side sessions**: PostgreSQL-backed session store with SHA-256 hashed tokens
- **Active session management**: View, revoke individual, or revoke all other sessions
- **Stale session invalidation**: Sessions auto-invalidated after password change
- **AuthzService**: Centralized authorization logic for all resources
- Granular permission checks (view, edit, delete, manage transactions)
- User impersonation for admin debugging
- Secure session management with CSRF protection

### Feature Toggles

- **Environment-based feature flags** for flexible deployment:
  - `ENABLE_CARDS` - Enable/disable loyalty cards feature
  - `ENABLE_VOUCHERS` - Enable/disable vouchers feature
  - `ENABLE_GIFT_CARDS` - Enable/disable gift cards feature
  - `ENABLE_LOCAL_LOGIN` - Enable/disable email/password authentication
  - `ENABLE_REGISTRATION` - Enable/disable user self-registration
  - `ENABLE_2FA` - Enable/disable two-factor authentication
  - `ENABLE_EXPIRY_REMINDERS` - Enable/disable automatic expiry reminders
- Runtime configuration via `/api/v1/config` endpoint
- Client-side feature detection (SvelteKit reads config on startup)
- Middleware-based feature enforcement

### Audit Logging

- **Automatic logging** of all delete operations
- Comprehensive audit trail with:
  - User ID, action type, resource type/ID
  - JSONB snapshot of deleted resource
  - IP address and user agent
  - Timestamp with timezone
- Admin panel access to audit logs
- Retention policy support
- Compliance-ready audit trail

### Observability & Monitoring

- **Prometheus Metrics** (`/metrics` endpoint):
  - HTTP request duration and counts
  - Resource counts (cards, vouchers, gift cards, users)
  - Active sessions and database connections
- **Health Checks**:
  - `/health` - Liveness probe (server running)
  - `/ready` - Readiness probe (database connectivity)
- **Structured Logging**: JSON logs with slog
- **OpenTelemetry**: Optional distributed tracing (OTEL_ENABLED)
- Performance monitoring (N+1 query prevention, database triggers)

### Admin Panel

- **User Management** (`/admin/users`): Search, filter, create, edit, promote/demote users
- **Merchant Management** (`/admin/merchants`): Create and edit merchants with branding
- **Audit Log Viewer** (`/admin/audit-log`): Review deletion history with filters
- **Email Templates** (`/admin/email-templates`): Preview email templates
- **System Health** (`/admin/system-health`): Real-time service monitoring with auto-refresh
- **Resource Restoration**: Recover soft-deleted items
- **User Impersonation**: Debug user-specific issues
- Admin dropdown menu in navigation for streamlined access
- Role-based access control (admin-only features)

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional, for Makefile shortcuts)

### Installation & Start

```bash
# 1. Clone repository
git clone <repository-url>
cd savvy

# 2. Configure environment variables
cp .env.example .env
# Edit .env with your settings

# 3. Start Docker containers (automatically builds everything)
make dev

# 4. Load test data (optional, in another terminal)
dde project:db:seed
```

**Application URL**: <https://savvy.test> (routed by dde traefik)

> **Note**: Local development without Docker is **not supported** due to Vite proxy requirements.
> The Vite dev server requires access to `http://api:8080` which only works in the Docker network.
> See [DEVELOPMENT.md](DEVELOPMENT.md) for details.

### Test Users

After seeding, the following test accounts are available (password: `test123`):

| Email                        | Role  | Description                  |
| ---------------------------- | ----- | ---------------------------- |
| `admin@example.com`          | Admin | Has admin rights + own items |
| `anna.mueller@example.com`   | User  | Has access to shared items   |
| `thomas.schmidt@example.com` | User  | Has access to shared items   |
| `maria.garcia@example.com`   | User  | Has own items                |

## Development

**IMPORTANT**: Always use Docker for development! Local npm/air commands won't work due to Vite proxy requirements.

```bash
make dev    # Start PostgreSQL + Go API (Air) + Vite Dev Server with Hot Reload
make test   # Run all tests
make build  # Build production binary with embedded client
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for full setup, all Makefile commands, and hot reload details.

## Project Structure

```
savvy/
├── cmd/                  # Application entrypoints (server, seed, migrate, e2e)
├── internal/
│   ├── setup/            # Server initialization (DI, routes, Echo config)
│   ├── handlers/api/     # JSON API handlers (cards, vouchers, gift cards, ...)
│   ├── handlers/shares/  # Share handler abstraction
│   ├── services/         # Business logic layer
│   ├── repository/       # Data access layer (GORM)
│   ├── models/           # GORM models
│   ├── middleware/       # HTTP middleware (auth, CSRF, sessions, ...)
│   ├── migrations/       # Database migrations (embedded)
│   └── ...               # config, metrics, i18n, email, audit, telemetry
├── client/               # SvelteKit Frontend
│   ├── src/routes/       # SvelteKit pages
│   ├── src/lib/          # Components, API clients, stores, i18n, types
│   └── tests/            # E2E tests (Playwright)
├── deploy/               # Helm charts, Kustomize overlays, Grafana dashboards
├── docker-compose.yml    # Docker services (dev)
├── Dockerfile            # Multi-stage production build
└── Makefile              # Development commands
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed package structure and architecture diagrams.

## Tech Stack

### Core Stack

| Component             | Technology             | Version | Purpose                  |
| --------------------- | ---------------------- | ------- | ------------------------ |
| **Backend Framework** | Echo                   | v4.15   | HTTP Router & Middleware |
| **ORM**               | GORM                   | v1.31   | PostgreSQL Abstraction   |
| **Frontend**          | SvelteKit + TypeScript | 2.51    | Modern SPA Framework     |
| **Build Tool**        | Vite                   | 7.3     | Fast Build & Dev Server  |
| **Styling**           | TailwindCSS            | 4.1     | Utility-First CSS        |
| **Database**          | PostgreSQL             | 16      | Primary Data Store       |
| **Language**          | Go                     | 1.26    | Backend Language         |
| **Language**          | TypeScript             | 5.9     | Frontend Language        |

### Features & Libraries

| Feature                | Technology                 | Purpose                         |
| ---------------------- | -------------------------- | ------------------------------- |
| **Auth (Sessions)**    | Gorilla Sessions + PGStore | Server-side Session Management  |
| **Auth (OAuth/OIDC)**  | go-oidc                    | OpenID Connect Provider         |
| **Barcode Scanning**   | barcode-detector           | BarcodeDetector Polyfill (WASM) |
| **Barcode Generation** | bwip-js                    | Server-side Barcode Rendering   |
| **Validation**         | go-playground/validator    | Input Validation                |
| **i18n**               | go-i18n                    | Internationalization            |
| **Metrics**            | Prometheus                 | Application Metrics             |
| **Tracing**            | OpenTelemetry              | Distributed Tracing             |
| **PWA**                | @vite-pwa/sveltekit        | Service Worker & Offline        |
| **Service Worker**     | Workbox                    | Caching & Offline Strategies    |

### Development & Testing

| Tool                  | Technology             | Purpose                      |
| --------------------- | ---------------------- | ---------------------------- |
| **Hot Reload (Go)**   | Air                    | Go Auto-Reload               |
| **Hot Reload (Vite)** | Vite HMR               | Frontend Hot Module Replace  |
| **Unit Testing (Go)** | testify                | Go Test Assertions           |
| **Unit Testing (JS)** | Vitest                 | Frontend Unit Tests          |
| **E2E Testing**       | Playwright             | End-to-End Browser Tests     |
| **Mocking**           | Mockery                | Mock Generation for Tests    |
| **Linting (Go)**      | golangci-lint + revive | Go Code Quality              |
| **Linting (TS)**      | svelte-check           | TypeScript/Svelte Validation |

## Database

The application uses **PostgreSQL** with 19 tables for users, merchants, cards, vouchers,
gift cards, sharing, favorites, notifications, audit logs, 2FA, email tokens, push subscriptions,
expiry reminders, and server-side sessions.

**For detailed database schema, ERD diagram, and table descriptions, see:**

[ARCHITECTURE.md - Database Schema](ARCHITECTURE.md#️-database-schema)

## Security

- Bcrypt password hashing
- Server-side sessions (PostgreSQL-backed with SHA-256 hashed tokens)
- Active session management (view, revoke, revoke all others)
- Stale session invalidation (auto-invalidate after password change)
- Two-factor authentication (TOTP with AES-256-GCM encryption)
- Email verification and secure password reset
- CSRF protection (Echo middleware)
- SQL injection protection (GORM parameterized queries)
- XSS protection (SvelteKit auto-escaping)
- UUID instead of integer IDs
- Granular permissions for sharing
- GDPR-compliant account deletion
- Rate limiting on auth endpoints

See [SECURITY.md](SECURITY.md) for complete security policy and vulnerability reporting.

## Deployment

The application is designed for **containerized deployment** with Traefik reverse proxy for
production use. Supports Docker Compose and Kubernetes (K3s/K8s).

**Production Architecture**:

```
Client (HTTPS) → Traefik (TLS) → Savvy (HTTP:8080) → PostgreSQL
```

**For detailed deployment instructions, environment variables, Traefik configuration,
and Kubernetes manifests, see:**

[OPERATIONS.md - Deployment Guide](OPERATIONS.md)

## Testing

```bash
make test                  # Run all Go tests
make test-core             # Run core tests (services + models)
cd client && npm test      # Run frontend unit tests (Vitest)
cd client && npm run test:e2e  # Run E2E tests (Playwright)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full testing guide, coverage details, and E2E test commands.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Contribution guidelines
- Development setup
- Code style requirements
- Pull request process
- Testing requirements

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for full release history and breaking changes.

## Documentation

### For Users

- **[README.md](README.md)** - This file: Quick Start, Features, Installation
- **[SUPPORT.md](SUPPORT.md)** - Support resources, FAQ, Troubleshooting
- **[SECURITY.md](SECURITY.md)** - Security policy, vulnerability reporting
- **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** - Community guidelines, code of conduct

### For Developers

- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines, code style, PR process
- **[AGENTS.md](AGENTS.md)** - AI agent documentation, offline architecture, cache validation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design, database schema, clean architecture, PWA
- **[OPERATIONS.md](OPERATIONS.md)** - Deployment (Traefik/K8s), monitoring, audit logging
- **[OBSERVABILITY.md](OBSERVABILITY.md)** - Observability stack, Prometheus, Grafana, Loki
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Docker development, Air hot reload, best practices
- **[RELEASE.md](RELEASE.md)** - Release process and versioning

### Project Management

- **[GOVERNANCE.md](GOVERNANCE.md)** - Project governance model, decision-making
- **[CHANGELOG.md](CHANGELOG.md)** - Release history and breaking changes
- **[LICENSE](LICENSE)** - MIT License
- **[NOTICE](NOTICE)** - Third-party software notices

### Technical Details

- **[internal/migrations/migrations.go](internal/migrations/migrations.go)** - Database migrations (embedded)

## Support

Need help? We have various support channels:

- **[SUPPORT.md](SUPPORT.md)** - Complete support guide with FAQ and troubleshooting
- **GitHub Discussions** - Community Q&A and discussions
- **GitHub Issues** - Bug reports and feature requests (use templates!)
- **Security Issues** - <security@sbaerlo.ch> (NEVER report publicly!)

See also: **[CONTRIBUTING.md](CONTRIBUTING.md)** for contribution guidelines

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Built with** Go + Echo + SvelteKit + TypeScript + TailwindCSS

**Deployed with** Docker + PostgreSQL (Traefik recommended for production)
