<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { notificationsApi } from '$lib/api/notifications';
	import NotificationTypeIcon from '$lib/components/NotificationTypeIcon.svelte';
	import SectionLabel from '$lib/components/ui/SectionLabel.svelte';
	import { notificationStore } from '$lib/stores/notifications';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { notificationTone } from '$lib/utils/notification-tone';
	import { platform } from '$lib/utils/platform';
	import type { NotificationDTO } from '$lib/types/api';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	const PAGE_SIZE = 20;
	const pageLogger = logger.child('NotificationsPage');
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Desktop renders the mockup layout (centred column, Neu/Früher sections,
	// boxed row actions). Android renders the M3 list, iOS the push screen with
	// grouped-inset sections and swipe row actions. `platform` is a module
	// constant, so these are plain consts, not $derived.
	const IS_DESKTOP = platform === 'other';
	const IS_ANDROID = platform === 'android';
	const IS_IOS = platform === 'ios';

	let notifications = $state<NotificationDTO[]>([]);
	let isLoadingNotifications = $state(true);
	let isLoadingMore = $state(false);
	let offset = $state(0);
	let hasMore = $state(true);

	const unreadNotifications = $derived(notifications.filter((n) => !n.is_read));
	const readNotifications = $derived(notifications.filter((n) => n.is_read));

	// Swipe-to-reveal for the iOS rows: unread rows expose mark-read plus delete
	// (2 x 70px), read rows only delete.
	const SWIPE_REVEAL = 140;
	const SWIPE_REVEAL_READ = 70;
	let swipedId = $state<string | null>(null);
	let touchStartX = 0;
	let touchStartY = 0;
	let touchAxis: 'none' | 'x' | 'y' = 'none';

	function onRowTouchStart(event: TouchEvent) {
		touchStartX = event.touches[0].clientX;
		touchStartY = event.touches[0].clientY;
		touchAxis = 'none';
	}

	function onRowTouchMove(event: TouchEvent, id: string) {
		const dx = event.touches[0].clientX - touchStartX;
		const dy = event.touches[0].clientY - touchStartY;
		// Lock the axis once, so a vertical scroll never opens the actions.
		if (touchAxis === 'none') {
			if (Math.abs(dx) < 10 && Math.abs(dy) < 10) return;
			touchAxis = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y';
		}
		if (touchAxis !== 'x') return;
		if (dx < -30) swipedId = id;
		else if (dx > 30) swipedId = null;
	}

	// Id of the row whose overflow menu is open (Android), or null.
	let openMenuId = $state<string | null>(null);

	// Any click outside the open overflow menu dismisses it (same pattern as
	// NotificationPanel and DesktopNav).
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (openMenuId && !target.closest('.notif-overflow')) openMenuId = null;
	}

	// Headline for the Android and iOS list rows (both mockups split headline and
	// detail across two lines). It names the kind of event; the detail line below
	// carries the actor and merchant, so the two must not repeat each other.
	function formatTypeTitle(notification: NotificationDTO): string {
		// $t() echoes the key back when it is missing, so compare against it
		// rather than relying on a falsy return for unknown notification types.
		const key = `notifications.typeTitle.${notification.type}`;
		const title = $t(key);
		return title === key ? $t('notifications.newNotification') : title;
	}

	onMount(async () => {
		// The auth guard lives in the page (it must also cover the desktop
		// skeleton, which does not mount this section).
		await loadNotifications();
		// The header count comes from the store (authoritative, server-side),
		// not from the paginated list, so make sure it is populated here.
		notificationStore.refreshUnreadCount();
	});

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

	function formatNotificationTitle(notification: NotificationDTO): string {
		const fromUser = notification.metadata.from_user_name || 'Someone';

		if (notification.type === 'share_received') {
			return `${fromUser} ${$t('notifications.sharedWithPlain', {
				resource: resourceLabel(notification)
			})}`;
		}
		if (notification.type === 'transfer_received') {
			return `${fromUser} ${$t('notifications.transferredToPlain', {
				resource: resourceLabel(notification)
			})}`;
		}
		return formatNotificationMessage(notification);
	}

	function resourceLabel(notification: NotificationDTO): string {
		return (
			$t(`notifications.resourceWithArticle.${notification.resource_type}`) ||
			$t(`common.${notification.resource_type}`) ||
			notification.resource_type
		);
	}

	// Desktop rows show a secondary line under the title. Share/transfer put the
	// merchant there; the other types have no separate detail, so the line is
	// omitted rather than duplicating the title.
	function formatNotificationDetail(notification: NotificationDTO): string {
		if (
			notification.type === 'share_received' ||
			notification.type === 'transfer_received'
		) {
			return (notification.metadata.merchant_name as string) || '';
		}
		return '';
	}

	function formatNotificationMessage(notification: NotificationDTO): string {
		const fromUser = notification.metadata.from_user_name || 'Someone';
		const resourceTypeLabel = resourceLabel(notification);

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

	// Called from the page's iOS actions snippet (part of the shell's title
	// row), which must mark the section's local rows read too.
	export async function handleMarkAllAsRead() {
		try {
			await notificationsApi.markAllAsRead();
			notifications = notifications.map((n) => ({ ...n, is_read: true }));
			notificationStore.refreshUnreadCount();
			toastStore.success(tr('notifications.toasts.markAllReadSuccess'));
		} catch {
			toastStore.error(tr('notifications.toasts.markAllReadError'));
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

{#snippet gearIcon(size: number)}
	<svg
		width={size}
		height={size}
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="1.9"
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
	>
		<circle cx="12" cy="12" r="3" />
		<path
			d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 11-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 008 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 11-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H2a2 2 0 010-4h.09A1.65 1.65 0 003.6 8a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 112.83-2.83l.06.06a1.65 1.65 0 001.82.33H8a1.65 1.65 0 001-1.51V2a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 112.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V8a1.65 1.65 0 001.51 1H22a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"
		/>
	</svg>
{/snippet}

{#snippet trashIcon()}
	<svg
		width="17"
		height="17"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
	>
		<path d="M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14" />
	</svg>
{/snippet}

<!-- Desktop row (mockup board B). `unread` drives dot, tint and title weight. -->
{#snippet desktopRow(notification: NotificationDTO, unread: boolean)}
	{@const tone = notificationTone(notification.type)}
	{@const detail = formatNotificationDetail(notification)}
	<div
		class="flex items-center gap-[15px] rounded-xl px-[18px] py-[15px] transition-colors {unread
			? 'border border-accent-100 bg-accent-50 hover:border-accent-200'
			: 'border border-border bg-surface hover:border-border-field'}"
	>
		<span
			class="size-[7px] flex-none rounded-full {unread
				? tone.dot
				: 'bg-transparent'}"
		></span>
		<button
			type="button"
			class="flex min-w-0 flex-1 items-center gap-[15px] text-left"
			onclick={() => handleNotificationClick(notification)}
		>
			<span
				class="flex size-11 flex-none items-center justify-center rounded-lg {tone.tile} {tone.ink} {unread
					? ''
					: 'opacity-70'}"
			>
				<NotificationTypeIcon type={notification.type} size={22} />
			</span>
			<span class="min-w-0 flex-1">
				<span
					class="block truncate text-[length:var(--text-body)] tracking-tight {unread
						? 'font-semibold text-text'
						: 'font-medium text-text-muted'}"
				>
					{formatNotificationTitle(notification)}
				</span>
				{#if detail}
					<span
						class="mt-0.5 block truncate text-[length:var(--text-body-sm)] leading-snug {unread
							? 'text-text-muted'
							: 'text-text-subtle'}"
					>
						{detail}
					</span>
				{/if}
			</span>
		</button>
		<span
			class="flex-none text-[length:var(--text-body-sm)] whitespace-nowrap text-text-faint"
		>
			{formatTimeAgo(notification.created_at)}
		</span>
		<div class="flex flex-none items-center gap-1">
			{#if unread}
				<button
					type="button"
					class="flex size-[34px] items-center justify-center rounded-md border border-border bg-surface text-accent-700 transition-colors hover:bg-accent-50"
					onclick={() => handleMarkAsRead(notification.id)}
					title={tr('notifications.markAsRead')}
					aria-label={tr('notifications.markAsRead')}
				>
					<svg
						width="17"
						height="17"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<path d="M20 6L9 17l-5-5" />
					</svg>
				</button>
			{/if}
			<button
				type="button"
				class="flex size-[34px] items-center justify-center rounded-md border border-border bg-surface text-danger-600 transition-colors hover:bg-danger-50"
				onclick={() => handleDelete(notification.id)}
				title={tr('notifications.delete')}
				aria-label={tr('notifications.delete')}
			>
				{@render trashIcon()}
			</button>
		</div>
	</div>
{/snippet}

<!-- Android M3 row (mockup): tonal type tile, headline plus detail line and a
     per-row overflow menu. -->
{#snippet androidRow(notification: NotificationDTO)}
	{@const tone = notificationTone(notification.type)}
	<div
		class="relative flex items-start gap-3.5 py-3.5 pr-4 pl-5 {notification.is_read
			? ''
			: 'bg-accent-100'}"
	>
		<button
			type="button"
			class="flex min-w-0 flex-1 items-start gap-3.5 text-left"
			onclick={() => handleNotificationClick(notification)}
		>
			<span
				class="flex h-notif-tile w-notif-tile shrink-0 items-center justify-center rounded-m3-full {tone.tile} {tone.ink} {notification.is_read
					? 'opacity-70'
					: ''}"
			>
				<NotificationTypeIcon type={notification.type} size={21} />
			</span>
			<span class="min-w-0 flex-1">
				<span class="flex items-baseline gap-2">
					<span
						class="min-w-0 flex-1 truncate text-body {notification.is_read
							? 'font-medium text-text-muted'
							: 'font-semibold text-text'}"
					>
						{formatTypeTitle(notification)}
					</span>
					<span
						class="shrink-0 text-eyebrow font-normal whitespace-nowrap text-text-faint"
					>
						{formatTimeAgo(notification.created_at)}
					</span>
				</span>
				<span
					class="mt-0.5 block text-body-sm {notification.is_read
						? 'text-text-subtle'
						: 'text-text-muted'}"
				>
					{formatNotificationMessage(notification)}
				</span>
			</span>
		</button>

		<!-- M3 overflow menu per row (mark as read / delete). -->
		<button
			type="button"
			class="notif-overflow -mt-0.5 -mr-1.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-m3-full text-text-faint"
			aria-label={tr('notifications.moreActions')}
			aria-expanded={openMenuId === notification.id}
			onclick={() =>
				(openMenuId = openMenuId === notification.id ? null : notification.id)}
		>
			<svg class="h-4.5 w-4.5" viewBox="0 0 24 24" fill="currentColor">
				<circle cx="12" cy="5" r="1.8" />
				<circle cx="12" cy="12" r="1.8" />
				<circle cx="12" cy="19" r="1.8" />
			</svg>
		</button>

		{#if openMenuId === notification.id}
			<div
				class="notif-overflow absolute top-11 right-3 z-30 min-w-44 overflow-hidden rounded-m3-md bg-m3-card py-1 shadow-m3-dialog"
			>
				{#if !notification.is_read}
					<button
						type="button"
						class="block w-full px-4 py-2.5 text-left text-body text-text"
						onclick={() => {
							openMenuId = null;
							handleMarkAsRead(notification.id);
						}}
					>
						{tr('notifications.markAsRead')}
					</button>
				{/if}
				<button
					type="button"
					class="block w-full px-4 py-2.5 text-left text-body text-danger-800"
					onclick={() => {
						openMenuId = null;
						handleDelete(notification.id);
					}}
				>
					{tr('notifications.delete')}
				</button>
			</div>
		{/if}
	</div>
{/snippet}

<!-- One row implementation for both iOS sections. Read rows keep the swipe
     wrapper so they stay deletable; only the mark-read action is unread-only. -->
{#snippet row(notification: NotificationDTO)}
	{@const t = notificationTone(notification.type)}
	{@const isRead = notification.is_read}
	<div
		class="relative overflow-hidden border-b border-border-soft last:border-b-0"
	>
		<!-- Swipe actions sit behind the row. -->
		<div class="absolute inset-y-0 right-0 flex items-stretch">
			{#if !isRead}
				<button
					type="button"
					onclick={() => {
						swipedId = null;
						handleMarkAsRead(notification.id);
					}}
					class="flex w-17.5 flex-col items-center justify-center gap-0.75 bg-accent-600 text-[length:var(--text-tag)] font-semibold text-white"
				>
					<svg
						class="h-4.25 w-4.25"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<path d="M20 6L9 17l-5-5" />
					</svg>
					{tr('notifications.ios.swipeRead')}
				</button>
			{/if}
			<button
				type="button"
				onclick={() => {
					swipedId = null;
					handleDelete(notification.id);
				}}
				class="flex w-17.5 flex-col items-center justify-center gap-0.75 bg-danger-600 text-[length:var(--text-tag)] font-semibold text-white"
			>
				<svg
					class="h-4.25 w-4.25"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path d="M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14" />
				</svg>
				{tr('notifications.ios.swipeDelete')}
			</button>
		</div>
		<div
			class="relative flex items-start gap-3 px-3.75 py-3.5 transition-transform {isRead
				? 'bg-surface'
				: 'bg-accent-50'}"
			style:transform={swipedId === notification.id
				? `translateX(-${isRead ? SWIPE_REVEAL_READ : SWIPE_REVEAL}px)`
				: 'translateX(0)'}
			ontouchstart={onRowTouchStart}
			ontouchmove={(e) => onRowTouchMove(e, notification.id)}
			onclick={() =>
				swipedId === notification.id
					? (swipedId = null)
					: handleNotificationClick(notification)}
			onkeydown={(e) =>
				e.key === 'Enter' && handleNotificationClick(notification)}
			role="button"
			tabindex="0"
		>
			<span
				class="flex h-9.5 w-9.5 shrink-0 items-center justify-center rounded-lg {t.tile} {t.ink} {isRead
					? 'opacity-75'
					: ''}"
			>
				<NotificationTypeIcon type={notification.type} />
			</span>
			<div class="min-w-0 flex-1">
				<div class="flex items-baseline gap-2">
					<span
						class="min-w-0 flex-1 text-subheading {isRead
							? 'font-medium text-text-muted'
							: 'text-text'}"
					>
						{formatTypeTitle(notification)}
					</span>
					<span
						class="shrink-0 whitespace-nowrap text-[length:var(--text-eyebrow)] text-text-faint"
					>
						{formatTimeAgo(notification.created_at)}
					</span>
				</div>
				<p
					class="mt-0.5 text-body-sm {isRead
						? 'text-text-subtle'
						: 'text-text-muted'}"
				>
					{formatNotificationMessage(notification)}
				</p>
			</div>
			{#if !isRead}
				<span class="mt-1.5 h-1.75 w-1.75 shrink-0 rounded-full {t.dot}"></span>
			{/if}
		</div>
	</div>
{/snippet}

{#if IS_DESKTOP}
	{#if isLoadingNotifications}
		<LoadingSpinner />
	{:else}
		<!-- Same two-column split as the settings tabs: the list fills the wide
		     left column, a w-90 sidebar with the archive note on the right. -->
		<div class="flex flex-col items-start gap-6 lg:flex-row">
			<div class="w-full min-w-0 flex-1">
				{#if notifications.length === 0}
					<!-- Board C — empty state -->
					<div class="flex flex-col items-center px-12 py-[72px] text-center">
						<span
							class="mb-6 flex size-[92px] items-center justify-center rounded-full border border-border bg-surface text-text-faint shadow-card"
						>
							<svg
								width="42"
								height="42"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="1.6"
								stroke-linecap="round"
								stroke-linejoin="round"
								aria-hidden="true"
							>
								<path
									d="M18 8a6 6 0 00-9.3-5M5.2 8.7C5 12.5 3 15 3 15h13M13.7 21a2 2 0 01-3.4 0"
								/>
								<path d="M3 3l18 18" />
							</svg>
						</span>
						<h1
							class="text-[length:var(--text-heading)] font-bold tracking-tight text-text-strong"
						>
							{tr('notifications.emptyTitle')}
						</h1>
						<div
							class="mt-2 max-w-[400px] text-[length:var(--text-body)] leading-relaxed text-text-muted"
						>
							{tr('notifications.emptyDescription')}
						</div>
						<a
							href={resolve('/notifications/settings')}
							class="mt-6 inline-flex items-center gap-2 rounded-full border border-border-field bg-surface px-[18px] py-[11px] text-[length:var(--text-label)] font-semibold text-accent-700"
						>
							{@render gearIcon(16)}
							{tr('notifications.settingsLink')}
						</a>
					</div>
				{:else}
					<!-- Board B — the header row is the shell's, above this content.
					     No NEU/FRÜHER kickers on desktop (same call as the dashboard's
					     at-checkout label); unread and read stay separate groups. -->
					{#if unreadNotifications.length > 0}
						<div class="flex flex-col gap-2">
							{#each unreadNotifications as notification (notification.id)}
								{@render desktopRow(notification, true)}
							{/each}
						</div>
					{/if}

					{#if readNotifications.length > 0}
						<div
							class="flex flex-col gap-2 {unreadNotifications.length > 0
								? 'mt-6'
								: ''}"
						>
							{#each readNotifications as notification (notification.id)}
								{@render desktopRow(notification, false)}
							{/each}
						</div>
					{/if}

					{#if hasMore}
						<div class="mt-5 text-center">
							<button
								class="text-[length:var(--text-label)] font-semibold text-accent-700 disabled:opacity-50"
								onclick={loadMore}
								disabled={isLoadingMore}
							>
								{isLoadingMore
									? tr('notifications.loadingMore')
									: tr('notifications.loadMore')}
							</button>
						</div>
					{/if}
				{/if}
			</div>
			<div class="flex w-full flex-col gap-5 lg:w-90 lg:flex-none">
				<div class="rounded-xl border border-border bg-white p-6">
					<h3 class="mb-1.5 text-subheading font-semibold text-text">
						{$t('settings.notifications.hintTitle')}
					</h3>
					<p class="text-body-sm text-text-muted">
						{tr('notifications.archiveHint')}
					</p>
				</div>
			</div>
		</div>
	{/if}
{:else if IS_ANDROID}
	{#if isLoadingNotifications}
		<LoadingSpinner />
	{:else if notifications.length === 0}
		<!-- Empty state: centred column with a settings shortcut. -->
		<div
			class="flex min-h-[60vh] flex-col items-center justify-center px-10 text-center"
		>
			<span
				class="mb-5.5 flex h-22 w-22 items-center justify-center rounded-m3-full bg-m3-surface-container text-text-faint"
			>
				<svg
					class="h-10 w-10"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.7"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path
						d="M18 8a6 6 0 00-9.3-5M5.2 8.7C5 12.5 3 15 3 15h13M13.7 21a2 2 0 01-3.4 0"
					/>
					<path d="M3 3l18 18" />
				</svg>
			</span>
			<p class="text-heading text-text-strong">
				{tr('notifications.emptyTitle')}
			</p>
			<p class="mt-2 max-w-empty-copy text-body-sm text-text-muted">
				{tr('notifications.emptyDescription')}
			</p>
			<a
				href={resolve('/notifications/settings')}
				class="mt-5.5 inline-flex items-center gap-2 rounded-m3-full bg-accent px-5 py-3 text-label text-on-accent shadow-fab"
			>
				{@render gearIcon(16)}
				{tr('notifications.settingsTitle')}
			</a>
		</div>
	{:else}
		<!-- Unread count + "mark all as read" pill, below the app bar. -->
		<div class="flex items-center justify-between px-4.5 pb-3">
			<p class="text-chip font-normal text-text-muted">
				{$t('notifications.unreadCount', {
					count: $notificationStore.unreadCount
				})}
			</p>
			<!-- Either signal keeps the action reachable: the server count can be
			     stale at 0 when the unread-count request failed while the list
			     itself loaded unread rows. -->
			{#if $notificationStore.unreadCount > 0 || unreadNotifications.length > 0}
				<button
					type="button"
					class="rounded-m3-full bg-accent-50 px-3.5 py-2 text-chip text-accent-700"
					onclick={handleMarkAllAsRead}
				>
					{tr('notifications.markAllAsRead')}
				</button>
			{/if}
		</div>

		{#if unreadNotifications.length > 0}
			<SectionLabel inset>{tr('notifications.sectionNew')}</SectionLabel>
			{#each unreadNotifications as notification (notification.id)}
				{@render androidRow(notification)}
			{/each}
		{/if}

		{#if readNotifications.length > 0}
			{#if unreadNotifications.length > 0}
				<div class="mx-5 my-2 h-px bg-border-soft"></div>
			{/if}
			<SectionLabel inset>{tr('notifications.sectionEarlier')}</SectionLabel>
			{#each readNotifications as notification (notification.id)}
				{@render androidRow(notification)}
			{/each}
		{/if}

		{#if hasMore}
			<div class="px-6 py-4 text-center">
				<button
					class="text-body text-accent disabled:opacity-50"
					onclick={loadMore}
					disabled={isLoadingMore}
				>
					{isLoadingMore
						? tr('notifications.loadingMore')
						: tr('notifications.loadMore')}
				</button>
			</div>
		{/if}

		<p class="mt-3.5 px-6 text-center text-eyebrow font-normal text-text-faint">
			{tr('notifications.archiveHint')}
		</p>
	{/if}
{:else if IS_IOS}
	{#if isLoadingNotifications}
		<LoadingSpinner />
	{:else if notifications.length === 0}
		<!-- Empty state: glass bell-off medallion, copy, settings pill. -->
		<!-- Mockup centers the empty state in the remaining screen height. -->
		<div
			class="flex min-h-[60vh] flex-col items-center justify-center px-6 text-center"
		>
			<span
				class="liquid-glass-card mb-5.5 flex h-21 w-21 items-center justify-center rounded-full text-text-faint"
			>
				<svg
					class="h-9.5 w-9.5"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.7"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path
						d="M18 8a6 6 0 00-9.3-5M5.2 8.7C5 12.5 3 15 3 15h13M13.7 21a2 2 0 01-3.4 0"
					/>
					<path d="M3 3l18 18" />
				</svg>
			</span>
			<p class="text-heading text-text-strong">
				{tr('notifications.emptyTitle')}
			</p>
			<p class="mt-2 max-w-62.5 text-label font-normal text-text-muted">
				{tr('notifications.emptyDescription')}
			</p>
			<a
				href={resolve('/notifications/settings')}
				class="liquid-glass-surface mt-5.5 inline-flex items-center gap-1.75 rounded-full px-4.5 py-2.75 text-label text-accent-700"
			>
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.9"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<circle cx="12" cy="12" r="3" />
					<path
						d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 11-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 008 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 11-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H2a2 2 0 010-4h.09A1.65 1.65 0 003.6 8a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 112.83-2.83l.06.06a1.65 1.65 0 001.82.33H8a1.65 1.65 0 001-1.51V2a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 112.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V8a1.65 1.65 0 001.51 1H22a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"
					/>
				</svg>
				{tr('notifications.settingsLink')}
			</a>
		</div>
	{:else}
		{#if unreadNotifications.length > 0}
			<SectionLabel inset>{tr('notifications.sectionNew')}</SectionLabel>
			<div class="overflow-hidden rounded-inset bg-surface">
				{#each unreadNotifications as notification (notification.id)}
					{@render row(notification)}
				{/each}
			</div>
		{/if}

		{#if readNotifications.length > 0}
			<SectionLabel inset spaced
				>{tr('notifications.sectionEarlier')}</SectionLabel
			>
			<div class="overflow-hidden rounded-inset bg-surface">
				{#each readNotifications as notification (notification.id)}
					{@render row(notification)}
				{/each}
			</div>
		{/if}

		{#if hasMore}
			<div class="pt-4 text-center">
				<button
					class="text-label text-accent disabled:opacity-50"
					onclick={loadMore}
					disabled={isLoadingMore}
				>
					{isLoadingMore
						? tr('notifications.loadingMore')
						: tr('notifications.loadMore')}
				</button>
			</div>
		{/if}

		<p
			class="mt-4 text-center text-[length:var(--text-eyebrow)] leading-normal text-text-faint"
		>
			{tr('notifications.archiveHint')}
		</p>
	{/if}
{/if}
