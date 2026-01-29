# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/sbaerlocher/savvy/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/sbaerlocher/savvy/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/sbaerlocher/savvy/releases/tag/v1.0.0
