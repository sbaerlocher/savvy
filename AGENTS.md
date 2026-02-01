# Savvy System - AI Agent Documentation

**Letzte Aktualisierung**: 2026-02-01
**Projekt-Typ**: Full-Stack Web Application
**Tech Stack**: Go + Echo + Templ + HTMX + Alpine.js + GORM + PostgreSQL
**Zweck**: Digitale Verwaltung von Kundenkarten, Gutscheinen und Geschenkkarten mit Sharing-Funktionalität

---

## 🎯 Dokumentations-Übersicht für AI-Agenten

Diese Datei dient als **zentrale Navigation** für AI-Agenten. Alle Details sind in spezialisierten Dokumenten organisiert.

### 📚 Dokumentationsstruktur

| Dokument | Zweck | Zielgruppe |
|----------|-------|------------|
| **AGENTS.md** | Zentrale Navigation, Quick Reference | AI-Agenten |
| [README.md](README.md) | Quick Start, Features, User Guide | Menschen (Entwickler, User) |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Technische Architektur, Diagramme, Performance | AI-Agenten + Entwickler |
| [OPERATIONS.md](OPERATIONS.md) | Audit Logging, Observability, Deployment | AI-Agenten + DevOps |
| [TODO.md](TODO.md) | Offene Aufgaben, Roadmap, Priorities | AI-Agenten + Entwickler |
| [docs/PWA.md](docs/PWA.md) | Progressive Web App Features, Offline-Modus | AI-Agenten + Entwickler |
| [migrations/README.md](migrations/README.md) | Datenbank-Schema Details | AI-Agenten + DB-Entwickler |

**Wichtig**: Redundanzen vermeiden! Details stehen NUR in den spezialisierten Dokumenten.

---

## 🚀 Quick Start für AI-Agenten

### 1. Projekt verstehen

**Lies ZUERST**: [README.md](README.md) für:
- ✅ Feature-Übersicht (Cards, Vouchers, Gift Cards, Sharing)
- ✅ Tech Stack Details
- ✅ Installation & Setup
- ✅ Database Schema (High-Level)

**Für tiefe technische Details**: [ARCHITECTURE.md](ARCHITECTURE.md)
**Für Deployment**: [OPERATIONS.md](OPERATIONS.md)

### 2. Code-Änderungen durchführen

**Architektur** (Clean Architecture mit 3 Layers):
```
Handlers (Presentation) → Services (Business Logic) → Repositories (Data Access)
```

**Wichtige Verzeichnisse**:
```
cmd/server/main.go                    # Entrypoint, Routing
internal/handlers/                    # HTTP Handlers (Echo Context)
  ├── cards/                         # Cards CRUD (modular, ~80 LOC/file)
  ├── vouchers/                      # Vouchers CRUD
  ├── gift_cards/                    # Gift Cards CRUD + Transactions
  ├── auth.go                        # Authentication
  ├── oauth.go                       # OAuth/OIDC
  ├── admin.go                       # Admin Panel
  ├── favorites.go                   # Favorites Toggle
  └── home.go                        # Dashboard

internal/services/                    # Business Logic
  ├── card_service.go                # Card business logic
  ├── voucher_service.go             # Voucher business logic
  ├── gift_card_service.go           # Gift Card business logic
  ├── favorite_service.go            # Favorites logic
  ├── merchant_service.go            # Merchant management
  ├── share_service.go               # Sharing logic
  ├── dashboard_service.go           # Dashboard queries
  ├── authz_service.go               # Authorization checks (✅ IMPLEMENTED)
  └── container.go                   # Dependency injection

internal/repository/                  # Data Access
  ├── card_repository.go             # Card GORM queries
  ├── voucher_repository.go          # Voucher GORM queries
  ├── gift_card_repository.go        # Gift Card GORM queries
  └── ...                            # Other repositories

internal/models/                      # GORM Models
  ├── user.go                        # User + Authentication
  ├── merchant.go                    # Merchant/Brands
  ├── user_favorite.go               # Polymorphic Favorites
  ├── card.go + voucher.go + gift_card.go
  └── *_share.go                     # Sharing models

internal/templates/                   # Templ Templates (Type-safe HTML)
  ├── layout.templ                   # Base + Nav + Alpine.js Functions
  ├── home.templ                     # Dashboard + Favorites
  ├── cards.templ                    # Cards UI
  └── ...                            # Other templates

internal/middleware/                  # Echo Middleware
  ├── auth.go                        # Authentication
  ├── admin.go                       # Admin check
  ├── feature.go                     # Feature toggles
  └── session.go                     # Session management

internal/config/                      # Configuration
  └── config.go                      # Environment variables

migrations/                           # Database Migrations
  └── *.up.sql / *.down.sql          # Gormigrate migrations
```

