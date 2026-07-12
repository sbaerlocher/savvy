#!/usr/bin/env bash
# @command db:migrate-status
# @description Show applied database migrations (runs in the api container)
set -euo pipefail

api_container=$(docker ps \
	--filter "label=com.docker.compose.project=savvy" \
	--filter "label=com.docker.compose.service=api" \
	--format '{{.Names}}' | head -1)

if [ -z "$api_container" ]; then
	echo "db:migrate-status: savvy api container not found" >&2
	echo "db:migrate-status: run 'dde project:up' first" >&2
	exit 1
fi

exec docker exec "$api_container" go run -mod=mod /app/cmd/migrate/main.go status
