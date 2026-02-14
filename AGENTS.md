# Savvy - AI Agent Documentation

**Last Updated**: 2026-03-04
**Version**: 2.0.0
**Project Type**: Full-Stack Web Application
**Tech Stack**: Go + Echo + SvelteKit + TypeScript + GORM + PostgreSQL
**Purpose**: Digital management of customer cards, vouchers, and gift cards with sharing functionality

---

## 🎯 Documentation Overview for AI Agents

This file serves as **central navigation** for AI agents. All details are organized in specialized documents.

### 📚 Documentation Structure

| Document                             | Purpose                                                       | Audience                   |
| ------------------------------------ | ------------------------------------------------------------- | -------------------------- |
| **AGENTS.md**                        | Central Navigation, Quick Reference                           | AI Agents                  |
| [README.md](README.md)               | Quick Start, Features, User Guide                             | Humans (Developers, Users) |
| [CONTRIBUTING.md](CONTRIBUTING.md)   | Contribution Guidelines, Code Style, PR Process               | Developers (Contributors)  |
| [ARCHITECTURE.md](ARCHITECTURE.md)   | Technical Architecture, Diagrams, Performance                 | AI Agents + Developers     |
| [DEVELOPMENT.md](DEVELOPMENT.md)     | Development Setup, Docker, Hot Reload, Testing, Commands      | AI Agents + Developers     |
| [OPERATIONS.md](OPERATIONS.md)       | Production Deployment, Monitoring, Incident Response, Scaling | AI Agents + DevOps         |
| [OBSERVABILITY.md](OBSERVABILITY.md) | OpenTelemetry, Grafana Stack, Logs, Traces, Metrics           | AI Agents + DevOps         |
| [SECURITY.md](SECURITY.md)           | Helm Chart Security, Kubernetes Best Practices, CIS Benchmark | DevOps + Security          |
| [SUPPORT.md](SUPPORT.md)             | Support Resources, FAQ, Troubleshooting                       | All (User Support)         |
| [GOVERNANCE.md](GOVERNANCE.md)       | Project Governance, Decision-Making                           | Contributors + Maintainers |
| [CHANGELOG.md](CHANGELOG.md)         | Release History, Breaking Changes                             | All (Version-Tracking)     |
| [RELEASE.md](RELEASE.md)             | Release Process, Versioning                                   | Maintainers                |
| [LICENSE](LICENSE)                   | MIT License                                                   | Legal                      |
| [NOTICE](NOTICE)                     | Third-Party Software Notices                                  | Legal + Compliance         |

**Important**: Avoid redundancies! Details are ONLY in the specialized documents.

---

## 🚀 Quick Start for AI Agents

### 1. Understanding the Project

**Read FIRST**: [README.md](README.md) for:

- ✅ Feature Overview (Cards, Vouchers, Gift Cards, Sharing)
- ✅ Tech Stack Details
- ✅ Installation & Setup
- ✅ Database Schema (High-Level)

**For deep technical details**: [ARCHITECTURE.md](ARCHITECTURE.md)
**For Deployment**: [OPERATIONS.md](OPERATIONS.md)

### 2. Important Project Features

**Key Features**:

- ✅ **SvelteKit SPA** - Modern TypeScript frontend with JSON API backend
- ✅ **Progressive Web App** - Offline-first with service worker and caching
- ✅ **Layered Architecture** - Clean separation with Handler → Service → Repository pattern
- ✅ **Barcode Features** - bwip-js for server-side generation, html5-qrcode for browser scanning
- ✅ **Granular Sharing** - Resource-specific permissions with transfer capabilities
- ✅ **Batch Operations** - Bulk delete, share, transfer, export (max 50 items per request)
- ✅ **Two-Factor Authentication** - TOTP with backup codes, QR code setup, rate-limited
- ✅ **Data Import/Export** - JSON and CSV import with preview, full data export, batch export
- ✅ **Email Verification** - Token-based email verification and password reset
- ✅ **Web Push Notifications** - VAPID-based push via service worker
- ✅ **Expiry Reminders** - Automatic multi-channel reminders before voucher/gift card expiry
- ✅ **Voucher Status** - Active/inactive/expired lifecycle with future-date support
- ✅ **Server-Side Sessions** - PostgreSQL-backed sessions with SHA-256 hashed tokens
- ✅ **Session Management** - View/revoke active sessions, stale session invalidation
- ✅ **Account Management** - Dedicated profile/security/notifications pages, GDPR-compliant deletion
- ✅ **Merchant Overview** - Browse data by merchant with aggregated counts and balances
- ✅ **Multi-language** - German, English, French support

**Frontend Architecture**:

```
Client (Browser) → SvelteKit SPA (TypeScript) → JSON API (/api/v1/*) → Go Backend
```

### 3. Making Code Changes

**⚠️ CRITICAL RULE: ALWAYS use Docker for Development**

```bash
# ✅ CORRECT: Start everything with Docker Compose
docker compose up

# ❌ WRONG: NEVER start local dev servers!
npm run dev          # NEVER! Causes proxy errors
cd client && npm run dev  # NEVER! Vite cannot access Docker network
air                  # NEVER! Go backend must run in Docker
```

**Why Docker?**

- ✅ Vite proxy only works in Docker network (`http://api:8080`)
- ✅ Consistent development environment (PostgreSQL, Go, Node.js)
- ✅ Prevents port conflicts and network issues
- ❌ Local processes cannot access Docker containers (`api:8080` is only resolvable in Docker network)

**When making changes:**

1. Edit files locally (e.g., in VSCode)
2. Air (Go) and Vite (Node.js) in Docker detect changes automatically (Hot Reload)
3. No manual restarts needed

**Details**: [DEVELOPMENT.md](DEVELOPMENT.md) - Air Hot Reload, Docker Best Practices

---

**Architecture** (Layered Architecture with 3 Layers):

```
Handlers (Presentation) → Services (Business Logic) → Repositories (Data Access)
```

**Important Directories**:

