#!/usr/bin/env bash
# @command e2e:test
# @description Run Playwright E2E tests against the running e2e stack
set -euo pipefail

# Resolve the project root from the plugin's own location so the command
# works regardless of the user's current working directory.
plugin_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$plugin_dir/../.." && pwd)"
cd "$project_root/client"

# Default to --project=chromium unless the caller explicitly passed a
# --project flag. This keeps the common case (running a single spec
# locally) fast — firefox/webkit are gated behind explicit opt-in
# because not every developer has those browsers installed.
has_project=false
for arg in "$@"; do
    if [[ "$arg" == --project || "$arg" == --project=* ]]; then
        has_project=true
        break
    fi
done

if ! $has_project; then
    set -- --project=chromium "$@"
fi

exec npx playwright test "$@"
