<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { authStore } from '$lib/stores/auth';
	import { notificationStore } from '$lib/stores/notifications';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import Section from './Section.svelte';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Desktop renders the mockup layout (centred column), Android the M3 list,
	// iOS the push screen. `platform` is a module constant, so these are plain
	// consts, not $derived.
	const IS_DESKTOP = platform === 'other';
	const IS_ANDROID = platform === 'android';
	const IS_IOS = platform === 'ios';

	// The actions snippets (desktop and iOS) call the section's mark-all
	// handler (API call, local row update, store refresh, toasts) through this
	// ref, so the page does not carry a copy that would leave the section's
	// rows stale.
	let section: Section | undefined;

	onMount(() => {
		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}
		// The eyebrow counts below read the store (authoritative, server-side);
		// populate it here so the desktop skeleton, which mounts no Section,
		// shows a count too.
		notificationStore.refreshUnreadCount();
	});

	// Back: return to where the user came from; fall back to the dashboard on a
	// deep link (push notification, bookmark, PWA start URL), where history.back()
	// would be a no-op and leave the chevron dead.
	function goBack() {
		if (history.length > 1) history.back();
		else goto(resolve('/dashboard'));
	}
</script>

<svelte:head>
	<title>{tr('notifications.title')} - {tr('common.appName')}</title>
</svelte:head>

{#if IS_DESKTOP}
	<!-- Full shell width like every other title row; the 680px reading column
	     is the section's own wrapper, not the shell's. -->
	<PageShell
		title={tr('notifications.title')}
		eyebrow={$t('notifications.unreadCount', {
			count: $notificationStore.unreadCount
		})}
		eyebrowVerbatim
		mobileActions={false}
	>
		{#snippet actions()}
			<a
				href={resolve('/notifications/settings')}
				class="title-action whitespace-nowrap text-text-muted"
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
				{tr('notifications.settingsTitle')}
			</a>
			{#if $notificationStore.unreadCount > 0}
				<button
					type="button"
					class="title-action whitespace-nowrap border-accent-600 bg-accent-600 font-semibold text-on-accent hover:bg-accent-700"
					onclick={() => section?.handleMarkAllAsRead()}
				>
					{tr('notifications.markAllAsRead')}
				</button>
			{/if}
		{/snippet}
		<Section bind:this={section} />
	</PageShell>
{:else if IS_ANDROID}
	<!-- Android M3: full-bleed list, no card chrome (mockup). `width="full"` is
	     the shell's own opt-out of the horizontal padding, replacing the page's
	     `-mx-4`; the rows reach the screen edge while the header keeps its own
	     inset below. No shell title for that reason — the header is padded, the
	     list is not. -->
	<PageShell
		width="full"
		title={tr('notifications.title')}
		mobileActions={false}
		onBack={goBack}
	>
		<Section />
	</PageShell>
{:else if IS_IOS}
	<!-- iOS push screen: in-flow glass header, grouped-inset sections, swipe
	     actions (screen-NotificationsIOS mockup). The unread count rides on the
	     eyebrow, mark-all on the actions slot. -->
	<PageShell
		title={tr('notifications.title')}
		eyebrow={$notificationStore.unreadCount > 0
			? $t('notifications.unreadCount', {
					count: $notificationStore.unreadCount
				})
			: $t('notifications.ios.allRead')}
		mobileActions={false}
		onBack={goBack}
	>
		{#snippet actions()}
			<!-- Enabled state comes from the store count (the section's local list
			     is not visible here); the store refreshes after every list op. -->
			<button
				type="button"
				onclick={() => section?.handleMarkAllAsRead()}
				disabled={$notificationStore.unreadCount === 0}
				aria-label={tr('notifications.markAllAsRead')}
				class="liquid-glass-surface mt-0.75 inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full {$notificationStore.unreadCount >
				0
					? 'text-accent-700'
					: 'text-text-faint'}"
			>
				<svg
					class="h-5 w-5"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2.3"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path d="M20 6L9 17l-5-5" />
				</svg>
			</button>
		{/snippet}

		<Section bind:this={section} />
	</PageShell>
{/if}
