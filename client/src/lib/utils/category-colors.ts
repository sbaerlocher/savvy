/**
 * Central color definitions for resource categories.
 * All Tailwind classes must be written out fully (no interpolation)
 * so the Tailwind compiler can detect them.
 */
export const categoryColors = {
	cards: {
		badge: 'bg-accent-100 text-accent-800',
		accent: 'bg-accent',
		filter: 'bg-accent-50 text-accent-hover',
		action: 'bg-accent-50 text-accent-hover hover:bg-accent-100'
	},
	vouchers: {
		badge: 'bg-emerald-100 text-emerald-800',
		accent: 'bg-emerald-500',
		filter: 'bg-emerald-50 text-emerald-700',
		action: 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100'
	},
	giftCards: {
		badge: 'bg-violet-100 text-violet-800',
		accent: 'bg-violet-500',
		filter: 'bg-violet-50 text-violet-700',
		action: 'bg-violet-50 text-violet-700 hover:bg-violet-100'
	}
} as const;
