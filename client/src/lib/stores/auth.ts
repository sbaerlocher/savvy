import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { authApi, adminApi, ApiError } from '$lib/api';
import type { UserDTO, LoginRequest, RegisterRequest } from '$lib/types/api';
import { logger } from '$lib/utils/logger';
import { offlineDB } from '$lib/stores/offline-db';
import { languageStore } from '$lib/stores/i18n';

const authLogger = logger.child('Auth');

interface AuthState {
	user: UserDTO | null;
	isAuthenticated: boolean;
	isLoading: boolean;
	error: string | null;
	requires2FA: boolean;
}

const AUTH_STORAGE_KEY = 'savvy_auth_state';

/**
 * SECURITY (SVL-003): Only store isAuthenticated flag, not user data
 * - Prevents XSS attacks from stealing user data via localStorage
 * - User data must be fetched from /api/v1/auth/me (validates session)
 */
function loadAuthFromStorage(): AuthState {
	if (!browser) {
		return {
			user: null,
			isAuthenticated: false,
			isLoading: false,
			error: null,
			requires2FA: false
		};
	}

	try {
		const stored = localStorage.getItem(AUTH_STORAGE_KEY);
		if (stored) {
			const parsed = JSON.parse(stored);
			return {
				isAuthenticated: parsed.isAuthenticated || false,
				user: null,
				isLoading: false,
				error: null,
				requires2FA: false
			};
		}
	} catch (error) {
		authLogger.error('Failed to load auth from storage:', error);
	}

	return {
		user: null,
		isAuthenticated: false,
		isLoading: false,
		error: null,
		requires2FA: false
	};
}

function saveAuthToStorage(state: AuthState) {
	if (!browser) return;

	try {
		localStorage.setItem(
			AUTH_STORAGE_KEY,
			JSON.stringify({
				isAuthenticated: state.isAuthenticated
			})
		);
	} catch (error) {
		authLogger.error('Failed to save auth to storage:', error);
	}
}

