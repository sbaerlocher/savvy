#!/usr/bin/env bash
# Lint Markdown files inside the dde client container.
#
#   .lefthook/markdownlint.sh <file.md>...   (paths relative to the repo root)
#
# Only the client container has Node, and it only mounts client/. The files and
# the root .markdownlint.json are therefore copied into a throwaway directory
# under client/ (git-ignored), linted there via `npx markdownlint-cli`, and the
# directory is removed again. The copies keep their repo-relative paths, so the
# reported locations match the real files.

set -euo pipefail

markdownlint_version=0.50.0

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
root=$(git rev-parse --show-toplevel)
cd "$root"

[ "$#" -gt 0 ] || exit 0

tmp=$(mktemp -d "$root/client/.markdownlint-hook.XXXXXX")
trap 'rm -rf "$tmp"' EXIT

cp .markdownlint.json "$tmp/"
for file in "$@"; do
	case "$file" in
	/* | .. | ../* | */../* | */..)
		echo "markdownlint: '$file' must be a path inside the repository" >&2
		exit 2
		;;
	esac
	mkdir -p "$tmp/$(dirname -- "$file")"
	cp -- "$file" "$tmp/$file"
done

cd "$tmp"
"$script_dir/dde-run.sh" client npx --yes "markdownlint-cli@$markdownlint_version" -- "$@"
