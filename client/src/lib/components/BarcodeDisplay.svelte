<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Barcode from './Barcode.svelte';
	import { get } from 'svelte/store';
	import { t, locale } from '$lib/stores/i18n';
	import { formatCurrency } from '$lib/utils/currency';
	import { logger } from '$lib/utils/logger';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const componentLogger = logger.child('BarcodeDisplay');

	// Get current locale for date formatting
	const currentLocale = $derived($locale || 'de-DE');

	interface Props {
		value: string;
		type?: string;
		height?: number;
		status?: string;
		statusBadge?: { class: string; text: string };
		pin?: string;
		validFrom?: string;
		validUntil?: string;
		minPurchaseInfo?: string;
		displayValue?: string;
		description?: string;
		balance?: string;
		expiresAt?: string;
		currency?: string;
	}

	let {
		value,
		type = 'CODE128',
		height = 128,
		status,
		statusBadge,
		pin,
		validFrom,
		validUntil,
		minPurchaseInfo,
		displayValue,
		description,
		balance,
		expiresAt,
		currency
	}: Props = $props();

	const hasValidityInfo = $derived(validFrom || validUntil);
	const showStatusBadge = $derived(status && status !== 'active');
	const showValidStatusBadge = $derived(status === 'valid');

	// Fullscreen state
	let isLandscape = $state(false);
	// Explicit user-triggered fullscreen (tap/click/keyboard), independent of
	// device orientation — works on desktop and portrait phones too.
	let manualFullscreen = $state(false);
	let resizeTimer: ReturnType<typeof setTimeout> | null = null;

	const showFullscreen = $derived(isLandscape || manualFullscreen);

	// 2D codes (QR, PDF417, …) stay square and benefit from a much larger
	// fullscreen size than the wide-but-short 1D barcodes.
	const is2DType = $derived(
		['QR', 'QRCODE', 'PDF417', 'DATAMATRIX', 'AZTEC', 'MAXICODE'].includes(
			type.toUpperCase()
		)
	);

	onMount(() => {
		// Only enable orientation detection on touch devices
		const isTouchDevice = 'ontouchstart' in window;
		componentLogger.debug('Touch device detected:', isTouchDevice);

		if (!isTouchDevice) {
			componentLogger.debug(
				'Not a touch device, skipping orientation detection'
			);
			return;
		}

		checkOrientation();

		window.addEventListener('orientationchange', handleOrientationChange);
		window.addEventListener('resize', handleResize);
	});

	onDestroy(() => {
		window.removeEventListener('orientationchange', handleOrientationChange);
		window.removeEventListener('resize', handleResize);
		if (resizeTimer) clearTimeout(resizeTimer);
	});

	function handleOrientationChange() {
		setTimeout(checkOrientation, 100);
	}

	function handleResize() {
		if (resizeTimer) clearTimeout(resizeTimer);
		resizeTimer = setTimeout(checkOrientation, 150);
	}

	function checkOrientation() {
		// Check using screen orientation API first, then fallback to dimensions
		const screenOrientationType = window.screen.orientation?.type;
		const isLandscapeByAPI =
			screenOrientationType?.includes('landscape') || false;
		const isLandscapeByDimensions = window.innerWidth > window.innerHeight;

		const isLandscapeOrientation = isLandscapeByAPI || isLandscapeByDimensions;
		const newIsLandscape = isLandscapeOrientation;

		componentLogger.debug('Orientation check:', {
			screenOrientationType,
			isLandscapeByAPI,
			isLandscapeByDimensions,
			windowWidth: window.innerWidth,
			windowHeight: window.innerHeight,
			newIsLandscape,
			currentIsLandscape: isLandscape
		});

		if (newIsLandscape !== isLandscape) {
			componentLogger.debug(
				'Orientation changed to:',
				newIsLandscape ? 'LANDSCAPE' : 'PORTRAIT'
			);
			isLandscape = newIsLandscape;
		}
	}

	function closeFullscreen() {
		isLandscape = false;
		manualFullscreen = false;
	}

	function openFullscreen() {
		manualFullscreen = true;
	}

	// Move focus to the overlay when it opens so a keyboard user can press
	// Escape to close it (the Escape handler only fires when the overlay has
	// focus). Without this, opening via the tap-to-enlarge button leaves focus
	// on the now-obscured button.
	let overlayEl = $state<HTMLDivElement>();
	$effect(() => {
		if (showFullscreen && overlayEl) {
			overlayEl.focus();
		}
	});
</script>