function clearAuthFromStorage() {
	if (!browser) return;

	try {
		localStorage.removeItem(AUTH_STORAGE_KEY);
	} catch (error) {
		authLogger.error('Failed to clear auth from storage:', error);
	}
}

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>(loadAuthFromStorage());

	/**
	 * Clear offline data (IndexedDB + Service Worker caches)
	 * Called on login/register to prevent stale data from previous user
	 */
	async function clearOfflineData() {
		if (browser && 'indexedDB' in window) {
			try {
				await offlineDB.clearAll();
				authLogger.info('Cleared IndexedDB offline data');
			} catch (error) {
				authLogger.warn('Failed to clear IndexedDB:', error);
			}
		}
	}

	return {
		subscribe,

		async checkAuth(): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const { user } = await authApi.me();
				authLogger.info(`Auth check successful: ${user.email}`);

				// Sync language from backend to frontend
				if (user.language) {
					languageStore.setFromBackend(user.language);
				}

				// Note: We do NOT clear offline data here to preserve PWA functionality
				// clearOfflineData() is only called on login/register to prevent stale data

				const newState: AuthState = {
					user,
					isAuthenticated: true,
					isLoading: false,
					error: null,
					requires2FA: false
				};
				set(newState);
				saveAuthToStorage(newState);
			} catch {
				if (!navigator.onLine) {
					update((state) => ({ ...state, isLoading: false }));
					return;
				}

				const newState: AuthState = {
					user: null,
					isAuthenticated: false,
					isLoading: false,
					error: null,
					requires2FA: false
				};
				set(newState);
				clearAuthFromStorage();
			}
		},

		async login(credentials: LoginRequest): Promise<void> {
			update((state) => ({
				...state,
				isLoading: true,
				error: null,
				requires2FA: false
			}));

			try {
				const response = await authApi.login(credentials);

				// Check if 2FA is required
				if ('requires_2fa' in response) {
					authLogger.info('Login requires 2FA verification');
					const newState: AuthState = {
						user: null,
						isAuthenticated: false,
						isLoading: false,
						error: null,
						requires2FA: true
					};
					set(newState);
					return;
				}

				const { user } = response;
				authLogger.info(`Login successful: ${user.email}`);

				// Sync language from backend to frontend
				if (user.language) {
					languageStore.setFromBackend(user.language);
				}

				// Clear offline data on login to prevent stale data from previous user
				await clearOfflineData();

				const newState: AuthState = {
					user,
					isAuthenticated: true,
					isLoading: false,
					error: null,
					requires2FA: false
				};
				set(newState);
				saveAuthToStorage(newState);
			} catch (error) {
				const errorMessage =
					error instanceof ApiError ? error.message : 'Login fehlgeschlagen';
				const newState: AuthState = {
					user: null,
					isAuthenticated: false,
					isLoading: false,
					error: errorMessage,
					requires2FA: false
				};
				set(newState);
				clearAuthFromStorage();
				throw error;
			}
		},

		async register(data: RegisterRequest): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const { user } = await authApi.register(data);

				// Clear offline data on register to ensure fresh start
				await clearOfflineData();

				const newState: AuthState = {
					user,
					isAuthenticated: true,
					isLoading: false,
					error: null,
					requires2FA: false
				};
				set(newState);
				saveAuthToStorage(newState);
			} catch (error) {
				const errorMessage =
					error instanceof ApiError
						? error.message
						: 'Registrierung fehlgeschlagen';
				const newState: AuthState = {
					user: null,
					isAuthenticated: false,
					isLoading: false,
					error: errorMessage,
					requires2FA: false
				};
				set(newState);
				clearAuthFromStorage();
				throw error;
			}
		},

		async logout(): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				await authApi.logout();
				authLogger.info('Logout successful');
			} catch (error) {
				authLogger.error('Logout error:', error);
			}

			// Clear offline data on logout
			await clearOfflineData();

			const newState: AuthState = {
				user: null,
				isAuthenticated: false,
				isLoading: false,
				error: null,
				requires2FA: false
			};
			set(newState);
			clearAuthFromStorage();
		},

		async verify2FA(data: {
			code?: string;
			backup_code?: string;
		}): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				await authApi.verify2FA(data);
				authLogger.info('2FA verification successful');

				// After successful 2FA, fetch user data
				const { user } = await authApi.me();

				// Sync language from backend to frontend
				if (user.language) {
					languageStore.setFromBackend(user.language);
				}

				// Clear offline data on login to prevent stale data from previous user
				await clearOfflineData();

				const newState: AuthState = {
					user,
					isAuthenticated: true,
					isLoading: false,
					error: null,
					requires2FA: false
				};
				set(newState);
				saveAuthToStorage(newState);
			} catch (error) {
				const errorMessage =
					error instanceof ApiError ? error.message : '2FA verification failed';
				update((state) => ({
					...state,
					isLoading: false,
					error: errorMessage
				}));
				throw error;
			}
		},

		clearError(): void {
			update((state) => ({ ...state, error: null }));
		},

		setUser(user: UserDTO): void {
			const newState: AuthState = {
				user,
				isAuthenticated: true,
				isLoading: false,
				error: null,
				requires2FA: false
			};
			set(newState);
			saveAuthToStorage(newState);
		},

		getCachedAuth(): AuthState | null {
			return loadAuthFromStorage();
		},

		async startImpersonation(userId: string): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				await adminApi.startImpersonation(userId);

				// Clear offline data on impersonation to prevent data leakage
				await clearOfflineData();

				const { user } = await authApi.me();

				const newState: AuthState = {
					user,
					isAuthenticated: true,
					isLoading: false,
					error: null,
					requires2FA: false
				};
				set(newState);
				saveAuthToStorage(newState);

				window.location.href = '/dashboard';
			} catch (error) {
				const errorMessage =
					error instanceof ApiError ? error.message : 'Impersonation failed';
				update((state) => ({
					...state,
					error: errorMessage,
					isLoading: false
				}));
				throw error;
			}
		},

		async stopImpersonation(): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				await adminApi.stopImpersonation();

				// Clear offline data when stopping impersonation
				await clearOfflineData();

				window.location.href = '/admin';
			} catch (error) {
				const errorMessage =
					error instanceof ApiError
						? error.message
						: 'Failed to stop impersonation';
				update((state) => ({
					...state,
					error: errorMessage,
					isLoading: false
				}));
				throw error;
			}
		}
	};
}

export const authStore = createAuthStore();
