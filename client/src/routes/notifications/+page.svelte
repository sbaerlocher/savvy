<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { notificationsApi } from '$lib/api/notifications';
	import NotificationTypeIcon from '$lib/components/NotificationTypeIcon.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { authStore } from '$lib/stores/auth';
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
	// boxed row actions). iOS/Android keep the existing list. `platform` is a
	// module constant, so this is a plain const, not $derived.
	const IS_DESKTOP = platform === 'other';
	const IS_ANDROID = platform === 'android';

	let notifications = $state<NotificationDTO[]>([]);
	let isLoadingNotifications = $state(true);
	let isLoadingMore = $state(false);
	let offset = $state(0);
	let hasMore = $state(true);

	const unreadNotifications = $derived(notifications.filter((n) => !n.is_read));
	const readNotifications = $derived(notifications.filter((n) => n.is_read));

	// Id of the row whose overflow menu is open (Android), or null.
	let openMenuId = $state<string | null>(null);

	// Any click outside the open overflow menu dismisses it (same pattern as
	// NotificationPanel and DesktopNav).
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (openMenuId && !target.closest('.notif-overflow')) openMenuId = null;
	}

	// Back: return to where the user came from; fall back to the dashboard on a
	// deep link (push notification, bookmark, PWA start URL), where history.back()
	// would be a no-op and leave the chevron dead.
	function goBack() {
		if (history.length > 1) history.back();
		else goto(resolve('/dashboard'));
	}

	// Headline for the Android list row (the mockup splits headline and detail
	// across two lines). It names the kind of event; the detail line below
	// carries the actor and merchant, so the two must not repeat each other.
	function formatAndroidTitle(notification: NotificationDTO): string {
		// $t() echoes the key back when it is missing, so compare against it
		// rather than relying on a falsy return for unknown notification types.
		const key = `notifications.typeTitle.${notification.type}`;
		const title = $t(key);
		return title === key ? $t('notifications.newNotification') : title;
	}

	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}
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
</script>

<svelte:head>
	<title>{tr('notifications.title')} - {tr('common.appName')}</title>
</svelte:head>

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
						{formatAndroidTitle(notification)}
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

{#if IS_DESKTOP}
	<div class="mx-auto max-w-7xl px-4">
		{#if isLoadingNotifications}
			<LoadingSpinner />
		{:else}
			<div class="mx-auto max-w-[680px] py-10">
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
					<!-- Board B — header row -->
					<div class="mb-1.5 flex items-end justify-between gap-5">
						<div>
							<h1
								class="text-[length:var(--text-title)] font-bold tracking-tight"
							>
								{tr('notifications.title')}
							</h1>
							<p class="mt-1 text-[length:var(--text-label)] text-text-muted">
								{$t('notifications.unreadCount', {
									count: $notificationStore.unreadCount
								})}
							</p>
						</div>
						<div class="flex items-center gap-2.5">
							<a
								href={resolve('/notifications/settings')}
								class="inline-flex items-center gap-[7px] rounded-full border border-border-field bg-surface px-[15px] py-[9px] text-[length:var(--text-label)] font-semibold text-text-muted"
							>
								{@render gearIcon(16)}
								{tr('notifications.settingsTitle')}
							</a>
							{#if unreadNotifications.length > 0}
								<button
									type="button"
									class="rounded-full bg-accent-600 px-4 py-2.5 text-[length:var(--text-label)] font-semibold text-on-accent shadow-sm transition-colors hover:bg-accent-700"
									onclick={handleMarkAllAsRead}
								>
									{tr('notifications.markAllAsRead')}
								</button>
							{/if}
						</div>
					</div>

					{#if unreadNotifications.length > 0}
						<div
							class="mt-6 mb-2.5 text-[length:var(--text-eyebrow)] font-semibold tracking-wider uppercase text-text-subtle"
						>
							{tr('notifications.sectionNew')}
						</div>
						<div class="flex flex-col gap-2">
							{#each unreadNotifications as notification (notification.id)}
								{@render desktopRow(notification, true)}
							{/each}
						</div>
					{/if}

					{#if readNotifications.length > 0}
						<div
							class="mt-6 mb-2.5 text-[length:var(--text-eyebrow)] font-semibold tracking-wider uppercase text-text-subtle"
						>
							{tr('notifications.sectionEarlier')}
						</div>
						<div class="flex flex-col gap-2">
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

					<p
						class="mt-[22px] text-center text-[length:var(--text-body-sm)] leading-relaxed text-text-faint"
					>
						{tr('notifications.archiveHint')}
					</p>
				{/if}
			</div>
		{/if}
	</div>
{:else if IS_ANDROID}
	<!-- Android M3: full-bleed list, no card chrome (mockup). -->
	<div class="-mx-4">
		<div class="px-3">
			<PageHeader
				title={tr('notifications.title')}
				mobileActions={false}
				onBack={goBack}
			/>
		</div>

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
				<p
					class="px-5 pt-2 pb-1.5 text-section-eyebrow uppercase text-text-subtle"
				>
					{tr('notifications.sectionNew')}
				</p>
				{#each unreadNotifications as notification (notification.id)}
					{@render androidRow(notification)}
				{/each}
			{/if}

			{#if readNotifications.length > 0}
				{#if unreadNotifications.length > 0}
					<div class="mx-5 my-2 h-px bg-border-soft"></div>
				{/if}
				<p
					class="px-5 pt-2 pb-1.5 text-section-eyebrow uppercase text-text-subtle"
				>
					{tr('notifications.sectionEarlier')}
				</p>
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

			<p
				class="mt-3.5 px-6 text-center text-eyebrow font-normal text-text-faint"
			>
				{tr('notifications.archiveHint')}
			</p>
		{/if}
	</div>
{:else}
	<div class="px-4 max-w-7xl mx-auto">
		<PageHeader title={tr('notifications.title')} />

		{#if isLoadingNotifications}
			<LoadingSpinner />
		{:else}
			<div class="overflow-hidden rounded-xl border border-border bg-white">
				<!-- Mark-all-as-read strip (page title lives in the PageHeader). -->
				{#if notifications.some((n) => !n.is_read)}
					<div
						class="flex items-center justify-end px-6 py-3 border-b border-border"
					>
						<button
							class="text-sm text-accent hover:text-accent-hover"
							onclick={handleMarkAllAsRead}
						>
							{tr('notifications.markAllAsRead')}
						</button>
					</div>
				{/if}

				<!-- Notification Items -->
				{#if notifications.length === 0}
					<div class="px-4 py-12 text-center text-text-subtle">
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
						<p>{tr('notifications.noNotifications')}</p>
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
											d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
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
											d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
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
									class="text-text-faint hover:text-danger-600 p-1.5 rounded-md hover:bg-danger-50 transition-colors"
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
					<div class="px-6 py-4 border-t border-border-soft text-center">
						<button
							class="text-sm text-accent hover:text-accent-hover font-medium disabled:opacity-50"
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
		{/if}
	</div>
{/if}
