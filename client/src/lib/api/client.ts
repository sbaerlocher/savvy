import type { ErrorResponse } from '$lib/types/api';
import { browser } from '$app/environment';
import { logger } from '$lib/utils/logger';

const API_BASE_URL = '/api/v1';

const apiLogger = logger.child('API');

/**
 * Gets CSRF token from cookie
 */
export function getCSRFToken(): string | null {
	const match = document.cookie.match(/csrf_token=([^;]+)/);
	return match ? match[1] : null;
}

/**
 * API Client Error
 */
export class ApiError extends Error {
	constructor(
		public status: number,
		public error: string,
		message: string
	) {
		super(message);
		this.name = 'ApiError';
	}
}

/**
 * Generic API request handler
 */
async function apiRequest<T>(
	endpoint: string,
	options: RequestInit = {},
	customFetch?: typeof fetch
): Promise<T> {
	const url = `${API_BASE_URL}${endpoint}`;
	const fetchFn = customFetch || fetch;

	const method = options.method || 'GET';
	const isMutation = ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method);

	const headers: Record<string, string> = {
		'Content-Type': 'application/json'
	};

	if (options.headers) {
		const existingHeaders = new Headers(options.headers);
		existingHeaders.forEach((value, key) => {
			headers[key] = value;
		});
	}

	// Block mutations when offline — defense-in-depth (UI also disables buttons)
	if (isMutation && browser && !navigator.onLine) {
		throw new ApiError(
			0,
			'offline',
			'Cannot perform this action while offline'
		);
	}

	// Add CSRF token for mutations
	if (isMutation) {
		const csrfToken = getCSRFToken();
		if (csrfToken) {
			headers['X-CSRF-Token'] = csrfToken;
		} else {
			throw new Error(`CSRF token missing for mutation: ${method} ${endpoint}`);
		}
	}

	apiLogger.debug(`${method} ${endpoint}`);

	const controller = new AbortController();
	const timeoutId = setTimeout(() => controller.abort(), 30_000);

	let response: Response;
	try {
		response = await fetchFn(url, {
			...options,
			headers,
			credentials: 'include',
			signal: options.signal ?? controller.signal
		});
	} catch (error) {
		clearTimeout(timeoutId);
		if (error instanceof DOMException && error.name === 'AbortError') {
			throw new ApiError(
				408,
				'timeout',
				'Request timed out — please try again'
			);
		}
		throw error;
	}
	clearTimeout(timeoutId);

	apiLogger.debug(`${method} ${endpoint} → ${response.status}`);

	// Handle non-JSON responses
	if (response.status === 204) {
		return {} as T;
	}

	const contentType = response.headers.get('content-type');
	if (!contentType || !contentType.includes('application/json')) {
		if (!response.ok) {
			throw new ApiError(
				response.status,
				'server_error',
				'Server returned non-JSON response'
			);
		}
		return {} as T;
	}

	let data;
	try {
		data = await response.json();
	} catch (error) {
		// Handle JSON parsing errors gracefully (malformed JSON)
		if (error instanceof SyntaxError) {
			apiLogger.error(`JSON parse error for ${method} ${endpoint}:`, error);
			throw new ApiError(
				response.status,
				'parse_error',
				'Server returned invalid JSON response'
			);
		}
		throw error;
	}

	if (!response.ok) {
		// Handle rate limiting with user-friendly message (SVL-SEC-003)
		if (response.status === 429) {
			const retryAfter = response.headers.get('Retry-After') || '60';
			const retrySeconds = parseInt(retryAfter, 10);
			const retryMinutes = Math.ceil(retrySeconds / 60);

			const message =
				retrySeconds < 60
					? `Too many requests. Please wait ${retrySeconds} seconds before trying again.`
					: `Too many requests. Please wait ${retryMinutes} minute${retryMinutes > 1 ? 's' : ''} before trying again.`;

			throw new ApiError(429, 'rate_limit_exceeded', message);
		}

		const error = data as ErrorResponse;
		throw new ApiError(response.status, error.error, error.message);
	}

	return data as T;
}

/**
 * API Client Options
 */
export interface ApiOptions {
	fetch?: typeof fetch;
}

/**
 * API Client — pure fetch wrapper with CSRF and error handling
 */
export const api = {
	get: <T>(
		endpoint: string,
		options?: RequestInit & ApiOptions
	): Promise<T> => {
		const { fetch: customFetch, ...fetchOptions } = options || {};
		return apiRequest<T>(
			endpoint,
			{ ...fetchOptions, method: 'GET' },
			customFetch
		);
	},

	post: <T>(
		endpoint: string,
		body?: unknown,
		options?: RequestInit & ApiOptions
	): Promise<T> => {
		const { fetch: customFetch, ...fetchOptions } = options || {};
		return apiRequest<T>(
			endpoint,
			{
				...fetchOptions,
				method: 'POST',
				body: body ? JSON.stringify(body) : undefined
			},
			customFetch
		);
	},

	put: <T>(
		endpoint: string,
		body?: unknown,
		options?: RequestInit & ApiOptions
	): Promise<T> => {
		const { fetch: customFetch, ...fetchOptions } = options || {};
		return apiRequest<T>(
			endpoint,
			{
				...fetchOptions,
				method: 'PUT',
				body: body ? JSON.stringify(body) : undefined
			},
			customFetch
		);
	},

	patch: <T>(
		endpoint: string,
		body?: unknown,
		options?: RequestInit & ApiOptions
	): Promise<T> => {
		const { fetch: customFetch, ...fetchOptions } = options || {};
		return apiRequest<T>(
			endpoint,
			{
				...fetchOptions,
				method: 'PATCH',
				body: body ? JSON.stringify(body) : undefined
			},
			customFetch
		);
	},

	delete: <T>(
		endpoint: string,
		options?: RequestInit & ApiOptions
	): Promise<T> => {
		const { fetch: customFetch, ...fetchOptions } = options || {};
		return apiRequest<T>(
			endpoint,
			{ ...fetchOptions, method: 'DELETE' },
			customFetch
		);
	}
};
