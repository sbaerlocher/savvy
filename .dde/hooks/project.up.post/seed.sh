#!/usr/bin/env bash
# Seed database after project:up
#
# The database itself is created by the project.up.pre/ensure-db.sh hook,
# which runs before compose starts the api so AUTO_MIGRATE never races a
# missing database. This post-hook only waits for the api to go healthy
# (migrations done) and then seeds:
#
#   1. Wait for the API healthcheck to go green.
#   2. Run the seeder with retry/backoff for transient failures.

set -euo pipefail

# Anchor docker compose to this project regardless of the caller's cwd —
# `docker compose ps -q api` would otherwise silently return nothing
# (read as "E2E mode") when the hook runs from elsewhere.
hook_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$hook_dir/../../.."

api_container=$(docker compose ps -q api 2>/dev/null || true)
if [ -z "$api_container" ]; then
	# E2E mode (or any non-dev project:up) does not bring up the api
	# service; the e2e binary handles its own migrations and seeding,
	# so this hook becomes a no-op rather than a hard failure.
	echo "E2E mode: dev api container absent, skipping dev seed"
	exit 0
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
