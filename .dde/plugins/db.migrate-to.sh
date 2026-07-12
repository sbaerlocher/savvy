#!/usr/bin/env bash
# @command db:migrate-to
# @description Migrate to a specific version (runs in the api container)
set -euo pipefail

target_version="${1:-}"

if [ -z "$target_version" ] || [ "$target_version" = "--help" ] || [ "$target_version" = "-h" ]; then
	cat <<EOF
Usage: dde project:db:migrate-to -- TARGET_VERSION

Migrates the database to the named migration version.

Example:
  dde project:db:migrate-to -- 202601230001_init_schema
EOF
	[ -z "$target_version" ] && exit 2 || exit 0
fi

# Anchor docker compose to this project regardless of the caller's cwd, and
# resolve the api service without hardcoding the compose project name
# (COMPOSE_PROJECT_NAME / checkout-dir overrides stay safe).
plugin_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$plugin_dir/../.."

if [ -z "$(docker compose ps -q api 2>/dev/null || true)" ]; then
	echo "db:migrate-to: api container not running" >&2
	echo "db:migrate-to: run 'dde project:up' first" >&2
	exit 1
fi

exec docker compose exec -T api go run -mod=mod /app/cmd/migrate/main.go to "$target_version"
