#!/usr/bin/env bash
# @command e2e:up
# @description Start the E2E test stack (detached by default, profile=e2e)
set -euo pipefail

# Detach by default, but step aside when the caller opts into an
# attached/watch run — compose rejects `-d` combined with those flags.
default_flags="-d"
for arg in "$@"; do
    case "$arg" in
    --attach | --attach=* | --watch | --abort-on-container-exit | --exit-code-from | --exit-code-from=*)
        default_flags=""
        ;;
    esac
done

# shellcheck disable=SC2086 -- intentionally unquoted: empty or "-d"
exec docker compose --profile e2e up $default_flags "$@"
