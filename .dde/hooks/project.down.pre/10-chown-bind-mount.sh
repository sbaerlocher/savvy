#!/usr/bin/env bash
# Hand root-owned files in the bind mount back to the host user before the
# stack goes away.
#
# dde starts the containers as root (user 0:0), so anything written into the
# bind mount as root — `dde project:exec --root`, git inside the container,
# tooling run as root — lands root-owned on the host and breaks host git
# (`.git/index`) and lefthook with EACCES.
#
# Runs in the api container: it bind-mounts the whole repo at /app, while the
# client container only sees ./client. In a git worktree dde also mounts the
# main clone's .git at its host path, so root git ops write there (e.g.
# .git/worktrees/<name>/index); that dir is included when it lies outside the
# project dir.
#
# /app/tmp is pruned: inside the container it is the anonymous volume, not the
# host dir (the host mountpoint is pre-created by
# project.up.pre/10-volume-mountpoints.sh).
#
# Best effort: when the api container is not running (e.g. e2e profile) the
# exec fails, a warning is printed and `project:down` continues.

set -euo pipefail

hook_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH='' cd -- "$hook_dir/../../.." && pwd -P)
cd "$project_dir"

paths=(/app)
if common_dir=$(git rev-parse --path-format=absolute --git-common-dir 2>/dev/null); then
	case "$common_dir" in
	"$project_dir"/*) ;;
	*) paths+=("$common_dir") ;;
	esac
fi

if ! dde project:exec --root --service=api -- find "${paths[@]}" \
	-path /app/tmp -prune \
	-o -user 0 -exec chown "$(id -u):$(id -g)" {} +; then
	echo "warn: could not chown root-owned files in the bind mount; run 'dde project:up && dde project:down' to retry" >&2
fi
