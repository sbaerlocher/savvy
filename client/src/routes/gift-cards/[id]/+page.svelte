<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { giftCardsApi, merchantsApi, ApiError } from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import GiftCardForm from '$lib/components/gift-cards/GiftCardForm.svelte';
	import GiftCardLedger from '$lib/components/gift-cards/GiftCardLedger.svelte';
	import type { GiftCardDTO, ShareDTO, MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ResourceDetail from '$lib/components/ui/ResourceDetail.svelte';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('GiftCardDetailsPage');

	const giftCardId = $derived($page.params.id);

	let giftCard = $state<GiftCardDTO | null>(null);
	let shares = $state<ShareDTO[]>([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Edit form fields
	let editMerchantId = $state('');
	let editCardNumber = $state('');
	let editInitialBalance = $state(0);
	let editCurrency = $state('CHF');
	let editPin = $state('');
	let editBarcodeType = $state('CODE128');
	let editExpiresAt = $state('');
	let editNotes = $state('');

	const isOffline = $derived(!$isOnline);

	onMount(async () => {
		await Promise.all([loadGiftCard(), loadMerchants()]);
	});

	async function loadGiftCard() {
		isLoading = true;
		try {
			if (!giftCardId) {
				toastStore.error(tr('giftCards.loadError'));
				goto(resolve('/gift-cards'));
				return;
			}

			// Phase 1: Show cached data immediately
			const cached = await giftCardsApi.getCached(giftCardId);
			if (cached) {
				giftCard = cached.gift_card;
				shares = cached.shares;
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					const fresh = await giftCardsApi.get(giftCardId);
					giftCard = fresh.gift_card;
					shares = fresh.shares || [];
				} catch (err: unknown) {
					if (
						err instanceof ApiError &&
						(err.status === 403 || err.status === 404)
					) {
						await offlineDB.deleteGiftCard(giftCardId);
						toastStore.error(tr('giftCards.loadError'));
						goto(resolve('/gift-cards'));
						return;
					}
					if (!cached) {
						toastStore.error(tr('giftCards.loadError'));
						goto(resolve('/gift-cards'));
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('giftCards.loadError'));
				goto(resolve('/gift-cards'));
			}
		} catch {
			toastStore.error(tr('giftCards.loadError'));
			goto(resolve('/gift-cards'));
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	async function loadMerchants() {
		// Merchants are only needed for editing, skip when offline
		if (!navigator.onLine) return;
		try {
			const response = await merchantsApi.list();
			merchants = response.merchants || [];
		} catch (err) {
			pageLogger.error('Failed to load merchants:', err);
		}
	}

	async function startEdit() {
		if (!giftCard) return;

		if (merchants.length === 0) {
			await loadMerchants();
		}

		editMerchantId = giftCard.merchant?.id || '';
		editCardNumber = giftCard.card_number;
		editInitialBalance = giftCard.initial_balance;
		editCurrency = giftCard.currency;
		editPin = giftCard.pin || '';
		editBarcodeType = giftCard.barcode_type || 'CODE128';
		editExpiresAt = giftCard.expires_at
			? giftCard.expires_at.split('T')[0]
			: '';
		editNotes = giftCard.notes || '';
	}

	async function saveEdit(close: () => void) {
		if (!giftCard || !giftCardId) return;
		try {
			const response = await giftCardsApi.update(giftCardId, {
				merchant_id: editMerchantId || undefined,
				card_number: editCardNumber,
				initial_balance: editInitialBalance,
				currency: editCurrency,
				pin: editPin || undefined,
				barcode_type: editBarcodeType,
				expires_at: editExpiresAt ? `${editExpiresAt}T00:00:00Z` : undefined,
				notes: editNotes || undefined
			});

			// Ensure permissions are set from the response
			if (response.permissions) {
				response.gift_card.permissions = response.permissions;
			}
			giftCard = response.gift_card;
			shares = response.shares || [];
			close();
			toastStore.success(tr('giftCards.updateSuccess'));
		} catch (err: unknown) {
			toastStore.error(
				err instanceof Error ? err.message : tr('giftCards.updateError')
			);
		}
	}
</script>

<svelte:head>
	<title
		>{giftCard
			? `${giftCard.merchant?.name || tr('giftCards.title')} - ${tr('common.appName')}`
			: `${tr('giftCards.title')} - ${tr('common.appName')}`}</title
	>
</svelte:head>

<div class="px-4 max-w-7xl mx-auto">
	{#if isRefreshing}
		<div class="mb-6 flex justify-end">
			<span class="text-xs text-text-faint animate-pulse"
				>{tr('common.refreshing')}</span
			>
		</div>
	{/if}

	{#if isLoading}
		<LoadingSpinner />
	{:else}
		<!-- Mounted unconditionally so ResourceDetail owns the not-found state
		     ({#if resource}…{:else}) — prevents the #121 white screen when the
		     resource is null and not loading (offline / 403 / 404). -->
		<ResourceDetail
			kind="gift_card"
			bind:resource={giftCard}
			bind:shares
			{isOffline}
			onStartEdit={startEdit}
		>
			{#snippet edit({ cancel, close })}
				<GiftCardForm
					bind:cardNumber={editCardNumber}
					bind:merchantId={editMerchantId}
					bind:initialBalance={editInitialBalance}
					bind:currency={editCurrency}
					bind:pin={editPin}
					bind:barcodeType={editBarcodeType}
					bind:expiresAt={editExpiresAt}
					bind:notes={editNotes}
					onSubmit={() => saveEdit(close)}
					onCancel={cancel}
					isLoading={false}
					submitLabel={tr('common.save')}
				/>
			{/snippet}
			{#snippet ledger()}
				<GiftCardLedger
					giftCard={giftCard!}
					{isOffline}
					onRefresh={loadGiftCard}
				/>
			{/snippet}
		</ResourceDetail>
	{/if}
</div>