**Details**: Siehe [ARCHITECTURE.md](ARCHITECTURE.md) - Package Structure (Zeile 168-221)

### 3. Feature Toggles

Das System unterstützt **5 Feature Toggles** via Environment Variables:

```bash
# Resource Toggles
ENABLE_CARDS=true                    # Cards feature
ENABLE_VOUCHERS=true                 # Vouchers feature
ENABLE_GIFT_CARDS=true               # Gift Cards feature

# Authentication Toggles
ENABLE_LOCAL_LOGIN=false             # Email/Password (false = OAuth only)
ENABLE_REGISTRATION=false            # User registration
```

**Implementation**:
- Middleware in [internal/middleware/feature.go](internal/middleware/feature.go)
- Template Conditionals in [internal/templates/layout.templ](internal/templates/layout.templ)
- Config Injection: `cmd/server/main.go` Lines 191-203

---

## 🏗️ Architektur-Highlights

### Clean Architecture Pattern

**Dependency Flow**:
```
Handlers → Services (Interfaces) → Repositories (Interfaces) → GORM Models → PostgreSQL
```

**WICHTIG**:
- Handler kennt NICHT Database (nur Services via Interfaces)
- Services kennen NICHT Echo Context (nur Repository Interfaces)
- Alle Services haben Interfaces → Testbar mit Mocks

**Details**: [ARCHITECTURE.md](ARCHITECTURE.md) - Clean Architecture Pattern (Zeile 87-126)

### Database Schema

**10 Haupttabellen** (siehe [migrations/README.md](migrations/README.md)):
1. `users` - Benutzer mit Auth
2. `merchants` - Händler (zentral für alle Typen)
3. `user_favorites` - **Polymorphic** Favorites (Cards, Vouchers, Gift Cards)
4. `cards` + `card_shares` - Kundenkarten + Sharing
5. `vouchers` + `voucher_shares` - Gutscheine + Sharing (read-only)
6. `gift_cards` + `gift_card_shares` + `gift_card_transactions` - Geschenkkarten + granulare Permissions

**Besonderheiten**:
- ✅ UUIDs statt Integer IDs (Security)
- ✅ Polymorphic Favorites (`resource_type` + `resource_id`)
- ✅ Database Trigger: `recalculate_gift_card_balance()` - Auto-update bei Transaktionen
- ✅ Database Trigger: `enforce_lowercase_email()` - Email Normalization

**ERD Diagramm**: [ARCHITECTURE.md](ARCHITECTURE.md) - Zeile 278-391

### Performance Optimierungen

**Dashboard**:
- **40% schneller**: N+1 Query Fix (10+ → 8 Queries)
- Parallelisierung mit Goroutines für Stats
- `GROUP BY` Aggregation für Favorites

**Gift Card Balance**:
- **78% schneller**: Database Trigger statt Runtime-Berechnung
- Balance wird bei Transaction INSERT/UPDATE/DELETE automatisch aktualisiert
- Keine `Preload("Transactions")` nötig

**Details**: [ARCHITECTURE.md](ARCHITECTURE.md) - Performance Optimizations (Zeile 627-700)

---

## 🔐 Sicherheit

**Implementierte Features**:
- ✅ Session-based Authentication (Gorilla Sessions)
- ✅ Bcrypt Password Hashing (DefaultCost)
- ✅ CSRF Protection (Echo Middleware + HTMX Integration)
- ✅ OAuth/OIDC Support (Provider-agnostisch)
- ✅ SQL Injection Prevention (GORM Parameterized Queries)
- ✅ XSS Protection (Templ Auto-Escaping)
- ✅ Granulare Sharing-Berechtigungen
- ✅ Audit Logging (alle Deletions)
- ✅ Rate Limiting (Auth Endpoints)
- ✅ Email Normalization (lowercase in DB)

**Details**:
- Architektur: [ARCHITECTURE.md](ARCHITECTURE.md) - Security Architecture (Zeile 427-543)
- Operations: [OPERATIONS.md](OPERATIONS.md) - Security (Zeile 207-283)

---

