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

api_container=$(docker ps \
	--filter "label=com.docker.compose.project=savvy" \
	--filter "label=com.docker.compose.service=api" \
	--format '{{.Names}}' | head -1)

if [ -z "$api_container" ]; then
	echo "db:migrate-to: savvy api container not found" >&2
	echo "db:migrate-to: run 'dde project:up' first" >&2
	exit 1
fi

exec docker exec "$api_container" go run -mod=mod /app/cmd/migrate/main.go to "$target_version"
