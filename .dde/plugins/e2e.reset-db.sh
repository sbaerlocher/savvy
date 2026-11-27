#!/usr/bin/env bash
# @command e2e:reset-db
# @description Drop and recreate the savvy_e2e database on the dde stock postgres
set -euo pipefail

DB_NAME=savvy_e2e

# shellcheck source=.dde/lib/postgres.sh
. "$(dirname "${BASH_SOURCE[0]}")/../lib/postgres.sh"
postgres_container=$(dde_postgres e2e:reset-db) || exit 1

# DROP DATABASE WITH (FORCE) requires PG 13+; the dde stock postgres is 18.x.
docker exec "$postgres_container" psql -U postgres -v ON_ERROR_STOP=1 \
    -c "DROP DATABASE IF EXISTS ${DB_NAME} WITH (FORCE);" \
    -c "CREATE DATABASE ${DB_NAME} OWNER postgres;"

echo "e2e:reset-db: ${DB_NAME} recreated"
