#!/usr/bin/env bash
# @command client:test
# @description Run frontend Vitest suite in the client container
set -euo pipefail

# Anchor docker compose to this project regardless of the caller's cwd, and
# resolve the client service without hardcoding the compose project name
# (COMPOSE_PROJECT_NAME / checkout-dir overrides stay safe).
plugin_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$plugin_dir/../.."

if [ -z "$(docker compose ps -q client 2>/dev/null || true)" ]; then
	echo "client:test: client container not running" >&2
	echo "client:test: run 'dde project:up' first" >&2
	exit 1
fi

# Run as the dde user (host uid) via su-exec, so $DDE_UID/$DDE_GID expand
# in-container (the host shell has them unset). The client adapter chowns the
# build-time (uid 1000) node_modules to the dde user and the dev server CMD runs
# under su-exec too, so the Vite caches (.vite, .vite-temp) are dde-owned — no
# more root exec needed. `-- run` forces vitest single-run (non-watch); the
# trailing `_ "$@"` forwards extra args to the inner sh -c's positional params.
exec docker compose exec -T client sh -c 'exec su-exec "$DDE_UID:$DDE_GID" npm run test -- run "$@"' _ "$@"