```
cmd/server/main.go                    # Entrypoint
internal/setup/                       # Server Setup (Layered Architecture)
  ├── dependencies.go                # DI Container, Database, Telemetry
  ├── routes.go                      # Route Registration
  └── server.go                      # Echo Configuration

internal/handlers/api/                # JSON API Handlers (20 files)
  ├── admin.go                       # Admin Operations
  ├── auth.go                        # Authentication
  ├── batch.go                       # Batch Delete/Share/Transfer/Export
  ├── cards.go                       # Cards CRUD
  ├── config.go                      # Feature Toggle Config
  ├── dashboard.go                   # Dashboard Stats
  ├── dto.go                         # Data Transfer Objects
  ├── export.go                      # Data Export
  ├── gift_cards.go                  # Gift Cards CRUD
  ├── helpers.go                     # Handler Utilities
  ├── import.go                      # Data Import (JSON/CSV)
  ├── mappers.go                     # Model ↔ DTO Mapping
  ├── merchants.go                   # Merchant Management
  ├── notifications.go               # Notifications API
  ├── profile.go                     # Profile Management
  ├── push.go                        # Web Push API
  ├── sessions.go                    # Session Management API
  ├── shared_users.go                # Share Recipients
  ├── totp.go                        # 2FA/TOTP Endpoints
  └── vouchers.go                    # Vouchers CRUD

internal/handlers/shares/             # Share Handler Abstraction
  ├── adapter.go                     # ShareAdapter Interface
  ├── base_handler.go                # Unified Share Logic
  └── *_adapter.go                   # Resource-Specific Adapters

internal/handlers/                    # Other Handlers
  ├── health.go                      # Health Checks
  ├── oauth.go                       # OAuth/OIDC
  └── spa.go                         # SvelteKit SPA Fallback

internal/services/                    # Business Logic (22+ files)
  ├── card_service.go                # Card business logic
  ├── voucher_service.go             # Voucher business logic
  ├── gift_card_service.go           # Gift Card business logic
  ├── favorite_service.go            # Favorites logic
  ├── merchant_service.go            # Merchant management
  ├── share_service.go               # Sharing logic
  ├── transfer_service.go            # Ownership Transfer
  ├── notification_service.go        # Notifications
  ├── dashboard_service.go           # Dashboard queries
  ├── authz_service.go               # Authorization checks
  ├── admin_service.go               # Admin Operations
  ├── user_service.go                # User Management
  ├── session_service.go             # Session Management (device/browser parsing)
  ├── totp_service.go                # 2FA/TOTP (setup, verify, backup codes)
  ├── import_service.go              # JSON/CSV Import
  ├── export_service.go              # JSON Export + Batch Export
  ├── email_token_service.go         # Email Verification & Password Reset
  ├── push_service.go                # Web Push Notifications (VAPID)
  ├── reminder_service.go            # Expiry Reminders (multi-channel)
  ├── account_service.go             # Account Deletion (GDPR)
  └── container.go                   # Dependency injection

internal/email/                       # Email Templates & Service
  ├── email_service.go               # SMTP Email Sending
  └── templates/                     # HTML Email Templates
      └── expiry_reminder.html       # Expiry Reminder Template

internal/repository/                  # Data Access (29 files, interface + impl pattern)
  ├── base_repository.go             # Base repository helpers
  ├── pagination.go                  # Pagination utilities
  ├── card_repository.go             # Card interface
  ├── card_repository_impl.go        # Card GORM queries
  ├── card_share_repository.go       # Card share interface
  ├── card_share_repository_impl.go  # Card share queries
  ├── voucher_repository.go          # Voucher interface
  ├── voucher_repository_impl.go     # Voucher GORM queries
  ├── voucher_share_repository.go    # Voucher share interface
  ├── voucher_share_repository_impl.go
  ├── gift_card_repository.go        # Gift card interface
  ├── gift_card_repository_impl.go   # Gift card GORM queries
  ├── gift_card_share_repository.go  # Gift card share interface
  ├── gift_card_share_repository_impl.go
  ├── merchant_repository.go         # Merchant interface
  ├── merchant_repository_impl.go    # Merchant queries
  ├── notification_repository.go     # Notification interface
  ├── notification_repository_impl.go
  ├── session_repository.go          # Session CRUD + bulk revocation
  ├── user_repository.go             # User interface
  ├── user_repository_impl.go        # User queries
  ├── favorite_repository.go         # Favorites interface
  ├── favorite_repository_impl.go    # Favorites queries
  ├── dashboard_repository.go        # Dashboard interface
  ├── dashboard_repository_impl.go   # Dashboard queries
  ├── transfer_repository.go         # Transfer interface
  ├── transfer_repository_impl.go    # Transfer queries
  ├── audit_log_repository.go        # Audit log interface
  └── audit_log_repository_impl.go   # Audit log queries

internal/models/                      # GORM Models (22+ files)
  ├── user.go                        # User + Authentication + PasswordChangedAt
  ├── user_totp.go                   # 2FA TOTP Settings (encrypted secrets)
  ├── session.go                     # Server-side Session model
  ├── merchant.go                    # Merchant/Brands
  ├── user_favorite.go               # Polymorphic Favorites
  ├── notification.go                # Notifications
  ├── email_token.go                 # Email Verification & Password Reset Tokens
  ├── push_subscription.go           # Web Push Subscriptions
  ├── expiry_reminder.go             # Expiry Reminder Tracking
  ├── batch.go                       # Batch Operation DTOs
  ├── card.go + voucher.go + gift_card.go
  └── *_share.go                     # Sharing models

internal/middleware/                  # Echo Middleware (16 files)
  ├── auth.go                        # Authentication + stale session detection
  ├── bodylimit.go                   # Request body size limiting
  ├── cors.go                        # CORS Configuration
  ├── csrf_api.go                    # CSRF for API
  ├── feature.go                     # Feature toggles
  ├── impersonate.go                 # Admin Impersonation
  ├── i18n.go                        # Internationalization
  ├── otel_logger.go                 # OpenTelemetry logging
  ├── pgstore.go                     # PostgreSQL Session Store (PGStore)
  ├── ratelimit.go                   # Rate Limiting
  ├── security.go                    # Security headers
  ├── session.go                     # Session Management
  ├── session_keys.go                # Centralized Session Key Constants
  ├── session_tracking.go            # Session Activity Tracking
  └── user_ratelimit.go              # Per-user rate limiting

internal/config/                      # Configuration
  └── config.go                      # Environment Variables

internal/migrations/                  # Database Migrations (Embedded in Go)
  └── migrations.go                  # Gormigrate-based Auto-Migration

client/                               # SvelteKit Frontend
  ├── src/routes/                    # SvelteKit Pages
  │   ├── +page.svelte              # Dashboard
  │   ├── cards/                     # Cards Routes
  │   ├── vouchers/                  # Vouchers Routes
  │   ├── gift-cards/                # Gift Cards Routes
  │   ├── merchants/                 # Merchant Overview & Detail
  │   ├── profile/                   # Profile Settings
  │   ├── security/                  # Security (Password, 2FA, Sessions)
  │   ├── notifications/             # Notifications List & Preferences
  │   ├── settings/                  # Settings Hub (redirects)
  │   ├── login/2fa/                 # 2FA Challenge Page
  │   ├── verify-email/              # Email Verification
  │   ├── forgot-password/           # Password Reset
  │   └── admin/                     # Admin Panel
  │       ├── users/                 # User Management
  │       ├── merchants/             # Merchant Management
  │       ├── audit-log/             # Audit Log Viewer
  │       ├── email-templates/       # Email Template Preview
  │       └── system-health/         # System Health Dashboard
  ├── src/lib/                       # Shared Components & Utilities
  │   ├── components/                # Reusable Components
  │   │   ├── TwoFactorSettings.svelte  # 2FA Management
  │   │   ├── ImportDialog.svelte    # Data Import Dialog
  │   │   ├── BatchPanel.svelte      # Batch Operations Panel
  │   │   ├── MerchantFilters.svelte # Merchant overview filter panel
  │   │   ├── TypeFilterButtons.svelte # Resource type filter buttons
  │   │   ├── PushNotificationSettings.svelte  # Push Settings
  │   │   ├── settings/              # Settings Sub-Components
  │   │   │   ├── ProfileSection.svelte      # Name editing
  │   │   │   ├── SecuritySection.svelte     # Account info, export, deletion
  │   │   │   ├── NotificationsSection.svelte # Notification preferences
  │   │   │   └── ToggleSwitch.svelte        # Accessible toggle (ARIA)
  │   │   └── ...                    # Other Components
  │   ├── stores/                    # Svelte Stores (auth, offline, i18n, push)
  │   ├── api/                       # API Client Modules
  │   │   ├── auth.ts               # Auth + 2FA API
  │   │   ├── import.ts             # Import/Export API
  │   │   ├── batch.ts              # Batch Operations API (incl. export)
  │   │   ├── export.ts             # Full Data Export API
  │   │   ├── sessions.ts           # Session Management API
  │   │   ├── push.ts               # Push Notifications API
  │   │   └── ...                    # Other API Modules
  │   ├── utils/                     # Helper Functions
  │   │   ├── merchant-aggregator.ts # Merchant data aggregation
  │   │   └── category-colors.ts   # Resource type color definitions
  │   ├── i18n/                      # Translations (TypeScript)
  │   └── types/                     # TypeScript Types
  ├── vite.config.ts                 # Vite Build Configuration
  ├── package.json                   # Node.js Dependencies
  └── tsconfig.json                  # TypeScript Configuration
```

**Details**: See [ARCHITECTURE.md](ARCHITECTURE.md) - Package Structure (Line 168-221)

### 3. Feature Toggles

The system supports **7 Feature Toggles** via Environment Variables:

```bash
# Resource Toggles
ENABLE_CARDS=true                    # Cards feature
ENABLE_VOUCHERS=true                 # Vouchers feature
ENABLE_GIFT_CARDS=true               # Gift Cards feature

# Authentication Toggles
ENABLE_LOCAL_LOGIN=false             # Email/Password (false = OAuth only)
ENABLE_REGISTRATION=false            # User registration
ENABLE_2FA=false                     # Two-Factor Authentication (TOTP)

# Notification Toggles
ENABLE_EXPIRY_REMINDERS=true         # Automatic expiry reminders
```

**Implementation**:

- Middleware in [internal/middleware/feature.go](internal/middleware/feature.go)
- API Config Endpoint: [internal/handlers/api/config.go](internal/handlers/api/config.go)
- Client-Side Toggle: SvelteKit reads `/api/v1/config` on startup

---

## 🏗️ Architecture Highlights

### Layered Architecture Pattern

**Dependency Flow**:

```
Handlers → Services (Interfaces) → Repositories (Interfaces) → GORM Models → PostgreSQL
```

**IMPORTANT**:

- Handlers do NOT know Database (only Services via Interfaces)
- Services do NOT know Echo Context (only Repository Interfaces)
- All Services have Interfaces → Testable with Mocks

**Details**: [ARCHITECTURE.md](ARCHITECTURE.md) - Layered Architecture Pattern (Line 87-126)

### Database Schema

**19 Tables** (embedded in [internal/migrations/migrations.go](internal/migrations/migrations.go)):

**Core Tables (initSchema):**

1. `users` - Users with Auth + `password_changed_at`
2. `merchants` - Merchants (central for all types)
3. `cards` + `card_shares` - Customer Cards + Sharing
4. `vouchers` + `voucher_shares` - Vouchers + Sharing (read-only)
5. `gift_cards` + `gift_card_shares` + `gift_card_transactions` - Gift Cards + granular Permissions

**Feature Tables (migrations):**

