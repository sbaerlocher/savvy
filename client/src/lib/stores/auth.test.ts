import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock API BEFORE imports
vi.mock('$lib/api', () => ({
	authApi: {
		me: vi.fn(),
		login: vi.fn(),
		register: vi.fn(),
		logout: vi.fn()
	},
	adminApi: {
		startImpersonation: vi.fn(),
		stopImpersonation: vi.fn()
	},
	ApiError: class ApiError extends Error {
		constructor(
			message: string,
			public status: number
		) {
			super(message);
		}
	}
}));

// Now import after mocks are set up
import { get } from 'svelte/store';
import { authStore } from './auth';
import { authApi } from '$lib/api';
import type { UserDTO } from '$lib/types/api';

describe('authStore', () => {
	const mockUser: UserDTO = {
		id: 'user-123',
		email: 'test@example.com',
		first_name: 'Test',
		last_name: 'User',
		is_admin: false,
		is_impersonating: false
	};

	beforeEach(() => {
		// Clear localStorage before each test
		localStorage.clear();
		vi.clearAllMocks();
	});

	describe('Security: localStorage (SVL-003)', () => {
		it('should NOT store user data in localStorage after login', async () => {
			// Mock successful login
			vi.mocked(authApi.login).mockResolvedValue({ user: mockUser });

			// Perform login
			await authStore.login({
				email: 'test@example.com',
				password: 'password'
			});

			// Check localStorage
			const stored = localStorage.getItem('savvy_auth_state');
			expect(stored).toBeTruthy();

			const parsed = JSON.parse(stored!);

			// ✅ SECURITY CHECK: User data MUST NOT be in localStorage
			expect(parsed.user).toBeUndefined();
			expect(parsed.email).toBeUndefined();
			expect(parsed.name).toBeUndefined();

			// ✅ Only isAuthenticated flag should be stored
			expect(parsed.isAuthenticated).toBe(true);
		});

		it('should NOT restore user data from localStorage on load', () => {
			// Simulate malicious localStorage injection (XSS attack)
			localStorage.setItem(
				'savvy_auth_state',
				JSON.stringify({
					isAuthenticated: true,
					user: {
						id: 'hacker-123',
						email: 'hacker@evil.com',
						is_admin: true // ❌ Attacker tries to escalate privileges
					}
				})
			);

			// Load auth from storage
			const cachedAuth = authStore.getCachedAuth();

			// ✅ SECURITY CHECK: User data MUST be null (not from localStorage)
			expect(cachedAuth?.user).toBeNull();
			expect(cachedAuth?.isAuthenticated).toBe(true); // Flag is OK
		});

		it('should only load isAuthenticated flag from localStorage', () => {
			localStorage.setItem(
				'savvy_auth_state',
				JSON.stringify({ isAuthenticated: true })
			);

			const cachedAuth = authStore.getCachedAuth();

			expect(cachedAuth?.isAuthenticated).toBe(true);
			expect(cachedAuth?.user).toBeNull();
		});
	});

	describe('Authentication Flow', () => {
		it('should successfully login and update store', async () => {
			vi.mocked(authApi.login).mockResolvedValue({ user: mockUser });

			await authStore.login({
				email: 'test@example.com',
				password: 'password'
			});

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(true);
			expect(state.user).toEqual(mockUser);
			expect(state.isLoading).toBe(false);
			expect(state.error).toBeNull();
		});

		it('should handle login failure', async () => {
			const error = new Error('Invalid credentials');
			vi.mocked(authApi.login).mockRejectedValue(error);

			await expect(
				authStore.login({ email: 'test@example.com', password: 'wrong' })
			).rejects.toThrow();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.user).toBeNull();
			expect(state.error).toBeTruthy();
		});

		it('should successfully logout and clear state', async () => {
			// Set initial state
			authStore.setUser(mockUser);
			expect(get(authStore).isAuthenticated).toBe(true);

			// Mock logout
			vi.mocked(authApi.logout).mockResolvedValue({ message: 'Logged out' });

			// Perform logout
			await authStore.logout();

			// Check state
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.user).toBeNull();
			expect(state.error).toBeNull();

			// Check localStorage
			const stored = localStorage.getItem('savvy_auth_state');
			expect(stored).toBeNull();
		});
	});

	describe('Offline Handling', () => {
		it('should keep cached auth when offline during checkAuth', async () => {
			// Set initial authenticated state
			authStore.setUser(mockUser);
			expect(get(authStore).isAuthenticated).toBe(true);

			// Mock offline state
			Object.defineProperty(navigator, 'onLine', {
				writable: true,
				value: false
			});

			// Mock API failure
			vi.mocked(authApi.me).mockRejectedValue(new Error('Network error'));

			// Check auth (should keep cached state)
			await authStore.checkAuth();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(true); // Should remain true (offline)
			expect(state.isLoading).toBe(false);
		});

		it('should clear auth when online but session invalid', async () => {
			// Set initial authenticated state
			authStore.setUser(mockUser);

			// Mock online state
			Object.defineProperty(navigator, 'onLine', {
				writable: true,
				value: true
			});

			// Mock API failure (session expired)
			vi.mocked(authApi.me).mockRejectedValue(new Error('Unauthorized'));

			// Check auth (should clear state)
			await authStore.checkAuth();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.user).toBeNull();
		});
	});

	describe('User Registration', () => {
		it('should successfully register and authenticate user', async () => {
			vi.mocked(authApi.register).mockResolvedValue({ user: mockUser });

			await authStore.register({
				email: 'new@example.com',
				password: 'password',
				first_name: 'New',
				last_name: 'User'
			});

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(true);
			expect(state.user).toEqual(mockUser);
		});

		it('should handle registration failure', async () => {
			const error = new Error('Email already exists');
			vi.mocked(authApi.register).mockRejectedValue(error);

			await expect(
				authStore.register({
					email: 'existing@example.com',
					password: 'password',
					first_name: 'User',
					last_name: 'Name'
				})
			).rejects.toThrow();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.error).toBeTruthy();
		});
	});

	describe('Error Handling', () => {
		it('should clear error', async () => {
			// Mock login failure to set error
			const error = new Error('Invalid credentials');
			vi.mocked(authApi.login).mockRejectedValue(error);

			try {
				await authStore.login({ email: 'test@example.com', password: 'wrong' });
			} catch (e) {
				// Expected to throw
			}

			// Verify error is set
			let state = get(authStore);
			expect(state.error).toBeTruthy();

			// Clear error
			authStore.clearError();

			// Verify error is cleared
			state = get(authStore);
			expect(state.error).toBeNull();
		});
	});
});
