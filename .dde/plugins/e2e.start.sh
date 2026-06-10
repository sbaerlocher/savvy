#!/usr/bin/env bash
# @command e2e:start
# @description Bring up a clean E2E stack (down + reset-db + up + wait)
set -euo pipefail

# Stop any leftover stack from a previous run so we always start clean.
dde project:e2e:down >/dev/null 2>&1 || true

# Drop+create the savvy_e2e database on the dde stock postgres.
dde project:e2e:reset-db

# Start app-e2e detached on the dde-services-savvy network.
dde project:e2e:up

# Wait until app-e2e is healthy (default 90 s).
dde project:e2e:wait -- --timeout 90

cat <<EOF
e2e:start: stack ready
  app:    https://e2e.savvy.test
  health: https://e2e.savvy.test/health
  db:     savvy_e2e on the dde stock postgres

Run tests with:  dde project:e2e:test
Tear down with:  dde project:e2e:down
EOF
