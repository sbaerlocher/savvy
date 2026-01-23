# 🎁 Savvy System

> Digitale Verwaltung von Kundenkarten, Gutscheinen und Geschenkkarten mit Sharing-Funktionalität

Ein modernes Web-basiertes System zur Verwaltung von Treuekarten, Rabatt-Gutscheinen und Prepaid-Geschenkkarten. Mit integriertem Barcode-Scanner, Transaktionsverlauf und flexiblem Sharing mit anderen Benutzern.

## ✨ Features

### 🎴 Kundenkarten (Savvy Cards)

- Digitale Speicherung von Treuekarten und Membership-Cards
- Barcode-Support (CODE128, QR, EAN13, EAN8)
- Barcode-Scanning via Smartphone/Webcam (ZXing)
- Status-Tracking (Aktiv, Inaktiv)
- Händler-Verwaltung mit Farben und Logos
- Teilen mit anderen Benutzern (mit Bearbeitungsrechten)

### 🎟️ Gutscheine (Vouchers)

- Rabatt-Gutscheine (Prozent, Festbetrag, Punkte-Multiplikator)
- Verschiedene Nutzungsmodelle:
  - Single-Use (einmalig)
  - One-per-Customer (einmal pro Kunde)
  - Multiple-Use (mehrfach mit/ohne Card-Tracking)
  - Unlimited (unbegrenzt)
- Gültigkeitszeitraum und Mindestbestellwert
- Barcode-Scanning für schnelle Erfassung
- Teilen (immer read-only für geteilte Benutzer)

### 💳 Geschenkkarten (Gift Cards)

- Prepaid-Guthaben mit automatischer Berechnung
- Transaktionsverlauf (Ausgaben und Aufladungen)
- PIN-Schutz optional
- Barcode-Scanning für Kartennummern
- Ablaufdatum-Verwaltung
- Teilen mit granularen Berechtigungen:
  - Bearbeiten (Details ändern)
  - Löschen (Karte entfernen)
  - Transaktionen verwalten (Ausgaben erfassen)

### 👥 Sharing & Permissions

- Alle drei Ressourcentypen können geteilt werden
- Flexible Berechtigungen pro Share
- Edit/Delete/View Permissions für Cards
- Transaction Management für Gift Cards
- Übersicht über geteilte Items im Dashboard
- User-spezifische Favoriten (geteilte Items können individuell favorisiert werden)
- Besitzer-Anzeige bei geteilten Items ("von [Name]")

### 📊 Dashboard

- Statistiken (Anzahl Cards/Vouchers/Gift Cards)
- Gesamtguthaben aller Gift Cards
- ⭐ Favoriten-System (Pinning) - Schnellzugriff zu häufig genutzten Items
- Zuletzt hinzugefügte Items (wenn keine Favoriten vorhanden)
- Schnellzugriff zum Erstellen neuer Items
- Mobile-optimierte Ansicht (Favoriten vor Statistiken)

### 🔍 Suchen & Filtern

- Volltextsuche nach Händler/Code
- Filtern nach Besitzer (Meine / Alle)
- Filtern nach Status (Aktiv / Abgelaufen)
- Sortieren nach Händler oder Datum
- Client-seitige Filterung (Alpine.js) für schnelle Ergebnisse

### 📱 Progressive Web App (PWA)

- ✅ **Installierbar**: Als App auf iOS/Android/Desktop installierbar
- ✅ **Offline-Modus**: Gecachte Daten offline verfügbar
- ✅ **Service Worker**: Network-First Strategie mit Cache-Fallback
- ✅ **Offline-Erkennung**: Visuelles Feedback bei Netzwerkproblemen
- ✅ **Automatische Updates**: Service Worker Updates transparent im Hintergrund

**Offline-Funktionen**:

- ✅ Karten/Gutscheine/Geschenkkarten ansehen (gecached)
- ✅ Barcode-Details anzeigen
- ✅ Filter & Sortierung (client-side)
- ✅ Favoriten durchsuchen
- ❌ Neue Items erstellen/bearbeiten (nur online)

Siehe [docs/PWA.md](docs/PWA.md) für Details.

## 🚀 Quick Start

### Voraussetzungen

