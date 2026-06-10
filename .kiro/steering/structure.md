# Project Structure

## Organization Philosophy

**Backend**: layered, interface-first. Each layer is its own package, and
dependencies flow strictly downward. Business logic does not know about
HTTP, and HTTP does not know about the database.

```
handlers/  →  services/  →  repository/  →  models/  →  GORM/Postgres
                  ↑                                ↑
            authz, mappers                   migrations
```

**Frontend**: SvelteKit file-based routing for pages, plus a flat
`lib/` split by responsibility (api / components / stores / utils / i18n /
types). Routes call into `$lib/api`, never `fetch()` directly.

The frontend lives under `client/` and is built into `internal/assets/client/`
so the Go binary can `//go:embed` it.

## Directory Patterns

### `cmd/`
**Purpose**: thin entry points only — wiring, flags, signal handling.
Notable subdirs: `cmd/server` (main HTTP server), `cmd/migrate`,
`cmd/seed`, `cmd/e2e`, `cmd/release-tool`.
**Rule**: no business logic in `cmd/`; delegate to `internal/setup` and
services.

### `internal/setup/`
**Purpose**: composition root. `dependencies.go` builds the DI container
(DB, telemetry, services, repos), `routes.go` registers all routes,
`server.go` configures Echo. New top-level wiring goes here, not in `main.go`.

### `internal/handlers/`
**Purpose**: HTTP layer. JSON endpoints live in `handlers/api/` (one file
per resource family). Cross-resource share endpoints use the adapter
pattern in `handlers/shares/` (`ShareAdapter` interface +
`base_handler.go`) — add a new shareable resource by writing an adapter,
not by duplicating handler code.
**Rule**: handlers validate input, call a service, map results to DTOs.
They do not import GORM or build queries.

### `internal/services/`
**Purpose**: business logic, one service per domain (cards, vouchers,
gift_cards, share, transfer, totp, push, reminder, account, …).
**Rules**:
- Every service exposes a `*ServiceInterface`; the implementation is in the
  same file or a `_impl` sibling.
- Authorization decisions go through `AuthzService.CheckXAccess()`, never
  ad-hoc owner checks in handlers.
- Services receive `context.Context` and repository interfaces; they never
  touch `echo.Context`.

### `internal/repository/`
**Purpose**: data access. Convention is **interface + `_impl.go` pair**
(e.g. `card_repository.go` + `card_repository_impl.go`). Pagination and
shared helpers live in `base_repository.go` and `pagination.go`.
**Rule**: only this layer writes GORM queries.

### `internal/models/`
**Purpose**: GORM structs and DB-backed types. UUID primary keys.
Shared/share/transaction tables follow `<resource>_share` and
`<resource>_transaction` naming.

### `internal/middleware/`
**Purpose**: Echo middleware (auth, CSRF, CORS, rate limiting, i18n,
session tracking, OTel). Session keys are centralized in
`session_keys.go` — never use raw string keys against the session store.

### `internal/migrations/`
**Purpose**: embedded Gormigrate migrations. Add new schema changes here,
not as separate SQL files. The binary auto-applies them when
`AUTO_MIGRATE=true`.

### `internal/email/` and `internal/email/templates/`
**Purpose**: SMTP service plus `//go:embed`'d HTML templates. New
notification types add a template here and call into `EmailService`.

### `internal/{audit, telemetry, metrics, validation, i18n, oauth, mocks, testutil}/`
**Purpose**: focused cross-cutting packages. Each owns one concern; do not
mix them with services or handlers.

### `client/src/routes/`
**Purpose**: SvelteKit pages and layouts. Subdirectories mirror the URL
(e.g. `cards/`, `gift-cards/`, `admin/users/`, `login/2fa/`). Settings are
split into dedicated pages (`/profile`, `/security`, `/notifications`)
rather than a single monolith.

