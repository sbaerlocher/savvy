# Product Overview

Savvy is a self-hosted personal management application for digital customer
loyalty cards, vouchers, and gift cards. It targets individuals (and small
households via sharing) who want one trusted place to keep, scan, and share
the various paper and plastic cards that accumulate over time.

## Core Capabilities

- **Three resource types** with a common merchant axis: cards (loyalty),
  vouchers (read-only redemption codes), gift cards (with balance and
  transaction history).
- **Granular sharing & ownership transfer** between users, with
  per-resource-type permission flags (e.g. `can_edit_transactions` only on
  gift cards, vouchers always read-only).
- **Barcode-first UX**: server-rendered barcodes for display, in-browser
  scanning via `BarcodeDetector` polyfill.
- **Offline-first PWA** so cards remain accessible without network; mutations
  are gated until back online.
- **Multi-channel reminders** for expiring vouchers/gift cards (in-app, web
  push, email) with per-user opt-in.

## Target Use Cases

- Replacing physical loyalty card wallets and email/PDF voucher folders.
- Sharing a household pool of gift cards while preserving ownership and audit
  trail.
- Getting reminded about expiring value before it is lost.
- Self-hosting (Docker / Kubernetes / Helm) for users who do not want to hand
  loyalty data to a third-party SaaS.

## Value Proposition

- **Self-hosted and privacy-respecting**: single Go binary with embedded
  SvelteKit assets, runs behind any reverse proxy.
- **Complete lifecycle, not just storage**: ownership transfer, sharing
  permissions, expiry reminders, audit logging, GDPR-compliant deletion.
- **Polished offline support**: cache validation on reconnect removes
  resources whose access was revoked or that were deleted while offline.
- **Operationally serious for a personal app**: OpenTelemetry traces,
  Prometheus metrics, two-tier readiness probes, an admin system-health
  dashboard.

---
_Focus on patterns and purpose, not exhaustive feature lists._