## 📊 Observability

**Stack**:
- **Metrics**: Prometheus (`/metrics` endpoint)
- **Logs**: Structured Logging (slog, JSON)
- **Traces**: OpenTelemetry (optional, via `OTEL_ENABLED=true`)
- **Health Checks**: `/health` (liveness), `/ready` (readiness)

**Key Metrics**:
- `http_request_duration_seconds`, `http_requests_total`
- `cards_total`, `vouchers_total`, `gift_cards_total`, `users_total`
- `active_sessions`, `db_connections_active`, `db_connections_idle`

**Details**: [OPERATIONS.md](OPERATIONS.md) - Observability (Zeile 146-204)

---

## 📱 Progressive Web App (PWA)

**Status**: ✅ Implemented (v1.1.0)

### Offline-Funktionalität

**Was funktioniert offline**:
- ✅ Karten/Gutscheine/Geschenkkarten ansehen (gecachte Daten)
- ✅ Geteilte Items anzeigen
- ✅ Favoriten durchsuchen
- ✅ Barcode-Details ansehen
- ✅ Dashboard mit Statistiken (cached)
- ✅ Filter & Sortierung (client-side)
- ✅ Barcode-Scanner (Camera API)

**Was NICHT offline funktioniert**:
- ❌ Neue Items erstellen
- ❌ Items bearbeiten/löschen
- ❌ Sharing verwalten
- ❌ Favoriten hinzufügen/entfernen
- ❌ Transaktionen (Gift Cards)

### Implementierung

**Service Worker**: `static/service-worker.js`
- **Strategie**: Network First, Cache Fallback
- **Gecachte Routes**: `/`, `/cards`, `/vouchers`, `/gift-cards`, `/cards/:id`, etc.
- **Cache-Version**: `savvy-v1.0.0`

**Offline-Erkennung**: Alpine.js Component in `layout.templ`
- Gelbes Banner bei Offline-Status
- Buttons automatisch deaktiviert
- "Erneut versuchen" Funktion

**UI-Anpassungen**:
- Alle Edit/Delete/Share/Create Buttons deaktiviert wenn offline
- Visual Feedback: `opacity-50 cursor-not-allowed`
- German Tooltips: "Bearbeiten nur online möglich", etc.

**Files**:
- Service Worker: [static/service-worker.js](static/service-worker.js)
- PWA Manifest: [static/manifest.json](static/manifest.json)
- Offline Page: [internal/templates/offline.templ](internal/templates/offline.templ)
- Layout Integration: [internal/templates/layout.templ](internal/templates/layout.templ) (Zeile 54-164)

**Details**: [docs/PWA.md](docs/PWA.md) - Vollständige PWA-Dokumentation

---

## 🎨 Frontend Patterns

### HTMX (Dynamic Updates ohne Page Reload)

```html
<!-- Delete mit Confirmation -->
<button
  hx-delete="/cards/123"
  hx-confirm="Karte wirklich löschen?"
  hx-target="closest div"
  hx-swap="outerHTML"
>
  Löschen
</button>
```

### Alpine.js (Client-Side State)

**Barcode Scanner**:
```html
<div x-data="cardForm()">  <!-- oder voucherForm(), giftCardForm() -->
  <input type="text" x-model="cardNumber" />
  <button @click="startScanning()">Scannen</button>
</div>
```

**Filter & Sort**:
```html
<div x-data="cardsFilter('user-id')" x-init="init()">
  <input x-model="search" @input="updateVisibility()" />
  <select x-model="sortBy" @change="updateSort()">...</select>
</div>
```

**Functions** definiert in [static/js/src/scanner.js](static/js/src/scanner.js):
- `window.cardForm()`, `window.voucherForm()`, `window.giftCardForm()` - Scanner (350 LOC)
- `window.emailAutocomplete()` - Email-Autocomplete für Sharing

**Build System**:
- Rollup bundelt `static/js/src/app.js` → `static/js/bundle.js`
- Dependencies: Alpine.js, HTMX, html5-qrcode
- Minification: Terser Plugin

---

## 🧪 Testing

**Status**: ✅ Vollständig implementiert (>70% Coverage erreicht)

**Coverage**:
- ✅ Service Tests: 68 Tests, **71.6% Coverage** (card, voucher, gift_card, merchant, favorite, dashboard, authz)
- ✅ Handler Tests: 122 Tests, **83.9% Average Coverage** (cards: 84.6%, vouchers: 85.6%, gift_cards: 81.6%)
- ✅ Model Tests: 38 Tests, **90.9% Coverage**
- ✅ Race Detection: Alle Tests bestehen mit `-race` Flag

