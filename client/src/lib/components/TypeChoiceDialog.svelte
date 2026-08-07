<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { authStore } from '$lib/stores/auth';
	import { configStore } from '$lib/stores/config';
	import { t } from '$lib/stores/i18n';
	import { platform } from '$lib/utils/platform';
	import { get } from 'svelte/store';

	interface Props {
		open: boolean;
		onClose: () => void;
	}

	let { open = $bindable(), onClose }: Props = $props();

	const tr = (key: string) => get(t)(key);

	// Move focus into the dialog when it opens so keyboard/screen-reader users
	// land inside it (and Escape works without a prior click).
	let dialogEl = $state<HTMLDivElement | null>(null);
	$effect(() => {
		if (open && dialogEl) dialogEl.focus();
	});

	// Resource types, feature-gated. Route = existing /{type}/new form.
	const types = $derived(
		[
			{
				key: 'card',
				href: resolve('/cards/new'),
				label: tr('common.card'),
				enabled: $configStore.features.cards,
				path: 'M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z'
			},
			{
				key: 'voucher',
				href: resolve('/vouchers/new'),
				label: tr('common.voucher'),
				enabled: $configStore.features.vouchers,
				path: 'M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z'
			},
			{
				key: 'gift_card',
				href: resolve('/gift-cards/new'),
				label: tr('common.gift_card'),
				enabled: $configStore.features.gift_cards,
				path: 'M12 8v13m0-13V6a2 2 0 112 2h-2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7'
			},
			{
				// Merchants are shared reference data managed by admins only, so this
				// entry is role-gated rather than feature-gated. Create lives here;
				// edit/delete live in the /admin/merchants table.
				key: 'merchant',
				href: resolve('/admin/merchants/new'),
				label: tr('common.merchant'),
				enabled: $authStore.user?.is_admin ?? false,
				path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5'
			}
		].filter((tpe) => tpe.enabled)
	);

	function handleClose() {
		open = false;
		onClose();
	}

	function choose(href: string) {
		open = false;
		onClose();
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- href is produced by resolve() above
		goto(href);
	}
</script>

<Modal
	{open}
	onclose={handleClose}
	layer="default"
	mobileLayout="sheet"
	backdrop="glass"
	label={tr('typeChoice.title')}
>
	<!-- Sheet (mobile) / Dialog (desktop) -->
	<div
		bind:this={dialogEl}
		tabindex="-1"
		class="pointer-events-auto w-full sm:max-w-md m-0 sm:m-4 p-6 {platform ===
		'ios'
			? 'liquid-glass-surface rounded-t-3xl sm:rounded-2xl'
			: platform === 'android'
				? 'bg-m3-surface-container rounded-t-[var(--radius-m3-xl)] sm:rounded-[var(--radius-m3-xl)] shadow-m3-dialog'
				: 'bg-white border border-border shadow-xl rounded-t-3xl sm:rounded-2xl'}"
	>
		{#if platform === 'ios'}
			<div
				class="mx-auto mb-4 h-1 w-10 rounded-full bg-[var(--color-glass-grabber)]"
			></div>
		{:else if platform === 'android'}
			<!-- M3 bottom-sheet drag handle -->
			<div class="mx-auto mb-4 h-1 w-8 rounded-full bg-border"></div>
		{/if}
		<h2 class="text-lg font-semibold text-text mb-1">
			{tr('typeChoice.title')}
		</h2>
		<p class="text-sm text-text-subtle mb-4">{tr('typeChoice.subtitle')}</p>

		<div class="flex flex-col gap-2">
			{#each types as type (type.key)}
				<button
					type="button"
					onclick={() => choose(type.href)}
					class="flex items-center gap-3 w-full px-4 py-3 rounded-xl border border-border hover:border-accent hover:bg-accent-50 transition-colors text-left"
				>
					<span
						class="flex items-center justify-center w-10 h-10 rounded-full bg-accent-100 text-accent-hover shrink-0"
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
								d={type.path}
							/>
						</svg>
					</span>
					<span class="font-medium text-text">{type.label}</span>
				</button>
			{/each}
		</div>
	</div>
</Modal>
