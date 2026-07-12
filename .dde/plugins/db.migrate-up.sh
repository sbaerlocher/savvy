#!/usr/bin/env bash
# @command db:migrate-up
# @description Apply all pending database migrations (runs in the api container)
set -euo pipefail

# Anchor docker compose to this project regardless of the caller's cwd, and
# resolve the api service without hardcoding the compose project name
# (COMPOSE_PROJECT_NAME / checkout-dir overrides stay safe).
plugin_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$plugin_dir/../.."

if [ -z "$(docker compose ps -q api 2>/dev/null || true)" ]; then
	echo "db:migrate-up: api container not running" >&2
	echo "db:migrate-up: run 'dde project:up' first" >&2
	exit 1
fi

# The database must exist before migrations can run; the app can't create it
# itself (it connects straight to it). db:ensure is idempotent.
"$plugin_dir/db.ensure.sh"

exec docker compose exec -T api go run -mod=mod /app/cmd/migrate/main.go up
