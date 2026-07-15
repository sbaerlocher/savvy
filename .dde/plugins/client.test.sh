#!/usr/bin/env bash
# @command client:test
# @description Run frontend Vitest suite in the client container
set -euo pipefail

# Anchor docker compose to this project regardless of the caller's cwd, and
# resolve the client service without hardcoding the compose project name
# (COMPOSE_PROJECT_NAME / checkout-dir overrides stay safe).
plugin_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$plugin_dir/../.."

if [ -z "$(docker compose ps -q client 2>/dev/null || true)" ]; then
	echo "client:test: client container not running" >&2
	echo "client:test: run 'dde project:up' first" >&2
	exit 1
fi

# Run as the container's default user (root, as dde starts the service). The
# dev server (PID 1) runs as root and owns the live Vite caches it creates in
# node_modules (.vite, .vite-temp); execing as any other user hits EACCES
# writing those. `-- run` forces vitest single-run (non-watch).
exec docker compose exec -T client npm run test -- run "$@"