<div class="bg-surface-1 rounded-lg p-4 text-center border-t border-border">
	<!-- Status Badge + Validity Period (for Vouchers) -->
	{#if showValidStatusBadge || showStatusBadge || hasValidityInfo}
		<div class="flex items-center justify-center gap-4 mb-4 flex-wrap">
			<!-- Valid Status Badge (green for vouchers) -->
			{#if showValidStatusBadge}
				<span
					class="inline-block px-3 py-1 text-xs rounded-full bg-green-100 text-green-800"
				>
					{tr('vouchers.status.valid')}
				</span>
			{:else if showStatusBadge && statusBadge}
				<!-- Other Status Badges -->
				<span
					class="inline-block px-3 py-1 text-xs rounded-full {statusBadge.class}"
				>
					{statusBadge.text}
				</span>
			{/if}

			<!-- Validity Period (for vouchers) -->
			{#if hasValidityInfo}
				<span class="text-xs text-text-muted">
					{#if validFrom && validUntil}
						{new Date(validFrom.split('T')[0]).toLocaleDateString(
							currentLocale
						)} - {new Date(validUntil.split('T')[0]).toLocaleDateString(
							currentLocale
						)}
					{:else if validUntil}
						{tr('vouchers.validUntil')}: {new Date(
							validUntil.split('T')[0]
						).toLocaleDateString(currentLocale)}
					{:else if validFrom}
						{tr('vouchers.validFrom')}: {new Date(
							validFrom.split('T')[0]
						).toLocaleDateString(currentLocale)}
					{/if}
					{#if minPurchaseInfo}
						· {minPurchaseInfo}
					{/if}
				</span>
			{:else if minPurchaseInfo}
				<span class="text-xs text-text-muted">{minPurchaseInfo}</span>
			{/if}
		</div>
	{/if}

	<!-- Barcode Image (tap to enlarge for scanning) -->
	<div class="flex justify-center mb-4">
		<button
			type="button"
			class="barcode-zoom-button"
			onclick={openFullscreen}
			aria-label={tr('dashboard.tapToEnlarge')}
			title={tr('dashboard.tapToEnlarge')}
		>
			<Barcode {value} {type} {height} />
		</button>
	</div>

	<!-- Code/Number -->
	<p
		class="font-mono text-sm sm:text-base md:text-lg font-semibold text-text break-all"
	>
		{value}
	</p>

	<!-- Additional Info Section -->
	{#if displayValue || description || pin}
		<div class="mt-4">
			<!-- Value + Description (for vouchers) -->
			{#if displayValue || description}
				<div class="flex items-baseline justify-center gap-3 flex-wrap">
					{#if displayValue}
						<p class="font-mono text-lg font-bold text-text">
							{displayValue}
						</p>
					{/if}
					{#if description}
						<p class="text-xs text-text-muted">{description}</p>
					{/if}
				</div>
			{/if}

			<!-- PIN (for gift cards) -->
			{#if pin}
				<p class="text-center">
					<span class="text-xs text-text-muted">{tr('giftCards.pin')}: </span>
					<span class="font-mono text-base font-semibold text-text">{pin}</span>
				</p>
			{/if}
		</div>
	{/if}
</div>

<!-- Barcode Fullscreen Overlay (landscape on mobile, or tap to enlarge) -->
{#if showFullscreen}
	<div
		bind:this={overlayEl}
		class="barcode-fullscreen-overlay"
		onclick={closeFullscreen}
		role="button"
		tabindex="0"
		onkeydown={(e) => e.key === 'Escape' && closeFullscreen()}
	>
		<div class="barcode-content-wrapper">
			<!-- Status Badge + Validity (above barcode) -->
			<div class="barcode-header-section">
				{#if showValidStatusBadge || showStatusBadge || hasValidityInfo || balance || expiresAt}
					<div class="barcode-header-info">
						{#if showValidStatusBadge}
							<span
								class="inline-block px-3 py-1 text-xs rounded-full bg-green-100 text-green-800"
							>
								{tr('vouchers.status.valid')}
							</span>
						{:else if showStatusBadge && statusBadge}
							<span
								class="inline-block px-3 py-1 text-xs rounded-full {statusBadge.class}"
							>
								{statusBadge.text}
							</span>
						{/if}

						{#if hasValidityInfo}
							<span class="barcode-info-validity">
								{#if validFrom && validUntil}
									{new Date(validFrom.split('T')[0]).toLocaleDateString(
										currentLocale
									)} - {new Date(validUntil.split('T')[0]).toLocaleDateString(
										currentLocale
									)}
								{:else if validUntil}
									{tr('vouchers.validUntil')}: {new Date(
										validUntil.split('T')[0]
									).toLocaleDateString(currentLocale)}
								{:else if validFrom}
									{tr('vouchers.validFrom')}: {new Date(
										validFrom.split('T')[0]
									).toLocaleDateString(currentLocale)}
								{/if}
								{#if minPurchaseInfo}
									· {minPurchaseInfo}
								{/if}
							</span>
						{:else if minPurchaseInfo}
							<span class="barcode-info-validity">{minPurchaseInfo}</span>
						{/if}

						{#if balance && currency}
							<span class="barcode-info-balance">
								{formatCurrency(parseFloat(balance), currency, $locale)}
							</span>
						{/if}

						{#if expiresAt}
							<span class="barcode-info-validity">
								{tr('giftCards.expiresAt')}: {new Date(
									expiresAt.split('T')[0]
								).toLocaleDateString(currentLocale)}
							</span>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Barcode Container (fixed height) -->
			<div class="barcode-container" class:is-2d={is2DType}>
				<Barcode
					{value}
					{type}
					width={is2DType ? 6 : 4}
					height={100}
					displayValue={false}
				/>
				<p class="barcode-value">{value}</p>
			</div>

			<!-- Footer Section (PIN + Value + Description) -->
			<div class="barcode-footer-section">
				<!-- PIN (for gift cards) -->
				{#if pin}
					<div class="barcode-pin-row">
						<span class="barcode-pin-label">{tr('giftCards.pin')}:</span>
						<span class="barcode-pin-value">{pin}</span>
					</div>
				{/if}

				<!-- Value + Description (below barcode, side by side) -->
				{#if displayValue || description}
					<div class="barcode-value-desc-row">
						{#if displayValue}
							<span class="barcode-info-value">{displayValue}</span>
						{/if}
						{#if description}
							<span class="barcode-info-desc">{description}</span>
						{/if}
					</div>
				{/if}
			</div>
		</div>
		<p class="barcode-close-hint">{tr('dashboard.tapToClose')}</p>
	</div>
{/if}

<style>
	/* Tap-to-enlarge button wrapping the inline barcode — visually transparent,
	   just makes the code a hit target. */
	.barcode-zoom-button {
		appearance: none;
		background: none;
		border: none;
		padding: 0;
		margin: 0;
		cursor: zoom-in;
		display: inline-flex;
	}

	/* Barcode Fullscreen Overlay */
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

	/* 2D codes (QR etc.) are square: let them grow to fill the available space
	   instead of the short 1D barcode height, so dense long-content codes stay
	   scannable. */
	.barcode-fullscreen-overlay .barcode-container.is-2d :global(canvas) {
		height: auto !important;
		max-height: min(80vw, 80vh) !important;
		max-width: min(80vw, 80vh) !important;
		width: auto !important;
	}

	.barcode-fullscreen-overlay .barcode-footer-section {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: flex-end;
		width: 100%;
		min-height: 4rem;
		flex-shrink: 0;
		gap: 0.5rem;
	}

	.barcode-fullscreen-overlay .barcode-value {
		font-family: monospace;
		font-size: 1rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.barcode-fullscreen-overlay .barcode-pin-row {
		display: flex;
		flex-direction: row;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		min-height: 1.75rem;
		width: 100%;
	}

	.barcode-fullscreen-overlay .barcode-pin-label {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.barcode-fullscreen-overlay .barcode-pin-value {
		font-family: monospace;
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
	}

	.barcode-fullscreen-overlay .barcode-value-desc-row {
		display: flex;
		flex-direction: row;
		align-items: baseline;
		justify-content: center;
		gap: 1rem;
		margin-top: 0;
		flex-wrap: wrap;
		min-height: 2rem;
		width: 100%;
	}

	.barcode-fullscreen-overlay .barcode-info-value {
		font-family: monospace;
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
	}

	.barcode-fullscreen-overlay .barcode-info-desc {
		font-size: 0.875rem;
		color: #6b7280;
		text-align: center;
	}

	.barcode-fullscreen-overlay .barcode-info-validity {
		font-size: 0.875rem;
		color: #6b7280;
		text-align: center;
	}

	.barcode-fullscreen-overlay .barcode-info-balance {
		font-family: monospace;
		font-size: 1rem;
		font-weight: 700;
		color: #059669;
		text-align: center;
	}

	.barcode-fullscreen-overlay .barcode-close-hint {
		color: #6b7280;
		font-size: 0.875rem;
		text-align: center;
		flex-shrink: 0;
		width: 100%;
		padding: 0.5rem 0;
	}
</style>