**Testbarkeit**:
- ✅ Alle Services haben Interfaces → Mock-basiertes Testing
- ✅ Repositories haben Interfaces → Testbar ohne echte DB
- ✅ AuthzService vollständig getestet (7 Tests mit PostgreSQL)

**Details**: [TODO.md](TODO.md) - Task 1 (Completed)

---

## 🚀 Deployment

**Production Setup**: App läuft hinter **Traefik Reverse Proxy**

**Architektur**:
```
Client (HTTPS) → Traefik (TLS Termination) → App (HTTP:8080) → PostgreSQL
```

**Container**:
- Docker Multi-Stage Build ([Dockerfile](Dockerfile))
- Docker Compose für Development ([docker-compose.yml](docker-compose.yml))
- **Traefik** als Reverse Proxy (TLS, HTTPS-Redirect, Load Balancing)

**Traefik Features**:
- ✅ **TLS-Terminierung**: Let's Encrypt Zertifikate
- ✅ **HTTPS-Redirect**: HTTP → HTTPS Redirect auf Proxy-Ebene
- ✅ **Header-Injection**: `X-Forwarded-Proto`, `X-Real-IP`, `X-Forwarded-For`
- ✅ **Load Balancing**: Für Multi-Instance Deployments

**Environment Variables**:
```bash
# Server
SERVER_PORT=8080
GO_ENV=development

# Database
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable

# Session
SESSION_SECRET=change-me-in-production

# OAuth (optional)
OAUTH_CLIENT_ID=...
OAUTH_CLIENT_SECRET=...
OAUTH_ISSUER=https://auth.example.com/application/o/app/

# Feature Toggles
ENABLE_CARDS=true
ENABLE_VOUCHERS=true
ENABLE_GIFT_CARDS=true
ENABLE_LOCAL_LOGIN=true
ENABLE_REGISTRATION=true

# Observability
OTEL_ENABLED=false
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
```

**Details**: [OPERATIONS.md](OPERATIONS.md) - Backup & Recovery (Zeile 381-425)

---

## 🎯 Wichtige Konzepte für AI-Agenten

### 1. Favoriten-System (Pinning)

**Polymorphisches Design**:
```go
type UserFavorite struct {
    UserID       uuid.UUID
    ResourceType string    // "card", "voucher", "gift_card"
    ResourceID   uuid.UUID
    DeletedAt    *time.Time // Soft delete für Toggle
}
```

**Toggle-Logik** (Clean Architecture):
```go
// Handler nutzt FavoriteService (nicht database.DB!)
func (h *FavoritesHandler) toggleFavorite(userID uuid.UUID, resourceType string, resourceID uuid.UUID) bool {
    ctx := context.Background()

    // ToggleFavorite handled die komplette Logik (Create/Restore/Delete)
    if err := h.favoriteService.ToggleFavorite(ctx, userID, resourceType, resourceID); err != nil {
        return false
    }

    isFavorite, err := h.favoriteService.IsFavorite(ctx, userID, resourceType, resourceID)
    return isFavorite
}
```

**Dashboard-Integration**:
- Favoriten ersetzen "Kürzlich hinzugefügt" wenn vorhanden
- Mobile: Favoriten erscheinen VOR Statistiken
- Besitzer-Anzeige bei geteilten Items: "von [Name]"

**Files**:
- Model: [internal/models/user_favorite.go](internal/models/user_favorite.go)
- Handler: [internal/handlers/favorites.go](internal/handlers/favorites.go)
- Template: [internal/templates/home.templ](internal/templates/home.templ) + *Show-Templates
- Migration: [migrations/000005_add_user_favorites.up.sql](migrations/000005_add_user_favorites.up.sql)

### 2. Sharing-System

**Granulare Berechtigungen**:
- **Cards**: `can_edit`, `can_delete`
- **Vouchers**: IMMER read-only (keine Edit-Rechte)
- **Gift Cards**: `can_edit`, `can_delete`, `can_edit_transactions`

