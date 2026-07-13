<script lang="ts">
	import { resolve } from '$app/paths';
	import NotificationPanel from '$lib/components/NotificationPanel.svelte';
	import { authStore } from '$lib/stores/auth';
	import { t } from '$lib/stores/i18n';
	import { showNewDialog } from '$lib/stores/newDialog';
	import { platform } from '$lib/utils/platform';
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
			triggerClass="notification-bell relative inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-white text-text-muted transition-colors hover:bg-surface-1"
			iconClass="h-5 w-5"
		/>
	{/if}
{/snippet}

<div class="flex items-center gap-2.5 sm:hidden">
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
		<!-- eslint-disable svelte/no-navigation-without-resolve -- base is resolve()d; ?search is a query string -->
		<a
			href={resolve('/wallet') + '?search=1'}
			data-testid="nav-search-mobile"
			aria-label={$t('common.search')}
			class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-white text-text-muted transition-colors hover:bg-surface-1"
		>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
			<svg
				class="h-5 w-5"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
				/>
			</svg>
		</a>
		{@render bell()}
	{/if}
</div>
