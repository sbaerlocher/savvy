import { browser } from '$app/environment';
import { ApiError, authApi } from '$lib/api';
import { logger } from '$lib/utils/logger';
import { redirect } from '@sveltejs/kit';

/**
 * SSR Configuration (SVL-001 Fix)
 *
 * CSR-Only Mode (SPA):
 * - Production: No Node.js runtime available
 * - Go Backend Guards (RequireAuth/RequireAdmin) protect all routes
 * - Security handled server-side BEFORE HTML is served
 *
 * Security Flow:
 * 1. Browser → GET /admin → Go Backend
 * 2. Go Middleware checks session (SetCurrentUser + RequireAdmin)
 * 3. If not admin → 403 or redirect BEFORE HTML sent
 * 4. If admin → ServeSPA() delivers index.html
 * 5. SvelteKit starts client-side (this load function runs)
 * 6. API calls → /api/v1/* → Go Backend validates again
 */
export const ssr = false; // ❌ CSR-only (SPA mode) - Go Backend handles auth

const layoutLogger = logger.child('Layout');

/**
 * Root layout load function (CLIENT-SIDE ONLY)
 * Fetches user data from API for authenticated routes
 */
export async function load({ url, fetch }) {
	// Public routes that don't require authentication
	const publicRoutes = [
		'/login',
		'/register',
		'/forgot-password',
		'/reset-password',
		'/verify-email',
		'/auth/oauth/login',
		'/auth/oauth/callback'
	];
	const isPublicRoute = publicRoutes.some((route) =>
		url.pathname.startsWith(route)
	);

	// Skip user loading for public routes
	if (isPublicRoute) {
		return {};
	}

	// Client-side: fetch user from API
	if (browser) {
		// Check offline status
		if (!navigator.onLine) {
			// Use debug level - this is expected behavior when offline, not an error
			layoutLogger.debug(
				'Offline: skipping user data fetch (will use cached auth state)'
			);
			return {};
		}

		// Online: fetch user from API (using SvelteKit's fetch)
		try {
			const { user } = await authApi.me({ fetch });
			return { user };
		} catch (error) {
			// Handle authentication errors
			if (error instanceof ApiError) {
				// 401 Unauthorized: User not logged in
				// 403 Forbidden: User doesn't have access
				if (error.status === 401 || error.status === 403) {
					layoutLogger.warn(
						`Auth failed (${error.status}): redirecting to /login`
					);
					throw redirect(
						303,
						`/login?redirect=${encodeURIComponent(url.pathname)}`
					);
				}
			}

			// API failed: Go Backend will have already handled auth check
			// If we reach here, user might be logged out
			layoutLogger.error('Failed to fetch user:', error);
			return {};
		}
	}

	return {};
}
