#!/usr/bin/env bash
# @command e2e:wait
# @description Wait until an E2E service is healthy (default service=app-e2e, timeout=90s)
set -euo pipefail

timeout=90
service="app-e2e"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --timeout)
            timeout="${2:?--timeout requires a value in seconds}"
            shift 2
            ;;
        --timeout=*)
            timeout="${1#*=}"
            shift
            ;;
        --service)
            service="${2:?--service requires a service name}"
            shift 2
            ;;
        --service=*)
            service="${1#*=}"
            shift
            ;;
        --help|-h)
            cat <<EOF
Usage: dde e2e:wait [--timeout SECONDS] [--service NAME]

Polls the named compose service (default: app-e2e) until its health status
is "healthy". On timeout, the last 100 lines of the service's logs are
written to stderr and the command exits 1.
EOF
            exit 0
            ;;
        *)
            echo "e2e:wait: unknown argument: $1" >&2
            exit 2
            ;;
    esac
done

if ! [[ "$timeout" =~ ^[0-9]+$ ]] || [[ "$timeout" -le 0 ]]; then
    echo "e2e:wait: --timeout must be a positive integer (got: $timeout)" >&2
    exit 2
fi

deadline=$(( $(date +%s) + timeout ))

# Resolve the container ID for the requested service. compose may not have
# created it yet on the very first iteration, so we re-resolve on each loop.
get_container_id() {
    docker compose --profile e2e ps -q "$service" 2>/dev/null || true
}

get_health_status() {
    local cid="$1"
    [[ -z "$cid" ]] && { echo "missing"; return; }
    # If the service has no healthcheck declared, .State.Health is null.
    docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$cid" 2>/dev/null || echo "missing"
}

while true; do
    cid=$(get_container_id)
    status=$(get_health_status "$cid")

    case "$status" in
        healthy)
            exit 0
            ;;
        none)
            echo "e2e:wait: service '$service' has no healthcheck; cannot determine readiness" >&2
            exit 1
            ;;
    esac

    now=$(date +%s)
    if (( now >= deadline )); then
        echo "e2e:wait: timeout after ${timeout}s waiting for '$service' (last status: ${status:-unknown})" >&2
        echo "----- last logs for '$service' -----" >&2
        docker compose --profile e2e logs --tail 100 "$service" >&2 2>&1 || true
        exit 1
    fi

    sleep 2
done
