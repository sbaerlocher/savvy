<script lang="ts">
	import { ICON_SEARCH } from '$lib/icons';
	import NotificationPanel from '$lib/components/NotificationPanel.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { showNewDialog } from '$lib/stores/newDialog';
	import { platform } from '$lib/utils/platform';

	// Android search opens an inline M3 search field in the header row; the
	// parent PageHeader owns that state and passes the opener down.
	let { onSearchOpen }: { onSearchOpen?: () => void } = $props();

	// Android M3 (mockup header): 44px round borderless icon buttons, 4px gap.
	// iOS/other keep the boxed 40px buttons. Module-constant platform → consts.
	const BOX_CLASS =
		platform === 'android'
			? 'inline-flex h-11 w-11 items-center justify-center rounded-full text-text-muted transition-colors hover:bg-surface-1'
			: 'inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-white text-text-muted transition-colors hover:bg-surface-1';
	const GAP_CLASS = platform === 'android' ? 'gap-1' : 'gap-2.5';
	const SEARCH_ICON = platform === 'android' ? 'h-5.5 w-5.5' : 'h-5 w-5';
	const BELL_ICON = platform === 'android' ? 'h-5.25 w-5.25' : 'h-5 w-5';
</script>

<!-- Mobile-only header actions, rendered on the page title row (see PageHeader)
     so the mockup's "title left · bell + '+' right" layout holds without a
     separate top header bar. Desktop keeps these in the global nav. -->
<!-- Action order follows the mockup's native chrome:
     iOS  → [bell] [+]      (add lives in the header)
     else → [search] [bell] (add is the Material FAB in the nav bar) -->
<!-- Boxed 40px header buttons matching the mockup:
     bell/search = white circle-ish box + border, "+" = filled teal box. -->
{#snippet bell()}
	{#if !$authStore.user?.is_impersonating}
		<NotificationPanel
			mode="link"
			triggerClass="notification-bell relative {BOX_CLASS}"
			iconClass={BELL_ICON}
		/>
	{/if}
{/snippet}

<div class="flex items-center {GAP_CLASS} sm:hidden">
	{#if platform === 'ios'}
		{@render bell()}
		<button
			type="button"
			onclick={() => ($showNewDialog = true)}
			data-testid="nav-new-mobile"
			aria-label={$t('common.new')}
			class="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-accent text-white shadow-[var(--shadow-accent)] transition-transform active:scale-95"
		>
			<svg
				class="h-5 w-5"
				fill="none"
				stroke="currentColor"
				stroke-width="2.3"
				viewBox="0 0 24 24"
			>
				<path stroke-linecap="round" d="M12 5v14M5 12h14" />
			</svg>
		</button>
	{:else}
		<button
			type="button"
			onclick={onSearchOpen}
			data-testid="nav-search-mobile"
			aria-label={$t('common.search')}
			class={BOX_CLASS}
		>
			<svg
				class={SEARCH_ICON}
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path stroke-linecap="round" stroke-linejoin="round" d={ICON_SEARCH} />
			</svg>
		</button>
		{@render bell()}
	{/if}
</div>
