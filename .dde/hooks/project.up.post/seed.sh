#!/bin/sh
# Seed database after project:up
#
# Postgres' initdb only creates the default 'postgres' database, so on a
# fresh data volume the 'savvy' DB is missing and the API crashes before
# AUTO_MIGRATE can run. This hook makes the bring-up idempotent:
#
#   1. Locate the dde-managed postgres container by label.
#   2. Create the savvy database if it doesn't exist.
#   3. If we just created it (or the API is unhealthy), restart api so
#      AUTO_MIGRATE picks up the schema.
#   4. Wait for the API healthcheck to go green.
#   5. Run the seeder with retry/backoff for transient failures.

set -e

DB_NAME=savvy

postgres_container=$(docker ps \
  --filter "label=dde.managed=true" \
  --filter "label=dde.service=postgres" \
  --format '{{.Names}}' | head -1)

if [ -z "$postgres_container" ]; then
  echo "Postgres container not found (label dde.service=postgres)"
  exit 1
fi

api_container=$(docker compose ps -q api 2>/dev/null || true)
if [ -z "$api_container" ]; then
  # E2E mode (or any non-dev project:up) does not bring up the api
  # service; the e2e binary handles its own migrations and seeding,
  # so this hook becomes a no-op rather than a hard failure.
  echo "E2E mode: dev api container absent, skipping dev seed"
  exit 0
fi

echo "Waiting for postgres to accept connections..."
for i in $(seq 1 30); do
  if docker exec "$postgres_container" pg_isready -U postgres -q; then
    break
  fi
  if [ "$i" = "30" ]; then
    echo "Postgres did not become ready within 60s"
    exit 1
  fi
  sleep 2
done

db_exists=$(docker exec "$postgres_container" \
  psql -U postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")

if [ "$db_exists" != "1" ]; then
  echo "Creating database '$DB_NAME'..."
  docker exec "$postgres_container" psql -U postgres -c "CREATE DATABASE $DB_NAME"
  echo "Restarting api so AUTO_MIGRATE runs against the new database..."
  docker restart "$api_container" >/dev/null
fi

echo "Waiting for API to be healthy..."
for i in $(seq 1 45); do
  status=$(docker inspect "$api_container" \
    --format '{{.State.Health.Status}}' 2>/dev/null || echo "missing")
  if [ "$status" = "healthy" ]; then
    break
  fi
  if [ "$i" = "45" ]; then
    echo "API did not become healthy within 90s (last status: $status)"
    exit 1
  fi
  sleep 2
done

echo "Seeding database..."
attempts=5
i=1
while [ "$i" -le "$attempts" ]; do
  if docker compose exec -T api go run ./cmd/seed/; then
    exit 0
  fi
  if [ "$i" -lt "$attempts" ]; then
    backoff=$((i * 3))
    echo "Seed attempt $i/$attempts failed, retrying in ${backoff}s..."
    sleep "$backoff"
  fi
  i=$((i + 1))
done

echo "Seed failed after $attempts attempts"
exit 1
