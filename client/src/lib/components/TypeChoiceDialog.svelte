<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
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
			}
		].filter((tpe) => tpe.enabled)
	);

	function choose(href: string) {
		open = false;
		onClose();
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- href is produced by resolve() above
		goto(href);
	}
</script>

{#if open}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-[60] flex items-end sm:items-center justify-center bg-black/40 backdrop-blur-sm"
		role="button"
		tabindex="0"
		aria-label={tr('common.close')}
		onclick={() => {
			open = false;
			onClose();
		}}
		onkeydown={(e) => {
			if (e.key === 'Escape') {
				open = false;
				onClose();
			}
		}}
	>
		<!-- Sheet (mobile) / Dialog (desktop) -->
		<div
			role="dialog"
			tabindex="-1"
			aria-modal="true"
			aria-label={tr('typeChoice.title')}
			class="w-full sm:max-w-md m-0 sm:m-4 p-6 rounded-t-3xl sm:rounded-2xl shadow-xl {platform ===
			'ios'
				? 'bg-white/80 backdrop-blur-xl backdrop-saturate-150 border border-white/40'
				: 'bg-white border border-gray-200'}"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			{#if platform === 'ios'}
				<div class="mx-auto mb-4 h-1 w-10 rounded-full bg-gray-300"></div>
			{/if}
			<h2 class="text-lg font-semibold text-gray-900 mb-1">
				{tr('typeChoice.title')}
			</h2>
			<p class="text-sm text-gray-500 mb-4">{tr('typeChoice.subtitle')}</p>

			<div class="flex flex-col gap-2">
				{#each types as type (type.key)}
					<button
						type="button"
						onclick={() => choose(type.href)}
						class="flex items-center gap-3 w-full px-4 py-3 rounded-xl border border-gray-200 hover:border-cyan-500 hover:bg-cyan-50 transition-colors text-left"
					>
						<span
							class="flex items-center justify-center w-10 h-10 rounded-full bg-cyan-100 text-cyan-700 shrink-0"
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
						<span class="font-medium text-gray-900">{type.label}</span>
					</button>
				{/each}
			</div>
		</div>
	</div>
{/if}