6. `user_favorites` - **Polymorphic** Favorites (Cards, Vouchers, Gift Cards)
7. `audit_logs` - Audit trail for all deletions
8. `notifications` - **In-App Notifications** for Shares/Transfers/Reminders
9. `user_totps` - **2FA/TOTP** config (encrypted secrets, hashed backup codes)
10. `email_tokens` - **Email Verification** & Password Reset tokens
11. `push_subscriptions` - **Web Push** API subscriptions (VAPID)
12. `expiry_reminder_sents` - Tracks sent **Expiry Reminders** (prevents duplicates)
13. `sessions` - **Server-side Sessions** (SHA-256 hashed tokens, IP/UA tracking)

**Special Features**:

- ✅ UUIDs instead of Integer IDs (Security)
- ✅ Polymorphic Favorites (`resource_type` + `resource_id`)
- ✅ Database Trigger: `recalculate_gift_card_balance()` - Auto-update on transactions
- ✅ Database Trigger: `enforce_lowercase_email()` - Email Normalization
- ✅ Embedded Migrations: Gormigrate-based auto-migration (no separate SQL files)
- ✅ AES-256-GCM encrypted TOTP secrets
- ✅ Bcrypt-hashed backup codes
- ✅ Server-side sessions with SHA-256 hashed tokens and automatic expired session cleanup

**ERD Diagram**: [ARCHITECTURE.md](ARCHITECTURE.md) - Database Schema

### Performance Optimizations

**Dashboard**:

- **40% faster**: N+1 Query Fix (10+ → 8 Queries)
- Parallelization with Goroutines for Stats
- `GROUP BY` Aggregation for Favorites

**Gift Card Balance**:

- **78% faster**: Database Trigger instead of Runtime calculation
- Balance is automatically updated on Transaction INSERT/UPDATE/DELETE
- No `Preload("Transactions")` needed

**Details**: [ARCHITECTURE.md](ARCHITECTURE.md) - Performance Optimizations (Line 627-700)

---

## 🔐 Security

**Implemented Features**:

- ✅ Server-side Sessions (PostgreSQL-backed PGStore with SHA-256 hashed tokens)
- ✅ Active Session Management (view, revoke, revoke all others)
- ✅ Stale Session Invalidation (auto-invalidate after password change)
- ✅ Bcrypt Password Hashing (DefaultCost)
- ✅ CSRF Protection (Echo Middleware + SvelteKit Integration)
- ✅ OAuth/OIDC Support (Provider-agnostic)
- ✅ Two-Factor Authentication (TOTP with AES-256-GCM encrypted secrets)
- ✅ Email Verification (token-based, SHA-256 hashed tokens)
- ✅ Password Reset (secure token flow with expiration)
- ✅ SQL Injection Prevention (GORM Parameterized Queries)
- ✅ XSS Protection (SvelteKit Auto-Escaping)
- ✅ Granular Sharing Permissions
- ✅ Audit Logging (all Deletions)
- ✅ Rate Limiting (Auth Endpoints, 2FA verification)
- ✅ Email Normalization (lowercase in DB)
- ✅ GDPR-compliant Account Deletion (full data removal)

**Details**:

- Architecture: [ARCHITECTURE.md](ARCHITECTURE.md) - Security Architecture (Line 427-543)
- Operations: [OPERATIONS.md](OPERATIONS.md) - Security (Line 207-283)

---

## 📊 Observability

**Stack**:

- **Metrics**: Prometheus (`/metrics` endpoint)
- **Logs**: Structured Logging (slog, JSON)
- **Traces**: OpenTelemetry (optional, via `OTEL_ENABLED=true`)
- **Health Checks**: Enhanced `/health` (liveness) and `/ready` (readiness with service validation)

**Health Endpoints**:

- **`/health`** - Simple liveness probe (returns `{"status": "ok"}`)
- **`/ready`** - Enhanced readiness probe with service health checks:
  - Two-tier model: Critical (database) vs Optional (SMTP, OAuth, VAPID, 2FA)
  - Parallel checks with goroutines (5s global timeout)
  - Detailed JSON response with per-service status, latency, errors
  - Status values: `"ready"` (all OK), `"degraded"` (optional service down), `"not_ready"` (database down)
  - HTTP 200 for ready/degraded, HTTP 503 for not_ready

**Admin System Health Dashboard** (`/admin/system-health`):

- Real-time monitoring of all services (database, SMTP, OAuth, VAPID, 2FA)
- Auto-refresh every 30 seconds (toggleable)
- Test email functionality for SMTP validation
- SVG status indicators with color coding
- Service cards showing status, enabled state, and error details

**Key Metrics**:

- `http_request_duration_seconds`, `http_requests_total`
- `cards_total`, `vouchers_total`, `gift_cards_total`
- `db_connections_active`, `db_connections_idle`
- `push_subscriptions_total`, `push_subscribed_users_total`, `email_verified_users_total`
- `push_notifications_enabled_total`, `email_notifications_enabled_total`
- `push_reminders_enabled_total`, `push_sharing_enabled_total`
- `email_reminders_enabled_total`, `email_sharing_enabled_total`
- `login_attempts_total`

**Details**: [OPERATIONS.md](OPERATIONS.md) - Observability

---

## 📱 Progressive Web App (PWA)

**Status**: ✅ Implemented with Custom Service Worker + Workbox Recipes

### Offline Functionality

**What works offline**:

- ✅ View cards/vouchers/gift cards (cached data)
- ✅ Display shared items
- ✅ Browse favorites
- ✅ View barcode details
- ✅ Dashboard with statistics (cached)
- ✅ Filter & sorting (client-side)
- ✅ Barcode scanner (Camera API)

**What does NOT work offline**:

- ❌ Create new items
- ❌ Edit/delete items
- ❌ Manage sharing
- ❌ Add/remove favorites
- ❌ Transactions (Gift Cards)
- ❌ Global Search (requires API call)

### Implementation

**Service Worker**: Custom service worker (`src/service-worker.ts`) with Workbox + `injectManifest` strategy

- **Strategy**: **NetworkFirst** with automatic warmup cache for offline-first functionality
- **Warmup Cache**: Automatic API precaching on install using `warmStrategyCache` (Workbox Recipes)
- **Cached Routes**: API responses, static assets, SvelteKit pages, navigation routes
- **Cache Names**: Separate caches for cards/vouchers/gift-cards (v4 versioning)
- **Registration**: Manual in `+layout.svelte` for full control (instead of auto-registration)

**Offline Store**: Manages online/offline state and cache validation (`client/src/lib/stores/offline.ts`)

**Architecture**:

```mermaid
graph TB
    subgraph Browser["Browser Environment"]
        SW[Custom Service Worker<br/>src/service-worker.ts<br/>NetworkFirst + Warmup]
        WB[Workbox Cache<br/>CacheStorage API]
        WARM[Warmup Cache<br/>warmStrategyCache<br/>Auto-preload on install]
        NET[Network<br/>/api/v1/*]

        subgraph OS["Offline Store"]
            OSD[Online/Offline Detection<br/>navigator.onLine]
            CV[Cache Validation<br/>automatic on online]
            POQ[Pending Operations Queue]
        end

        IDB[IndexedDB<br/>Cards/Vouchers/Gift Cards/Pending Ops]
        API[API Client<br/>cardsApi/vouchersApi/giftCardsApi]
        UI[UI Components<br/>Lists/Offline Banner]
    end

    SW --> WARM
    WARM --> WB
    SW <--> WB
    WB <--> NET
    SW --> OS
    OS --> IDB
    OS --> API
    OS --> UI
```

**Cache Validation** (Automatic):

When user comes back online, the system automatically validates cached data:

1. **Event Detection**: `window.addEventListener('online', handleOnline)`
2. **Sync Pending Operations**: Upload queued mutations (create/update/delete)
3. **Validate Cache**: Compare IndexedDB with server data
4. **Remove Deleted Items**: Delete cards/vouchers/gift-cards no longer on server
5. **Update Permissions**: Refresh data to reflect access changes
6. **Background Execution**: Non-blocking, runs in background

```typescript
// client/src/lib/stores/offline.ts
async function handleOnline() {
  update((state) => ({ ...state, isOnline: true }));

  // 1. Sync pending operations first (mutations)
  await syncPendingOperations();

  // 2. Validate cache (remove deleted items, update permissions)
  await validateCache();
}

async function validateCache() {
  // Validate all resource types in parallel
  await Promise.allSettled([
    validateCachedCards(), // Remove deleted, update cache
    validateCachedVouchers(), // Remove deleted, update cache
    validateCachedGiftCards() // Remove deleted, update cache
  ]);
}

// Example: Validate Cards
async function validateCachedCards() {
  const cachedCards = await offlineDB.getAllCards();
  if (cachedCards.length === 0) return;

  // Fetch current data from server
  const response = await cardsApi.list();
  const serverCardIds = new Set(response.cards.map((c) => c.id));

  // Remove deleted cards from IndexedDB
  for (const cached of cachedCards) {
    if (!serverCardIds.has(cached.id)) {
      await offlineDB.deleteCard(cached.id);
      logger.info(`Removed deleted card from cache: ${cached.id}`);
    }
  }

  // Update cache with fresh data (includes permission updates)
  await offlineDB.saveManyCards(response.cards);
}
```