- Docker & Docker Compose
- Go 1.23+ (für lokale Entwicklung)
- Node.js 18+ & npm (für Frontend-Build)
- Make (optional, für Makefile-Commands)

### Installation & Start

```bash
# 1. Repository klonen
git clone <repository-url>
cd savvy

# 2. Environment-Variablen konfigurieren
cp .env.example .env
# Edit .env with your settings

# 3. Frontend-Bundles bauen
npm install
npm run build

# 4. Docker Container starten
docker compose up -d

# 5. Test-Daten laden (optional)
make seed-docker
```

**Anwendung öffnen**: <http://localhost:8080>

### Test-Benutzer

Nach dem Seeding stehen folgende Test-Accounts zur Verfügung (Passwort: `test123`):

| Email                        | Rolle | Beschreibung                    |
| ---------------------------- | ----- | ------------------------------- |
| `admin@example.com`          | Admin | Hat Admin-Rechte + eigene Items |
| `anna.mueller@example.com`   | User  | Hat Zugriff auf geteilte Items  |
| `thomas.schmidt@example.com` | User  | Hat Zugriff auf geteilte Items  |
| `maria.garcia@example.com`   | User  | Hat eigene Items                |

## 💻 Entwicklung

### Lokale Entwicklungsumgebung

```bash
# 1. Dependencies installieren
go mod download
npm install

# 2. Templ CLI installieren
go install github.com/a-h/templ/cmd/templ@latest

# 3. Air installieren (Hot Reload)
go install github.com/air-verse/air@latest

# 4. Datenbank starten
docker compose up -d postgres

# 5. Frontend Bundles bauen (initial)
npm run build

# 6. Development Server mit Hot Reload
# Air triggert automatisch npm run build bei JS/CSS-Änderungen
air
```

**Hinweis**: Air überwacht automatisch:
- `internal/templates/**/*.templ` → Templ Generierung
- `static/js/src/**/*.js` → JS Bundle Rebuild
- `static/css/src/**/*.css` → CSS Bundle Rebuild
- `**/*.go` → Go Binary Rebuild

Siehe [BUILD.md](BUILD.md) für Details zum Build-System.

### Makefile Commands

```bash
# Docker Compose
make up          # Start all services
make down        # Stop all services
make logs        # View logs
make restart     # Restart services

# Development
make dev         # Start with hot reload (Air)
make seed        # Seed test data (local)
make seed-docker # Seed test data (Docker)
make test        # Run tests

# Database
make db-shell    # PostgreSQL shell
make db-reset    # Reset database (⚠️ deletes all data)

# Application
make shell       # Application shell
make build       # Build binary
```

### Code-Änderungen

**Templates (`.templ` Dateien)**:

```bash
# Nach Änderungen an .templ files
templ generate

# Air reloaded automatisch
```

**Models**:

```bash
# GORM AutoMigrate läuft beim Server-Start
# Oder manuelle Migration in migrations/ erstellen
```

**Handlers**:

```bash
# Air reloaded automatisch bei Änderungen
```

## 📁 Projekt-Struktur

