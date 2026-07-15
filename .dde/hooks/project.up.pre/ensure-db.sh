#!/usr/bin/env bash
# Create the app database on the dde stock postgres BEFORE the api starts.
#
# Postgres' initdb only creates the default 'postgres' database. On a fresh
# data volume — or in a git worktree, where dde rewrites DATABASE_URL to a
# per-branch db name (savvy vs savvy_feat_x) — the target db is missing, so
# the api crash-loops (AUTO_MIGRATE runs against a non-existent database)
# before the project.up.post seed hook ever gets a chance to create it.
#
# Running as a pre-up hook removes that race: the db exists before compose
# starts the api. The db name is read from the api service's DATABASE_URL as
# resolved by `docker compose config`, which honours dde's per-worktree
# COMPOSE_FILE override — so this stays correct in both main and worktrees.

set -euo pipefail

# Anchor docker compose to this project regardless of the caller's cwd.
hook_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$hook_dir/../../.."

# DATABASE_URL from compose config is the source of truth (worktree-aware).
db_url=$(docker compose config --format json 2>/dev/null |
	python3 -c "import sys,json; print(json.load(sys.stdin)['services']['api']['environment'].get('DATABASE_URL',''))" \
		2>/dev/null || true)

if [ -z "$db_url" ]; then
	# No api service (e.g. e2e-only profile) — the e2e binary owns its own db.
	echo "ensure-db: no api DATABASE_URL in compose config, skipping"
	exit 0
fi

# Strip query string, then take the path segment after the last slash.
db_name=${db_url%%\?*}
db_name=${db_name##*/}
if [ -z "$db_name" ]; then
	echo "ensure-db: could not parse database name from DATABASE_URL" >&2
	exit 1
fi

postgres_container=$(docker ps \
	--filter "label=dde.managed=true" \
	--filter "label=dde.service=postgres" \
	--format '{{.Names}}' | head -1)

if [ -z "$postgres_container" ]; then
	echo "ensure-db: dde stock postgres not found (label dde.service=postgres)" >&2
	echo "ensure-db: run 'dde system:up' first" >&2
	exit 1
fi

# Postgres is normally already healthy (dde system service); wait briefly.
for i in $(seq 1 10); do
	if docker exec "$postgres_container" pg_isready -U postgres -q; then
		break
	fi
	if [ "$i" = "10" ]; then
		echo "ensure-db: postgres did not become ready within 20s" >&2
		exit 1
	fi
	sleep 2
done

# CREATE DATABASE has no IF NOT EXISTS, so guard on pg_database.
exists=$(docker exec "$postgres_container" psql -U postgres -tAc \
	"SELECT 1 FROM pg_database WHERE datname = '${db_name}';")
if [ "$exists" = "1" ]; then
	echo "ensure-db: ${db_name} already exists"
	exit 0
fi

docker exec "$postgres_container" psql -U postgres -v ON_ERROR_STOP=1 \
	-c "CREATE DATABASE ${db_name} OWNER postgres;"
echo "ensure-db: ${db_name} created"
