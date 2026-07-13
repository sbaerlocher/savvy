<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { profileApi, type ProfileDTO } from '$lib/api';
	import { notificationsApi } from '$lib/api/notifications';
	import NotificationsSection from '$lib/components/settings/NotificationsSection.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { authStore } from '$lib/stores/auth';
	import { notificationStore } from '$lib/stores/notifications';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import type { NotificationDTO } from '$lib/types/api';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const PAGE_SIZE = 20;
	const pageLogger = logger.child('NotificationsPage');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	let profile = $state<ProfileDTO | null>(null);
	let isLoadingProfile = $state(true);
	let notifications = $state<NotificationDTO[]>([]);
	let isLoadingNotifications = $state(true);
	let isLoadingMore = $state(false);
	let offset = $state(0);
	let hasMore = $state(true);

	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}
		await Promise.all([loadProfile(), loadNotifications()]);
	});

	async function loadProfile() {
		try {
			const response = await profileApi.get();
			profile = response.profile;
		} catch (error) {
			pageLogger.error('Failed to load profile', { error });
			toastStore.error(tr('common.error'));
		} finally {
			isLoadingProfile = false;
		}
	}

	async function loadNotifications() {
		try {
			const response = await notificationsApi.list(PAGE_SIZE, 0);
			notifications = response.notifications || [];
			offset = notifications.length;
			hasMore = notifications.length >= PAGE_SIZE;
		} catch (error) {
			pageLogger.error('Failed to load notifications', { error });
		} finally {
			isLoadingNotifications = false;
		}
	}

	async function loadMore() {
		if (isLoadingMore) return;
		isLoadingMore = true;
		try {
			const response = await notificationsApi.list(PAGE_SIZE, offset);
			const newItems = response.notifications || [];
			notifications = [...notifications, ...newItems];
			offset += newItems.length;
			hasMore = newItems.length >= PAGE_SIZE;
		} catch (error) {
			pageLogger.error('Failed to load more notifications', { error });
		} finally {
			isLoadingMore = false;
		}
	}

	function formatNotificationMessage(notification: NotificationDTO): string {
		const fromUser = notification.metadata.from_user_name || 'Someone';
		const resourceTypeLabel =
			$t(`notifications.resourceWithArticle.${notification.resource_type}`) ||
			$t(`common.${notification.resource_type}`) ||
			notification.resource_type;

		if (notification.type === 'share_received') {
			return `${fromUser} ${$t('notifications.sharedWith', { resource: resourceTypeLabel })}`;
		} else if (notification.type === 'transfer_received') {
			return `${fromUser} ${$t('notifications.transferredTo', { resource: resourceTypeLabel })}`;
		} else if (notification.type === 'expiry_reminder') {
			const merchantName =
				(notification.metadata.merchant_name as string) || '';
			const daysLeft = notification.metadata.days_left as number;
			return $t('notifications.expiryReminder', {
				resource: resourceTypeLabel,
				merchant: merchantName,
				days: String(daysLeft)
			});
		} else if (notification.type === 'validity_start') {
			const merchantName =
				(notification.metadata.merchant_name as string) || '';
			return $t('notifications.validityStart', {
				merchant: merchantName
			});
		}

		return $t('notifications.newNotification');
	}

	function getNotificationLink(notification: NotificationDTO) {
		const id = notification.resource_id;
		switch (notification.resource_type) {
			case 'voucher':
				return `/vouchers/${id}` as const;
			case 'gift_card':
				return `/gift-cards/${id}` as const;
			default:
				return `/cards/${id}` as const;
		}
	}

	async function handleNotificationClick(notification: NotificationDTO) {
		if (!notification.is_read) {
			await handleMarkAsRead(notification.id);
		}
		goto(resolve(getNotificationLink(notification)));
	}

	function formatTimeAgo(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

		if (seconds < 60) return $t('notifications.timeAgo.justNow');
		if (seconds < 3600)
			return $t('notifications.timeAgo.minutesAgo', {
				count: Math.floor(seconds / 60)
			});
		if (seconds < 86400)
			return $t('notifications.timeAgo.hoursAgo', {
				count: Math.floor(seconds / 3600)
			});
		if (seconds < 604800)
			return $t('notifications.timeAgo.daysAgo', {
				count: Math.floor(seconds / 86400)
			});
		return date.toLocaleDateString();
	}

	async function handleMarkAsRead(id: string) {
		try {
			await notificationsApi.markAsRead(id);
			notifications = notifications.map((n) =>
				n.id === id ? { ...n, is_read: true } : n
			);
			notificationStore.refreshUnreadCount();
		} catch {
			toastStore.error(tr('notifications.toasts.markReadError'));
		}
	}

	async function handleDelete(id: string) {
		try {
			await notificationsApi.delete(id);
			notifications = notifications.filter((n) => n.id !== id);
			notificationStore.refreshUnreadCount();
			toastStore.success(tr('notifications.toasts.deleteSuccess'));
		} catch {
			toastStore.error(tr('notifications.toasts.deleteError'));
		}
	}

	async function handleMarkAllAsRead() {
		try {
			await notificationsApi.markAllAsRead();
			notifications = notifications.map((n) => ({ ...n, is_read: true }));
			notificationStore.refreshUnreadCount();
			toastStore.success(tr('notifications.toasts.markAllReadSuccess'));
		} catch {
			toastStore.error(tr('notifications.toasts.markAllReadError'));
		}
	}

	function handleProfileUpdated(updatedProfile: ProfileDTO) {
		profile = updatedProfile;
		authStore.checkAuth();
	}
