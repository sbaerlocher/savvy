import { api, type ApiOptions } from './client';

export interface ProfileDTO {
	id: string;
	email: string;
	first_name: string;
	last_name: string;
	language: string;
	auth_provider: string;
	push_notifications_enabled: boolean;
	email_notifications_enabled: boolean;
	push_reminders_enabled: boolean;
	push_sharing_enabled: boolean;
	email_reminders_enabled: boolean;
	email_sharing_enabled: boolean;
	email_verified: boolean;
	email_verified_at?: string;
	created_at: string;
	updated_at: string;
}

export const profileApi = {
	/**
	 * Get current user's profile
	 */
	async get(options?: ApiOptions): Promise<{ profile: ProfileDTO }> {
		return api.get<{ profile: ProfileDTO }>('/profile', options);
	},

	/**
	 * Update profile (name, language)
	 */
	async update(
		data: {
			first_name?: string;
			last_name?: string;
			language?: string;
			push_notifications_enabled?: boolean;
			email_notifications_enabled?: boolean;
			push_reminders_enabled?: boolean;
			push_sharing_enabled?: boolean;
			email_reminders_enabled?: boolean;
			email_sharing_enabled?: boolean;
		},
		options?: ApiOptions
	): Promise<{ profile: ProfileDTO }> {
		return api.patch<{ profile: ProfileDTO }>('/profile', data, options);
	},

	/**
	 * Change password (local auth only)
	 */
	async changePassword(
		data: { current_password: string; new_password: string },
		options?: ApiOptions
	): Promise<{ message: string }> {
		return api.post<{ message: string }>(
			'/profile/change-password',
			data,
			options
		);
	},

	/**
	 * Delete account permanently
	 */
	async deleteAccount(
		data: { password: string; confirmation: string },
		options?: ApiOptions
	): Promise<{ message: string }> {
		return api.post<{ message: string }>(
			'/profile/delete-account',
			data,
			options
		);
	}
};
