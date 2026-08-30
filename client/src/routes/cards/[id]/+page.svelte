<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { get } from 'svelte/store';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { cardsApi, merchantsApi, ApiError } from '$lib/api';
	import { offlineDB } from '$lib/stores/offline-db';
	import { toastStore } from '$lib/stores/toast';
	import CardForm from '$lib/components/cards/CardForm.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import ResourceDetail from '$lib/components/ui/ResourceDetail.svelte';
	import ResourceDetailActions from '$lib/components/ui/ResourceDetailActions.svelte';

	import type { CardDTO, ShareDTO, MerchantDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';

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
	const pageLogger = logger.child('CardDetailsPage');

	const cardId = $derived($page.params.id!);

	let card = $state<CardDTO | null>(null);
	let shares = $state<ShareDTO[]>([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let merchants = $state<MerchantDTO[]>([]);

	// Edit form fields
	let editMerchantId = $state('');
	let editProgram = $state('');
	let editCardNumber = $state('');
	let editBarcodeType = $state('CODE128');
	let editStatus = $state('active');
	let editNotes = $state('');

	const isOffline = $derived(!$isOnline);

	onMount(async () => {
		await Promise.all([loadCard(), loadMerchants()]);
	});

	async function loadCard() {
		isLoading = true;
		try {
			// Phase 1: Show cached data immediately
			const cached = await cardsApi.getCached(cardId);
			if (cached) {
				card = cached.card;
				shares = cached.shares;
				isLoading = false;
				isRefreshing = true;
			}

			// Phase 2: Fetch fresh data from network
			if (navigator.onLine) {
				try {
					const fresh = await cardsApi.get(cardId);
					card = fresh.card;
					shares = fresh.shares || [];
				} catch (err: unknown) {
					if (
						err instanceof ApiError &&
						(err.status === 403 || err.status === 404)
					) {
						await offlineDB.deleteCard(cardId);
						toastStore.error(tr('cards.loadError'));
						goto(resolve('/cards'));
						return;
					}
					if (!cached) {
						toastStore.error(tr('cards.loadError'));
						goto(resolve('/cards'));
						return;
					}
					// Transient error with cached data available - show warning, don't redirect
					toastStore.warning(tr('common.offlineMode'));
				}
			} else if (!cached) {
				toastStore.error(tr('cards.loadError'));
				goto(resolve('/cards'));
			}
		} catch {
			toastStore.error(tr('cards.loadError'));
			goto(resolve('/cards'));
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
		if (!card) return;

		// Ensure merchants are loaded before entering edit mode
		if (merchants.length === 0) {
			await loadMerchants();
		}

		editMerchantId = card.merchant?.id || '';
		editProgram = card.program || '';
		editCardNumber = card.card_number;
		editBarcodeType = card.barcode_type || 'CODE128';
		editStatus = card.status || 'active';
		editNotes = card.notes || '';
	}

	async function saveEdit(close: () => void) {
		if (!card) return;
		try {
			const response = await cardsApi.update(cardId, {
				merchant_id: editMerchantId || undefined,
				program: editProgram || undefined,
				card_number: editCardNumber,
				barcode_type: editBarcodeType,
				status: editStatus,
				notes: editNotes || undefined
			});

			// Ensure permissions are set from the response
			if (response.permissions) {
				response.card.permissions = response.permissions;
			}
			card = response.card;
			shares = response.shares || [];
			close();
			toastStore.success(tr('cards.updateSuccess'));
		} catch (err: unknown) {
			pageLogger.error('Save error:', err);
			toastStore.error(
				err instanceof Error ? err.message : tr('cards.updateError')
			);
		}
	}
</script>

<svelte:head>
	<title
		>{card
			? `${card.merchant?.name || tr('cards.title')} - ${tr('common.appName')}`
			: `${tr('cards.title')} - ${tr('common.appName')}`}</title
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
			{tr('common.backToCard')}
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
		? tr('cards.editCard')
		: (card?.merchant?.name ?? tr('cards.title'))}
	mobileActions={false}
	back={IS_ANDROID || IS_IOS ? undefined : backLine}
	onBack={IS_ANDROID || IS_IOS
		? () => (isEditing ? detail?.cancelEdit() : detail?.goBack())
		: undefined}
>
	{#snippet actions()}
		{#if card && !isEditing}
			<ResourceDetailActions
				isFavorite={!!card.is_favorite}
				{isTogglingFavorite}
				canEdit={!IS_ANDROID && !!card.permissions?.can_edit}
				canDelete={!!card.permissions?.can_delete}
				{isOffline}
				favoriteAddLabel={tr('common.addToFavorites')}
				favoriteRemoveLabel={tr('common.removeFromFavorites')}
				onFavorite={() => detail?.toggleFavorite()}
				onEdit={() => detail?.startEdit()}
				onDelete={() => detail?.promptDelete()}
				onMore={(IS_IOS && card.permissions?.is_owner) ||
				(IS_ANDROID &&
					(card.permissions?.is_owner || card.permissions?.can_delete))
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
			kind="card"
			bind:resource={card}
			bind:shares
			{isOffline}
			onStartEdit={startEdit}
		>
			{#snippet edit({ cancel, close, deleteAction })}
				<CardForm
					bind:cardNumber={editCardNumber}
					bind:merchantId={editMerchantId}
					bind:program={editProgram}
					bind:barcodeType={editBarcodeType}
					bind:status={editStatus}
					bind:notes={editNotes}
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