### `client/src/lib/`
**Purpose**: shared frontend code, organized by responsibility:
- `api/` — typed API clients (`cardsApi`, `vouchersApi`, …). All HTTP
  calls go through here.
- `components/` — reusable Svelte components, including `settings/`
  sub-components and the offline indicator.
- `stores/` — Svelte stores (`auth`, `offline`, `notifications`, …) and
  the IndexedDB wrapper (`offline-db.ts`).
- `i18n/` — TS translations for DE/EN/FR.
- `utils/`, `types/` — pure helpers and shared TS types.

### `client/src/service-worker.ts`
**Purpose**: custom Workbox service worker (`injectManifest` strategy).
Cache strategy and warmup live here. Bump cache version when changing
strategy.

### `client/tests/e2e/`
**Purpose**: Playwright specs (one file per feature). `global.setup.ts`
brings up Docker; `global.teardown.ts` tears it down. Tests are written
against the running app, not mocked.

### `deploy/`
**Purpose**: Helm chart, Kustomize overlays, Grafana dashboards. Changes
to runtime config go here, not to ad-hoc YAML.

## Naming Conventions

- **Go files**: `snake_case.go`. Interface and implementation in matching
  pairs: `foo_repository.go` + `foo_repository_impl.go`. Tests as
  `*_test.go` next to the code under test.
- **Go types**: `PascalCase` for exported, `camelCase` for unexported.
  Service interfaces end in `ServiceInterface`, repositories in
  `RepositoryInterface` or `Repository`.
- **Svelte components**: `PascalCase.svelte` (e.g.
  `BarcodeScanner.svelte`, `OfflineIndicator.svelte`).
- **TS modules**: `kebab-case.ts` for utilities (`merchant-aggregator.ts`),
  `camelCase` exports.
- **SvelteKit routes**: lowercase URL segments, `+page.svelte` /
  `+layout.svelte` per Kit convention.
- **Env vars**: `SCREAMING_SNAKE_CASE`. Feature flags use the `ENABLE_`
  prefix.

## Import Organization

**Go**: module path is `savvy`. Always import internal packages from the
module root:

```go
import (
    "savvy/internal/services"
    "savvy/internal/models"
)
```

**TypeScript / Svelte**: use the configured aliases from
`client/svelte.config.js`; don't reach into `src/lib` with relative paths.

```ts
import { cardsApi } from '$lib/api/cards';
import OfflineIndicator from '$components/OfflineIndicator.svelte';
import { offlineStore } from '$lib/stores/offline';
```

**Path Aliases**:
- `$lib` → `client/src/lib`
- `$components` → `client/src/lib/components`

## Code Organization Principles

- **Strict layer dependencies**: handlers → services → repositories →
  models. Skipping a layer (e.g. a handler reading from `database.DB`) is
  a regression — route it through a service.
- **Interfaces at every seam**: every service and every repository is
  defined by an interface so it can be mocked. Add the interface first,
  the implementation second.
- **Authorization is centralized**: use `AuthzService.CheckXAccess()` for
  every protected resource read/write. Don't reinvent the
  owner-or-share permission check inline.
- **Adapters for cross-resource features**: sharing, batch operations and
  similar features that span cards / vouchers / gift cards should add an
  adapter implementing a small interface, not a new parallel handler tree.
- **Frontend mutations only via `$lib/api`**: routes and components must
  not call `fetch` directly; all API calls go through typed clients so
  offline gating, error handling, and types stay consistent.
- **Offline-first contract**: any new resource type that should work
  offline must (1) cache via the service worker's NetworkFirst rule,
  (2) be persisted in IndexedDB via `offline-db.ts`, and (3) participate
  in the cache-validation flow on the `online` event.
- **i18n is mandatory**: any new user-visible string ships with DE/EN/FR.
  Backend strings go through the `i18n` package; frontend strings through
  `$lib/i18n`.

---
_Document patterns, not file trees. New files following patterns shouldn't require updates._