```
savvy/
├── cmd/
│   ├── server/           # Application entrypoint
│   └── seed/             # Database seeding script
│
├── internal/
│   ├── config/           # Configuration (env vars)
│   ├── database/         # GORM connection
│   ├── handlers/         # HTTP handlers (Controllers)
│   │   ├── home.go       # Dashboard
│   │   ├── auth.go       # Login/Logout/Register
│   │   ├── admin.go      # Admin Panel
│   │   ├── favorites.go  # Favorites Toggle (Pinning)
│   │   ├── cards.go      # Cards CRUD
│   │   ├── card_shares.go # Card Sharing Management
│   │   ├── vouchers.go   # Vouchers CRUD
│   │   ├── voucher_shares.go # Voucher Sharing
│   │   ├── gift_cards.go # Gift Cards CRUD + Transactions
│   │   └── gift_card_shares.go # Gift Card Sharing
│   ├── middleware/       # Auth & session middleware
│   ├── models/           # GORM models
│   │   ├── user.go       # User model
│   │   ├── merchant.go   # Merchant model
│   │   ├── user_favorite.go # User Favorites (polymorphic)
│   │   ├── card.go       # Card + CardShare models
│   │   ├── voucher.go    # Voucher + VoucherShare models
│   │   ├── gift_card.go  # GiftCard + Shares
│   │   └── gift_card_transaction.go # Transaction History
│   └── templates/        # Templ templates
│       ├── layout.templ  # Base layout + Nav + Alpine.js
│       ├── home.templ    # Dashboard + Favorites
│       ├── auth.templ    # Login/Register
│       ├── admin.templ   # Admin Panel
│       ├── cards.templ   # Cards UI + Favorite Button
│       ├── card_shares.templ # Card Sharing UI
│       ├── vouchers.templ # Vouchers UI + Favorite Button
│       ├── voucher_shares.templ # Voucher Sharing UI
│       ├── gift_cards.templ # Gift Cards UI + Favorite Button
│       └── gift_card_shares.templ # Gift Card Sharing UI
│
├── migrations/           # Database migrations
│   ├── README.md        # Schema documentation
│   ├── 000001_init_schema.up.sql
│   ├── 000002_add_gift_card_share_permissions.up.sql
│   └── 000005_add_user_favorites.up.sql
│
├── static/
│   ├── css/             # TailwindCSS
│   └── js/              # HTMX
│
├── .air.toml            # Hot reload config
├── docker-compose.yml   # Docker services
├── Dockerfile           # Multi-stage build
├── Makefile             # Development commands
├── go.mod / go.sum      # Go dependencies
└── README.md            # This file
```

## 🛠️ Tech Stack

| Komponente            | Technologie                  | Zweck                        |
| --------------------- | ---------------------------- | ---------------------------- |
| **Backend Framework** | Echo v4                      | HTTP Router & Middleware     |
| **ORM**               | GORM v2                      | PostgreSQL Abstraction       |
| **Templates**         | Templ                        | Type-safe Go HTML Templates  |
| **Frontend**          | HTMX + Alpine.js             | Dynamic UI ohne Page Reload  |
| **Styling**           | TailwindCSS                  | Utility-First CSS            |
| **Barcode**           | ZXing JS + boombuler/barcode | Scanning & Generation        |
| **Auth**              | Gorilla Sessions             | Session-based Authentication |
| **Database**          | PostgreSQL 16                | Primary Data Store           |
| **Hot Reload**        | Air                          | Development Auto-Reload      |

## 🗄️ Datenbank-Schema

### Tabellen-Übersicht

```
users (1) ──┬─< cards (N)
            ├─< vouchers (N)
            ├─< gift_cards (N)
            ├─< card_shares (N)
            ├─< voucher_shares (N)
            ├─< gift_card_shares (N)
            └─< user_favorites (N)  [NEW - Polymorphic]

merchants (1) ──┬─< cards (N)
                ├─< vouchers (N)
                └─< gift_cards (N)

gift_cards (1) ─< gift_card_transactions (N)

cards (1) ─< card_shares (N)
vouchers (1) ─< voucher_shares (N)
gift_cards (1) ─< gift_card_shares (N)
```

### Haupttabellen

1. **users** - Benutzer-Accounts mit Authentication
2. **merchants** - Händler/Marken mit Farben und Logos
3. **user_favorites** - User-spezifische Favoriten (polymorphic: Cards, Vouchers, Gift Cards)
4. **cards** - Kundenkarten mit Barcode
5. **card_shares** - Sharing von Cards (mit can_edit, can_delete)
6. **vouchers** - Gutscheine mit Nutzungslimits
7. **voucher_shares** - Sharing von Vouchers (read-only)
8. **gift_cards** - Geschenkkarten mit Guthaben
9. **gift_card_transactions** - Transaktionsverlauf
10. **gift_card_shares** - Sharing von Gift Cards (mit can_edit, can_delete, can_edit_transactions)

Details siehe: [migrations/README.md](migrations/README.md)

## 🔐 Sicherheit

- ✅ Bcrypt Password Hashing
- ✅ Session-based Authentication
- ✅ CSRF Protection (Echo Middleware)
- ✅ SQL Injection Protection (GORM)
- ✅ XSS Protection (Templ Auto-Escaping)
- ✅ UUID statt Integer IDs
- ✅ Granulare Berechtigungen für Sharing

## 🚀 Deployment

### Docker Production Build

```bash
# Build image
docker build -t savvy:latest .

# Run with environment variables
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_USER=savvy_user \
  -e DB_PASSWORD=secure_password \
  -e DB_NAME=savvy_db \
  -e SESSION_SECRET=your-secret-key \
  savvy:latest
```

