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
# The command runs as the container's default user (uid 1000 = the dde/host
# user), never with --root, so files it rewrites (gofmt, prettier, eslint
# --fix) stay owned by the host user.
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

toplevel=$(cd "$(git rev-parse --show-toplevel)" && pwd -P)
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

if ! dde project:exec --service="$service" -- true </dev/null >/dev/null 2>&1; then
	{
		echo "lefthook: dde service '$service' is not running for $toplevel"
		echo "lefthook: start the stack with 'dde project:up', or skip this hook once with"
		echo "lefthook:   LEFTHOOK_EXCLUDE=<command name from lefthook.yml> git commit|push"
	} >&2
	exit 1
fi

# shellcheck disable=SC2016 # expanded by the container shell
exec dde project:exec --service="$service" -- sh -c 'cd "$1" && shift && exec "$@"' _ "$workdir" "$@" </dev/null
