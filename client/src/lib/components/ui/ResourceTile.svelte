<script lang="ts">
	import { t } from '$lib/stores/i18n';
	import { resolve } from '$app/paths';
	import Barcode from '$lib/components/Barcode.svelte';
	import type { TileModel } from '$lib/utils/tile-model';
	import type { BarcodeModalItem } from '$lib/components/dashboard/BarcodeModal.svelte';

	let {
		model,
		showBarcode = false,
		compact = false,
		selectMode = false,
		selected = false,
		onSelect,
		onShowBarcode
	}: {
		model: TileModel;
		showBarcode?: boolean;
		compact?: boolean;
		selectMode?: boolean;
		selected?: boolean;
		onSelect?: (id: string) => void;
		onShowBarcode?: (item: BarcodeModalItem) => void;
	} = $props();

	// Per-type icon (neutral line icons, Direction B — no emoji).
	const iconPaths: Record<TileModel['type'], string> = {
		card: 'M3 10h18M7 15h1m4 0h1m-7 4h12a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
		voucher:
			'M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 010 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 010-4V7a2 2 0 00-2-2H5z',
		gift_card:
			'M12 8v13m0-13V6a2 2 0 112-2 2 2 0 01-2 2zm0 0V5.5A2.5 2.5 0 109.5 8H12zm-7 4h14M5 12a2 2 0 110-4h14a2 2 0 110 4M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7'
	};

	const typeLabel = $derived(
		model.type === 'card'
			? $t('dashboard.cardType')
			: model.type === 'voucher'
				? $t('dashboard.voucherType')
				: $t('dashboard.giftCardType')
	);

	// Resolve-friendly literal per type (resolve needs the concrete [id] route).
	const href = $derived(
		model.type === 'card'
			? resolve(`/cards/${model.id}`)
			: model.type === 'voucher'
				? resolve(`/vouchers/${model.id}`)
				: resolve(`/gift-cards/${model.id}`)
	);

	const contentClass = $derived(
		`flex flex-col text-left ${compact ? 'p-3' : 'p-4'} ${
			model.isActive ? '' : 'opacity-50 grayscale'
		}`
	);

	function handleBarcodeClick(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		onShowBarcode?.(model.barcodeModalItem);
	}
</script>

<div
	class="group relative flex flex-col overflow-hidden rounded-xl border border-border/80 bg-white transition hover:border-border-field {selectMode &&
	selected
		? 'ring-2 ring-accent'
		: ''}"
	style="border-left: 3px solid color-mix(in srgb, {model.merchantColor} 70%, transparent)"
