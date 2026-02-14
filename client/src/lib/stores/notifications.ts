import { writable, derived, get } from 'svelte/store';
import type { NotificationDTO } from '$lib/types/api';
import { notificationsApi } from '$lib/api/notifications';
import { toastStore } from './toast';
import { t } from './i18n';
import { logger } from '$lib/utils/logger';

const tr = (key: string, params?: Record<string, string | number>) =>
	get(t)(key, params);

const notifLogger = logger.child('Notifications');

export interface NotificationState {
	notifications: NotificationDTO[];
	unreadCount: number;
	isLoading: boolean;
	isOpen: boolean;
}

function createNotificationStore() {
	const { subscribe, set, update } = writable<NotificationState>({
		notifications: [],
		unreadCount: 0,
		isLoading: false,
		isOpen: false
	});

	let pollInterval: ReturnType<typeof setInterval> | null = null;

	return {
		subscribe,

		async load(): Promise<void> {
			notifLogger.debug('Loading notifications...');
			update((state) => ({ ...state, isLoading: true }));

			try {
				const [notificationsRes, unreadRes] = await Promise.all([
					notificationsApi.list(50, 0),
					notificationsApi.getUnreadCount()
				]);

				notifLogger.info('Notifications loaded:', {
					count: notificationsRes.notifications.length,
					unread: unreadRes.count,
					notifications: notificationsRes.notifications
				});

				update((state) => ({
					...state,
					notifications: notificationsRes.notifications,
					unreadCount: unreadRes.count,
					isLoading: false
				}));
			} catch (error) {
				notifLogger.error('Failed to load notifications:', error);
				update((state) => ({ ...state, isLoading: false }));
			}
		},

		async refreshUnreadCount(): Promise<void> {
			try {
				const res = await notificationsApi.getUnreadCount();
				update((state) => ({ ...state, unreadCount: res.count }));
			} catch (error) {
				notifLogger.error('Failed to refresh unread count:', error);
			}
		},

		async markAsRead(notificationId: string): Promise<void> {
			try {
				await notificationsApi.markAsRead(notificationId);

				update((state) => ({
					...state,
					notifications: state.notifications.map((n) =>
						n.id === notificationId
							? { ...n, is_read: true, read_at: new Date().toISOString() }
							: n
					),
					unreadCount: Math.max(0, state.unreadCount - 1)
				}));
			} catch (error) {
				notifLogger.error('Failed to mark notification as read:', error);
				toastStore.error(tr('notifications.toasts.markReadError'));
			}
		},

		async markAllAsRead(): Promise<void> {
			try {
				await notificationsApi.markAllAsRead();

				update((state) => ({
					...state,
					notifications: state.notifications.map((n) => ({
						...n,
						is_read: true,
						read_at: n.read_at || new Date().toISOString()
					})),
					unreadCount: 0
				}));

				toastStore.success(tr('notifications.toasts.markAllReadSuccess'));
			} catch (error) {
				notifLogger.error('Failed to mark all as read:', error);
				toastStore.error(tr('notifications.toasts.markAllReadError'));
			}
		},

		async delete(notificationId: string): Promise<void> {
			try {
				await notificationsApi.delete(notificationId);

				update((state) => {
					const deletedNotification = state.notifications.find(
						(n) => n.id === notificationId
					);
					const wasUnread = deletedNotification && !deletedNotification.is_read;

					return {
						...state,
						notifications: state.notifications.filter(
							(n) => n.id !== notificationId
						),
						unreadCount: wasUnread
							? Math.max(0, state.unreadCount - 1)
							: state.unreadCount
					};
				});

				toastStore.success(tr('notifications.toasts.deleteSuccess'));
			} catch (error) {
				notifLogger.error('Failed to delete notification:', error);
				toastStore.error(tr('notifications.toasts.deleteError'));
			}
		},

		togglePanel(): void {
			update((state) => ({ ...state, isOpen: !state.isOpen }));
		},

		closePanel(): void {
			update((state) => ({ ...state, isOpen: false }));
		},

		startPolling(intervalMs = 300000): void {
			notifLogger.info(
				'Starting notification polling (interval:',
				intervalMs / 1000,
				'seconds)'
			);
			if (pollInterval) {
				clearInterval(pollInterval);
			}

			this.load();

			pollInterval = setInterval(() => {
				notifLogger.debug('Polling for notification updates...');
				this.refreshUnreadCount();
			}, intervalMs);
		},

		stopPolling(): void {
			if (pollInterval) {
				clearInterval(pollInterval);
				pollInterval = null;
			}
		},

		reset(): void {
			this.stopPolling();
			set({
				notifications: [],
				unreadCount: 0,
				isLoading: false,
				isOpen: false
			});
		}
	};
}

export const notificationStore = createNotificationStore();

export const unreadNotifications = derived(notificationStore, ($store) =>
	$store.notifications.filter((n) => !n.is_read)
);
