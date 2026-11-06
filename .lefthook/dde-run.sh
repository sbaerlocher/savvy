#!/usr/bin/env bash
# Run a lefthook command inside a dde container instead of on the host.
#
#   .lefthook/dde-run.sh <service> <cmd> [args...]
#
# The host cwd (lefthook's `root:`) is mapped to the container path, so
# relative {staged_files}/{push_files} keep working:
#   api    -> the repo root is mounted at /app
#   client -> client/ is mounted at /app/client (the client container sees
#             nothing outside client/)
#
# `dde project:exec` runs as the host uid, never with --root here, so files a
# command rewrites (gofmt, prettier, eslint --fix) stay owned by the host user.
# The api image owns its Go caches as uid 1000; for any other host uid the
# command gets per-uid GOMODCACHE/GOCACHE under /tmp instead of EACCES.
#
# Outcomes:
#   - dde not installed   -> skip with a notice, exit 0 (contributors without
#                            dde; CI does not run lefthook)
#   - service not running -> fail with a hint; a hook that could not run must
#                            not pass a commit/push silently
#   - otherwise           -> the command's exit code
#
# dde resolves the stack from the cwd, so in a git worktree the worktree's own
# stack must be up (the main clone's containers mount the main clone's files).

set -euo pipefail

if [ "$#" -lt 2 ]; then
	echo "usage: $0 <service> <cmd> [args...]" >&2
	exit 2
fi

service=$1
shift

if ! command -v dde >/dev/null 2>&1; then
	echo "lefthook: dde not installed, skipping '$*' (runs in the dde '$service' container)" >&2
	exit 0
fi

# The repo root is this script's parent dir. Not `git rev-parse
# --show-toplevel`: git exports GIT_DIR to some hooks (pre-push in a worktree),
# and with GIT_DIR set that returns the cwd instead of the repo root.
toplevel=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd -P)
here=$(pwd -P)
case "$here" in
"$toplevel") rel=. ;;
"$toplevel"/*) rel=${here#"$toplevel"/} ;;
*)
	echo "lefthook: '$here' is outside the repository $toplevel" >&2
	exit 2
	;;
esac

case "$service" in
api) base=/app ;;
client)
	case "$rel" in
	client) rel=. ;;
	client/*) rel=${rel#client/} ;;
	*)
		echo "lefthook: '$PWD' is outside client/, not visible in the client container" >&2
		exit 2
		;;
	esac
	base=/app/client
	;;
*)
	echo "lefthook: unknown service '$service' (expected api or client)" >&2
	exit 2
	;;
esac

workdir=$base
[ "$rel" = . ] || workdir=$base/$rel

if ! probe=$(dde project:exec --service="$service" -- true </dev/null 2>&1); then
	{
		[ -z "$probe" ] || printf '%s\n' "$probe"
		echo "lefthook: cannot exec into dde service '$service' for $toplevel"
		echo "lefthook: start the stack with 'dde project:up', or skip this hook once with"
		echo "lefthook:   LEFTHOOK_EXCLUDE=<command name from lefthook.yml> git commit|push"
	} >&2
	exit 1
fi

# Runs in the container shell: fall back to per-uid Go caches when the image's
# (owned by uid 1000) are not writable for the exec user.
# shellcheck disable=SC2016 # expanded by the container shell
in_container='
cd "$1" && shift
if command -v go >/dev/null 2>&1; then
	[ -w "$(go env GOMODCACHE)" ] || export GOMODCACHE="/tmp/gomodcache-$(id -u)"
	[ -w "$(go env GOCACHE)" ] || export GOCACHE="/tmp/gocache-$(id -u)"
fi
exec "$@"'

exec dde project:exec --service="$service" -- sh -c "$in_container" _ "$workdir" "$@" </dev/null
