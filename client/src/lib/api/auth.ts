import { api, type ApiOptions } from './client';
import type { UserDTO, LoginRequest, RegisterRequest } from '$lib/types/api';

export const authApi = {
	/**
	 * Login with email and password
	 */
	async login(
		credentials: LoginRequest,
		options?: ApiOptions
	): Promise<{ user: UserDTO } | { requires_2fa: true }> {
		return api.post<{ user: UserDTO } | { requires_2fa: true }>(
			'/auth/login',
			credentials,
			options
		);
	},

	/**
	 * Register a new user
	 */
	async register(
		data: RegisterRequest,
		options?: ApiOptions
	): Promise<{ user: UserDTO }> {
		return api.post<{ user: UserDTO }>('/auth/register', data, options);
	},

	/**
	 * Logout current user
	 */
	async logout(options?: ApiOptions): Promise<{ message: string }> {
		return api.post<{ message: string }>('/auth/logout', undefined, options);
	},

	/**
	 * Get current user info
	 */
	async me(options?: ApiOptions): Promise<{ user: UserDTO }> {
		return api.get<{ user: UserDTO }>('/auth/me', options);
	},

	/**
	 * Request a new verification email
	 */
	async requestVerification(
		options?: ApiOptions
	): Promise<{ message: string; email_verified?: boolean }> {
		return api.post<{ message: string; email_verified?: boolean }>(
			'/auth/request-verification',
			undefined,
			options
		);
	},

	/**
	 * Verify email with token
	 */
	async verifyEmail(
		token: string,
		options?: ApiOptions
	): Promise<{ message: string; email_verified: boolean }> {
		return api.post<{ message: string; email_verified: boolean }>(
			'/auth/verify-email',
			{ token },
			options
		);
	},

	/**
	 * Unsubscribe from notification emails via one-click token
	 */
	async unsubscribeNotifications(
		token: string,
		options?: ApiOptions
	): Promise<{ message: string; email_sharing_enabled: boolean }> {
		return api.post<{ message: string; email_sharing_enabled: boolean }>(
			'/auth/unsubscribe-notifications',
			{ token },
			options
		);
	},

	/**
	 * Unsubscribe from expiry reminder emails via one-click token
	 */
	async unsubscribeReminders(
		token: string,
		options?: ApiOptions
	): Promise<{ message: string; email_reminders_enabled: boolean }> {
		return api.post<{ message: string; email_reminders_enabled: boolean }>(
			'/auth/unsubscribe-reminders',
			{ token },
			options
		);
	},

	/**
	 * Request a password reset email
	 */
	async forgotPassword(
		email: string,
		options?: ApiOptions
	): Promise<{ message: string }> {
		return api.post<{ message: string }>(
			'/auth/forgot-password',
			{ email },
			options
		);
	},

	/**
	 * Reset password with token
	 */
	async resetPassword(
		token: string,
		password: string,
		options?: ApiOptions
	): Promise<{ message: string }> {
		return api.post<{ message: string }>(
			'/auth/reset-password',
			{ token, password },
			options
		);
	},

	/**
	 * Verify 2FA challenge (login step 2)
	 */
	async verify2FA(
		data: { code?: string; backup_code?: string },
		options?: ApiOptions
	): Promise<{ message: string; authenticated: boolean }> {
		return api.post<{ message: string; authenticated: boolean }>(
			'/auth/2fa/challenge',
			data,
			options
		);
	},

	/**
	 * Get 2FA status
	 */
	async get2FAStatus(
		options?: ApiOptions
	): Promise<{ enabled: boolean; is_local_auth: boolean }> {
		return api.get<{ enabled: boolean; is_local_auth: boolean }>(
			'/auth/2fa/status',
			options
		);
	},

	/**
	 * Start 2FA setup
	 */
	async setup2FA(options?: ApiOptions): Promise<{
		secret: string;
		qr_code_url: string;
		backup_codes: string[];
	}> {
		return api.post<{
			secret: string;
			qr_code_url: string;
			backup_codes: string[];
		}>('/auth/2fa/setup', undefined, options);
	},

	/**
	 * Verify and enable 2FA
	 */
	async verify2FASetup(
		code: string,
		options?: ApiOptions
	): Promise<{ message: string; enabled: boolean }> {
		return api.post<{ message: string; enabled: boolean }>(
			'/auth/2fa/verify',
			{ code },
			options
		);
	},

	/**
	 * Disable 2FA
	 */
	async disable2FA(
		code: string,
		options?: ApiOptions
	): Promise<{ message: string; enabled: boolean }> {
		return api.post<{ message: string; enabled: boolean }>(
			'/auth/2fa/disable',
			{ code },
			options
		);
	},

	/**
	 * Regenerate 2FA backup codes
	 */
	async regenerateBackupCodes(
		code: string,
		options?: ApiOptions
	): Promise<{ backup_codes: string[] }> {
		return api.post<{ backup_codes: string[] }>(
			'/auth/2fa/backup-codes',
			{ code },
			options
		);
	}
};
