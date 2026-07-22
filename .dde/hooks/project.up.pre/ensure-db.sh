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
# starts the api. The base db name comes from the api service's DATABASE_URL
# in docker-compose.yml; dde's per-worktree rewrite lives in a temp compose
# override that is deleted after project:up, so `docker compose config` never
# sees it. The worktree suffix is therefore re-derived via
# `dde project:describe` the same way dde builds it: <base>_<suffix> with
# hyphens mapped to underscores.

set -euo pipefail

# Anchor docker compose to this project regardless of the caller's cwd.
hook_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$hook_dir/../../.."

# Resolved base compose config (worktree rewrite applied further below). jq
# is already a common dev/docker dependency; avoid adding a python3 host
# requirement no other dde hook has.
compose_config=$(docker compose config --format json 2>/dev/null || true)

# No api service (e.g. e2e-only profile) — the e2e binary owns its own db.
if ! printf '%s' "$compose_config" | jq -e '.services.api' >/dev/null 2>&1; then
	echo "ensure-db: no api service in compose config, skipping"
	exit 0
fi

# api exists, so a missing/empty DATABASE_URL is a real error — fail loud
# rather than swallowing it and letting the api crash-loop against a missing
# db. `// empty` keeps a null value from becoming the literal string "None".
db_url=$(printf '%s' "$compose_config" | jq -r '.services.api.environment.DATABASE_URL // empty')
if [ -z "$db_url" ]; then
	echo "ensure-db: api service has no DATABASE_URL set" >&2
	exit 1
fi

# Strip query string, then take the path segment after the last slash.
db_name=${db_url%%\?*}
db_name=${db_name##*/}
if [ -z "$db_name" ]; then
	echo "ensure-db: could not parse database name from DATABASE_URL" >&2
	exit 1
fi

# In a worktree dde rewrites the db name to <base>_<suffix>; mirror that.
# Unlike `docker compose config` above this is NOT guarded with `|| true`:
# an empty suffix is indistinguishable from "main clone", so a failing
# describe in a worktree would silently create the base db and reintroduce
# the crash-loop — fail loud instead. Note the naming scheme (<base>_<suffix>,
# hyphens → underscores) is dde-private; project:describe does not expose the
# resolved db name, so keep this in sync with dde's worktree rewrite.
if ! describe_json=$(dde project:describe --output=json 2>/dev/null); then
	echo "ensure-db: dde project:describe failed — cannot derive worktree db name" >&2
	exit 1
fi
worktree_suffix=$(printf '%s' "$describe_json" |
	jq -r '.data.worktree.suffix // empty')
if [ -n "$worktree_suffix" ]; then
	db_name="${db_name}_$(printf '%s' "$worktree_suffix" | tr '-' '_')"
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

# CREATE DATABASE has no IF NOT EXISTS, so guard on pg_database. The name is
# a branch-derived identifier; quote it so hyphens/leading digits stay valid.
exists=$(docker exec "$postgres_container" psql -U postgres -tAc \
	"SELECT 1 FROM pg_database WHERE datname = '${db_name}';")
if [ "$exists" = "1" ]; then
	echo "ensure-db: ${db_name} already exists"
	exit 0
fi

docker exec "$postgres_container" psql -U postgres -v ON_ERROR_STOP=1 \
	-c "CREATE DATABASE \"${db_name}\" OWNER postgres;"
echo "ensure-db: ${db_name} created"
