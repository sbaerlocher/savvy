/**
 * Route inventory for the structural baseline.
 *
 * Kept in sync with the SvelteKit route tree by `npm run structure:routes`
 * (scripts/check-routes.ts), which fails when a `+page.svelte` has no entry
 * here or an entry points at a page that no longer exists — an uncovered
 * route would otherwise look exactly like a covered one.
 *
 * Parameterised routes resolve at runtime against the seeded data, because
 * cmd/seed/main.go assigns random UUIDs (no fixed IDs to hardcode).
 */

export type Platform = 'ios' | 'android' | 'desktop';

export type RouteKind =
	/** Renders inside the app shell: nav, footer, `<main>` container. */
	| 'shell'
	/** Full-bleed screens without the shell chrome (auth, error, offline). */
	| 'standalone'
	/** Renders no markup at all — redirects in onMount. */
	| 'redirect';

export interface RouteSpec {
	/** URL path, or a `:param` template resolved at runtime. */
	path: string;
	/** Short id used in snapshot filenames. */
	id: string;
	kind: RouteKind;
	/** Requires an authenticated session. */
	auth: boolean;
	/** Requires an admin session. */
	admin?: boolean;
	/** Platforms where this route renders markup. Defaults to all three. */
	platforms?: Platform[];
	/** Why this route is excluded from the shared-container structure rules. */
	exempt?: string;
	/**
	 * The PageShell width the route opts into. `narrow` is the shell's own
	 * 680px reading column (PageShell.svelte WIDTH_CLASS) — still the one
	 * shared definition, just the narrow variant of it, so the container
	 * assertion expects 680 instead of the full shell width on desktop.
	 */
	width?: 'narrow';
	/**
	 * Why this route carries no screenshot baseline. Set only for routes whose
	 * rendered height changes between runs for reasons unrelated to layout —
	 * a screenshot that cannot repeat is not a regression guard, it is noise
	 * that trains you to ignore red. These routes stay fully covered by the
	 * structure rules, which assert on layout rather than pixels.
	 */
	noScreenshot?: string;
}

export const ALL_PLATFORMS: Platform[] = ['ios', 'android', 'desktop'];

