// Pure filter/sort helpers for the wallet-style list screens (wallet +
// merchant detail). Extracted from WalletView.svelte so they are testable and
// framework-free: no Svelte state, no stores, no API. Component-scoped values
// (the filter state and currentUserId) are passed in as arguments instead of
// closed over, keeping these functions pure.

// Helper: check if item expires within N days
export function expiresWithinDays(
	dateStr: string | undefined,
	days: number
): boolean {
	if (!dateStr) return false;
	const expiry = new Date(dateStr);
	const now = new Date();
	const diffMs = expiry.getTime() - now.getTime();
	return diffMs > 0 && diffMs <= days * 24 * 60 * 60 * 1000;
}

// Shared filter: ownership, favorites, expiring
export function applyCommonFilters<
	T extends { owner?: { id?: string }; is_favorite: boolean }
>(
	items: T[],
	getExpiryDate: (item: T) => string | undefined,
	ownerFilter: string,
	favoritesOnly: boolean,
	expiringFilter: string,
	currentUserId: string | undefined
): T[] {
	let result = items;
	if (ownerFilter === 'mine') {
		result = result.filter(
			(item) => !item.owner || item.owner.id === currentUserId
		);
	} else if (ownerFilter === 'shared') {
		result = result.filter(
			(item) => item.owner && item.owner.id !== currentUserId
		);
	}
	if (favoritesOnly) {
		result = result.filter((item) => item.is_favorite);
	}
	if (expiringFilter === '7') {
		result = result.filter((item) => expiresWithinDays(getExpiryDate(item), 7));
	} else if (expiringFilter === '30') {
		result = result.filter((item) =>
			expiresWithinDays(getExpiryDate(item), 30)
		);
	}
	return result;
}

// Sort helper
export function sortItems<T>(
	items: T[],
	getDate: (item: T) => string,
	getValue: (item: T) => number,
	getExpiry: (item: T) => string | undefined,
	sortBy: string
): T[] {
	return [...items].sort((a, b) => {
		switch (sortBy) {
			case 'newest':
				return new Date(getDate(b)).getTime() - new Date(getDate(a)).getTime();
			case 'oldest':
				return new Date(getDate(a)).getTime() - new Date(getDate(b)).getTime();
			case 'value-desc':
				return getValue(b) - getValue(a);
			case 'value-asc':
				return getValue(a) - getValue(b);
			case 'expiry-asc': {
				const ea = getExpiry(a);
				const eb = getExpiry(b);
				if (!ea && !eb) return 0;
				if (!ea) return 1;
				if (!eb) return -1;
				return new Date(ea).getTime() - new Date(eb).getTime();
			}
			default:
				return 0;
		}
	});
}

// Card status filter. Cards carry a manual status (active | inactive |
// expired | lost | blocked): 'active' must match exactly, 'inactive' groups
// every non-active status so expired/lost/blocked cards stay reachable.
export function matchesCardStatus(
	status: string | undefined,
	filter: string
): boolean {
	if (filter === 'active') return status === 'active';
	if (filter === 'inactive') return status !== 'active';
	return true;
}

export function searchMerchant(
	name: string | undefined,
	q: string,
	matchMerchantName: boolean
): boolean {
	return matchMerchantName ? !!name?.toLowerCase().includes(q) : false;
}
