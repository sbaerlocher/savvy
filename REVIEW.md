# Code Review Guidelines

## Scope

In scope:

- Go backend under `internal/` and `cmd/` (handlers, services,
  repositories, middleware, migrations, models)
- SvelteKit frontend under `client/src/` (routes, components, stores,
  API clients, i18n)
- Database migrations under `internal/migrations/`
- Helm chart and Kustomize overlays under `deploy/`
- CI/CD workflow changes under `.github/workflows/`
- Renovate configuration updates

Out of scope:

- Auto-generated lock files (`go.sum`, `client/package-lock.json`)
- Embedded frontend build output under `internal/assets/client/`
- Renovate dependency-only PRs (patch/minor with automerge enabled)

## Required checks

- No secrets committed — secrets via External Secrets Operator +
  Bitwarden (K8s), Repository Secrets (CI); gitleaks gates every PR
- Go: `go test ./...` passes (incl. `-race`), `golangci-lint` clean
- Frontend: TypeScript type check passes (`npm run typecheck`),
  Prettier-formatted, build succeeds (`npm run build`)
- E2E suite green (`dde project:e2e:test`)
- Migrations are forward-only and idempotent — reviewed for blast
  radius before merge; one file per migration, registry order unchanged
- Layered architecture respected — Handlers never touch the DB directly,
  Services never touch the Echo context (interfaces only)
- Authorization goes through `AuthzService` — no duplicated permission
  logic in handlers
- Sharing permissions correct per resource (vouchers always read-only)

## Severity levels

| Level        | Meaning                                             | Merge impact       |
| ------------ | --------------------------------------------------- | ------------------ |
| Bug          | Incorrect behavior or broken contract               | Blocks merge       |
| Nit          | Minor issue — suboptimal but not incorrect          | Non-blocking       |
| Pre-existing | Issue present before this PR; flagged for awareness | No action required |

## Skip

- Renovate PRs with `automerge: true` (patch/minor) after CI passes
- Documentation-only changes with no functional impact
- Formatting-only diffs already enforced by Prettier / gofmt
