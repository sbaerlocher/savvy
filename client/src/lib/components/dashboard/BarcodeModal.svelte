<script lang="ts" module>
	export type BarcodeModalItem = {
		type: 'card' | 'voucher' | 'gift_card';
		value: string;
		barcodeType?: string;
		merchantName?: string;
		displayValue?: string;
		description?: string;
		pin?: string;
		balance?: string;
		currency?: string;
		expiresAt?: string;
		validFrom?: string;
		validUntil?: string;
		status?: string;
	};
</script>

<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { t, locale } from '$lib/stores/i18n';
	import { formatCurrency } from '$lib/utils/currency';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import Barcode from '$lib/components/Barcode.svelte';

	let {
		item,
		onClose
	}: {
		item: BarcodeModalItem | null;
		onClose: () => void;
	} = $props();

	const currentLocale = $derived($locale || 'de-DE');

	// Landscape detection for fullscreen barcode
	let isLandscape = $state(false);
	let resizeTimer: ReturnType<typeof setTimeout> | null = null;

	function checkOrientation() {
		const screenOrientationType = window.screen.orientation?.type;
		const isLandscapeByAPI =
			screenOrientationType?.includes('landscape') || false;
		const isLandscapeByDimensions = window.innerWidth > window.innerHeight;
		const newIsLandscape = isLandscapeByAPI || isLandscapeByDimensions;
		if (newIsLandscape !== isLandscape) {
			isLandscape = newIsLandscape;
		}
	}

	function handleOrientationChange() {
		setTimeout(checkOrientation, 100);
	}

	function handleResize() {
		if (resizeTimer) clearTimeout(resizeTimer);
		resizeTimer = setTimeout(checkOrientation, 150);
	}

	onMount(() => {
		const isTouchDevice = 'ontouchstart' in window;
		if (isTouchDevice) {
			checkOrientation();
			window.addEventListener('orientationchange', handleOrientationChange);
			window.addEventListener('resize', handleResize);
		}
	});

	onDestroy(() => {
		window.removeEventListener('orientationchange', handleOrientationChange);
		window.removeEventListener('resize', handleResize);
		if (resizeTimer) clearTimeout(resizeTimer);
	});
</script>

