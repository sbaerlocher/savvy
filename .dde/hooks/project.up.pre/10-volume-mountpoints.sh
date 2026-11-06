#!/usr/bin/env bash
# Pre-create the host-side mountpoints of the anonymous volumes declared in
# docker-compose.yml (api: /app/tmp, client: /app/client/node_modules).
#
# When such a mountpoint is missing inside the bind mount, Docker creates it as
# root, leaving a root-owned ./tmp or ./client/node_modules on the host. Host
# tools then fail with EACCES (lefthook, `cd client && npm ci`), and the dir
# cannot be fixed from inside the container because the volume shadows it.
# Creating the dirs as the host user first keeps them user-owned; the volume
# content itself stays inside the container.
#
# Independent of the other project.up.pre hooks, so ordering does not matter.
# Keep the list in sync with the anonymous volumes in docker-compose.yml.

set -euo pipefail

hook_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
cd "$hook_dir/../../.."

mkdir -p tmp client/node_modules
