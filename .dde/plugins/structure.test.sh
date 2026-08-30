#!/usr/bin/env bash
# @command structure:test
# @description Run the structural baseline suite against the running dev stack
set -euo pipefail

# Resolve the project root from the plugin's own location so the command works
# regardless of the caller's cwd (same pattern as e2e.test.sh).
plugin_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$plugin_dir/../.." && pwd)"
cd "$project_root/client"

# Playwright runs on the host (not in a container), so it needs host-side
# node_modules plus the chromium binary. A fresh clone or worktree has neither.
if [[ ! -d node_modules/@playwright/test ]]; then
	echo "structure:test: host node_modules missing @playwright/test, running npm ci" >&2
	npm ci --loglevel=error
	npx playwright install chromium
fi

# playwright.config.ts is TypeScript and pulls in client/tsconfig.json, which
# extends the gitignored ./.svelte-kit/tsconfig.json.
if [[ ! -f .svelte-kit/tsconfig.json ]]; then
	echo "structure:test: .svelte-kit/tsconfig.json missing, running svelte-kit sync" >&2
	npx svelte-kit sync
fi

# The structure suite targets the DEV stack, not the e2e stack — it baselines
# the app as developed, and dev is where the refactor happens.
#
# dde serves each worktree under its own branch slug, so the host is derived
# from the current branch rather than hardcoded (a hardcoded savvy.test would
# silently baseline the main checkout from inside a worktree).
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

# The route list drives the whole baseline; a route added without an entry
# would silently go uncovered, so the guard runs before the suite.
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

exec npx playwright test --project=structure "$@"
