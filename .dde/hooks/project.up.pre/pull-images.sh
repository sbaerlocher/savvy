#!/usr/bin/env bash
# dde's dev-layer build probes every compose image (all profiles) with a
# 30-second `docker run … exit 0` timeout. On a cold image cache — fresh
# CI runner, new machine — pulling the larger observability images
# (grafana, loki, tempo) inside that probe exceeds the timeout and fails
# `project:up` before the stack ever starts. Pre-pull all pinned images
# so the probe only ever runs already-present ones. `--ignore-buildable`
# skips the locally built services (api, client, app-e2e); on a warm
# cache the digest-pinned pull is a cheap no-op.
set -euo pipefail

docker compose --profile '*' pull --ignore-buildable --quiet
