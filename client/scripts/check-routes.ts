/**
 * Guard: every `+page.svelte` in the route tree must appear in
 * tests/structure/routes.ts, and every listed route must still exist.
 *
 * The route list drives the whole structural baseline, so a route added
 * without a baseline entry would silently go uncovered. Run via
 * `npm run structure:check-routes`.
 */
import { readdirSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import { DYNAMIC_ROUTES, STATIC_ROUTES } from '../tests/structure/routes';

const here = fileURLToPath(new URL('.', import.meta.url));
const ROUTES_DIR = join(here, '..', 'src', 'routes');

/** Collect every `+page.svelte` and turn its directory into a URL path. */
function collectPagePaths(dir: string, acc: string[] = []): string[] {
	for (const entry of readdirSync(dir)) {
		const full = join(dir, entry);
		if (statSync(full).isDirectory()) {
			collectPagePaths(full, acc);
		} else if (entry === '+page.svelte') {
			const rel = relative(ROUTES_DIR, dir).replace(/\\/g, '/');
			// SvelteKit `[id]` → our `:id` template notation.
			const url = '/' + rel.replace(/\[([^\]]+)\]/g, ':$1');
			acc.push(url === '/.' ? '/' : url);
		}
	}
	return acc;
}

const onDisk = new Set(collectPagePaths(ROUTES_DIR));
const listed = new Set(
	[...STATIC_ROUTES, ...DYNAMIC_ROUTES].map((r) => r.path)
);

const missing = [...onDisk].filter((p) => !listed.has(p)).sort();
const stale = [...listed].filter((p) => !onDisk.has(p)).sort();

if (missing.length || stale.length) {
	if (missing.length) {
		console.error('Routes on disk but missing from tests/structure/routes.ts:');
		for (const p of missing) console.error(`  + ${p}`);
	}
	if (stale.length) {
		console.error(
			'Routes listed in tests/structure/routes.ts but not on disk:'
		);
		for (const p of stale) console.error(`  - ${p}`);
	}
	process.exit(1);
}

console.log(`✓ route list matches the tree (${onDisk.size} pages)`);
