/**
 * SvelteKit Client Hooks (SVL-005 Fix)
 * Client-Side Error Handling
 *
 * Purpose:
 * - Global error handler for client-side errors
 * - Structured error logging
 * - Error ID generation for debugging
 * - Optional: Integration with error tracking services (Sentry, Rollbar)
 * - OpenTelemetry observability (traces, logs, metrics)
 */

import type { HandleClientError } from '@sveltejs/kit';
import { logger } from '$lib/utils/logger';
import { initOpenTelemetry } from '$lib/otel/instrumentation';

const errorLogger = logger.child('ClientError');

// Initialize OpenTelemetry (runs once on app startup)
initOpenTelemetry();

/**
 * Check if error is a chunk loading failure (offline navigation)
 * These happen when navigating to unvisited routes while offline
 */
function isChunkLoadError(error: unknown): boolean {
	if (error instanceof Error) {
		// Common patterns for chunk loading failures:
		// - "Failed to fetch dynamically imported module" (Chrome/Edge)
		// - "error loading dynamically imported module" (Firefox)
		// - "Importing a module script failed" (Safari)
		const message = error.message.toLowerCase();
		return (
			message.includes('dynamically imported') ||
			message.includes('dynamic import') ||
			message.includes('importing a module script failed') ||
			(message.includes('failed to fetch') && message.includes('module')) ||
			message.includes('chunk')
		);
	}
	return false;
}

/**
 * Global client error handler
 * Catches all unhandled errors in client-side code
 */
export const handleError: HandleClientError = async ({
	error,
	event,
	status,
	message
}) => {
	// Generate unique error ID for tracking
	const errorId = crypto.randomUUID();

	// Check if this is an offline chunk loading error
	const isOfflineChunkError = isChunkLoadError(error) && !navigator.onLine;

	if (isOfflineChunkError) {
		// Offline chunk loading - log as debug, not error (expected behavior)
		errorLogger.debug('Offline navigation to unvisited route:', {
			errorId,
			path: event.url.pathname,
			message: 'Route chunk not cached - redirecting to /offline'
		});

		// Return friendly offline message (user will be redirected to /offline by SW)
		return {
			message: 'Diese Seite ist offline nicht verfügbar',
			errorId
		};
	}

	// Real errors - log normally
	errorLogger.error('Client error occurred:', {
		errorId,
		status,
		message,
		path: event.url.pathname,
		error:
			error instanceof Error
				? {
						name: error.name,
						message: error.message,
						stack: error.stack
					}
				: error
	});

	// Optional: Send to external error tracking service
	// Example: Sentry, Rollbar, LogRocket, etc.
	// await reportToSentry(error, errorId, { path: event.url.pathname, status });

	// Return user-friendly error message
	// errorId allows correlation between user reports and logs
	return {
		message: message || 'Ein unerwarteter Fehler ist aufgetreten',
		errorId
	};
};
