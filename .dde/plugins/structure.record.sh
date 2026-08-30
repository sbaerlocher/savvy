#!/usr/bin/env bash
# @command structure:record
# @description Record the structural screenshot + axe baseline (needs approval)
set -euo pipefail

plugin_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$plugin_dir/../.." && pwd)"
cd "$project_root/client"

if [[ ! -d node_modules/@playwright/test ]]; then
	echo "structure:record: host node_modules missing @playwright/test, running npm ci" >&2
	npm ci --loglevel=error
	npx playwright install chromium
fi

if [[ ! -f .svelte-kit/tsconfig.json ]]; then
	npx svelte-kit sync
fi

# Same branch-slug derivation as structure:test — see the comment there.
if [[ -z "${BASE_URL:-}" ]]; then
	branch="$(git -C "$project_root" rev-parse --abbrev-ref HEAD)"
	slug="${branch//\//-}"
	if [[ "$branch" == "main" ]]; then
		BASE_URL="https://savvy.test"
	else
		BASE_URL="https://${slug}.savvy.test"
	fi
fi
export BASE_URL
echo "structure: target ${BASE_URL}" >&2

# Recording overwrites the reference the verification run compares against.
# Doing that accidentally destroys exactly the regression the suite exists to
# catch, so this is gated behind an explicit confirmation.
if [[ "${STRUCTURE_RECORD_CONFIRM:-}" != "yes" ]]; then
	cat >&2 <<'MSG'
structure:record overwrites the visual + axe baseline.

Only run this to create the FIRST baseline, or after each changed image has
been reviewed and approved individually. Never run it to "make the suite pass".

To proceed:
  STRUCTURE_RECORD_CONFIRM=yes dde project:structure:record
MSG
	exit 1
fi

# Both recording and verification must run against the SAME data, or the diff
# reports "the seed changed", not "the layout changed". Test runs themselves
# mutate state (each login adds a session, notifications accumulate), so the
# database is reset and re-seeded first unless the caller opts out.
if [[ "${STRUCTURE_SKIP_RESET:-}" != "1" ]]; then
	(cd "$project_root" && dde project:structure:reset-db)
fi

# Health is probed here with --resolve (the dde traefik always sits on
# 127.0.0.1), because macOS's mDNSResponder occasionally stalls on the
# /etc/resolver/test lookup while the stack itself is healthy. The generic
# globalSetup curl would inherit that stall, so it is skipped; Chromium gets
# the same mapping via --host-resolver-rules in playwright.config.ts.
host="${BASE_URL#https://}"
if ! curl -ksSf -o /dev/null --max-time 5 --resolve "${host}:443:127.0.0.1" "${BASE_URL}/health"; then
	echo "structure: ${BASE_URL}/health not reachable (via 127.0.0.1)" >&2
	echo "structure: run 'dde project:up' first" >&2
	exit 1
fi
export SKIP_E2E_SETUP=true

npx tsx scripts/check-routes.ts
npx playwright test --project=structure tests/structure/baseline.spec.ts --update-snapshots
AXE_RECORD=1 npx playwright test --project=structure tests/structure/axe.spec.ts