export const STATIC_ROUTES: RouteSpec[] = [
	// --- shell routes (the refactor's actual target) -----------------------
	{ path: '/dashboard', id: 'dashboard', kind: 'shell', auth: true },
	{ path: '/wallet', id: 'wallet', kind: 'shell', auth: true },
	{
		path: '/merchants',
		id: 'merchants',
		kind: 'shell',
		auth: true,
		noScreenshot:
			'merchant list hydrates from the offline cache; a run occasionally captures it still empty, so the height is not repeatable'
	},
	{
		path: '/profile',
		id: 'profile',
		kind: 'shell',
		auth: true,
		noScreenshot:
			'renders the active-session list; every test run adds a session, so the page height differs between runs'
	},
	{
		path: '/security',
		id: 'security',
		kind: 'shell',
		auth: true,
		noScreenshot:
			'renders the active-session list; every test run adds a session, so the page height differs between runs'
	},
	{
		path: '/notifications',
		id: 'notifications',
		kind: 'shell',
		auth: true,
		noScreenshot:
			'notification list hydrates from the offline cache; a run occasionally captures it still empty, so the height is not repeatable'
	},
	{
		path: '/notifications/settings',
		id: 'notifications-settings',
		kind: 'shell',
		auth: true
	},
	{ path: '/cards/new', id: 'cards-new', kind: 'shell', auth: true },
	{ path: '/vouchers/new', id: 'vouchers-new', kind: 'shell', auth: true },
	{ path: '/gift-cards/new', id: 'gift-cards-new', kind: 'shell', auth: true },

	// --- admin shell routes -------------------------------------------------
	{
		path: '/admin/users',
		id: 'admin-users',
		kind: 'shell',
		auth: true,
		admin: true
	},
	{
		path: '/admin/users/new',
		id: 'admin-users-new',
		kind: 'shell',
		auth: true,
		admin: true
	},
	{
		path: '/admin/merchants',
		id: 'admin-merchants',
		kind: 'shell',
		auth: true,
		admin: true
	},
	{
		path: '/admin/merchants/new',
		id: 'admin-merchants-new',
		kind: 'shell',
		auth: true,
		admin: true
	},
	{
		path: '/admin/audit-log',
		id: 'admin-audit-log',
		kind: 'shell',
		auth: true,
		admin: true,
		noScreenshot:
			'the audit log records the test run visiting it, so the list grows between capture and verification'
	},
	{
		path: '/admin/email-templates',
		id: 'admin-email-templates',
		kind: 'shell',
		auth: true,
		admin: true
	},
	{
		path: '/admin/system-health',
		id: 'admin-system-health',
		kind: 'shell',
		auth: true,
		admin: true
	},

	// --- legacy redirect ----------------------------------------------------
	{
		path: '/settings',
		id: 'settings',
		kind: 'redirect',
		auth: true,
		exempt: 'legacy route; redirects to /profile (or the old tab targets) in onMount'
	},

	// --- standalone: auth screens ------------------------------------------
	// Android renders these full-bleed: the shell drops `<main>` padding and
	// the footer (see +layout.svelte ANDROID_FULL_BLEED).
	{ path: '/login', id: 'login', kind: 'standalone', auth: false },
	{ path: '/login/2fa', id: 'login-2fa', kind: 'standalone', auth: false },
	{ path: '/register', id: 'register', kind: 'standalone', auth: false },
	{
		path: '/forgot-password',
		id: 'forgot-password',
		kind: 'standalone',
		auth: false
	},
	{
		path: '/reset-password',
		id: 'reset-password',
		kind: 'standalone',
		auth: false
	},
	{
		path: '/verify-email',
		id: 'verify-email',
		kind: 'standalone',
		auth: false
	},
	{ path: '/unsubscribe', id: 'unsubscribe', kind: 'standalone', auth: false },

	// --- standalone: state screens -----------------------------------------
	{ path: '/offline', id: 'offline', kind: 'standalone', auth: false },

	// --- redirect-only: no markup to baseline ------------------------------
	{
		path: '/',
		id: 'root',
		kind: 'redirect',
		auth: false,
		exempt: 'redirects to /dashboard or /login in onMount'
	},
	{
		path: '/admin',
		id: 'admin-root',
		kind: 'redirect',
		auth: true,
		admin: true,
		exempt: 'redirects to /admin/users in onMount'
	}
];

/**
 * Routes with a `:id` segment. The seed assigns random UUIDs, so the concrete
 * path is resolved per run by navigating the list page and taking the first
 * item — see `resolveDynamicRoutes()` in helpers.ts.
 */
export const DYNAMIC_ROUTES: RouteSpec[] = [
	{ path: '/cards/:id', id: 'cards-detail', kind: 'shell', auth: true },
	{ path: '/vouchers/:id', id: 'vouchers-detail', kind: 'shell', auth: true },
	{
		path: '/gift-cards/:id',
		id: 'gift-cards-detail',
		kind: 'shell',
		auth: true
	},
	{
		path: '/merchants/:id',
		id: 'merchants-detail',
		kind: 'shell',
		auth: true,
		noScreenshot:
			'lists the merchant’s resources from the offline cache, same non-repeatable height as /merchants'
	},
	{
		path: '/admin/users/:id/edit',
		id: 'admin-users-edit',
		kind: 'shell',
		auth: true,
		admin: true
	},
	{
		path: '/admin/merchants/:id/edit',
		id: 'admin-merchants-edit',
		kind: 'shell',
		auth: true,
		admin: true
	}
];

/** Routes that must satisfy the shared-container structure rules. */
export function shellRoutes(routes: RouteSpec[] = STATIC_ROUTES): RouteSpec[] {
	return routes.filter((r) => r.kind === 'shell');
}

export function platformsFor(route: RouteSpec): Platform[] {
	return route.platforms ?? ALL_PLATFORMS;
}