**Permission Check Pattern**:
```go
// 1. Prüfe Ownership
isOwner := resource.UserID != nil && *resource.UserID == user.ID

// 2. Falls nicht Owner, prüfe Share
if !isOwner {
    var share models.CardShare
    err := database.DB.Where("card_id = ? AND shared_with_id = ?",
                             resourceID, user.ID).First(&share).Error
    if err != nil {
        return http.StatusForbidden
    }
    // 3. Prüfe Permission
    canEdit = share.CanEdit
}
```

**✅ Authorization Service** (`internal/services/authz_service.go`, 154 LOC):
```go
// Zentrale Permission-Checks für alle Ressourcen
type AuthzServiceInterface interface {
    CheckCardAccess(ctx, userID, cardID) (*ResourcePermissions, error)
    CheckVoucherAccess(ctx, userID, voucherID) (*ResourcePermissions, error)
    CheckGiftCardAccess(ctx, userID, giftCardID) (*ResourcePermissions, error)
}

// ResourcePermissions enthält alle Access-Flags
type ResourcePermissions struct {
    CanView             bool
    CanEdit             bool
    CanDelete           bool
    CanEditTransactions bool // Gift Cards only
    IsOwner             bool
}
```

**Status**: ✅ Vollständig implementiert und in ALLEN 27 Handlern integriert (v1.4.0)
- Eliminiert duplicate Permission-Logic
- Konsistente Authorization-Checks über alle Ressourcen
- 7 Unit Tests mit PostgreSQL (Owner, SharedUser, Permissions)

### 3. Barcode-Scanning

**ZXing Integration**:
- Browser-basiert (ZXing JS)
- Kamera-Zugriff via MediaDevices API
- Unterstützte Formate: CODE128, QR, EAN13, EAN8

**Alpine.js Functions** in [layout.templ](internal/templates/layout.templ):
- `window.cardForm()` - Card Scanner Logic
- `window.voucherForm()` - Voucher Scanner Logic
- `window.giftCardForm()` - Gift Card Scanner Logic

**HTTPS Required**: Browser camera access requires HTTPS (except localhost)

### 4. Audit Logging

**Automatisch bei allen Deletions**:
```go
// Service-Layer handled Deletion mit Audit Logging
cardService.DeleteCard(ctx, cardID)  // → Service → Repository → GORM Hook → AuditLog Entry

// Alternativ: Direktes Audit Logging via AdminService
adminService.CreateAuditLog(ctx, &auditLog)
```

**Audit Log Schema**:
```sql
audit_logs:
  - user_id (wer hat gelöscht)
  - action ("delete", "hard_delete", "restore")
  - resource_type ("cards", "vouchers", etc.)
  - resource_id (UUID)
  - resource_data (JSONB Snapshot)
  - ip_address + user_agent
  - created_at
```

**Details**: [OPERATIONS.md](OPERATIONS.md) - Audit Logging (Zeile 18-143)

---

## 📝 Changelog

### Version 1.6.0 (2026-02-01) ✅ CURRENT
- ✅ **Clean Architecture Completion** - Alle 34 database.DB Aufrufe aus Handlers eliminiert
  - AdminService erstellt (226 LOC) - User Management, Audit Logs, Resource Restoration
  - ShareService erweitert - GetSharedUsers() für Shared Users Autocomplete
  - HealthHandler, SharedUsersHandler, AdminHandler vollständig refactored
  - 100% Clean Architecture: Handlers → Services → Repositories
  - Production-Ready Score: 8.9/10 → 9.1/10 (Wartbarkeit: 10/10)

### Version 1.5.0 (2026-02-01)
- ✅ **Production Secrets Validation** - Automatische Validierung verhindert Deployment mit Default-Secrets
  - ValidateProduction() prüft SESSION_SECRET (min. 32 Zeichen)
  - ValidateProduction() prüft OAUTH_CLIENT_SECRET (min. 16 Zeichen) wenn OAuth aktiv
  - 11 Tests (9 Unit Tests + 2 Integration Tests)

### Version 1.4.0 (2026-01-31)
- ✅ **AuthzService Integration** - Vollständig in ALLEN 27 Handlern integriert, eliminiert duplicate Permission-Logic
- ✅ **Handler Testing** - 122 Tests, 83.9% Average Coverage (Cards: 84.6%, Vouchers: 85.6%, Gift Cards: 81.6%)
- ✅ **Service Testing** - 68 Tests, 71.6% Coverage (Target >70% erreicht)
- ✅ **CSP Implementation** - Content Security Policy mit OAuth-Support

