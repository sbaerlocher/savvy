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
	import Barcode from '$lib/components/Barcode.svelte';

	let {
		item,
		onClose
	}: {
		item: BarcodeModalItem | null;
		onClose: () => void;
	} = $props();

	const currentLocale = $derived($locale || 'de-DE');

	// Orientation detection. On touch devices in portrait we rotate the barcode
	// 90° so the user turns the phone sideways to scan. Desktop never rotates.
	let isLandscape = $state(false);
	let isTouchDevice = $state(false);
	let resizeTimer: ReturnType<typeof setTimeout> | null = null;

	// Rotate only on a touch device held in portrait.
	const rotateBarcode = $derived(isTouchDevice && !isLandscape);

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
		isTouchDevice = 'ontouchstart' in window;
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

<!-- Fullscreen barcode overlay. Always opens in this large view; the barcode
     is rotated 90° when the device is in portrait so the user turns the phone
     sideways to scan (no waiting for a physical orientation change). -->
{#if item}
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
					<h2 class="text-lg font-bold text-text">
						{item.merchantName}
					</h2>
				{/if}
				<div class="barcode-header-info">
					{#if item.status === 'valid'}
						<span
							class="inline-block px-3 py-1 text-xs rounded-full bg-success-100 text-success-800"
						>
							{$t('vouchers.status.valid')}
						</span>
					{/if}
					{#if item.validFrom || item.validUntil}
						<span class="text-sm text-text-muted">
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
						<span class="font-mono text-base font-bold text-success-600">
							{formatCurrency(parseFloat(item.balance), item.currency, $locale)}
						</span>
					{/if}
					{#if item.expiresAt}
						<span class="text-sm text-text-muted">
							{$t('giftCards.expiresAt')}: {new Date(
								item.expiresAt.split('T')[0]
							).toLocaleDateString(currentLocale)}
						</span>
					{/if}
				</div>
			</div>

			<!-- Barcode (large) -->
			<div class="barcode-container" class:is-portrait={rotateBarcode}>
				<Barcode
					value={item.value}
					type={item.barcodeType || 'CODE128'}
					width={4}
					height={100}
					displayValue={false}
				/>
				<p class="font-mono text-base font-semibold text-text">
					{item.value}
				</p>
			</div>

			<!-- Footer: PIN + Value + Description -->
			<div class="barcode-footer-section">
				{#if item.pin}
					<div class="flex items-center justify-center gap-2">
						<span class="text-sm text-text-muted">{$t('giftCards.pin')}:</span>
						<span class="font-mono text-lg font-semibold text-text"
							>{item.pin}</span
						>
					</div>
				{/if}
				{#if item.displayValue || item.description}
					<div class="flex items-baseline justify-center gap-4 flex-wrap">
						{#if item.displayValue}
							<span class="font-mono text-lg font-semibold text-text"
								>{item.displayValue}</span
							>
						{/if}
						{#if item.description}
							<span class="text-sm text-text-muted">{item.description}</span>
						{/if}
					</div>
				{/if}
			</div>
		</div>
		<p class="text-sm text-text-subtle text-center w-full py-2">
			{$t('dashboard.tapToClose')}
		</p>
	</div>
{/if}

<style>
	.barcode-fullscreen-overlay {
		position: fixed;
		inset: 0;
		z-index: 9999;
		background: var(--color-surface);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
		/* Fullscreen overlay pins content top and bottom, so pad past the notch
		   and home indicator on devices with a safe area. */
		padding: max(1rem, env(safe-area-inset-top))
			max(1rem, env(safe-area-inset-right))
			max(1rem, env(safe-area-inset-bottom))
			max(1rem, env(safe-area-inset-left));
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

	/* Portrait on touch: rotate the barcode + its value 90° so the widest
	   dimension runs down the (taller) screen — user turns the phone sideways. */
	.barcode-fullscreen-overlay .barcode-container.is-portrait {
		transform: rotate(90deg);
		/* Swap the box's effective axes so the rotated content uses the tall side. */
		width: 100vh;
		height: auto;
		flex: none;
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