>
	<!-- Status overlay (centered) when not active — sits OUTSIDE the dimmed
	     content so the badge stays at full opacity, like the prototype. -->
	{#if !model.isActive && model.statusBadge}
		<div
			class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center"
		>
			<span
				class="rounded-full border border-border-field bg-white/90 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-text-ink2 shadow-sm"
			>
				{model.statusBadge}
			</span>
		</div>
	{/if}

	<!-- Info area: a real link for navigation; a toggle button in select mode.
	     The barcode is a sibling button so no interactive element is nested. -->
	{#if selectMode}
		<button
			type="button"
			class={contentClass}
			data-owner={model.shareState.kind === 'sharedFrom' ? 'shared' : 'owned'}
			data-resource-type={model.type}
			onclick={() => onSelect?.(model.id)}
		>
			{@render tileBody()}
		</button>
	{:else}
		<!-- eslint-disable svelte/no-navigation-without-resolve -- href is already produced by resolve() above -->
		<a
			{href}
			class={contentClass}
			data-owner={model.shareState.kind === 'sharedFrom' ? 'shared' : 'owned'}
			data-resource-type={model.type}
		>
			{@render tileBody()}
		</a>
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
	{/if}

	{#snippet tileBody()}
		<!-- Header: icon + type label + merchant name | amount + expiry -->
		<div class="flex items-start gap-3">
			<div
				class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-border-soft text-text-subtle"
			>
				<svg
					class="h-5 w-5"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.75"
						d={iconPaths[model.type]}
					/>
				</svg>
			</div>

			<div class="min-w-0 flex-1">
				<!-- Type label + compact share status share one row (prototype:
				     lock private / people + N shared-out / people + first name
				     received). Full text stays on the detail page. -->
				<div
					class="flex items-center gap-2 text-[0.65rem] font-semibold uppercase tracking-wider text-text-faint"
				>
					<span>{typeLabel}</span>
					<span class="flex items-center gap-1 normal-case tracking-normal">
						{#if model.shareState.kind === 'private'}
							<svg
								class="h-3.5 w-3.5 shrink-0"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
								aria-hidden="true"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
								/>
							</svg>
						{:else}
							<svg
								class="h-3.5 w-3.5 shrink-0"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
								aria-hidden="true"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M17 20h5v-2a4 4 0 00-3-3.87M9 20H4v-2a4 4 0 013-3.87m6-1.13a4 4 0 10-4-4 4 4 0 004 4zm6 0a4 4 0 00-3-6.85"
								/>
							</svg>
							{#if model.shareState.kind === 'sharedWith'}
								<span class="tabular-nums">{model.shareState.count}</span>
							{:else}
								<span class="truncate"
									>{$t('tile.sharedFrom', {
										name: model.shareState.firstName
									})}</span
								>
							{/if}
						{/if}
					</span>
				</div>
				<p
					class="truncate font-semibold text-text transition group-hover:text-accent-hover"
				>
					{model.merchantName}
				</p>
				{#if model.identifier}
					<p class="truncate text-sm text-text-subtle">{model.identifier}</p>
				{/if}
			</div>

			<!-- Kennzahl + expiry badge (fixed top-right slot) -->
			<div class="flex shrink-0 flex-col items-end gap-1 text-right">
				{#if model.amount}
					<p class="text-lg font-bold tabular-nums text-text">
						{model.amount}
					</p>
				{/if}
				{#if model.expiryBadge}
					<span
						class="rounded-full px-2 py-0.5 text-[0.7rem] font-medium {model.expiryUrgent
							? 'bg-amber-50 text-amber-700'
							: 'bg-border-soft text-text-subtle'}"
					>
						{model.expiryBadge}
					</span>
				{:else if model.notYetValid}
					<span
						class="rounded-full bg-border-soft px-2 py-0.5 text-[0.7rem] font-medium text-text-subtle"
					>
						{model.notYetValid}
					</span>
				{/if}
			</div>
		</div>

		<!-- Footer: masked number | usage marker -->
		<div class="mt-1 flex items-center justify-between gap-2">
			<p class="font-mono text-xs text-text-faint">
				{#if model.maskedNumber}{model.maskedNumber}{:else}&nbsp;{/if}
			</p>
			{#if model.usageMarker}
				<span
					class="text-[0.65rem] font-semibold uppercase tracking-wider text-text-faint"
				>
					{model.usageMarker}
				</span>
			{/if}
		</div>
	{/snippet}

	<!-- Barcode box (optional, context-driven via showBarcode). With
	     onShowBarcode it is a button that enlarges via modal; without it the
	     barcode is already shown inline, so render a static box (no fake tap). -->
	{#if showBarcode && model.barcodeValue}
		{@const barcodeValue = model.barcodeValue}
		{@const boxClass = `mx-3 mb-3 flex flex-col items-center justify-center rounded-lg border border-border-soft bg-surface-1 p-3 ${model.isActive ? '' : 'opacity-50 grayscale'}`}
		{#snippet barcodeContent()}
			<!-- Barcode image only; the raw value already shows masked in the
			     footer (·· 1234), so no duplicate full-value text here. -->
			<Barcode
				value={barcodeValue}
				type={model.barcodeType}
				height={compact ? 50 : 64}
				maxHeight={compact ? 50 : 64}
			/>
		{/snippet}
		{#if onShowBarcode}
			<button
				type="button"
				onclick={handleBarcodeClick}
				class="{boxClass} transition hover:bg-accent-50"
				title={$t('dashboard.tapToEnlarge')}
				aria-label={$t('dashboard.showBarcode')}
			>
				{@render barcodeContent()}
			</button>
		{:else}
			<div class={boxClass}>
				{@render barcodeContent()}
			</div>
		{/if}
	{/if}
</div>
