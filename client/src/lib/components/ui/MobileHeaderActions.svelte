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
<div class="flex items-center gap-2 sm:hidden">
	{#if !$authStore.user?.is_impersonating}
		<NotificationPanel />
	{/if}

	{#if platform === 'ios'}
		<button
			type="button"
			onclick={() => ($showNewDialog = true)}
			data-testid="nav-new-mobile"
			aria-label={$t('common.new')}
			class="inline-flex items-center justify-center p-1 text-cyan-600"
		>
			<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 4v16m8-8H4"
				/>
			</svg>
		</button>
	{:else}
		<!-- eslint-disable svelte/no-navigation-without-resolve -- base is resolve()d; ?search is a query string -->
		<a
			href={resolve('/wallet') + '?search=1'}
			data-testid="nav-search-mobile"
			aria-label={$t('common.search')}
			class="inline-flex items-center justify-center p-1 text-gray-600"
		>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
			<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
				/>
			</svg>
		</a>
	{/if}
</div>
