#!/usr/bin/env bash
# @command e2e:down
# @description Stop and remove the E2E test stack (app-e2e)
set -euo pipefail

# -v is a no-op now that postgres-e2e was retired in favour of the dde
# stock postgres + a per-run savvy_e2e database; the flag is kept so any
# future profile-local volume gets cleared automatically.
exec docker compose --profile e2e down -v "$@"
