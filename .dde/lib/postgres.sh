# shellcheck shell=bash
# Shared by the dde hooks and plugins that talk to the dde stock postgres.
#
# dde_postgres <log-prefix>
#   Prints the name of the dde stock postgres container once it accepts TCP
#   connections. Postgres is not one of dde's global services, so
#   `dde system:up` never starts it; when no container is running this starts
#   it via `dde system:service:up postgres`.
#
#   The probe goes over TCP (-h 127.0.0.1) on purpose: on a fresh volume the
#   image runs initdb with a temporary server that only listens on the unix
#   socket and then restarts, so a socket probe can report ready too early.
#   60s covers initdb on slow hosts.
dde_postgres() {
	local prefix=$1 container i
	for i in $(seq 1 30); do
		container=$(docker ps \
			--filter "label=dde.managed=true" \
			--filter "label=dde.service=postgres" \
			--format '{{.Names}}' | head -1)
		if [ -n "$container" ] &&
			docker exec "$container" pg_isready -h 127.0.0.1 -U postgres -q; then
			printf '%s\n' "$container"
			return 0
		fi
		if [ -z "$container" ] && [ "$i" = 1 ]; then
			echo "$prefix: dde stock postgres not running, starting it" >&2
			dde system:service:up postgres >&2 || true
		fi
		sleep 2
	done
	echo "$prefix: dde stock postgres not ready within 60s (label dde.service=postgres)" >&2
	echo "$prefix: check 'dde system:service:up postgres'" >&2
	return 1
}