### Version 1.3.0 (2026-01-30)
- ✅ **Share Handler Abstraction** - Adapter pattern eliminates 70% code duplication
- ✅ **RESTful Compliance** - 5 update operations changed from POST to PATCH
- ✅ **Testing Infrastructure** - AuthzService tests with PostgreSQL

### Version 1.2.0 (2026-01-27)
- ✅ **PWA Implementation** - Service Worker, Manifest, Offline-Mode
- ✅ **JavaScript Extraction** - Modular Build System (Rollup + Terser)
- ✅ **AuthzService Creation** - Zentrale Authorization-Logic (154 LOC)

### Version 1.1.0 (2026-01-26)
- ✅ **Feature Toggles** - ENV-basierte Toggles für Cards, Vouchers, Gift Cards, Local Login, Registration
- ✅ **Observability** - Prometheus Metrics, Health Checks, Structured Logging
- ✅ **Performance** - Dashboard 40% faster, Gift Card Balance 78% faster
- ✅ **Mobile Optimization** - Responsive Design
- ✅ **OAuth/OIDC** - Provider-agnostische Auth

### Version 1.0.0 (2026-01-25)
- ✅ **Clean Architecture** - Service Layer + Repository Pattern
- ✅ **Favoriten-System** - Polymorphic Pinning
- ✅ **Internationalization** - German, English, French
- ✅ **Audit Logging** - Deletion Tracking
- ✅ **Sharing** - Granulare Permissions

**Voller Changelog**: [README.md](README.md) - Changelog

---

## 🎯 Offene Aufgaben

**Production Readiness Score**: **8.9/10** ✅ Production-Ready

**CRITICAL (vor Production)**:

- ⚠️ **Production Deployment**: Reverse Proxy Setup, TLS, Database Backups, Monitoring, Log Aggregation

**MEDIUM Priority**:

- ⚠️ **Security Hardening**: Additional HTTP Headers (XSS-Protection, X-Frame-Options, HSTS)
- ⚠️ **CI/CD Pipeline**: GitHub Actions für Tests + Deployment
- ⚠️ **Kubernetes Manifests**: Production Deployment Setup (Deployment, Ingress, ConfigMap)

**LOW Priority**:

- ⚠️ **Handler Refactoring**: Entfernung direkter database.DB Zugriffe (AuthzService nutzt noch GORM direkt)
- ⚠️ **Main.go Refactoring**: Setup-Logik in separate Packages auslagern

**COMPLETED** ✅:

- ✅ **Testing**: >70% Coverage erreicht (Service: 71.6%, Handler: 83.9%, Model: 90.9%)
- ✅ **AuthzService**: Vollständig in ALLEN 27 Handlern integriert (v1.4.0)
- ✅ **Migration Strategy**: Gormigrate implementiert mit AUTO_MIGRATE Flag
- ✅ **HTTPS Enforcement**: Via Traefik Reverse Proxy (TLS-Terminierung, HTTP→HTTPS Redirect)
- ✅ **Secrets Validation**: Production-Checks für SESSION_SECRET und OAUTH_CLIENT_SECRET (v1.5.0)
- ✅ **CSP**: Content Security Policy mit OAuth-Support
- ✅ **JavaScript Extraction**: Modular Build System (Rollup + Terser)
- ✅ **PWA Implementation**: Service Worker, Manifest, Offline-Mode
- ✅ **SameSite Cookies**: SameSite=Lax (OAuth-kompatibel, CSRF-Protection)

**Details**: [TODO.md](TODO.md) - Vollständige Roadmap

---

## 🛠️ Troubleshooting

### "Templ generation failed"

```bash
go install github.com/a-h/templ/cmd/templ@latest
templ generate
```

### "Database connection refused"

```bash
docker compose ps
docker compose logs postgres
```

### "Barcode scanner not working"

- HTTPS Required (Browser camera access)
- ZXing Script muss in layout.templ geladen sein
- User muss Camera Permission gewähren

---

## 📚 Weitere Ressourcen

- **Clean Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **Database Schema**: [migrations/README.md](migrations/README.md)
- **Deployment**: [OPERATIONS.md](OPERATIONS.md)
- **Roadmap**: [TODO.md](TODO.md)
- **User Guide**: [README.md](README.md)

---

**Ende der AI Agent Dokumentation**

Dieses Projekt folgt Clean Architecture mit Go + Echo + Templ + HTMX + Alpine.js.
Alle technischen Details sind in spezialisierten Dokumenten organisiert - vermeide Redundanzen!