**Cache Validation Scenarios**:

| Scenario           | Cached Data       | Server Response    | Action                   |
| ------------------ | ----------------- | ------------------ | ------------------------ |
| Card deleted       | Card A in cache   | Card A not in list | ❌ Remove from IndexedDB |
| Permission revoked | Card B in cache   | Card B not in list | ❌ Remove from IndexedDB |
| Card updated       | Card C (old data) | Card C (new data)  | ✅ Update in IndexedDB   |
| New card added     | No Card D         | Card D in list     | ✅ Add to IndexedDB      |

**Offline Detection**: Real-time status tracking

```typescript
// Svelte Store for Offline Status
export const offlineStore = writable({
  isOnline: !navigator.onLine,
  isSyncing: false,
  pendingCount: 0,
  lastSync: null
});

// Event listeners for online/offline
window.addEventListener("online", handleOnline);
window.addEventListener("offline", handleOffline);
```

**UI Adaptations**:

- Offline banner via `<OfflineIndicator />` Component
- All mutations disabled when offline (via Svelte reactive statements)
- Visual Feedback: disabled buttons, offline badge, lock icons
- Read-only mode: View cached data, no create/edit/delete/share

**Caching Strategy Details**:

**NetworkFirst** (API Routes):

- Tries network first with 5-second timeout
- Falls back to cache if network fails
- Auto-caches successful responses
- **Result**: No 500 errors offline, data always available

**Warmup Cache** (Automatic on Install):

```typescript
// Runs automatically when Service Worker installs
warmStrategyCache({
  urls: ["/api/v1/cards", "/api/v1/vouchers", "/api/v1/gift-cards", "/api/v1/dashboard"],
  strategy: NetworkFirst
});
```

- Preloads critical API routes in background
- No user interaction needed
- Enables offline mode from first visit

**Storage Layers**:

1. **Service Worker Cache** (CacheStorage API)
   - HTTP responses for network requests
   - Managed by custom service worker (injectManifest)
   - Expiration: 1 day (API), 1 hour (Dashboard/Merchants)

2. **IndexedDB** (via offline-db.ts)
   - Structured data (Cards, Vouchers, Gift Cards)
   - Pending operations queue (for sync)
   - No expiration (manual validation on online)

**Files**:

- Service Worker: [client/src/service-worker.ts](client/src/service-worker.ts) (Custom SW with Workbox Recipes)
- PWA Config: [client/vite.config.ts](client/vite.config.ts) (`injectManifest` strategy, globPatterns)
- SW Registration: [client/src/routes/+layout.svelte](client/src/routes/+layout.svelte) (Manual registration)
- Offline Store: [client/src/lib/stores/offline.ts](client/src/lib/stores/offline.ts) (cache validation logic)
- IndexedDB Wrapper: [client/src/lib/stores/offline-db.ts](client/src/lib/stores/offline-db.ts)
- API Clients: [client/src/lib/api/cards.ts](client/src/lib/api/cards.ts) (offline-first pattern)
- Offline Component: [client/src/lib/components/OfflineIndicator.svelte](client/src/lib/components/OfflineIndicator.svelte)

---

## 🎨 Frontend Architecture

### SvelteKit SPA

The system uses a **modern SvelteKit frontend** with JSON API backend:

**Architecture**:

- **Client**: SvelteKit SPA (`/client/`) with TypeScript
- **Backend**: Go/Echo JSON API (`/api/v1/`)
- **Build**: Vite bundles SvelteKit → `internal/assets/client/`
- **Deployment**: Embedded SvelteKit in Go Binary

**Client Structure**:

```typescript
// API Client (client/src/lib/api/)
import { cardsApi } from '$lib/api/cards';

const cards = await cardsApi.getAll();
await cardsApi.create({ merchant_name: 'IKEA', ... });
await cardsApi.delete(cardId);

// Svelte Stores (client/src/lib/stores/)
import { authStore } from '$lib/stores/auth';
import { offlineStore } from '$lib/stores/offline';
import { notificationStore } from '$lib/stores/notifications';

// Components (client/src/lib/components/)
<CardForm bind:card on:submit={handleSubmit} />
<BarcodeScanner on:scan={handleScan} />
<NotificationPanel />
<OfflineIndicator />

// Routes (client/src/routes/)
+page.svelte              # Dashboard
cards/+page.svelte        # Cards List
cards/[id]/+page.svelte   # Card Details
```

**Barcode Scanner** (TypeScript):

```typescript
// client/src/lib/components/BarcodeScanner.svelte
import { Html5Qrcode } from "html5-qrcode";

let scanning = false;
let qrScanner: Html5Qrcode;

async function startScanning() {
  scanning = true;
  qrScanner = new Html5Qrcode("scanner");
  await qrScanner.start({ facingMode: "environment" }, { fps: 10 }, (decodedText) => {
    dispatch("scan", { barcode: decodedText });
    stopScanning();
  });
}
```

**Build System**:

- **Development**: `docker compose up` (Vite dev server on port 5173 in Docker)
- **Production**: `npm run build:embed` → copies to `internal/assets/client/`
- **Go Embedding**: `//go:embed` in main.go serves SvelteKit build
- **PWA**: `@vite-pwa/sveltekit` plugin generates service worker

---

## 🧪 Testing

**Status**: ✅ Comprehensively Tested

**Backend Coverage** (Last Updated: 2026-03-04):

- ✅ **Repository Tests: 97.2%** (excellent!)
- ✅ **Models: 100%** (perfect!)
- ✅ **i18n: 97.0%** (excellent!)
- ✅ **Metrics: 96.0%** (excellent!)
- ✅ **Database: 88.9%** (very good)
- ✅ **Middleware: 87.4%** (very good)
- ✅ **Validation: 100%** (perfect!)
- ✅ **Audit: 75.5%** (good)
- ✅ **API Handler Tests: 80.4%** (target reached!)
- ✅ **Service Tests: 80.0%** (target reached!)
- ✅ **Race Detection**: All tests pass with `-race` flag

**Frontend E2E Tests** (Last Updated: 2026-02-20):

- ✅ **23 E2E Test Files** (Playwright, Chromium + Firefox + Mobile Chrome)
- ✅ **Authentication Tests**: Login, logout, registration, OAuth, admin access, protected routes
- ✅ **Dashboard Tests**: Stats, navigation, favorites, timezone handling
- ✅ **CRUD Tests**: Cards, vouchers, gift cards (create, read, update, delete)
- ✅ **Sharing Tests**: Permissions, transfer ownership, share management
- ✅ **Favorites Tests**: Toggle favorites, display in dashboard
- ✅ **Batch Operations Tests**: Bulk delete, share, transfer
- ✅ **Two-Factor Tests**: 2FA settings, challenge page, backup codes
- ✅ **Import/Export Tests**: JSON/CSV import, preview, error handling
- ✅ **Offline Tests**: PWA offline mode, cache validation, sync
- ✅ **Error Handling Tests**: Network errors, API errors (500, 400, 403, 404), timeout handling
- ✅ **Notification Tests**: Notification panel, mark as read, navigation
- ✅ **Form Validation Tests**: Input validation, date ranges, currency, barcode types
- ✅ **Internationalization Tests**: Language switching, translations (DE, EN, FR)
- ✅ **Admin Panel Tests**: User management, merchant management, audit log
- ✅ **Profile Tests**: Profile editing, account info, data export
- ✅ **Security Tests**: Password change, session management, 2FA settings
- ✅ **Config/Features Tests**: Feature toggles, OAuth config

**Testability**:

- ✅ All Services have Interfaces → Mock-based Testing
- ✅ Repositories have Interfaces → Testable without real DB
- ✅ AuthzService fully tested (7 Tests with PostgreSQL)
- ✅ CI/CD runs tests automatically on every PR
- ✅ Playwright E2E tests with Docker-based test environment

### Running E2E Tests

**Available npm scripts**:

```bash
# Run all E2E tests (full suite)
npm run test:e2e

# Run individual test file (recommended for development)
npm run test:e2e:run -- <test-file>

# Browser-specific tests
npm run test:e2e:chromium    # Chromium only
npm run test:e2e:firefox     # Firefox only
npm run test:e2e:mobile      # Mobile Chrome

# Additional utilities
npm run test:e2e:ui          # Open Playwright UI mode
npm run test:e2e:debug       # Debug mode
npm run test:e2e:headed      # Show browser window
npm run playwright:report    # Show HTML test report
```

**Running individual test files** (Playwright CLI wrapper):

The `test:e2e:run` script provides direct access to Playwright CLI with argument forwarding:

```bash
# Run a single test file
npm run test:e2e:run -- auth.spec.ts

# List tests without running (NO Docker startup, instant)
npm run test:e2e:run -- auth.spec.ts --list
npm run test:e2e:run -- cards.spec.ts --list

# Run with specific browser
npm run test:e2e:run -- cards.spec.ts --project=chromium
npm run test:e2e:run -- vouchers.spec.ts --project=firefox

# Run in headed mode (show browser)
npm run test:e2e:run -- import.spec.ts --headed

# Debug mode (step through tests)
npm run test:e2e:run -- sharing.spec.ts --debug

# Combine multiple flags
npm run test:e2e:run -- batch-operations.spec.ts --project=chromium --headed
```

**Important**: The `--list` flag is optimized to skip Docker environment setup, making test listing instant.

**Available test files** (23 test files in `client/tests/e2e/`):

- `admin.spec.ts`, `auth.spec.ts`, `batch-operations.spec.ts`, `cards.spec.ts`
- `config-and-features.spec.ts`, `dashboard.spec.ts`, `error-handling.spec.ts`
- `favorites.spec.ts`, `form-validation.spec.ts`, `gift-cards.spec.ts`
- `import.spec.ts`, `internationalization.spec.ts`, `merchant-management.spec.ts`
- `notifications.spec.ts`, `offline-storage.spec.ts`, `pagination.spec.ts`
- `password-reset.spec.ts`, `profile.spec.ts`, `security.spec.ts`
- `sharing.spec.ts`, `two-factor.spec.ts`, `verify-email.spec.ts`
- `vouchers.spec.ts`

**Docker Environment Setup** (Best Practice):

- **global.setup.ts** ([client/tests/global.setup.ts](client/tests/global.setup.ts)): Runs ONCE before all tests
  - Starts PostgreSQL + app-e2e Docker containers
  - Resets database to clean state (removes volume + recreates)
  - Waits for services to be healthy
- **global.teardown.ts** ([client/tests/global.teardown.ts](client/tests/global.teardown.ts)): Runs ONCE after all tests
  - Stops Docker containers
  - Optional cleanup via environment variables
- **Browser Projects**: Run in parallel after globalSetup (chromium, firefox, Mobile Chrome)
- **Conditional Skip**: Setup/teardown automatically skipped for `--list`, `--help`, `--version` flags
- **Environment Variables**:
  - `SKIP_E2E_SETUP=true` - Skip Docker setup (assumes environment already running)
  - `E2E_KEEP_CONTAINERS=true` - Keep containers running after tests (for debugging)
  - `E2E_VERBOSE_LOGS=true` - Show detailed Docker logs
  - `E2E_REMOVE_VOLUMES=true` - Remove Docker volumes on teardown

**Architecture**: Standard Playwright setup (globalSetup/globalTeardown only, no Project Dependencies)

**Example workflows**:

```bash
# 1. List available tests (instant, no Docker startup)
npm run test:e2e:run -- --list

# 2. Run single test file across all browsers (33 tests)
npm run test:e2e:run -- tests/e2e/auth.spec.ts

# 3. Run with specific browser only (11 tests)
npm run test:e2e:run -- tests/e2e/auth.spec.ts --project=chromium

# 4. Run multiple specific browsers (22 tests)
npm run test:e2e:run -- tests/e2e/auth.spec.ts --project=chromium --project=firefox

# 5. Run with grep pattern across all files
npm run test:e2e:run -- --grep "Authentication"

# 6. Run with debugging (headed mode, step-through)
npm run test:e2e:run -- tests/e2e/cards.spec.ts --headed --debug

# 7. Keep containers running for manual debugging
E2E_KEEP_CONTAINERS=true npm run test:e2e:run -- tests/e2e/auth.spec.ts

# ⚠️ WICHTIG: Vollständiger Pfad erforderlich (tests/e2e/...)
# ❌ FALSCH: auth.spec.ts (findet keine Tests)
# ✅ RICHTIG: tests/e2e/auth.spec.ts
```

---

## 🚀 Deployment

**Production Setup**: App runs behind **Traefik Reverse Proxy**

**Architecture**:

```
Client (HTTPS) → Traefik (TLS Termination) → Savvy App (HTTP:3000) → PostgreSQL
```

**Container**:

- Docker Multi-Stage Build ([Dockerfile](Dockerfile))
  - **Frontend Build Stage**: `client/` → Node.js + Vite → compiled SvelteKit
  - **Backend Build Stage**: Go + embedded client assets → binary
  - **Production Stage**: Distroless with binary (~50MB)
- Docker Compose for Development ([docker-compose.yml](docker-compose.yml))
  - **api**: Go backend with Air hot reload (port 8080)
  - **client**: Vite dev server (port 5173)
  - **postgres**: PostgreSQL 16 (port 5432)

**Development Workflow**:

```bash
# Local Development (separate frontend + backend)
docker compose up              # Vite (5173) + Go API (8080)
cd client && npm run dev       # Or: manual Vite dev server

# Production Build (embedded frontend)
cd client && npm run build:embed  # Build + copy to internal/assets/client/
make build                        # Go binary with embedded SvelteKit
```

**Kubernetes/Helm** (optional):

- Full Helm chart in `/deploy/helm/` (23 templates)
- Kustomize overlays: `dev`, `staging`, `production`
- Traefik IngressRoute, ConfigMap, Secrets

**Environment Variables**:

```bash
# Application
GO_ENV=production
SERVER_PORT=3000              # Production: 3000, Development: 8080
LOG_LEVEL=INFO                # DEBUG, INFO, WARN, ERROR
TIMEZONE=Europe/Zurich        # IANA timezone for date calculations (default: UTC)

# Database
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
AUTO_MIGRATE=true             # Auto-run database migrations on startup

# Session
SESSION_SECRET=change-me-in-production  # min 32 chars (validated)
SESSION_MAX_AGE=604800        # Session duration in seconds (default: 7 days)

# OAuth (optional)
OAUTH_CLIENT_ID=...
OAUTH_CLIENT_SECRET=...           # min 16 chars (validated if OAuth enabled)
OAUTH_ISSUER=https://auth.example.com/application/o/app/
OAUTH_ADMIN_EMAILS=admin@example.com  # Comma-separated admin emails

# Feature Toggles
ENABLE_CARDS=true
ENABLE_VOUCHERS=true
ENABLE_GIFT_CARDS=true
ENABLE_LOCAL_LOGIN=true
ENABLE_REGISTRATION=true

# Two-Factor Authentication (optional)
ENABLE_2FA=false              # Enable TOTP-based 2FA
TOTP_ISSUER=Savvy             # Issuer name shown in authenticator apps
TOTP_ENCRYPTION_KEY=...       # AES-256 key (32 bytes, required if 2FA enabled)

# Email/SMTP (optional)
SMTP_HOST=smtp.example.com
SMTP_PORT=587                 # Default: 587
SMTP_USERNAME=...
SMTP_PASSWORD=...
SMTP_FROM_EMAIL=noreply@example.com  # Required if SMTP enabled
SMTP_FROM_NAME=Savvy          # Sender display name
SMTP_USE_TLS=true             # Default: true

# Web Push Notifications (optional)
VAPID_PUBLIC_KEY=...          # VAPID public key
VAPID_PRIVATE_KEY=...         # VAPID private key
VAPID_SUBJECT=mailto:admin@example.com

# Expiry Reminders (optional)
ENABLE_EXPIRY_REMINDERS=true  # Default: true
REMINDER_DAYS_BEFORE=7,3,1    # Days before expiry (comma-separated)
REMINDER_CHECK_TIME=08:00         # Daily check time (HH:MM in configured TIMEZONE)

# Observability
OTEL_ENABLED=false
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

**Details**: [OPERATIONS.md](OPERATIONS.md) - Operations Guide

---

## 🎯 Important Concepts for AI Agents

### 1. Favorites System (Pinning)

**Polymorphic Design**:

```go
type UserFavorite struct {
    UserID       uuid.UUID
    ResourceType string    // "card", "voucher", "gift_card"
    ResourceID   uuid.UUID
    DeletedAt    *time.Time // Soft delete for Toggle
}
```

**Toggle Logic** (Layered Architecture):

```go
// Handler uses FavoriteService (not database.DB!)
func (h *FavoritesHandler) toggleFavorite(userID uuid.UUID, resourceType string, resourceID uuid.UUID) bool {
    ctx := context.Background()

    // ToggleFavorite handles the complete logic (Create/Restore/Delete)
    if err := h.favoriteService.ToggleFavorite(ctx, userID, resourceType, resourceID); err != nil {
        return false
    }

    isFavorite, err := h.favoriteService.IsFavorite(ctx, userID, resourceType, resourceID)
    return isFavorite
}
```

**Dashboard Integration**:

- Favorites replace "Recently Added" when available
- Mobile: Favorites appear BEFORE statistics
- Owner display for shared items: "by [Name]"

**Files**:

- Model: [internal/models/user_favorite.go](internal/models/user_favorite.go)
- Repository: [internal/repository/favorite_repository.go](internal/repository/favorite_repository.go)
- Service: [internal/services/favorite_service.go](internal/services/favorite_service.go)
- API: Favorites embedded in resource responses (Cards, Vouchers, Gift Cards)

### 2. Sharing System

**Granular Permissions**:

- **Cards**: `can_edit`, `can_delete`
- **Vouchers**: ALWAYS read-only (no edit rights)
- **Gift Cards**: `can_edit`, `can_delete`, `can_edit_transactions`

**Permission Check Pattern**:

```go
// 1. Check Ownership
isOwner := resource.UserID != nil && *resource.UserID == user.ID

