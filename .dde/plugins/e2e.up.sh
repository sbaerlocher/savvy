#!/usr/bin/env bash
# @command e2e:up
# @description Start the E2E test stack (detached, profile=e2e)
set -euo pipefail

exec docker compose --profile e2e up -d "$@"
