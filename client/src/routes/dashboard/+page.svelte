<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import { isOnline } from '$lib/stores/offline';
	import { t, locale } from '$lib/stores/i18n';
	import { dashboardApi } from '$lib/api';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { formatCurrency } from '$lib/utils/currency';
	import { categoryColors } from '$lib/utils/category-colors';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import BarcodeDisplay from '$lib/components/BarcodeDisplay.svelte';
	import Barcode from '$lib/components/Barcode.svelte';
	import type {
		DashboardResponse,
		CardDTO,
		VoucherDTO,
		GiftCardDTO
	} from '$lib/types/api';

	const pageLogger = logger.child('DashboardPage');
	const currentLocale = $derived($locale || 'de-DE');

	let data = $state<DashboardResponse | null>(null);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let error = $state<string | null>(null);

	// Quick Barcode Modal state
	type BarcodeModalItem = {
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
	let barcodeModalItem = $state<BarcodeModalItem | null>(null);

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

	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto(resolve('/login'));
			return;
		}

		// Setup orientation detection on touch devices
		const isTouchDevice = 'ontouchstart' in window;
		if (isTouchDevice) {
			checkOrientation();
			window.addEventListener('orientationchange', handleOrientationChange);
			window.addEventListener('resize', handleResize);
		}

		await loadDashboard();
	});

	onDestroy(() => {
		window.removeEventListener('orientationchange', handleOrientationChange);
		window.removeEventListener('resize', handleResize);
		if (resizeTimer) clearTimeout(resizeTimer);
	});

	async function loadDashboard() {
		isLoading = true;
		error = null;
		try {
			// Phase 1: Show cached data immediately
			const cached = await dashboardApi.getCached();
			if (cached) {
				data = cached;
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					data = await dashboardApi.get();
				} catch (err) {
					if (!cached) {
						error = $t('common.error');
						pageLogger.error('Failed to load dashboard', { error: err });
					}
				}
			} else if (!cached) {
				error = $t('common.error');
			}
		} catch (err) {
			error = $t('common.error');
			pageLogger.error('Failed to load dashboard', { error: err });
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	// Quick Barcode Modal functions
	function showCardBarcode(card: CardDTO, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		barcodeModalItem = {
			type: 'card',
			value: card.card_number,
			barcodeType: card.barcode_type,
			merchantName: card.merchant?.name
		};
		pageLogger.debug('Opening barcode modal for card:', card.id);
	}

	function showVoucherBarcode(voucher: VoucherDTO, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		const displayValue =
			voucher.type === 'percentage'
				? `${voucher.value}%`
				: voucher.type === 'fixed_amount'
					? formatCurrency(voucher.value, voucher.currency, $locale)
					: `${voucher.value}x Punkte`;

		barcodeModalItem = {
			type: 'voucher',
			value: voucher.code,
			barcodeType: voucher.barcode_type,
			merchantName: voucher.merchant?.name,
			displayValue,
			description: voucher.description,
			validFrom: voucher.valid_from,
			validUntil: voucher.valid_until,
			status: voucher.status
		};
		pageLogger.debug('Opening barcode modal for voucher:', voucher.id);
	}

	function showGiftCardBarcode(giftCard: GiftCardDTO, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		barcodeModalItem = {
			type: 'gift_card',
			value: giftCard.card_number,
			barcodeType: giftCard.barcode_type,
			merchantName: giftCard.merchant?.name,
			pin: giftCard.pin,
			balance: giftCard.current_balance.toFixed(2),
			currency: giftCard.currency,
			expiresAt: giftCard.expires_at || undefined,
			status: giftCard.status
		};
		pageLogger.debug('Opening barcode modal for gift card:', giftCard.id);
	}

	function closeBarcodeModal() {
		barcodeModalItem = null;
		pageLogger.debug('Closed barcode modal');
	}

	const isOffline = $derived(!$isOnline);
</script>

<svelte:head>
	<title>{$t('dashboard.title')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 max-w-7xl mx-auto">
	{#if isLoading}
		<LoadingSpinner />
	{:else if error}
		<div class="bg-white rounded-lg shadow-md p-6 text-center">
			<p class="text-red-600 mb-4">{error}</p>
			<button onclick={loadDashboard} class="btn btn-primary"
				>{$t('common.retry')}</button
			>
		</div>
	{:else if data}
		<!-- Welcome Header with inline Stats -->
		<div class="mb-8">
			<div class="flex items-center gap-3">
				<h1 class="text-3xl font-bold text-gray-900">
					{$t('dashboard.welcome', { name: $authStore.user?.first_name || '' })}
				</h1>
				{#if isRefreshing}
					<span class="text-xs text-gray-400 animate-pulse"
						>{$t('common.refreshing')}</span
					>
				{/if}
			</div>
			<div
				class="mt-2 hidden lg:flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-600"
			>
				<a
					href={resolve('/cards')}
					data-testid="dashboard-stat-cards"
					class="hover:text-cyan-600 transition"
				>
					<span class="font-semibold text-gray-900"
						>{data.stats.cards_count}</span
					>
					{$t('dashboard.stats.cards')}
				</a>
				<span class="text-gray-300">|</span>
				<a
					href={resolve('/vouchers')}
					data-testid="dashboard-stat-vouchers"
					class="hover:text-green-600 transition"
				>
					<span class="font-semibold text-gray-900"
						>{data.stats.vouchers_count}</span
					>
					{$t('dashboard.stats.vouchers')}
				</a>
				<span class="text-gray-300">|</span>
				<a
					href={resolve('/gift-cards')}
					data-testid="dashboard-stat-gift-cards"
					class="hover:text-red-600 transition"
				>
					<span class="font-semibold text-gray-900"
						>{data.stats.gift_cards_count}</span
					>
					{$t('dashboard.stats.giftCards')}
				</a>
				{#if data.stats.total_balance > 0}
					<span class="text-gray-300">|</span>
					<span>
						<span class="font-semibold text-gray-900"
							>CHF {Math.round(data.stats.total_balance)}</span
						>
						{$t('dashboard.stats.totalBalance')}
					</span>
				{/if}
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Left column: Favorites (2/3 width) -->
			<div class="lg:col-span-2 space-y-6">
				<!-- Favorites / Recently Added -->
				<div
					class="bg-white rounded-lg shadow-md p-6"
					data-testid="favorites-section"
				>
					<h2 class="text-xl font-semibold text-gray-900 mb-4">
						{#if data.has_favorites}
							{$t('dashboard.favorites')}
						{:else}
							{$t('dashboard.recentlyAdded')}
						{/if}
					</h2>
					<div data-testid="favorites-list">
						{#if data.recent_cards.length === 0 && data.recent_vouchers.length === 0 && data.recent_gift_cards.length === 0}
							<div class="text-center py-8">
								<p class="text-gray-500 text-sm">
									{$t('dashboard.noActivity')}
								</p>
								<p class="text-gray-400 text-xs mt-1 mb-4">
									{$t('dashboard.noActivityHint')}
								</p>
								<a
									href={resolve('/cards/new')}
									class="inline-flex items-center gap-1 text-sm font-medium text-cyan-600 hover:text-cyan-800 transition"
								>
									{$t('dashboard.getStarted')} →
								</a>
							</div>
						{:else}
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
								<!-- Cards -->
								{#if data.recent_cards.length > 0 && (!data.has_favorites || data.has_card_favorites)}
									{#each data.recent_cards.slice(0, 3) as card (card.id)}
										<div
											class="group flex rounded-xl border border-gray-100 bg-gray-50/60 overflow-hidden hover:shadow-md hover:bg-white transition"
										>
											<div
												class="w-1.5 flex-shrink-0"
												style="background-color: {card.merchant?.color ||
													'#6B7280'}"
											></div>
											<a
												href={resolve(`/cards/${card.id}`)}
												class="p-3 flex-1 min-w-0"
											>
												<p
													class="font-semibold text-gray-900 text-sm truncate group-hover:text-cyan-600 transition"
												>
													{card.merchant?.name || $t('dashboard.cardType')}
												</p>
												<div
													class="flex items-baseline justify-between gap-2 mt-1"
												>
													{#if card.card_number}
														<p class="text-xs text-gray-500 font-mono">
															····{card.card_number.slice(-4)}
														</p>
													{/if}
													{#if card.owner && card.owner.id !== $authStore.user?.id}
														<p class="text-xs text-gray-400 truncate">
															{$t('dashboard.sharedBy', {
																name:
																	card.owner.first_name ||
																	card.owner.email ||
																	'User'
															})}
														</p>
													{/if}
												</div>
											</a>
											<button
												onclick={(e) => showCardBarcode(card, e)}
												class="p-3 flex-shrink-0 text-gray-400 hover:text-cyan-600 hover:bg-cyan-50 transition"
												title={$t('dashboard.showBarcode')}
												aria-label={$t('dashboard.showBarcode')}
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
														d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"
													/>
												</svg>
											</button>
										</div>
									{/each}
								{/if}

								<!-- Vouchers -->
								{#if data.recent_vouchers.length > 0 && (!data.has_favorites || data.has_voucher_favorites)}
									{#each data.recent_vouchers.slice(0, 3) as voucher (voucher.id)}
										<div
											class="group flex rounded-xl border border-gray-100 bg-gray-50/60 overflow-hidden hover:shadow-md hover:bg-white transition"
										>
											<div
												class="w-1.5 flex-shrink-0"
												style="background-color: {voucher.merchant?.color ||
													'#6B7280'}"
											></div>
											<a
												href={resolve(`/vouchers/${voucher.id}`)}
												class="p-3 flex-1 min-w-0"
											>
												<p
													class="font-semibold text-gray-900 text-sm truncate group-hover:text-green-600 transition"
												>
													{voucher.merchant?.name ||
														$t('dashboard.voucherType')}
												</p>
												<div
													class="flex items-baseline justify-between gap-2 mt-1"
												>
													<p class="text-xs text-gray-600 font-medium">
														{#if voucher.type === 'percentage'}
															{voucher.value}%
														{:else if voucher.type === 'fixed_amount'}
															{formatCurrency(
																voucher.value,
																voucher.currency,
																$locale
															)}
														{:else if voucher.type === 'points_multiplier'}
															{voucher.value}{$t(
																'vouchers.types.pointsMultiplierDisplay'
															)}
														{:else if voucher.type === 'bonus_points'}
															+{voucher.value}{$t(
																'vouchers.types.bonusPointsDisplay'
															)}
														{/if}
													</p>
													{#if voucher.owner && voucher.owner.id !== $authStore.user?.id}
														<p class="text-xs text-gray-400 truncate">
															{$t('dashboard.sharedBy', {
																name:
																	voucher.owner.first_name ||
																	voucher.owner.email ||
																	'User'
															})}
														</p>
													{/if}
													{#if voucher.valid_until}
														<p
															class="text-xs text-orange-600 whitespace-nowrap"
														>
															{$t('dashboard.validUntil', {
																date: new Date(
																	voucher.valid_until.split('T')[0]
																).toLocaleDateString(currentLocale)
															})}
														</p>
													{/if}
												</div>
											</a>
											<button
												onclick={(e) => showVoucherBarcode(voucher, e)}
												class="p-3 flex-shrink-0 text-gray-400 hover:text-green-600 hover:bg-green-50 transition"
												title={$t('dashboard.showBarcode')}
												aria-label={$t('dashboard.showBarcode')}
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
														d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"
													/>
												</svg>
											</button>
										</div>
									{/each}
								{/if}

								<!-- Gift Cards -->
								{#if data.recent_gift_cards.length > 0 && (!data.has_favorites || data.has_gift_card_favorites)}
									{#each data.recent_gift_cards.slice(0, 3) as giftCard (giftCard.id)}
										<div
											class="group flex rounded-xl border border-gray-100 bg-gray-50/60 overflow-hidden hover:shadow-md hover:bg-white transition"
										>
											<div
												class="w-1.5 flex-shrink-0"
												style="background-color: {giftCard.merchant?.color ||
													'#6B7280'}"
											></div>
											<a
												href={resolve(`/gift-cards/${giftCard.id}`)}
												class="p-3 flex-1 min-w-0"
											>
												<p
													class="font-semibold text-gray-900 text-sm truncate group-hover:text-red-600 transition"
												>
													{giftCard.merchant?.name ||
														$t('dashboard.giftCardType')}
												</p>
												<div
													class="flex items-baseline justify-between gap-2 mt-1"
												>
													<p class="text-sm text-green-600 font-semibold">
														{formatCurrency(
															giftCard.current_balance,
															giftCard.currency,
															$locale
														)}
													</p>
													{#if giftCard.owner && giftCard.owner.id !== $authStore.user?.id}
														<p class="text-xs text-gray-400 truncate">
															{$t('dashboard.sharedBy', {
																name:
																	giftCard.owner.first_name ||
																	giftCard.owner.email ||
																	'User'
															})}
														</p>
													{/if}
													{#if giftCard.expires_at}
														<p
															class="text-xs text-orange-600 whitespace-nowrap"
														>
															{$t('dashboard.validUntil', {
																date: new Date(
																	giftCard.expires_at.split('T')[0]
																).toLocaleDateString(currentLocale)
															})}
														</p>
													{/if}
												</div>
											</a>
											<button
												onclick={(e) => showGiftCardBarcode(giftCard, e)}
												class="p-3 flex-shrink-0 text-gray-400 hover:text-red-600 hover:bg-red-50 transition"
												title={$t('dashboard.showBarcode')}
												aria-label={$t('dashboard.showBarcode')}
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
														d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"
													/>
												</svg>
											</button>
										</div>
									{/each}
								{/if}
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Right column: Quick Actions + Tips (1/3 width) -->
			<div class="lg:col-span-1 space-y-6">
				<!-- Quick Actions -->
				<div class="bg-white rounded-lg shadow-md p-6">
					<h2 class="text-xl font-semibold text-gray-900 mb-3">
						{$t('dashboard.quickActions')}
					</h2>
					<div class="flex flex-col gap-2">
						<a
							href={resolve('/cards/new')}
							onclick={(e) => {
								if (isOffline) e.preventDefault();
							}}
							class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition {isOffline
								? 'opacity-50 cursor-not-allowed bg-gray-50 text-gray-400'
								: categoryColors.cards.action}"
						>
							<span>+</span>
							{$t('dashboard.addCard')}
						</a>
						<a
							href={resolve('/vouchers/new')}
							onclick={(e) => {
								if (isOffline) e.preventDefault();
							}}
							class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition {isOffline
								? 'opacity-50 cursor-not-allowed bg-gray-50 text-gray-400'
								: categoryColors.vouchers.action}"
						>
							<span>+</span>
							{$t('dashboard.addVoucher')}
						</a>
						<a
							href={resolve('/gift-cards/new')}
							onclick={(e) => {
								if (isOffline) e.preventDefault();
							}}
							class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition {isOffline
								? 'opacity-50 cursor-not-allowed bg-gray-50 text-gray-400'
								: categoryColors.giftCards.action}"
						>
							<span>+</span>
							{$t('dashboard.addGiftCard')}
						</a>
					</div>
				</div>

				<!-- Tips (Desktop only) -->
				<div class="bg-white rounded-lg shadow-md p-6 hidden lg:block">
					<h2 class="text-lg font-semibold text-gray-900 mb-3">
						{$t('dashboard.tipsTitle')}
					</h2>
					<div class="space-y-3">
						<div class="flex items-start gap-2">
							<svg
								class="w-5 h-5 text-cyan-500 mt-0.5 flex-shrink-0"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
								/>
							</svg>
							<p class="text-xs text-gray-600">{$t('dashboard.tip1')}</p>
						</div>
						<div class="flex items-start gap-2">
							<svg
								class="w-5 h-5 text-green-500 mt-0.5 flex-shrink-0"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
								/>
							</svg>
							<p class="text-xs text-gray-600">{$t('dashboard.tip2')}</p>
						</div>
						<div class="flex items-start gap-2">
							<svg
								class="w-5 h-5 text-yellow-500 mt-0.5 flex-shrink-0"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
								/>
							</svg>
							<p class="text-xs text-gray-600">{$t('dashboard.tip3')}</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Quick Barcode Modal (portrait / desktop) -->
{#if barcodeModalItem && !isLandscape}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
		onclick={closeBarcodeModal}
		onkeydown={(e) => e.key === 'Escape' && closeBarcodeModal()}
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
				onclick={closeBarcodeModal}
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
			{#if barcodeModalItem.merchantName}
				<h2
					id="barcode-modal-title"
					class="text-xl font-bold text-gray-900 mb-4 pr-8"
				>
					{barcodeModalItem.merchantName}
				</h2>
			{/if}

			<!-- Barcode Display -->
			<BarcodeDisplay
				value={barcodeModalItem.value}
				type={barcodeModalItem.barcodeType}
				status={barcodeModalItem.status}
				pin={barcodeModalItem.pin}
				validFrom={barcodeModalItem.validFrom}
				validUntil={barcodeModalItem.validUntil}
				displayValue={barcodeModalItem.displayValue}
				description={barcodeModalItem.description}
				balance={barcodeModalItem.balance}
				expiresAt={barcodeModalItem.expiresAt}
				currency={barcodeModalItem.currency}
			/>

			<!-- Hint -->
			<p class="text-xs text-gray-500 text-center mt-4">
				{$t('dashboard.barcodeHint')}
			</p>
		</div>
	</div>
{/if}

<!-- Fullscreen Barcode Overlay (landscape on touch device) -->
{#if barcodeModalItem && isLandscape}
	<div
		class="barcode-fullscreen-overlay"
		onclick={closeBarcodeModal}
		onkeydown={(e) => e.key === 'Escape' && closeBarcodeModal()}
		role="button"
		tabindex="0"
	>
		<div class="barcode-content-wrapper">
			<!-- Header: Merchant Name + Status/Validity Info -->
			<div class="barcode-header-section">
				{#if barcodeModalItem.merchantName}
					<h2 class="text-lg font-bold text-gray-900">
						{barcodeModalItem.merchantName}
					</h2>
				{/if}
				<div class="barcode-header-info">
					{#if barcodeModalItem.status === 'valid'}
						<span
							class="inline-block px-3 py-1 text-xs rounded-full bg-green-100 text-green-800"
						>
							{$t('vouchers.status.valid')}
						</span>
					{/if}
					{#if barcodeModalItem.validFrom || barcodeModalItem.validUntil}
						<span class="text-sm text-gray-600">
							{#if barcodeModalItem.validFrom && barcodeModalItem.validUntil}
								{new Date(
									barcodeModalItem.validFrom.split('T')[0]
								).toLocaleDateString(currentLocale)} - {new Date(
									barcodeModalItem.validUntil.split('T')[0]
								).toLocaleDateString(currentLocale)}
							{:else if barcodeModalItem.validUntil}
								{$t('vouchers.validUntil')}: {new Date(
									barcodeModalItem.validUntil.split('T')[0]
								).toLocaleDateString(currentLocale)}
							{:else if barcodeModalItem.validFrom}
								{$t('vouchers.validFrom')}: {new Date(
									barcodeModalItem.validFrom.split('T')[0]
								).toLocaleDateString(currentLocale)}
							{/if}
						</span>
					{/if}
					{#if barcodeModalItem.balance && barcodeModalItem.currency}
						<span class="font-mono text-base font-bold text-green-600">
							{formatCurrency(
								parseFloat(barcodeModalItem.balance),
								barcodeModalItem.currency,
								$locale
							)}
						</span>
					{/if}
					{#if barcodeModalItem.expiresAt}
						<span class="text-sm text-gray-600">
							{$t('giftCards.expiresAt')}: {new Date(
								barcodeModalItem.expiresAt.split('T')[0]
							).toLocaleDateString(currentLocale)}
						</span>
					{/if}
				</div>
			</div>

			<!-- Barcode (large) -->
			<div class="barcode-container">
				<Barcode
					value={barcodeModalItem.value}
					type={barcodeModalItem.barcodeType || 'CODE128'}
					width={4}
					height={100}
					displayValue={false}
				/>
				<p class="font-mono text-base font-semibold text-gray-900">
					{barcodeModalItem.value}
				</p>
			</div>

			<!-- Footer: PIN + Value + Description -->
			<div class="barcode-footer-section">
				{#if barcodeModalItem.pin}
					<div class="flex items-center justify-center gap-2">
						<span class="text-sm text-gray-600">{$t('giftCards.pin')}:</span>
						<span class="font-mono text-lg font-semibold text-gray-900"
							>{barcodeModalItem.pin}</span
						>
					</div>
				{/if}
				{#if barcodeModalItem.displayValue || barcodeModalItem.description}
					<div class="flex items-baseline justify-center gap-4 flex-wrap">
						{#if barcodeModalItem.displayValue}
							<span class="font-mono text-lg font-semibold text-gray-900"
								>{barcodeModalItem.displayValue}</span
							>
						{/if}
						{#if barcodeModalItem.description}
							<span class="text-sm text-gray-600"
								>{barcodeModalItem.description}</span
							>
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
