<script lang="ts">
	import { t, locale } from '$lib/stores/i18n';
	import { authStore } from '$lib/stores/auth';
	import { resolve } from '$app/paths';
	import { formatCurrency } from '$lib/utils/currency';
	import { logger } from '$lib/utils/logger';
	import type {
		DashboardResponse,
		CardDTO,
		VoucherDTO,
		GiftCardDTO
	} from '$lib/types/api';
	import type { BarcodeModalItem } from './BarcodeModal.svelte';

	let {
		data,
		onShowBarcode
	}: {
		data: DashboardResponse;
		onShowBarcode: (item: BarcodeModalItem) => void;
	} = $props();

	const pageLogger = logger.child('FavoritesSection');
	const currentLocale = $derived($locale || 'de-DE');

	function showCardBarcode(card: CardDTO, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		onShowBarcode({
			type: 'card',
			value: card.card_number,
			barcodeType: card.barcode_type,
			merchantName: card.merchant?.name
		});
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

		onShowBarcode({
			type: 'voucher',
			value: voucher.code,
			barcodeType: voucher.barcode_type,
			merchantName: voucher.merchant?.name,
			displayValue,
			description: voucher.description,
			validFrom: voucher.valid_from,
			validUntil: voucher.valid_until,
			status: voucher.status
		});
		pageLogger.debug('Opening barcode modal for voucher:', voucher.id);
	}

	function showGiftCardBarcode(giftCard: GiftCardDTO, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		onShowBarcode({
			type: 'gift_card',
			value: giftCard.card_number,
			barcodeType: giftCard.barcode_type,
			merchantName: giftCard.merchant?.name,
			pin: giftCard.pin,
			balance: giftCard.current_balance.toFixed(2),
			currency: giftCard.currency,
			expiresAt: giftCard.expires_at || undefined,
			status: giftCard.status
		});
		pageLogger.debug('Opening barcode modal for gift card:', giftCard.id);
	}
</script>

<div class="bg-white rounded-lg shadow-md p-6" data-testid="favorites-section">
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
							class="group flex rounded-lg bg-white shadow-lg hover:shadow-xl overflow-hidden transition"
							style="border-left: 6px solid {card.merchant?.color || '#6B7280'}"
						>
							<a href={resolve(`/cards/${card.id}`)} class="p-3 flex-1 min-w-0">
								<p
									class="font-semibold text-gray-900 text-sm truncate group-hover:text-cyan-600 transition"
								>
									{card.merchant?.name || $t('dashboard.cardType')}
								</p>
								<div class="flex items-baseline justify-between gap-2 mt-1">
									{#if card.card_number}
										<p class="text-xs text-gray-500 font-mono">
											····{card.card_number.slice(-4)}
										</p>
									{/if}
									{#if card.owner && card.owner.id !== $authStore.user?.id}
										<p class="text-xs text-gray-400 truncate">
											{$t('dashboard.sharedBy', {
												name:
													card.owner.first_name || card.owner.email || 'User'
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
							class="group flex rounded-lg bg-white shadow-lg hover:shadow-xl overflow-hidden transition"
							style="border-left: 6px solid {voucher.merchant?.color ||
								'#6B7280'}"
						>
							<a
								href={resolve(`/vouchers/${voucher.id}`)}
								class="p-3 flex-1 min-w-0"
							>
								<p
									class="font-semibold text-gray-900 text-sm truncate group-hover:text-green-600 transition"
								>
									{voucher.merchant?.name || $t('dashboard.voucherType')}
								</p>
								<div class="flex items-baseline justify-between gap-2 mt-1">
									<p class="text-xs text-gray-600 font-medium">
										{#if voucher.type === 'percentage'}
											{voucher.value}%
										{:else if voucher.type === 'fixed_amount'}
											{formatCurrency(voucher.value, voucher.currency, $locale)}
										{:else if voucher.type === 'points_multiplier'}
											{voucher.value}{$t(
												'vouchers.types.pointsMultiplierDisplay'
											)}
										{:else if voucher.type === 'bonus_points'}
											+{voucher.value}{$t('vouchers.types.bonusPointsDisplay')}
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
										<p class="text-xs text-orange-600 whitespace-nowrap">
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
							class="group flex rounded-lg bg-white shadow-lg hover:shadow-xl overflow-hidden transition"
							style="border-left: 6px solid {giftCard.merchant?.color ||
								'#6B7280'}"
						>
							<a
								href={resolve(`/gift-cards/${giftCard.id}`)}
								class="p-3 flex-1 min-w-0"
							>
								<p
									class="font-semibold text-gray-900 text-sm truncate group-hover:text-red-600 transition"
								>
									{giftCard.merchant?.name || $t('dashboard.giftCardType')}
								</p>
								<div class="flex items-baseline justify-between gap-2 mt-1">
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
										<p class="text-xs text-orange-600 whitespace-nowrap">
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
