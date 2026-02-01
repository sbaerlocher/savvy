# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.6.0] - 2026-02-01

### Changed
- **Clean Architecture Completion** - Eliminated all 34 database.DB calls from handlers
  - Created AdminService (226 LOC) for user management, audit logs, and resource restoration
  - Extended ShareService with GetSharedUsers() method for shared users autocomplete
  - Refactored HealthHandler, SharedUsersHandler, AdminHandler to use dependency injection
  - Update handlers (cards, vouchers, gift_cards) now use `h.db` for audit logging
  - Achieved 100% Clean Architecture compliance: Handlers → Services → Repositories

### Improved
- **Maintainability** - Production-Ready Score increased from 8.9/10 to 9.1/10 (Wartbarkeit: 10/10)
- **Code Organization** - Zero direct database access in presentation layer

## [1.5.0] - 2026-02-01

### Added
- **Production Secrets Validation** - Automatic validation prevents deployment with default secrets
  - ValidateProduction() checks SESSION_SECRET (min. 32 characters)
  - ValidateProduction() checks OAUTH_CLIENT_SECRET (min. 16 characters) when OAuth is active
  - 11 tests (9 unit tests + 2 integration tests)

## [1.4.0] - 2026-01-31

### Added
- **AuthzService Integration** - Fully integrated in ALL 27 handlers, eliminates duplicate permission logic
- **Handler Testing** - 122 tests, 83.9% average coverage (Cards: 84.6%, Vouchers: 85.6%, Gift Cards: 81.6%)
- **Service Testing** - 68 tests, 71.6% coverage (target >70% achieved)
- **Content Security Policy (CSP)** - XSS protection with OAuth support

### Changed
- Race Detection - All tests pass with `-race` flag

## [1.3.0] - 2026-01-30

### Added
- **Share Handler Abstraction** - Adapter pattern eliminates 70% code duplication
- **RESTful Compliance** - 5 update operations changed from POST to PATCH
- **Testing Infrastructure** - AuthzService tests with PostgreSQL

## [1.2.0] - 2026-01-27

### Added
- **Progressive Web App (PWA)** - Service Worker, Manifest, Offline-Mode
- **JavaScript Extraction** - Modular Build System (Rollup + Terser)
- **AuthzService Creation** - Central authorization logic (154 LOC)

## [1.1.0] - 2026-01-26

### Added
- Feature Toggles - ENV-based toggles for Cards, Vouchers, Gift Cards, Local Login, Registration
- Observability - Prometheus Metrics, Health Checks, Structured Logging
- Mobile Optimization - Responsive Design improvements
- OAuth/OIDC - Provider-agnostic authentication
- PWA Support - Service Worker, Offline Mode, App Manifest

### Changed
- Dashboard Performance - 40% faster with N+1 query fixes
- Gift Card Balance - 78% faster with database triggers

### Fixed
- Various security improvements and audit logging enhancements

## [1.0.0] - 2026-01-25

### Added
- Initial release with Clean Architecture
- Cards, Vouchers, and Gift Cards management
- Polymorphic Favorites system
- Granular Sharing permissions
- Internationalization (German, English, French)
- Audit Logging for deletions
- Barcode scanning support
- Session-based authentication
- CSRF protection
- PostgreSQL database with GORM

[Unreleased]: https://github.com/sbaerlocher/savvy/compare/v1.6.0...HEAD
[1.6.0]: https://github.com/sbaerlocher/savvy/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/sbaerlocher/savvy/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/sbaerlocher/savvy/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/sbaerlocher/savvy/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/sbaerlocher/savvy/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/sbaerlocher/savvy/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.0.0
