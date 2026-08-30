#!/usr/bin/env bash
# @command structure:reset-db
# @description Drop, recreate and re-seed the dev database for a clean structural run
set -euo pipefail

DB_NAME=savvy

plugin_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$plugin_dir/../.." && pwd)"

postgres_container=$(docker ps \
	--filter "label=dde.managed=true" \
	--filter "label=dde.service=postgres" \
	--format '{{.Names}}' | head -1)

if [ -z "$postgres_container" ]; then
	echo "structure:reset-db: dde stock postgres container not found" >&2
	echo "structure:reset-db: run 'dde project:up' first" >&2
	exit 1
fi

for i in $(seq 1 10); do
	if docker exec "$postgres_container" pg_isready -U postgres -q; then
		break
	fi
	if [ "$i" = "10" ]; then
		echo "structure:reset-db: postgres did not become ready within 20s" >&2
		exit 1
	fi
	sleep 2
done

# `db:migrate-reset` prompts for confirmation and dies on EOF, so it cannot be
# used non-interactively. DROP ... WITH (FORCE) needs PG 13+; dde ships 18.x.
docker exec "$postgres_container" psql -U postgres -v ON_ERROR_STOP=1 \
	-c "DROP DATABASE IF EXISTS ${DB_NAME} WITH (FORCE);" \
	-c "CREATE DATABASE ${DB_NAME} OWNER postgres;" >/dev/null

# The api container auto-migrates on boot (AUTO_MIGRATE=true); restarting it is
# the cheapest way to rebuild the schema the seed then fills.
cd "$project_root"
docker compose restart api >/dev/null

for i in $(seq 1 30); do
	if docker compose exec -T api wget --no-verbose --tries=1 --spider \
		http://localhost:8080/health 2>/dev/null; then
		break
	fi
	if [ "$i" = "30" ]; then
		echo "structure:reset-db: api did not come back within 60s" >&2
		exit 1
	fi
	sleep 2
done

dde project:db:seed >/dev/null
echo "structure:reset-db: ${DB_NAME} recreated and seeded"
