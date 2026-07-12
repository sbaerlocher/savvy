import { redirect } from '@sveltejs/kit';

// Cards list merged into the unified /wallet overview (type-filtered).
export function load() {
	throw redirect(307, '/wallet?type=cards');
}
