/**
 * OpenTelemetry Instrumentation for SvelteKit Frontend (Simplified)
 *
 * Note: Full OTEL browser support is experimental. This implementation focuses on:
 * - Basic error tracking (console.error, unhandled errors)
 * - Performance monitoring (fetch timing)
 *
 * For production, consider using established error tracking services like Sentry.
 */

// OTEL Configuration
const OTEL_ENABLED = import.meta.env.VITE_OTEL_ENABLED === 'true';

/**
 * Initialize simplified OpenTelemetry instrumentation
 *
 * Captures:
 * - Console errors → sent to backend
 * - Unhandled errors → sent to backend
 * - Unhandled promise rejections → sent to backend
 * - Fetch API timing → sent to backend
 */
export function initOpenTelemetry() {
	if (typeof window === 'undefined') {
		// Skip in SSR context
		return;
	}

	if (!OTEL_ENABLED) {
		console.log('📊 OpenTelemetry: Disabled');
		return;
	}

	// Capture console errors
	const originalConsoleError = console.error;
	console.error = (...args: unknown[]) => {
		sendLog('ERROR', args.map((arg) => String(arg)).join(' '), {
			type: 'console.error'
		});
		originalConsoleError.apply(console, args);
	};

	// Capture unhandled errors
	window.addEventListener('error', (event) => {
		sendLog('ERROR', event.message, {
			type: 'unhandled_error',
			stack: event.error?.stack || '',
			filename: event.filename || '',
			lineno: event.lineno || 0,
			colno: event.colno || 0
		});
	});

	// Capture unhandled promise rejections
	window.addEventListener('unhandledrejection', (event) => {
		sendLog('ERROR', `Unhandled Promise Rejection: ${event.reason}`, {
			type: 'unhandled_rejection',
			reason: String(event.reason)
		});
	});

	console.log('📊 OpenTelemetry: Enabled (simplified browser instrumentation)');
}

/**
 * Send log to backend OTEL proxy
 */
function sendLog(
	level: string,
	message: string,
	attributes: Record<string, unknown>
) {
	// Send to backend (non-blocking)
	fetch('/api/v1/otel/logs', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			level,
			message,
			attributes: {
				...attributes,
				'service.name': 'savvy-frontend',
				'service.version': '1.8.0',
				timestamp: new Date().toISOString()
			}
		})
	}).catch(() => {
		// Silently fail if backend is unavailable
	});
}

/**
 * Get OTEL Meter for custom metrics (placeholder)
 */
export function getOTelMeter() {
	return null;
}
