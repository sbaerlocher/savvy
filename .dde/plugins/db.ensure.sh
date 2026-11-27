#!/usr/bin/env bash
# @command db:ensure
# @description Create the app database on the dde stock postgres if it doesn't exist yet
set -euo pipefail

# Anchor docker compose to this project regardless of the caller's cwd
# (COMPOSE_PROJECT_NAME / checkout-dir overrides stay safe).
plugin_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$plugin_dir/../.."

# The api container's DATABASE_URL is the source of truth for the db name:
# dde derives it per-worktree (e.g. savvy vs savvy_feat_dashboard), so we must
# not hardcode it here.
if [ -z "$(docker compose ps -q api 2>/dev/null || true)" ]; then
	echo "db:ensure: api container not running" >&2
	echo "db:ensure: run 'dde project:up' first" >&2
	exit 1
fi

db_url=$(docker compose exec -T api printenv DATABASE_URL 2>/dev/null || true)
if [ -z "$db_url" ]; then
	echo "db:ensure: DATABASE_URL not set in api container" >&2
	exit 1
fi

# Strip query string, then take the path segment after the last slash.
db_name=${db_url%%\?*}
db_name=${db_name##*/}
if [ -z "$db_name" ]; then
	echo "db:ensure: could not parse database name from DATABASE_URL" >&2
	exit 1
fi

# shellcheck source=.dde/lib/postgres.sh
. "$(dirname "${BASH_SOURCE[0]}")/../lib/postgres.sh"
postgres_container=$(dde_postgres db:ensure) || exit 1

# CREATE DATABASE has no IF NOT EXISTS, so guard on pg_database.
exists=$(docker exec "$postgres_container" psql -U postgres -tAc \
	"SELECT 1 FROM pg_database WHERE datname = '${db_name}';")
if [ "$exists" = "1" ]; then
	echo "db:ensure: ${db_name} already exists"
	exit 0
fi

docker exec "$postgres_container" psql -U postgres -v ON_ERROR_STOP=1 \
	-c "CREATE DATABASE ${db_name} OWNER postgres;"
echo "db:ensure: ${db_name} created"