### Kubernetes (K3s)

Siehe [AGENTS.md](AGENTS.md) für Kubernetes Deployment-Beispiele.

### Environment Variables

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=savvy_user
DB_PASSWORD=change-me
DB_NAME=savvy_db
DB_SSLMODE=disable

# Session
SESSION_SECRET=change-me-in-production
SESSION_NAME=savvy_session

# Admin (for initial setup)
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=change-me
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/models -run TestCard_GetColor
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Go: `gofmt` + `golangci-lint`
- Templ: `templ fmt`
- Commit Messages: Conventional Commits

## 📝 Changelog

### Version 1.2.0 (2026-01-27) ✅ CURRENT

**New Features**

- ✅ **Progressive Web App (PWA)** - Vollständige PWA-Implementierung
  - Service Worker mit Network-First Strategie
  - Installierbar auf iOS/Android/Desktop
  - Offline-Modus für gecachte Daten
  - Offline-Erkennung mit visuellem Feedback
  - Automatische Background-Updates
- ✅ **Authorization Service** - Zentrale Authorization-Logic (154 LOC)
  - Interface-basiertes Design für Testbarkeit
  - Resource-spezifische Permission-Checks
  - Ownership + Share-based Access Control
  - Im Container registriert und einsatzbereit

**Improvements**

- ✅ **JavaScript Extraction** - Modular Build System
  - Rollup-basierte Build Pipeline
  - Separate Module: scanner.js (350 LOC), offline.js, precache.js
  - Terser Minification (~150KB Bundle)
  - Hot Reload via `npm run watch`
- ✅ **Build Pipeline** - PostCSS + TailwindCSS + Rollup
- ✅ **Documentation Update** - AGENTS.md, ARCHITECTURE.md, TODO.md aktualisiert

### Version 1.1.0 (2026-01-26)

**New Features**

- ✅ **Favoriten-System (Pinning)** - User-spezifische Favoriten für schnellen Zugriff
- ✅ **OAuth/OIDC Authentication** - Provider-agnostische OAuth-Integration
- ✅ **Feature Toggles** - ENV-basierte Toggles für 5 Features
- ✅ **Observability** - Prometheus Metrics, Health Checks, Structured Logging
- ✅ **Mobile Optimization** - Responsive Design für alle Seiten

**Improvements**

- ✅ **Performance**: Dashboard 40% faster, Gift Card Balance 78% faster
- ✅ CSRF-Token-Handling für HTMX-Requests
- ✅ Soft-Delete-Handling für Favoriten (Toggle-Logik)

### Version 1.0.0 (2026-01-25)

**Initial Release**

- ✅ Cards Management (CRUD + Sharing)
- ✅ Vouchers Management (CRUD + Sharing)
- ✅ Gift Cards Management (CRUD + Transactions + Sharing)
- ✅ Barcode Scanning (ZXing) für alle drei Typen
- ✅ Dashboard mit Statistiken
- ✅ Admin Panel mit User Management
- ✅ Filter & Search auf allen Index-Seiten
- ✅ Merchant-System mit Farben
- ✅ Docker Compose Setup
- ✅ GORM AutoMigrate
- ✅ Seed Data Script

### Geplante Features

- 🔄 QR-Code Export
- 🔄 CSV Import/Export
- 🔄 PWA Support (Offline-Fähigkeit)
- 🔄 Push Notifications (Gift Card Balance)
- 🔄 Authentik OAuth Integration
- 🔄 API for Mobile Apps

## 📚 Dokumentation

- **AGENTS.md** - Technische Dokumentation für AI-Agenten und Entwickler
- **migrations/README.md** - Datenbank-Schema Dokumentation
- **CLAUDE.md** - Claude Code Integration

## 📧 Support

Bei Fragen oder Problemen:

- GitHub Issues: [Create an issue](../../issues)
- Dokumentation: Siehe [AGENTS.md](AGENTS.md)

## 📄 License

MIT License - siehe [LICENSE](LICENSE) file für Details.

---

**Entwickelt mit** Go + Echo + Templ + HTMX + Alpine.js

**Deployed auf** Kubernetes (K3s) + PostgreSQL
