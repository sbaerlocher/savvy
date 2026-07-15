# shellcheck shell=sh
# Fix node_modules ownership so the dde user (host uid) can write Vite caches
# (.vite / .vite-temp). The image installs deps as uid 1000 (the node user), so
# dde exec — which runs as the host uid — hits EACCES writing those caches when
# running vitest/svelte-check. Chown once at container start, before the CMD.
detect() {
	[ -d /app/client/node_modules ]
}

configure() {
	# Chown numerically to the exact uid/gid the dev-server CMD and
	# client.test.sh drop to via su-exec ($DDE_UID:$DDE_GID). Owning by number
	# instead of name guarantees alignment even if the `dde` user's name→uid
	# mapping ever diverges from DDE_UID/DDE_GID, and avoids resolving a group.
	uid="${DDE_UID:-}"
	gid="${DDE_GID:-}"
	if [ -z "$uid" ] || [ -z "$gid" ]; then
		echo "client adapter: DDE_UID/DDE_GID unset, skipping node_modules chown" >&2
		return 0
	fi
	# Log on failure instead of swallowing: a failing chown leaves the EACCES
	# this adapter exists to fix, so a broken setup must stay diagnosable.
	if ! chown -R "$uid:$gid" /app/client/node_modules; then
		echo "client adapter: chown of /app/client/node_modules to $uid:$gid failed" >&2
	fi
}
