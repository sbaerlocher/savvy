<script lang="ts">
	import {
		ICON_BELL,
		ICON_CHECK_CIRCLE,
		ICON_CLOSE,
		ICON_SHARE,
		ICON_TRANSFER
	} from '$lib/icons';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { t } from '$lib/stores/i18n';
	import { notificationStore } from '$lib/stores/notifications';
	import type { NotificationDTO } from '$lib/types/api';
	import { resourceDetailPath } from '$lib/resource/routes';
	import { onDestroy, onMount } from 'svelte';
	import { platform } from '$lib/utils/platform';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	// triggerClass lets callers restyle the bell button (mobile header uses a
	// 40px boxed variant to match the mockup); defaults to the desktop look.
	// mode: 'panel' (desktop) opens the dropdown; 'link' (mobile) navigates
	// straight to the notifications page instead — the full screen reads better
	// on small viewports than a wide dropdown.
	let {
		triggerClass = 'notification-bell relative inline-flex items-center justify-center p-1 text-text-muted transition-colors hover:text-text-strong',
		iconClass = 'w-6 h-6',
		mode = 'panel'
	}: {
		triggerClass?: string;
		iconClass?: string;
		mode?: 'panel' | 'link';
	} = $props();

	const notifications = $derived($notificationStore.notifications);
	const unreadCount = $derived($notificationStore.unreadCount);
	const isOpen = $derived($notificationStore.isOpen);
	const isLoading = $derived($notificationStore.isLoading);

	onMount(() => {
		notificationStore.startPolling();
	});

	onDestroy(() => {
		notificationStore.stopPolling();
	});

	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (
			isOpen &&
			!target.closest('.notification-panel') &&
			!target.closest('.notification-bell')
		) {
			notificationStore.closePanel();
		}
	}

	function formatNotificationMessage(notification: NotificationDTO): string {
		const fromUser = notification.metadata.from_user_name || 'Someone';
		const resourceTypeLabel =
			$t(`notifications.resourceWithArticle.${notification.resource_type}`) ||
			$t(`common.${notification.resource_type}`) ||
			notification.resource_type;

		if (notification.type === 'share_received') {
			const merchant = (notification.metadata.merchant_name as string) || '';
			return merchant
				? `${fromUser} ${$t('notifications.sharedWith', { resource: resourceTypeLabel, merchant })}`
				: `${fromUser} ${$t('notifications.sharedWithPlain', { resource: resourceTypeLabel })}`;
		} else if (notification.type === 'transfer_received') {
			const merchant = (notification.metadata.merchant_name as string) || '';
			return merchant
				? `${fromUser} ${$t('notifications.transferredTo', { resource: resourceTypeLabel, merchant })}`
				: `${fromUser} ${$t('notifications.transferredToPlain', { resource: resourceTypeLabel })}`;
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

	async function handleNotificationClick(notification: NotificationDTO) {
		if (!notification.is_read) {
			await notificationStore.markAsRead(notification.id);
		}
		notificationStore.closePanel();

		const path = resourceDetailPath(
			notification.resource_type,
			notification.resource_id
		);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- resourceDetailPath() already returns a resolve()'d path
		goto(path);
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
</script>

<svelte:window onclick={handleClickOutside} />

<!-- Bell content: icon + unread badge, shared by both trigger variants. -->
{#snippet bellContent()}
	<svg class={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path
			stroke-linecap="round"
			stroke-linejoin="round"
			stroke-width="2"
			d={ICON_BELL}
		/>
	</svg>

	{#if unreadCount > 0}
		<span
			class="absolute -top-1.5 -right-1.5 inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 text-[length:var(--text-tag)] font-bold leading-none text-white bg-danger-600 rounded-full ring-2 ring-white"
		>
			{unreadCount > 99 ? '99+' : unreadCount}
		</span>
	{/if}
{/snippet}

<!-- Notification Bell -->
{#if mode === 'link'}
	<!-- eslint-disable svelte/no-navigation-without-resolve -- href from resolve() -->
	<a
		href={resolve('/notifications')}
		class={triggerClass}
		aria-label={$t('notifications.title')}
	>
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
		{@render bellContent()}
	</a>
{:else}
	<button
		class={triggerClass}
		onclick={() => notificationStore.togglePanel()}
		aria-label={$t('notifications.title')}
	>
		{@render bellContent()}
	</button>
{/if}

<!-- Notification Panel -->
{#if isOpen}
	<div
		class="notification-panel fixed left-4 right-4 sm:absolute sm:left-auto sm:right-0 mt-2 sm:w-96 z-50 max-w-[90vw] overflow-hidden {platform ===
		'ios'
			? 'liquid-glass-menu rounded-lg'
			: platform === 'android'
				? 'bg-m3-surface-container rounded-m3-lg shadow-m3-dialog'
				: 'bg-white border border-border shadow-xl rounded-lg'}"
	>
		<!-- Header -->
		<div
			class="flex items-center justify-between px-4 py-3 border-b {platform ===
			'ios'
				? 'border-[var(--color-glass-menu-edge)]'
				: 'border-border'}"
		>
			<h3 class="text-lg font-semibold text-text">
				{$t('notifications.title')}
			</h3>

			<div class="flex items-center gap-1">
				{#if notifications.length > 0}
					<button
						class="text-accent hover:text-accent-hover p-1.5 rounded-md hover:bg-accent-50 transition-colors"
						onclick={() => notificationStore.markAllAsRead()}
						title={$t('notifications.markAllAsRead')}
						aria-label={$t('notifications.markAllAsRead')}
					>
						<svg
							class="w-5 h-5"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M2 13l4 4L15 8"
							/>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 13l4 4L22 8"
							/>
						</svg>
					</button>
				{/if}
				<button
					class="text-accent hover:text-accent-hover p-1.5 rounded-md hover:bg-accent-50 transition-colors"
					onclick={() => {
						notificationStore.closePanel();
						goto(resolve('/notifications'));
					}}
					title={$t('notifications.viewAll')}
					aria-label={$t('notifications.viewAll')}
				>
					<svg
						class="w-5 h-5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
						/>
					</svg>
				</button>
			</div>
		</div>

		<!-- Notification List -->
		<div class="max-h-96 overflow-y-auto">
			{#if isLoading}
				<LoadingSpinner />
			{:else if notifications.length === 0}
				<div class="px-4 py-8 text-center text-text-subtle">
					<svg
						class="w-12 h-12 mx-auto mb-2 text-text-placeholder"
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
					<p>{$t('notifications.noNotifications')}</p>
				</div>
			{:else}
				{#each notifications as notification (notification.id)}
					<div
						class="flex items-start px-4 py-3 hover:bg-surface-1 cursor-pointer border-b border-border-soft last:border-b-0 transition-colors {notification.is_read
							? 'opacity-60'
							: 'bg-accent-50'}"
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
									class="w-5 h-5 text-accent"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d={ICON_SHARE}
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
										d={ICON_TRANSFER}
									/>
								</svg>
							{:else if notification.type === 'expiry_reminder'}
								<svg
									class="w-5 h-5 text-warning-500"
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
									class="w-5 h-5 text-success-500"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d={ICON_CHECK_CIRCLE}
									/>
								</svg>
							{:else}
								<svg
									class="w-5 h-5 text-text-faint"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d={ICON_BELL}
									/>
								</svg>
							{/if}
						</div>

						<!-- Content -->
						<div class="flex-1 min-w-0">
							<p class="text-sm text-text">
								{formatNotificationMessage(notification)}
							</p>
							<p class="text-xs text-text-subtle mt-1">
								{formatTimeAgo(notification.created_at)}
							</p>
						</div>

						<!-- Actions -->
						<div class="shrink-0 ml-2 flex items-center gap-1">
							{#if !notification.is_read}
								<button
									class="text-accent hover:text-accent-hover p-1.5 rounded-md hover:bg-accent-100 transition-colors"
									onclick={(e) => {
										e.stopPropagation();
										notificationStore.markAsRead(notification.id);
									}}
									aria-label={$t('notifications.markAsRead')}
								>
									<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
										<circle cx="12" cy="12" r="5" />
									</svg>
								</button>
							{/if}

							<button
								class="text-text-faint hover:text-danger-600 p-1.5 rounded-md hover:bg-danger-50 transition-colors"
								onclick={(e) => {
									e.stopPropagation();
									notificationStore.delete(notification.id);
								}}
								aria-label={$t('notifications.delete')}
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
										d={ICON_CLOSE}
									/>
								</svg>
							</button>
						</div>
					</div>
				{/each}
			{/if}
		</div>
	</div>
{/if}
