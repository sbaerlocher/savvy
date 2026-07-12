#!/usr/bin/env bash
# @command db:seed
# @description Seed the database with test data (runs in the api container)
set -euo pipefail

api_container=$(docker ps \
	--filter "label=com.docker.compose.project=savvy" \
	--filter "label=com.docker.compose.service=api" \
	--format '{{.Names}}' | head -1)

if [ -z "$api_container" ]; then
	echo "db:seed: savvy api container not found" >&2
	echo "db:seed: run 'dde project:up' first" >&2
	exit 1
fi

exec docker exec "$api_container" go run -mod=mod /app/cmd/seed/main.go