<!-- Quick Barcode Modal (portrait / desktop) -->
{#if item && !isLandscape}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
		onclick={onClose}
		onkeydown={(e) => e.key === 'Escape' && onClose()}
		role="dialog"
		aria-modal="true"
		aria-labelledby="barcode-modal-title"
		tabindex="-1"
	>
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="bg-white rounded-xl shadow-2xl max-w-md w-full p-6 relative"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="document"
		>
			<!-- Close Button -->
			<button
				onclick={onClose}
				class="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition"
				aria-label={$t('common.close')}
			>
				<svg
					class="w-6 h-6"
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

			<!-- Merchant Name -->
			{#if item.merchantName}
				<h2
					id="barcode-modal-title"
					class="text-xl font-bold text-gray-900 mb-4 pr-8"
				>
					{item.merchantName}
				</h2>
			{/if}

			<!-- Barcode Display -->
			<BarcodeDisplay
				value={item.value}
				type={item.barcodeType}
				status={item.status}
				pin={item.pin}
				validFrom={item.validFrom}
				validUntil={item.validUntil}
				displayValue={item.displayValue}
				description={item.description}
				balance={item.balance}
				expiresAt={item.expiresAt}
				currency={item.currency}
			/>

			<!-- Hint -->
			<p class="text-xs text-gray-500 text-center mt-4">
				{$t('dashboard.barcodeHint')}
			</p>
		</div>
	</div>
{/if}

<!-- Fullscreen Barcode Overlay (landscape on touch device) -->
{#if item && isLandscape}
	<div
		class="barcode-fullscreen-overlay"
		onclick={onClose}
		onkeydown={(e) => e.key === 'Escape' && onClose()}
		role="button"
		tabindex="0"
	>
		<div class="barcode-content-wrapper">
			<!-- Header: Merchant Name + Status/Validity Info -->
			<div class="barcode-header-section">
				{#if item.merchantName}
					<h2 class="text-lg font-bold text-gray-900">
						{item.merchantName}
					</h2>
				{/if}
				<div class="barcode-header-info">
					{#if item.status === 'valid'}
						<span
							class="inline-block px-3 py-1 text-xs rounded-full bg-green-100 text-green-800"
						>
							{$t('vouchers.status.valid')}
						</span>
					{/if}
					{#if item.validFrom || item.validUntil}
						<span class="text-sm text-gray-600">
							{#if item.validFrom && item.validUntil}
								{new Date(item.validFrom.split('T')[0]).toLocaleDateString(
									currentLocale
								)} - {new Date(
									item.validUntil.split('T')[0]
								).toLocaleDateString(currentLocale)}
							{:else if item.validUntil}
								{$t('vouchers.validUntil')}: {new Date(
									item.validUntil.split('T')[0]
								).toLocaleDateString(currentLocale)}
							{:else if item.validFrom}
								{$t('vouchers.validFrom')}: {new Date(
									item.validFrom.split('T')[0]
								).toLocaleDateString(currentLocale)}
							{/if}
						</span>
					{/if}
					{#if item.balance && item.currency}
						<span class="font-mono text-base font-bold text-green-600">
							{formatCurrency(parseFloat(item.balance), item.currency, $locale)}
						</span>
					{/if}
					{#if item.expiresAt}
						<span class="text-sm text-gray-600">
							{$t('giftCards.expiresAt')}: {new Date(
								item.expiresAt.split('T')[0]
							).toLocaleDateString(currentLocale)}
						</span>
					{/if}
				</div>
			</div>

			<!-- Barcode (large) -->
			<div class="barcode-container">
				<Barcode
					value={item.value}
					type={item.barcodeType || 'CODE128'}
					width={4}
					height={100}
					displayValue={false}
				/>
				<p class="font-mono text-base font-semibold text-gray-900">
					{item.value}
				</p>
			</div>

			<!-- Footer: PIN + Value + Description -->
			<div class="barcode-footer-section">
				{#if item.pin}
					<div class="flex items-center justify-center gap-2">
						<span class="text-sm text-gray-600">{$t('giftCards.pin')}:</span>
						<span class="font-mono text-lg font-semibold text-gray-900"
							>{item.pin}</span
						>
					</div>
				{/if}
				{#if item.displayValue || item.description}
					<div class="flex items-baseline justify-center gap-4 flex-wrap">
						{#if item.displayValue}
							<span class="font-mono text-lg font-semibold text-gray-900"
								>{item.displayValue}</span
							>
						{/if}
						{#if item.description}
							<span class="text-sm text-gray-600">{item.description}</span>
						{/if}
					</div>
				{/if}
			</div>
		</div>
		<p class="text-sm text-gray-500 text-center w-full py-2">
			{$t('dashboard.tapToClose')}
		</p>
	</div>
{/if}

<style>
	.barcode-fullscreen-overlay {
		position: fixed;
		inset: 0;
		z-index: 9999;
		background: #fff;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
		padding: 1rem;
		touch-action: manipulation;
		overflow: hidden;
	}

	.barcode-fullscreen-overlay .barcode-content-wrapper {
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		align-items: center;
		width: 100%;
		max-width: 600px;
		height: 100%;
	}

	.barcode-fullscreen-overlay .barcode-header-section {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: flex-start;
		width: 100%;
		min-height: 3rem;
		flex-shrink: 0;
		gap: 0.5rem;
	}

	.barcode-fullscreen-overlay .barcode-header-info {
		display: flex;
		flex-direction: row;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		flex-wrap: wrap;
		width: 100%;
	}

	.barcode-fullscreen-overlay .barcode-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex: 1;
		width: 100%;
		gap: 0.5rem;
	}

	.barcode-fullscreen-overlay .barcode-container :global(canvas) {
		max-height: 160px !important;
		height: 160px !important;
		width: auto !important;
		object-fit: contain;
	}

	.barcode-fullscreen-overlay .barcode-footer-section {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: flex-end;
		width: 100%;
		min-height: 3rem;
		flex-shrink: 0;
		gap: 0.5rem;
	}
</style>