// 2. If not Owner, check Share
if !isOwner {
    var share models.CardShare
    err := database.DB.Where("card_id = ? AND shared_with_id = ?",
                             resourceID, user.ID).First(&share).Error
    if err != nil {
        return http.StatusForbidden
    }
    // 3. Check Permission
    canEdit = share.CanEdit
}
```

**✅ Authorization Service** (`internal/services/authz_service.go`, 154 LOC):

```go
// Central Permission Checks for all Resources
type AuthzServiceInterface interface {
    CheckCardAccess(ctx, userID, cardID) (*ResourcePermissions, error)
    CheckVoucherAccess(ctx, userID, voucherID) (*ResourcePermissions, error)
    CheckGiftCardAccess(ctx, userID, giftCardID) (*ResourcePermissions, error)
}

// ResourcePermissions contains all Access Flags
type ResourcePermissions struct {
    CanView             bool
    CanEdit             bool
    CanDelete           bool
    CanEditTransactions bool // Gift Cards only
    IsOwner             bool
}
```

**Status**: ✅ Fully implemented and integrated in ALL 27 handlers- Eliminates duplicate Permission logic

- Consistent Authorization checks across all resources
- 7 Unit Tests with PostgreSQL (Owner, SharedUser, Permissions)

### 3. Barcode Features

**Two complementary technologies**:

1. **Browser Scanning** (html5-qrcode):
   - Browser-based (html5-qrcode library)
   - Camera access via MediaDevices API
   - Supported formats: CODE128, QR, EAN13, EAN8, UPC
   - **HTTPS Required**: Browser camera access requires HTTPS (except localhost)

2. **Server Generation** (bwip-js):
   - Server-side barcode generation
   - Higher quality and more formats
   - No client-side dependencies
   - Embedded in API responses

**SvelteKit Component** ([client/src/lib/components/BarcodeScanner.svelte](client/src/lib/components/BarcodeScanner.svelte)):

```typescript
import { Html5Qrcode } from "html5-qrcode";

let scanning = false;
let scanner: Html5Qrcode;

async function startScanning() {
  scanner = new Html5Qrcode("scanner");
  await scanner.start({ facingMode: "environment" }, { fps: 10 }, (decodedText) =>
    dispatch("scan", { barcode: decodedText })
  );
}
```

**Usage in Forms**:

```svelte
<BarcodeScanner on:scan={(e) => cardNumber = e.detail.barcode} />
```

**Barcode Display Component** ([client/src/lib/components/Barcode.svelte](client/src/lib/components/Barcode.svelte)):

```svelte
<!-- Displays server-generated barcodes -->
<img src="/api/v1/barcode?value={cardNumber}&type={barcodeType}" alt="Barcode" />
```

### 4. Audit Logging

**Automatic on all Deletions**:

```go
// Service Layer handles Deletion with Audit Logging
cardService.DeleteCard(ctx, cardID)  // → Service → Repository → GORM Hook → AuditLog Entry

// Alternative: Direct Audit Logging via AdminService
adminService.CreateAuditLog(ctx, &auditLog)
```

**Audit Log Schema**:

```sql
audit_logs:
  - user_id (who deleted)
  - action ("delete", "hard_delete", "restore")
  - resource_type ("cards", "vouchers", etc.)
  - resource_id (UUID)
  - resource_data (JSONB Snapshot)
  - ip_address + user_agent
  - created_at