</script>

<svelte:head>
	<title>{tr('notifications.title')} - {tr('common.appName')}</title>
</svelte:head>

<div class="px-4 max-w-7xl mx-auto">
	<PageHeader title={tr('notifications.title')} />

	{#if isLoadingProfile || isLoadingNotifications}
		<LoadingSpinner />
	{:else}
		<div class="flex flex-col lg:flex-row gap-6 items-start">
			<!-- Settings (on mobile: first, on desktop: right side) -->
			<div class="w-full lg:w-1/3 lg:order-2">
				{#if profile}
					<NotificationsSection
						{profile}
						onProfileUpdated={handleProfileUpdated}
					/>
				{/if}
			</div>

			<!-- Notification List (on mobile: second, on desktop: left side) -->
			<div class="w-full lg:w-2/3 lg:order-1">
				<div class="bg-white rounded-xl shadow-lg overflow-hidden">
					<!-- Header -->
					<div
						class="flex items-center justify-between px-6 py-4 border-b border-gray-200"
					>
						<h2 class="text-lg font-semibold text-gray-900">
							{tr('notifications.title')}
						</h2>
						{#if notifications.some((n) => !n.is_read)}
							<button
								class="text-sm text-cyan-600 hover:text-cyan-700"
								onclick={handleMarkAllAsRead}
							>
								{tr('notifications.markAllAsRead')}
							</button>
						{/if}
					</div>

					<!-- Notification Items -->
					{#if notifications.length === 0}
						<div class="px-4 py-12 text-center text-gray-500">
							<svg
								class="w-12 h-12 mx-auto mb-2 text-gray-300"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="1.5"
									d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
								/>
							</svg>
							<p>{tr('notifications.noNotifications')}</p>
						</div>
					{:else}
						{#each notifications as notification (notification.id)}
							<div
								class="flex items-start px-4 py-3 hover:bg-gray-50 cursor-pointer border-b border-gray-100 last:border-b-0 transition-colors {notification.is_read
									? 'opacity-60'
									: 'bg-cyan-50'}"
								onclick={() => handleNotificationClick(notification)}
								onkeydown={(e) =>
									e.key === 'Enter' && handleNotificationClick(notification)}
								role="button"
								tabindex="0"
							>
								<!-- Icon -->
								<div class="shrink-0 mr-3 mt-0.5">
									{#if notification.type === 'share_received'}
										<svg
											class="w-5 h-5 text-cyan-500"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
											/>
										</svg>
									{:else if notification.type === 'transfer_received'}
										<svg
											class="w-5 h-5 text-purple-500"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
											/>
										</svg>
									{:else if notification.type === 'expiry_reminder'}
										<svg
											class="w-5 h-5 text-amber-500"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
											/>
										</svg>
									{:else if notification.type === 'validity_start'}
										<svg
											class="w-5 h-5 text-emerald-500"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
											/>
										</svg>
									{:else}
										<svg
											class="w-5 h-5 text-gray-400"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
											/>
										</svg>
									{/if}
								</div>

								<!-- Content -->
								<div class="flex-1 min-w-0">
									<p class="text-sm text-gray-900">
										{formatNotificationMessage(notification)}
									</p>
									<p class="text-xs text-gray-500 mt-1">
										{formatTimeAgo(notification.created_at)}
									</p>
								</div>

								<!-- Actions -->
								<div class="shrink-0 ml-2 flex items-center gap-1">
									{#if !notification.is_read}
										<button
											class="text-cyan-500 hover:text-cyan-700 p-1.5 rounded-md hover:bg-cyan-100 transition-colors"
											onclick={(e) => {
												e.stopPropagation();
												handleMarkAsRead(notification.id);
											}}
											aria-label={tr('notifications.markAsRead')}
										>
											<svg
												class="w-4 h-4"
												fill="currentColor"
												viewBox="0 0 24 24"
											>
												<circle cx="12" cy="12" r="5" />
											</svg>
										</button>
									{/if}

									<button
										class="text-gray-400 hover:text-red-600 p-1.5 rounded-md hover:bg-red-50 transition-colors"
										onclick={(e) => {
											e.stopPropagation();
											handleDelete(notification.id);
										}}
										aria-label={tr('notifications.delete')}
									>
										<svg
											class="w-4 h-4"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M6 18L18 6M6 6l12 12"
											/>
										</svg>
									</button>
								</div>
							</div>
						{/each}
					{/if}

					<!-- Load More -->
					{#if hasMore && notifications.length > 0}
						<div class="px-6 py-4 border-t border-gray-100 text-center">
							<button
								class="text-sm text-cyan-600 hover:text-cyan-700 font-medium disabled:opacity-50"
								onclick={loadMore}
								disabled={isLoadingMore}
							>
								{isLoadingMore
									? tr('notifications.loadingMore')
									: tr('notifications.loadMore')}
							</button>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
