// Module-level filter state for the wallet list. Living at module scope (not in
// the page component) means it survives navigation/remount — so returning from a
// resource detail lands back on the same filtered view. Merchant-detail keeps
// its own local filters; this store is the wallet context only.
export const walletFilters = $state({
	searchInput: '',
	typeFilter: 'all',
	statusFilter: 'active',
	sortBy: 'newest',
	ownerFilter: 'all',
	favoritesOnly: false,
	expiringFilter: 'all',
	// Last scroll position of the wallet list, restored after the list renders
	// (SvelteKit's own scroll restore fires before the async list has loaded).
	scrollY: 0
});