```

**Details**: [OPERATIONS.md](OPERATIONS.md) - Audit Logging (Line 18-143)

### 5. Two-Factor Authentication (TOTP)

**Architecture**:

- **Standard**: RFC 6238 (TOTP) with HMAC-SHA1
- **Backup Codes**: 10 codes per user (format: `XXXX-XXXX`, alphanumeric, no ambiguous chars)
- **Encryption**: AES-256-GCM for TOTP secrets in DB
- **Hashing**: Bcrypt for backup codes (stored as JSON array)
- **Rate Limiting**: Applied to verification endpoints

**Flow**:

```
1. User enables 2FA → GenerateSetup() → QR code + backup codes
2. User scans QR → enters code → VerifyAndEnable() → 2FA active
3. Login → password OK → requires_2fa: true → /login/2fa → Verify() → session created
4. Backup code → VerifyBackupCode() → code consumed (one-time use)
```

**Service Interface** (`internal/services/totp_service.go`):

```go
type TOTPServiceInterface interface {
    GenerateSetup(ctx, userID, email) (*TOTPSetup, error)
    VerifyAndEnable(ctx, userID, code) error
    Verify(ctx, userID, code) error
    VerifyBackupCode(ctx, userID, code) error
    Disable(ctx, userID, code) error
    IsEnabled(ctx, userID) (bool, error)
    RegenerateBackupCodes(ctx, userID, code) ([]string, error)
}
```

**Files**:

- Service: [internal/services/totp_service.go](internal/services/totp_service.go)
- Handler: [internal/handlers/api/totp.go](internal/handlers/api/totp.go)
- Model: [internal/models/user_totp.go](internal/models/user_totp.go)
- Frontend: [client/src/lib/components/TwoFactorSettings.svelte](client/src/lib/components/TwoFactorSettings.svelte)
- Challenge Page: [client/src/routes/login/2fa/](client/src/routes/login/2fa/)

### 6. Data Import/Export

**Import Formats**:

1. **JSON Export Format** - Full import (cards, vouchers, gift cards)
2. **CSV** - Resource-specific import with column mapping

**Import Flow**:

```
1. Upload file → PreviewJSON() → counts (cards: N, vouchers: N, gift_cards: N)
2. User confirms → ImportJSON() / ImportCardsCSV() → ImportResult
3. Result: success counts + skipped + detailed errors with row numbers
```

**Features**:

- Preview before import (no data changes)
- Partial success (continues on error, reports failures)
- Merchant auto-resolution/creation
- CSV column validation

**Export**: Full JSON export of all user data (`GET /api/v1/export`)

**Files**:

- Import Service: [internal/services/import_service.go](internal/services/import_service.go)
- Export Service: [internal/services/export_service.go](internal/services/export_service.go)
- Handler: [internal/handlers/api/import.go](internal/handlers/api/import.go)
- Frontend: [client/src/lib/components/ImportDialog.svelte](client/src/lib/components/ImportDialog.svelte)

### 7. Batch Operations

**Supported Operations** (12 endpoints):

| Operation | Cards | Vouchers       | Gift Cards |
| --------- | ----- | -------------- | ---------- |
| Delete    | ✅    | ✅             | ✅         |
| Share     | ✅    | ✅ (read-only) | ✅         |
| Transfer  | ✅    | ✅             | ✅         |
| Export    | ✅    | ✅             | ✅         |

**Limits**: Max **50 items** per request

**Semantics**:

- **Delete**: All-or-nothing (checks all permissions before any deletion)
- **Share/Transfer**: Partial success (continues on error, reports failures)
- **Export**: JSON file download with `Content-Disposition` header (checks view permissions)

**Response**:

```json
{ "success_count": 42, "failed": [{ "id": "uuid", "error": "reason" }] }
```

**Files**:

- Handler: [internal/handlers/api/batch.go](internal/handlers/api/batch.go)
- DTOs: [internal/models/batch.go](internal/models/batch.go)

### 8. Web Push Notifications

**Implementation**: VAPID-based Web Push via `webpush-go` library

**Features**:

- Multi-device support (sends to all user subscriptions)
- Auto-cleanup of expired subscriptions (410 Gone)
- JSON payload with title, body, URL, icon
- TTL: 86400 seconds (24 hours)

**Service Interface** (`internal/services/push_service.go`):

```go
type PushServiceInterface interface {
    Subscribe(ctx, userID, endpoint, p256dh, auth, userAgent) error
    Unsubscribe(ctx, endpoint) error
    SendPushToUser(ctx, userID, title, body, url) error
    GetVAPIDPublicKey() string
    IsEnabled() bool
}
```

**Files**:

- Service: [internal/services/push_service.go](internal/services/push_service.go)
- Model: [internal/models/push_subscription.go](internal/models/push_subscription.go)
- Frontend: [client/src/lib/components/PushNotificationSettings.svelte](client/src/lib/components/PushNotificationSettings.svelte)

### 9. Expiry Reminders

**Multi-Channel Reminders** before voucher/gift card expiry:

**Channels**:

1. **In-App Notifications** - Always created
2. **Push Notifications** - If VAPID configured AND `PushNotificationsEnabled && PushRemindersEnabled`
3. **Email** - If SMTP configured AND `EmailNotificationsEnabled && EmailRemindersEnabled`

**Per-Channel Notification Preferences** (6 toggles on User model):

| Field                       | Controls                          | Default |
| --------------------------- | --------------------------------- | ------- |
| `PushNotificationsEnabled`  | Global push channel on/off        | `true`  |
| `EmailNotificationsEnabled` | Global email channel on/off       | `true`  |
| `PushRemindersEnabled`      | Push for expiry & validity start  | `true`  |
| `PushSharingEnabled`        | Push for share & transfer         | `true`  |
| `EmailRemindersEnabled`     | Email for expiry & validity start | `true`  |
| `EmailSharingEnabled`       | Email for share & transfer        | `true`  |

**Delivery Gating Logic**:

| Notification Type | Push Condition                   | Email Condition                    |
| ----------------- | -------------------------------- | ---------------------------------- |
| Expiry/Validity   | `Push…Enabled && Push…Reminders` | `Email…Enabled && Email…Reminders` |
| Share/Transfer    | `Push…Enabled && Push…Sharing`   | `Email…Enabled && Email…Sharing`   |

**Settings UI**: Nested toggle layout — category toggles only visible
when parent channel is enabled.

**Unsubscribe**: Email unsubscribe links disable the specific
per-channel category (`EmailRemindersEnabled` or
`EmailSharingEnabled`), not the global channel.

**Configuration**:

- Default windows: 7, 3, 1 days before expiry
- Configurable via `REMINDER_DAYS_BEFORE` env var
- Daily check time: `REMINDER_CHECK_TIME` (default: "08:00" in configured `TIMEZONE`)

**Duplicate Prevention**: `expiry_reminder_sents` table tracks
(user_id, resource_type, resource_id, days_before) → no duplicate reminders

**Localization**: Supports DE, EN, FR for emails and notifications

**Prometheus Metrics** (9 notification gauges):

- `push_subscriptions_total`, `push_subscribed_users_total`, `email_verified_users_total`
- `push_notifications_enabled_total`, `email_notifications_enabled_total`
- `push_reminders_enabled_total`, `push_sharing_enabled_total`
- `email_reminders_enabled_total`, `email_sharing_enabled_total`

**Grafana Dashboard**: "Notifications" section in
[savvy-overview.json](deploy/grafana/savvy-overview.json) with:

- Stat panels: Push Subscriptions, Push Users, Email Verified Users
- Bar gauge: All 6 preference toggle counts

**Files**:

- Service: [internal/services/reminder_service.go](internal/services/reminder_service.go)
- Notification: [internal/services/notification_service.go](internal/services/notification_service.go)
- Email Token: [internal/services/email_token_service.go](internal/services/email_token_service.go)
- Model: [internal/models/expiry_reminder.go](internal/models/expiry_reminder.go)
- User Model: [internal/models/user.go](internal/models/user.go) (6 preference fields)
- Metrics: [internal/metrics/metrics.go](internal/metrics/metrics.go) (`NotificationMetrics` struct)
- Email Template: [internal/email/templates/expiry_reminder.html](internal/email/templates/expiry_reminder.html)
- Dashboard: [deploy/grafana/savvy-overview.json](deploy/grafana/savvy-overview.json)

### 10. Account Management

**Email Verification & Password Reset**:

- Token-based flows with SHA-256 hashed tokens in DB
- Configurable expiration
- Multi-language email templates (DE, EN, FR)

**GDPR-Compliant Account Deletion** (9 stages):

1. Remove incoming shares (shared WITH user)
2. Remove user preferences (favorites, notifications)
3. Remove outgoing shares (OF user's resources)
4. Delete owned resources (cards, vouchers, gift cards)
5. Remove auth data (email tokens, push subscriptions, TOTP)
6. Create audit log entry + nullify user references in audit logs
7. Nullify transaction references
8. Hard DELETE user record (unscoped)
9. Send async confirmation email

**Files**:

- Account Service: [internal/services/account_service.go](internal/services/account_service.go)
- Email Token Service: [internal/services/email_token_service.go](internal/services/email_token_service.go)
- Email Service: [internal/email/email.go](internal/email/email.go)
- Email Templates: [internal/email/templates/](internal/email/templates/)

### 11. Email Templates & SMTP Configuration

**Email System Architecture**:

- **SMTP Service**: RFC-compliant SMTP implementation with STARTTLS support (Port 587)
- **Template Engine**: Go `html/template` with embedded FS (`//go:embed templates/*.html`)
- **Dynamic Logo URL**: Logo URL derived from `FrontendURL` config (no hardcoded domains)
- **Multilingual**: German, English, French support for all email types
- **Brand Colors**: Savvy cyan (#06b6d4) gradient with consistent styling

**Email Templates** (5 templates):

1. **Test Email** (`test_email.html`) - SMTP validation with status box
2. **Password Reset** (`password_reset.html`) - Password reset with CTA button
3. **Email Verification** (`email_verification.html`) - Email verification with CTA button
4. **Account Deleted** (`account_deleted.html`) - Deletion confirmation
5. **Expiry Reminder** (`expiry_reminder.html`) - Voucher/Gift Card expiry notification

**Template Features**:

- Unified Savvy branding with cyan (#06b6d4 → #0891b2) gradient header
- Dynamic logo URL: `{{.LogoURL}}` (from `FrontendURL` config)
- Responsive design (max-width: 600px, mobile-friendly)
- Consistent structure: Header → Content → Footer
- Localized strings passed via `emailStrings.Data` map

**SMTP Configuration**:

```bash
SMTP_HOST=smtp.mailchannels.net    # SMTP server host
SMTP_PORT=587                       # Port 587 = STARTTLS (RFC 4409)
SMTP_USERNAME=...                   # Authentication username
SMTP_PASSWORD=...                   # Authentication password
SMTP_FROM_EMAIL=noreply@example.com # Sender address
SMTP_FROM_NAME=Savvy                # Sender display name
SMTP_USE_TLS=true                   # Enable STARTTLS (default: true)
```

**STARTTLS Protocol** (Port 587):

```go
// Correct implementation for Port 587 (RFC 4409)
conn, err := net.Dial("tcp", addr)           // 1. Plain TCP connection
client, err := smtp.NewClient(conn, host)    // 2. SMTP client
err := client.StartTLS(tlsConfig)            // 3. STARTTLS upgrade to TLS
err := client.Auth(auth)                     // 4. Authentication
err := client.Mail(from)                     // 5. MAIL FROM
err := client.Rcpt(to)                       // 6. RCPT TO
w, err := client.Data()                      // 7. DATA
_, err := w.Write(msg)                       // 8. Write email body
err := w.Close()                             // 9. Close message
err := client.Quit()                         // 10. Quit
```

**Important**: Port 587 requires STARTTLS (not direct TLS like port 465).
Using `tls.Dial()` directly will fail with "Client Quit Before Message" error.

**Testing SMTP**:

- Admin Dashboard: `/admin/system-health` → "Send Test Email" button
- Health Check: `GET /ready` shows SMTP status (`healthy`, `unhealthy`, `not_configured`)
- Pod Logs: `kubectl logs -n savvy -l app.kubernetes.io/name=savvy --tail=100 | grep -i smtp`

**Files**:

- Email Service: [internal/email/email.go](internal/email/email.go)
- Templates: [internal/email/templates/](internal/email/templates/)
- Health Service: [internal/services/health_service.go](internal/services/health_service.go)
- Admin Handler: [internal/handlers/api/admin.go](internal/handlers/api/admin.go) (`SendTestEmail()`)

### 12. Server-Side Session Management

**Architecture**: PostgreSQL-backed session store replacing cookie-based storage

**Key Components**:

- **PGStore** (`internal/middleware/pgstore.go`): Full `gorilla/sessions.Store` implementation
  - SHA-256 hashed token stored in DB, raw token in cookie
  - 512-bit cryptographically random session tokens
  - IP address and User-Agent tracked per session
  - `LastActiveAt` throttled to once per 60 seconds
- **Session Keys** (`internal/middleware/session_keys.go`): Centralized constants + typed helpers
  - `CreateUserSession`, `Create2FAPendingSession`, `DestroySession`
  - `CreateImpersonationSession`, `StopImpersonationSession`
- **Session Service** (`internal/services/session_service.go`): Business layer
  - User-agent parsing: device type + browser detection
  - `ListUserSessions`, `RevokeSession`, `RevokeOtherSessions`
- **Session Repository** (`internal/repository/session_repository.go`): CRUD + bulk revocation

**API Endpoints**:

- `GET /api/v1/profile/sessions` — List active sessions (device, browser, IP, `is_current`)
- `DELETE /api/v1/profile/sessions/:id` — Revoke specific session
- `POST /api/v1/profile/sessions/revoke-others` — Revoke all except current

**Security Features**:

- Stale session invalidation: `session_created_at` vs `user.PasswordChangedAt`
- Password change → `RevokeOtherSessions()` automatically
- Password reset → `RevokeAllSessions()` automatically
- Hourly background goroutine cleans up expired sessions

**Files**:

- PGStore: [internal/middleware/pgstore.go](internal/middleware/pgstore.go)
- Session Keys: [internal/middleware/session_keys.go](internal/middleware/session_keys.go)
- Session Model: [internal/models/session.go](internal/models/session.go)
- Repository: [internal/repository/session_repository.go](internal/repository/session_repository.go)
- Service: [internal/services/session_service.go](internal/services/session_service.go)
- Handler: [internal/handlers/api/sessions.go](internal/handlers/api/sessions.go)
- Frontend API: [client/src/lib/api/sessions.ts](client/src/lib/api/sessions.ts)
- Security Page: [client/src/routes/security/+page.svelte](client/src/routes/security/+page.svelte)

---

## 📝 Changelog

### Unreleased

**Server-Side Sessions, Settings Refactoring, Merchant Overview & Batch Export**:

- ✅ **Server-Side Sessions** - PostgreSQL-backed session store (PGStore)
  - SHA-256 hashed tokens, 512-bit random, IP/UA tracking
  - Session management API (list, revoke, revoke all others)
  - Stale session invalidation after password change
  - Hourly expired session cleanup goroutine
- ✅ **Settings Page Refactoring** - Monolith split into dedicated pages
  - `/profile` — Profile editing, account info, export, deletion
  - `/security` — Password change, 2FA, active sessions
  - `/notifications` — Notification list + preferences
  - 4 new reusable components (ProfileSection, SecuritySection, NotificationsSection, ToggleSwitch)
- ✅ **Merchant Overview** - New public merchant browsing
  - `/merchants` — Merchant list with aggregated counts and balances
  - `/merchants/:id` — Merchant detail with all resources
  - Client-side aggregation utility
- ✅ **Batch Export** - Export selected items as JSON download
  - 3 new endpoints: `POST /api/v1/{resource}/batch/export`
  - View permission validation per item
- ✅ **Admin Users Page** - Extracted to `/admin/users`
- ✅ **Navigation Updates** - Merchants link, admin dropdown menu, 4-column mobile grid
- ✅ **Session Key Centralization** - Typed constants replacing magic strings
- ✅ **Docker Watch Mode** - `develop.watch` configuration
- ✅ **i18n** - ~167 new keys per language (DE/EN/FR)

**Enhanced Monitoring, Notifications & Email System**:

- ✅ **Per-Channel Notification Preferences** - 6 granular toggles
  - 2 channel toggles (Push, Email) + 4 category toggles (Reminders, Sharing)
  - Nested Settings UI with conditional subcategory visibility
  - 9 Prometheus metrics + Grafana dashboard "Notifications" section
  - Updated unsubscribe flow, 20 Go tests + 7 E2E tests
- ✅ **Enhanced Health Checks** - Comprehensive service validation in `/ready` endpoint
  - Two-tier readiness: critical (database) vs optional (SMTP, OAuth, VAPID, 2FA)
  - Parallel checks with goroutines (5s timeout)
  - Detailed JSON response with per-service status, latency, errors
- ✅ **Admin System Health Dashboard** - Real-time monitoring at `/admin/system-health`
  - Auto-refresh every 30 seconds (toggleable)
  - Test email functionality for SMTP validation
  - SVG status indicators with color coding
- ✅ **SMTP STARTTLS Fix** - Corrected Port 587 implementation
  - Changed from direct TLS (`tls.Dial()`) to STARTTLS protocol
  - Fixes "Client Quit Before Message" error
  - RFC 4409 compliant
- ✅ **Email Template Redesign** - Unified Savvy branding
  - Cyan brand color (#06b6d4) instead of blue
  - Dynamic logo URL from `FrontendURL` config (no hardcoded domains)
  - Consistent gradient header across all 5 templates
  - Multilingual support (DE/EN/FR)

### Version 1.1.0 (2026-02-20)

**Feature Release** - Major new features added:

- ✅ **Two-Factor Authentication** - TOTP with backup codes, QR setup, AES-256 encryption
- ✅ **Data Import/Export** - JSON and CSV import with preview, full JSON export
- ✅ **Batch Operations** - Bulk delete, share, transfer (max 50 items)
- ✅ **Web Push Notifications** - VAPID-based push via service worker
- ✅ **Expiry Reminders** - Multi-channel reminders (in-app, push, email)
- ✅ **Email Verification** - Token-based verification and password reset
- ✅ **Account Management** - Profile editing, password change, GDPR deletion
- ✅ **Voucher Status** - Active/inactive/expired lifecycle
- ✅ **E2E Test Suite** - 15 Playwright test files covering all features

### Version 1.0.0 (2026-02-14)

**Initial Public Release** - First stable version with core feature set:

- ✅ SvelteKit SPA, PWA, Layered Architecture
- ✅ Cards, Vouchers, Gift Cards with CRUD
- ✅ Sharing, Transfer, Favorites, Notifications
- ✅ Session auth, OAuth/OIDC, CSRF, CSP
- ✅ Barcode generation/scanning, i18n (DE/EN/FR)

**Full Changelog**: [CHANGELOG.md](CHANGELOG.md)

---

## 🎯 Open Tasks

**Production Readiness Score**: **9.8/10** ✅ Production-Ready

**LOW Priority** (Post-Release):

- ⚠️ **Shares Handler Tests**: 0% coverage (complex adapter pattern)
- ⚠️ **Performance**: Add pagination, caching (Redis)
- ⚠️ **Documentation**: API Documentation (OpenAPI/Swagger)

**COMPLETED** ✅:

- ✅ **v2.0.0 Release Preparation**: Code quality, CORS externalization, CI linting, PGStore tests
- ✅ **Global Search**: Cross-resource search with filters and Playwright E2E tests
- ✅ **Structured Logging**: Migrated all log.Println to slog structured logging
- ✅ **CORS Externalized**: Configurable via CORS_ALLOWED_ORIGINS env var
- ✅ **Server-Side Sessions**: PostgreSQL-backed PGStore with session management API
- ✅ **Settings Refactoring**: Split into /profile, /security, /notifications pages
- ✅ **Merchant Overview**: Browse data by merchant with aggregated stats
- ✅ **Batch Export**: Export selected items as JSON download
- ✅ **Two-Factor Authentication**: TOTP, backup codes, AES-256 encryption
- ✅ **Data Import/Export**: JSON/CSV import with preview, full export
- ✅ **Batch Operations**: Bulk delete, share, transfer, export (max 50 items)
- ✅ **Web Push Notifications**: VAPID-based push via service worker
- ✅ **Expiry Reminders**: Multi-channel (in-app, push, email)
- ✅ **Email Verification & Password Reset**: Token-based, SMTP
- ✅ **Account Management**: Profile, password change, GDPR deletion
- ✅ **Voucher Status**: Active/inactive/expired lifecycle
- ✅ **CI/CD Pipeline**: 4 GitHub Actions workflows (ci.yml, e2e.yml, security.yml, release.yml)
- ✅ **Security Headers**: X-XSS-Protection, X-Content-Type-Options, X-Frame-Options, HSTS
- ✅ **E2E Testing**: 23 Playwright test files covering all major features
- ✅ **AuthzService**: Fully integrated in ALL 27 handlers
- ✅ **PWA Implementation**: Service Worker, Manifest, Offline-Mode
- ✅ **Svelte 5 Migration**: Modern {@render children()} syntax

---

## 🛠️ Troubleshooting

### "Build failed"

Check the build logs for errors. The build process includes client build and Go compilation.

### "Database connection refused"

```bash
docker compose ps
docker compose logs postgres
```

### "Barcode scanner not working"

- HTTPS Required (Browser camera access)
- html5-qrcode library must be loaded (SvelteKit component)
- User must grant Camera Permission

---

## 📚 Additional Resources

### For AI Agents & Developers

- **Layered Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **Database Schema**: See Mermaid ER diagram in [ARCHITECTURE.md](ARCHITECTURE.md) or [README.md](README.md)
- **Deployment**: [OPERATIONS.md](OPERATIONS.md)
- **Development Setup**: [DEVELOPMENT.md](DEVELOPMENT.md)

### For Contributors

- **Contribution Guidelines**: [CONTRIBUTING.md](CONTRIBUTING.md)
- **Code Style & PR Process**: see CONTRIBUTING.md
- **Governance**: [GOVERNANCE.md](GOVERNANCE.md)

### For Users

- **User Guide**: [README.md](README.md)
- **Support Resources**: [SUPPORT.md](SUPPORT.md)
- **Security Policy**: [SECURITY.md](SECURITY.md)
- **Release History**: [CHANGELOG.md](CHANGELOG.md)

---

**End of AI Agent Documentation**

This project follows Layered Architecture with Go + Echo + SvelteKit + TypeScript.
All technical details are organized in specialized documents - avoid redundancies!
