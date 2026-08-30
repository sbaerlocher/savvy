<script lang="ts">
	import { cardsApi, giftCardsApi, merchantsApi, vouchersApi } from '$lib/api';
	import WalletView, {
		type WalletFilterState
	} from '$lib/components/ui/WalletView.svelte';
	import { authStore } from '$lib/stores/auth';
	import { locale, t } from '$lib/stores/i18n';
	import { isOnline } from '$lib/stores/offline';
	import { toastStore } from '$lib/stores/toast';
	import type {
		CardDTO,
		GiftCardDTO,
		MerchantDTO,
		VoucherDTO
	} from '$lib/types/api';
	import { extractMerchantFromItems } from '$lib/utils/merchant-aggregator';
	import { page } from '$app/stores';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { ICON_PENCIL, ICON_SEARCH, ICON_STOREFRONT } from '$lib/icons';
	import { platform } from '$lib/utils/platform';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	// Android renders this screen as Material 3 (mockup screen-MerchantsAndroid,
	// board 2): a back row above the merchant name, the edit action as a tonal
	// chip, a tonal search field and a visible type chip row.
	// `platform` is a module constant, so a plain const, not $derived.
	const IS_ANDROID = platform === 'android';
	// iOS draws the same three parts as its own chrome (mockup
	// screen-MerchantsIOS, board 2): a back row above the title, a glass edit
	// pill beside it and a glass search field.
	const IOS = platform === 'ios';

	// Desktop puts a back link above the merchant name and the type chips on the
	// toolbar row (mockup screen-MerchantsDesktop, board 2).
	const IS_DESKTOP = platform === 'other';

	const isAdmin = $derived($authStore.user?.is_admin || false);
	const currentUserId = $derived($authStore.user?.id);
	const currentLocale = $derived($locale || 'de-DE');
	const isOffline = $derived(!$isOnline);

	let merchant = $state<MerchantDTO | null>(null);
	let cards = $state<CardDTO[]>([]);
	let vouchers = $state<VoucherDTO[]>([]);
	let giftCards = $state<GiftCardDTO[]>([]);
	let isLoading = $state(true);

	// Merchant-detail keeps its own filter scope (must NOT share the wallet's
	// module-level walletFilters). Inline $state, not a factory — a $state-returning
	// factory in a .svelte.ts module breaks the SSR build.
	const filters = $state<WalletFilterState>({
		searchInput: '',
		typeFilter: 'all',
		statusFilter: 'active',
		sortBy: 'newest',
		ownerFilter: 'all',
		favoritesOnly: false,
		expiringFilter: 'all',
		scrollY: 0
	});

	const enabledTypes = { cards: true, vouchers: true, giftCards: true };

	// Android draws the edit action as a tonal M3 assist chip on the phone and
	// falls back to the shared grey button from `sm` up (mockup).
	// Desktop/sm+: the shared title-row action chrome; the native max-sm
	// overrides below keep their own chip shapes.
	const EDIT_BUTTON = 'title-action whitespace-nowrap text-text-muted';
	const EDIT_ACTION_CLASS = IS_ANDROID
		? `max-sm:inline-flex max-sm:h-8 max-sm:shrink-0 max-sm:items-center max-sm:gap-1.25 max-sm:rounded-m3-full max-sm:border-none max-sm:bg-m3-card max-sm:px-3.5 max-sm:text-label max-sm:text-text-ink2 ${EDIT_BUTTON}`
		: IOS
			? `max-sm:liquid-glass-card max-sm:inline-flex max-sm:shrink-0 max-sm:items-center max-sm:gap-1.25 max-sm:rounded-full max-sm:px-3 max-sm:py-1.75 max-sm:text-label max-sm:text-accent-700 ${EDIT_BUTTON}`
			: EDIT_BUTTON;

	const merchantId = $derived(get(page).params.id ?? '');

	async function loadData() {
		isLoading = true;
		try {
			const [merchantRes, cardsRes, vouchersRes, giftCardsRes] =
				await Promise.all([
					merchantsApi.get(merchantId).catch(() => null),
					cardsApi.list().catch(() => ({ cards: [] })),
					vouchersApi.list().catch(() => ({ vouchers: [] })),
					giftCardsApi.list().catch(() => ({ gift_cards: [] }))
				]);

			// Filter by merchant_id
			cards = cardsRes.cards.filter((c) => c.merchant_id === merchantId);
			vouchers = vouchersRes.vouchers.filter(
				(v) => v.merchant_id === merchantId
			);
			giftCards = giftCardsRes.gift_cards.filter(
				(g) => g.merchant_id === merchantId
			);

			// Use merchant from API, or extract from cached items as fallback
			merchant =
				merchantRes?.merchant ??
				extractMerchantFromItems(merchantId, cards, vouchers, giftCards);
		} catch {
			toastStore.error(tr('merchantOverview.detail.loadError'));
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<svelte:head>
	<title
		>{merchant?.name ?? tr('merchantOverview.detail.title')} - {tr(
			'common.appName'
		)}</title
	>
</svelte:head>

{#if isLoading}
	<PageShell>
		<LoadingSpinner />
	</PageShell>
{:else if merchant}
	<!-- Declared outside the WalletView tag: a snippet placed inside a component
	     becomes one of its props, and WalletView has no `editAction` prop — it is
	     only referenced by the `header` snippet below. -->
	{#snippet editAction()}
		<a
			href={resolve(`/admin/merchants/${merchant?.id}/edit`)}
			class={EDIT_ACTION_CLASS}
		>
			{#if IS_ANDROID || IOS}
				<svg
					class="hidden h-3.25 w-3.25 shrink-0 max-sm:block"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path d={ICON_PENCIL} />
				</svg>
			{/if}
			{tr('common.edit')}
		</a>
	{/snippet}

	<!-- One title row for every platform, rendered by the shell: back link in
	     the eyebrow slot, merchant name as the title, the edit chip as the
	     title-row action. -->
	<PageShell
		title={merchant?.name ?? ''}
		actions={isAdmin && merchant ? editAction : undefined}
		mobileActions={false}
	>
		{#snippet back()}
			<a
				href={resolve('/merchants')}
				class="text-text-subtle hover:text-text-ink2"
				>{tr('common.backToOverview')}</a
			>
		{/snippet}
		<WalletView
			{filters}
			{cards}
			{vouchers}
			{giftCards}
			{enabledTypes}
			{currentUserId}
			{currentLocale}
			{isOffline}
			onReload={loadData}
			barcodeStorageKey="savvy_merchant_detail_show_barcodes"
			idPrefix="detail"
			matchMerchantName={false}
			typeFilterPlacement={IS_ANDROID || IS_DESKTOP ? 'top' : 'batch-header'}
			filterShowAll={IS_ANDROID || IS_DESKTOP}
			inlineTypeToolbar
			androidBarcodeInTypeRow
			iosToolbarVariant={IOS ? 'detail' : 'wallet'}
			barcodeButtonVariant="icon"
		>
			{#snippet emptyIcon()}
				<svg
					class="w-16 h-16 mx-auto text-text-placeholder mb-4 {IS_ANDROID
						? 'max-sm:mx-0 max-sm:mb-0 max-sm:h-10.5 max-sm:w-10.5'
						: ''} {IOS
						? 'max-sm:mb-0 max-sm:h-10.5 max-sm:w-10.5 max-sm:text-text-faint'
						: ''}"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d={ICON_STOREFRONT}
					/>
				</svg>
			{/snippet}

			{#snippet searchField()}
				<!-- Merchant detail: always-visible plain search input. The native
				     platforms draw their own field from the mockups instead — tonal M3
				     on Android, glass on iOS. -->
				{#if IOS}
					<label
						class="liquid-glass-card mb-2.5 flex h-10 items-center gap-2 rounded-[var(--radius-lg)] px-3.25 sm:hidden"
					>
						<svg
							class="h-4 w-4 shrink-0 text-text-faint"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path d={ICON_SEARCH} />
						</svg>
						<input
							type="search"
							bind:value={filters.searchInput}
							placeholder={tr('common.search')}
							aria-label={tr('common.search')}
							class="min-w-0 flex-1 bg-transparent text-body text-text placeholder:text-text-placeholder focus:outline-none"
						/>
					</label>
				{/if}
				{#if IS_ANDROID}
					<div
						class="mb-3 flex h-11 items-center gap-2 rounded-m3-md bg-m3-card px-3.5 sm:hidden"
					>
						<svg
							class="h-4.25 w-4.25 shrink-0 text-text-faint"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path d={ICON_SEARCH} />
						</svg>
						<input
							type="search"
							bind:value={filters.searchInput}
							placeholder={tr('common.search')}
							aria-label={tr('common.search')}
							class="min-w-0 flex-1 bg-transparent text-body text-text placeholder:text-text-placeholder focus:outline-none"
						/>
					</div>
				{/if}
				<div class="mb-6 {IS_ANDROID || IOS ? 'max-sm:hidden' : ''}">
					<input
						type="search"
						bind:value={filters.searchInput}
						placeholder={tr('common.search')}
						class="w-full px-4 py-2 bg-white border border-border-field rounded-md focus:ring-accent focus:border-accent"
					/>
				</div>
			{/snippet}
		</WalletView>
	</PageShell>
{/if}
