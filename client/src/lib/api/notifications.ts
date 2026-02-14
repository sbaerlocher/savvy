import { api } from './client';
import type {
	NotificationDTO,
	NotificationListResponse,
	NotificationUnreadCountResponse
} from '$lib/types/api';

export const notificationsApi = {
	/**
	 * Get all notifications for the current user
	 */
	async list(limit = 50, offset = 0): Promise<NotificationListResponse> {
		return api.get<NotificationListResponse>(
			`/notifications?limit=${limit}&offset=${offset}`
		);
	},

	/**
	 * Get count of unread notifications
	 */
	async getUnreadCount(): Promise<NotificationUnreadCountResponse> {
		return api.get<NotificationUnreadCountResponse>(
			'/notifications/unread-count'
		);
	},

	/**
	 * Mark a notification as read
	 */
	async markAsRead(notificationId: string): Promise<{ message: string }> {
		return api.post<{ message: string }>(
			`/notifications/${notificationId}/read`
		);
	},

	/**
	 * Mark all notifications as read
	 */
	async markAllAsRead(): Promise<{ message: string }> {
		return api.post<{ message: string }>('/notifications/read-all');
	},

	/**
	 * Delete a notification
	 */
	async delete(notificationId: string): Promise<{ message: string }> {
		return api.delete<{ message: string }>(`/notifications/${notificationId}`);
	}
};
