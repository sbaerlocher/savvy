#!/usr/bin/env bash
# @command e2e:reset-db
# @description Drop and recreate the savvy_e2e database on the dde stock postgres
set -euo pipefail

DB_NAME=savvy_e2e

postgres_container=$(docker ps \
    --filter "label=dde.managed=true" \
    --filter "label=dde.service=postgres" \
    --format '{{.Names}}' | head -1)

if [ -z "$postgres_container" ]; then
    echo "e2e:reset-db: dde stock postgres container not found (label dde.service=postgres)" >&2
    echo "e2e:reset-db: run 'dde system:up' or 'dde project:up' first" >&2
    exit 1
fi

# Wait briefly for postgres to accept connections (typically already healthy).
for i in $(seq 1 10); do
    if docker exec "$postgres_container" pg_isready -U postgres -q; then
        break
    fi
    if [ "$i" = "10" ]; then
        echo "e2e:reset-db: postgres did not become ready within 20s" >&2
        exit 1
    fi
    sleep 2
done

# DROP DATABASE WITH (FORCE) requires PG 13+; the dde stock postgres is 18.x.
docker exec "$postgres_container" psql -U postgres -v ON_ERROR_STOP=1 \
    -c "DROP DATABASE IF EXISTS ${DB_NAME} WITH (FORCE);" \
    -c "CREATE DATABASE ${DB_NAME} OWNER postgres;"

echo "e2e:reset-db: ${DB_NAME} recreated"
