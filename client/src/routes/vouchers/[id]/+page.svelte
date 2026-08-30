<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { vouchersApi, merchantsApi, ApiError } from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import VoucherForm from '$lib/components/vouchers/VoucherForm.svelte';
	import ResourceDetail from '$lib/components/ui/ResourceDetail.svelte';
	import ResourceDetailActions from '$lib/components/ui/ResourceDetailActions.svelte';
	import type { VoucherDTO, ShareDTO, MerchantDTO } from '$lib/types/api';

	import { platform } from '$lib/utils/platform';

	const IS_IOS = platform === 'ios';
	// Android edits through the FAB (ResourceDetail), so the title row hides
	// the edit button; share, transfer and delete live behind the more menu.
	const IS_ANDROID = platform === 'android';

	// Desktop title row: edit-aware title, back link and the action buttons
	// live on the shell; ResourceDetail keeps owning the behaviour (bound
	// state + instance methods).
	let detail: ResourceDetail | undefined = $state();
	let isEditing = $state(false);
	let isTogglingFavorite = $state(false);

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('VoucherDetailPage');

	const voucherId = $derived($page.params.id);

	let voucher = $state<VoucherDTO | null>(null);
	let shares = $state<ShareDTO[]>([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Edit form fields
	let editMerchantId = $state('');
	let editCode = $state('');
	let editType = $state('percentage');
	let editValue = $state(0);
	let editCurrency = $state('CHF');
	let editMinPurchaseAmount = $state(0);
	let editBarcodeType = $state('CODE128');
	let editUsageLimitType = $state('single_use');
	let editValidFrom = $state('');
	let editValidUntil = $state('');
	let editStatus = $state('active');
	let editDescription = $state('');

	const isOffline = $derived(!$isOnline);

	onMount(async () => {
		await Promise.all([loadVoucher(), loadMerchants()]);
	});

	async function loadVoucher() {
		isLoading = true;
		try {
			if (!voucherId) {
				toastStore.error(tr('vouchers.loadError'));
				goto(resolve('/vouchers'));
				return;
			}

			// Phase 1: Show cached data immediately
			const cached = await vouchersApi.getCached(voucherId);
			if (cached) {
				voucher = cached.voucher;
				shares = cached.shares;
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					const fresh = await vouchersApi.get(voucherId);
					voucher = fresh.voucher;
					shares = fresh.shares || [];
				} catch (err: unknown) {
					if (
						err instanceof ApiError &&
						(err.status === 403 || err.status === 404)
					) {
						await offlineDB.deleteVoucher(voucherId);
						toastStore.error(tr('vouchers.loadError'));
						goto(resolve('/vouchers'));
						return;
					}
					if (!cached) {
						toastStore.error(tr('vouchers.loadError'));
						goto(resolve('/vouchers'));
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('vouchers.loadError'));
				goto(resolve('/vouchers'));
			}
		} catch {
			toastStore.error(tr('vouchers.loadError'));
			goto(resolve('/vouchers'));
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
			pageLogger.error('Failed to load merchants', { error: err });
		}
	}

	async function startEdit() {
		if (!voucher) return;

		if (merchants.length === 0) {
			await loadMerchants();
		}

		editMerchantId = voucher.merchant?.id || '';
		editCode = voucher.code;
		editType = voucher.type || 'percentage';
		editValue = voucher.value;
		editCurrency = voucher.currency || 'CHF';
		editMinPurchaseAmount = voucher.min_purchase_amount || 0;
		// Ensure barcode_type always has a valid value (handle null, undefined, empty string)
		editBarcodeType =
			(voucher.barcode_type && voucher.barcode_type.trim()) || 'CODE128';
		editUsageLimitType = voucher.usage_limit_type || 'single_use';
		editValidFrom = voucher.valid_from ? voucher.valid_from.split('T')[0] : '';
		editValidUntil = voucher.valid_until
			? voucher.valid_until.split('T')[0]
			: '';
		editStatus = voucher.status || 'active';
		editDescription = voucher.description || '';
	}

	async function saveEdit(close: () => void) {
		if (!voucher || !voucherId) return;

		// Validate required fields
		if (!editValidUntil) {
			toastStore.error(tr('vouchers.validUntilRequired'));
			return;
		}

		try {
			const response = await vouchersApi.update(voucherId, {
				merchant_id: editMerchantId || undefined,
				code: editCode,
				type: editType,
				value: editValue,
				currency: editCurrency || undefined,
				min_purchase_amount: editMinPurchaseAmount || undefined,
				barcode_type: editBarcodeType,
				usage_limit_type: editUsageLimitType,
				valid_from: editValidFrom ? `${editValidFrom}T00:00:00Z` : undefined,
				valid_until: `${editValidUntil}T23:59:59Z`,
				status: editStatus,
				description: editDescription || undefined
			});

			// Ensure permissions are set from the response
			if (response.permissions) {
				response.voucher.permissions = response.permissions;
			}
			voucher = response.voucher;
			shares = response.shares || [];
			close();
			toastStore.success(tr('vouchers.updateSuccess'));
		} catch (err) {
			const message = err instanceof Error ? err.message : '';
			toastStore.error(message || tr('vouchers.updateError'));
		}
	}
</script>

<svelte:head>
	<title
		>{voucher
			? `${voucher.merchant?.name || tr('common.voucher')} - ${tr('common.appName')}`
			: `${tr('common.voucher')} - ${tr('common.appName')}`}</title
	>
</svelte:head>

<!-- Desktop: the shell renders the title (skeleton phase — the section stays
     unmounted). Native: ResourceDetail renders the screen's own heading. -->
{#snippet backLine()}
	{#if isEditing}
		<button
			type="button"
			onclick={() => detail?.cancelEdit()}
			class="text-text-subtle hover:text-text-ink2"
		>
			{tr('common.backToVoucher')}
		</button>
	{:else}
		<!-- history.back() with a wallet fallback: a link to the type route
		     would redirect through /wallet?type=… and overwrite the user's
		     wallet filter. -->
		<button
			type="button"
			onclick={() => detail?.goBack()}
			class="text-text-subtle hover:text-text-ink2"
		>
			{tr('common.backToOverview')}
		</button>
	{/if}
{/snippet}

<PageShell
	title={isEditing
		? tr('vouchers.editVoucher')
		: (voucher?.merchant?.name ?? tr('common.voucher'))}
	mobileActions={false}
	back={IS_ANDROID || IS_IOS ? undefined : backLine}
	onBack={IS_ANDROID || IS_IOS
		? () => (isEditing ? detail?.cancelEdit() : detail?.goBack())
		: undefined}
>
	{#snippet actions()}
		{#if voucher && !isEditing}
			<ResourceDetailActions
				isFavorite={!!voucher.is_favorite}
				{isTogglingFavorite}
				canEdit={!IS_ANDROID && !!voucher.permissions?.can_edit}
				canDelete={!!voucher.permissions?.can_delete}
				{isOffline}
				favoriteAddLabel={tr('common.addToFavorites')}
				favoriteRemoveLabel={tr('common.removeFromFavorites')}
				onFavorite={() => detail?.toggleFavorite()}
				onEdit={() => detail?.startEdit()}
				onDelete={() => detail?.promptDelete()}
				onMore={(IS_IOS && voucher.permissions?.is_owner) ||
				(IS_ANDROID &&
					(voucher.permissions?.is_owner || voucher.permissions?.can_delete))
					? () => detail?.openMoreMenu()
					: undefined}
			/>
		{/if}
	{/snippet}
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
			bind:this={detail}
			bind:isEditing
			bind:isTogglingFavorite
			kind="voucher"
			bind:resource={voucher}
			bind:shares
			{isOffline}
			shareMode="readonly"
			onStartEdit={startEdit}
		>
			{#snippet edit({ cancel, close, deleteAction })}
				<VoucherForm
					bind:code={editCode}
					bind:merchantId={editMerchantId}
					bind:type={editType}
					bind:value={editValue}
					bind:currency={editCurrency}
					bind:minPurchaseAmount={editMinPurchaseAmount}
					bind:barcodeType={editBarcodeType}
					bind:validFrom={editValidFrom}
					bind:validUntil={editValidUntil}
					bind:usageLimitType={editUsageLimitType}
					bind:description={editDescription}
					onSubmit={() => saveEdit(close)}
					onCancel={cancel}
					isLoading={false}
					submitLabel={tr('common.save')}
					trailingActions={deleteAction}
					pairedLayout
				/>
			{/snippet}
		</ResourceDetail>
	{/if}
</PageShell>
