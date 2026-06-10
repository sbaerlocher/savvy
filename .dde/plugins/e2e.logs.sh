#!/usr/bin/env bash
# @command e2e:logs
# @description Tail logs from E2E services (default tail=50)
set -euo pipefail

tail="${1:-50}"
shift || true

exec docker compose --profile e2e logs --tail "$tail" "$@"
