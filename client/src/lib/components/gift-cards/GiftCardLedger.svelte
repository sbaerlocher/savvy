<script lang="ts">
	import { ICON_LOCK } from '$lib/icons';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import { t, locale } from '$lib/stores/i18n';
	import { giftCardsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { MERCHANT_DEFAULT_COLOR } from '$lib/utils/merchant-color';
	import { platform } from '$lib/utils/platform';
	import { formatDisplayDate, todayInputValue } from '$lib/utils/date';
	import type { GiftCardDTO, TransactionDTO } from '$lib/types/api';

	interface Props {
		giftCard: GiftCardDTO;
		isOffline: boolean;
		/** Called after add/delete transaction so the parent can refresh the balance. */
		onRefresh?: () => void | Promise<void>;
	}

	let { giftCard, isOffline, onRefresh }: Props = $props();

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const currentLocale = $derived($locale || 'de-DE');
	const pageLogger = logger.child('GiftCardLedger');

	// ponytail: gift-card only reads cached.transactions on load; fresh fetch does
	// NOT refresh them, only onMount's loadTransactions does — preserved as-is.
	// Seeded once in onMount (untracked) so the cached list shows before the fetch.
	let transactions = $state<TransactionDTO[]>([]);
	let showTransactionForm = $state(false);

	// Transaction form
	let transactionAmount = $state(0);
	let transactionDescription = $state('');
	let transactionDate = $state(todayInputValue());

	let showDeleteTransactionModal = $state(false);
	let transactionToDelete: string | null = null;

	// Android and iOS render the ledger as a second card stacked under the
	// barcode card (screen-ResourceDetailAndroid frame 3 /
	// screen-ResourceDetailIOS); desktop renders it in the right column with the
	// gift-card gold accent and a mono balance (screen-ResourceDetailDesktop,
	// board 3). `platform` is a module constant, so plain consts.
	const isAndroid = platform === 'android';
	const IS_DESKTOP = platform === 'other';
	const IS_IOS = platform === 'ios';
	// Desktop and iOS lead with the currency ("CHF 85.50"); Android keeps the
	// trailing-code form.
	const currencyFirst = IS_DESKTOP || IS_IOS;

	const percentageRemaining = $derived(
		giftCard
			? Math.round((giftCard.current_balance / giftCard.initial_balance) * 100)
			: 0
	);

	onMount(() => {
		// Show cached transactions immediately, then refresh from network.
		transactions = giftCard.transactions || [];
		loadTransactions();
	});

	async function loadTransactions() {
		try {
			if (!giftCard.id) return;
			const response = await giftCardsApi.listTransactions(giftCard.id);
			transactions = response.transactions || [];
		} catch (err) {
			pageLogger.error('Failed to load transactions:', err);
		}
	}

	async function handleAddTransaction() {
		try {
			if (!giftCard.id) {
				toastStore.error(tr('giftCards.transactions.error'));
				return;
			}
			// Convert YYYY-MM-DD to RFC3339 format (ISO 8601 with timezone)
			const transactionDateISO = `${transactionDate}T00:00:00Z`;

			// Automatisch debit (Ausgabe) verwenden
			await giftCardsApi.createTransaction(giftCard.id, {
				type: 'debit',
				amount: transactionAmount,
				description: transactionDescription || undefined,
				transaction_date: transactionDateISO
			});
			toastStore.success(tr('giftCards.transactions.createSuccess'));
			showTransactionForm = false;
			transactionAmount = 0;
			transactionDescription = '';
			transactionDate = todayInputValue();
			await Promise.all([onRefresh?.(), loadTransactions()]);
		} catch (err: unknown) {
			toastStore.error(
				err instanceof Error
					? err.message
					: tr('giftCards.transactions.createError')
			);
		}
	}

	function promptDeleteTransaction(transactionId: string) {
		transactionToDelete = transactionId;
		showDeleteTransactionModal = true;
	}

	async function confirmDeleteTransaction() {
		if (!transactionToDelete || !giftCard.id) return;
		try {
			await giftCardsApi.deleteTransaction(giftCard.id, transactionToDelete);
			toastStore.success(tr('giftCards.transactions.deleteSuccess'));
			showDeleteTransactionModal = false;
			await Promise.all([onRefresh?.(), loadTransactions()]);
		} catch (err: unknown) {
			toastStore.error(
				err instanceof Error
					? err.message
					: tr('giftCards.transactions.deleteError')
			);
		} finally {
			transactionToDelete = null;
			showDeleteTransactionModal = false;
		}
	}
</script>

<!-- Balance & Transactions Box -->
<div
	class={isAndroid
		? 'rounded-m3-lg bg-m3-card p-5'
		: IS_DESKTOP
			? 'rounded-2xl border border-border bg-white p-5'
			: IS_IOS
				? 'rounded-[var(--radius-inset)] bg-surface p-5'
				: 'rounded-xl border border-border bg-white p-6'}
>
	<!-- Balance Display -->
	<div class={isAndroid || IS_DESKTOP ? 'mb-0' : 'mb-6'}>
		<p
			class="{isAndroid
				? 'text-label font-normal'
				: 'text-sm'} text-text-ink2 mb-1"
		>
			{tr('giftCards.balance.current')}
		</p>
		<p
			class={isAndroid || IS_IOS
				? 'text-giftcard-ink text-screen-title font-mono font-semibold'
				: IS_DESKTOP
					? 'text-giftcard-ink text-display font-mono font-semibold'
					: 'text-3xl font-bold'}
			style={isAndroid || IS_DESKTOP || IS_IOS
				? undefined
				: `color: ${giftCard.merchant?.color || MERCHANT_DEFAULT_COLOR}`}
		>
			<!-- The desktop and iOS mockups lead with the currency ("CHF 85.50",
			     "Ausgangswert: CHF 150.00", "−CHF 14.50"); Android keeps the
			     trailing form. -->
			{#if currencyFirst}
				{giftCard.currency}
				{giftCard.current_balance.toFixed(2)}
			{:else}
				{giftCard.current_balance.toFixed(2)}
				{giftCard.currency}
			{/if}
		</p>
		<!-- Progress Bar -->
		<div
			class="mt-3 overflow-hidden bg-border rounded-full {isAndroid || IS_IOS
				? 'h-2.25'
				: IS_DESKTOP
					? 'h-2.5'
					: 'h-3'}"
		>
			<div
				class="h-full rounded-full transition-all {isAndroid ||
				IS_DESKTOP ||
				IS_IOS
					? 'bg-giftcard-line'
					: percentageRemaining > 50
						? 'bg-success-500'
						: percentageRemaining > 20
							? 'bg-warning-500'
							: 'bg-danger-600'}"
				style="width: {percentageRemaining}%"
			></div>
		</div>
		<p
			class={IS_DESKTOP
				? 'mt-2 text-body-sm text-text-subtle'
				: 'text-xs text-text-subtle mt-2'}
		>
			{tr('giftCards.balance.initial')}:
			{#if currencyFirst}
				{giftCard.currency}
				{giftCard.initial_balance.toFixed(2)}
			{:else}
				{giftCard.initial_balance.toFixed(2)}
				{giftCard.currency}
			{/if}
		</p>
	</div>

	<!-- Transactions Section -->
	<div
		class={isAndroid || IS_IOS
			? 'border-border-soft mt-4.5 border-t pt-4'
			: IS_DESKTOP
				? 'border-border-soft mt-4.5 border-t pt-4.5'
				: 'border-t pt-6'}
	>
		<div
			class="flex justify-between items-center {isAndroid ||
			IS_DESKTOP ||
			IS_IOS
				? 'mb-3'
				: 'mb-4'}"
		>
			<h3
				class={isAndroid || IS_DESKTOP || IS_IOS
					? 'text-label text-text-ink2'
					: 'text-sm font-medium text-text-ink2'}
			>
				{tr('giftCards.transactions.title')}
			</h3>
			{#if giftCard.permissions?.can_edit_transactions && !showTransactionForm}
				<button
					data-testid="add-transaction"
					onclick={() => (showTransactionForm = true)}
					disabled={isOffline}
					class={isAndroid
						? 'bg-danger-600 text-chip rounded-m3-full inline-flex items-center gap-1 px-3.25 py-1.5 whitespace-nowrap text-white disabled:opacity-50'
						: IS_DESKTOP
							? `h-7 rounded-sm border border-accent-200 bg-accent-50 px-3 text-chip text-accent-700 hover:bg-accent-100 whitespace-nowrap flex items-center gap-1.5 ${
									isOffline
										? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
										: ''
								}`
							: `btn btn-xs btn-danger whitespace-nowrap flex items-center gap-1.5 ${
									isOffline
										? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
										: ''
								}`}
				>
					{#if isOffline}
						<svg
							class="w-3.5 h-3.5"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d={ICON_LOCK}
							></path>
						</svg>
					{:else}
						<span>+</span>
					{/if}
					{tr('common.new')}
				</button>
			{/if}
		</div>

		{#if showTransactionForm}
			<div
				class={isAndroid
					? 'border-danger-200 bg-danger-50 rounded-m3-md mb-3.5 flex flex-col gap-3 border p-3.5'
					: 'border border-danger-200 bg-danger-50 rounded-lg p-4 space-y-4 mb-4'}
			>
				<div>
					<label
						for="transactionDate-input"
						class="text-body-sm text-text-ink2 mb-1.5 block font-medium"
					>
						{tr('giftCards.transactions.date')} *
					</label>
					<input
						id="transactionDate-input"
						type="date"
						required
						bind:value={transactionDate}
						max={todayInputValue()}
						class={isAndroid
							? 'text-subheading rounded-m3-sm border-border-field bg-m3-card text-text w-full border px-3 py-2.5 font-normal'
							: 'input w-full text-base bg-white'}
					/>
				</div>

				<div>
					<label
						for="transactionAmount-input"
						class="text-body-sm text-text-ink2 mb-1.5 block font-medium"
					>
						{tr('giftCards.transactions.amount')} *
					</label>
					<input
						id="transactionAmount-input"
						type="number"
						step="0.01"
						min="0.01"
						required
						bind:value={transactionAmount}
						class={isAndroid
							? 'text-subheading rounded-m3-sm border-border-field bg-m3-card text-text w-full border px-3 py-2.5 font-normal'
							: 'input bg-white'}
						placeholder="10.00"
					/>
				</div>

				<div>
					<label
						for="transactionDescription-input"
						class="text-body-sm text-text-ink2 mb-1.5 block font-medium"
					>
						{tr('giftCards.transactions.description')}
					</label>
					<input
						id="transactionDescription-input"
						type="text"
						bind:value={transactionDescription}
						class={isAndroid
							? 'text-subheading rounded-m3-sm border-border-field bg-m3-card text-text w-full border px-3 py-2.5 font-normal'
							: 'input bg-white'}
						placeholder={tr('giftCards.transactions.descriptionPlaceholder')}
					/>
				</div>

				<!-- Android puts cancel/save right-aligned as M3 text + filled
				     buttons; the other platforms keep the stretched pair. -->
				<div
					class={isAndroid
						? 'flex items-center justify-end gap-2'
						: 'flex gap-2'}
				>
					{#if isAndroid}
						<button
							type="button"
							onclick={() => (showTransactionForm = false)}
							class="text-label text-accent-700 rounded-m3-full inline-flex h-10 items-center px-3"
						>
							{tr('common.cancel')}
						</button>
						<button
							type="button"
							onclick={handleAddTransaction}
							disabled={isOffline}
							class="bg-accent-600 text-on-accent text-label rounded-m3-full inline-flex h-10 items-center px-6 disabled:opacity-50"
						>
							{tr('common.save')}
						</button>
					{:else}
						<button
							onclick={handleAddTransaction}
							disabled={isOffline}
							class="btn btn-danger flex-1 {isOffline
								? 'opacity-50 cursor-not-allowed'
								: ''}"
						>
							{tr('common.save')}
						</button>
						<button
							onclick={() => (showTransactionForm = false)}
							class="btn btn-ghost"
						>
							{tr('common.cancel')}
						</button>
					{/if}
				</div>
			</div>
		{/if}

		{#if transactions.length > 0}
			<div
				class={isAndroid || IS_DESKTOP ? 'flex flex-col gap-2' : 'space-y-2'}
			>
				{#each transactions as transaction (transaction.id)}
					<!-- Both mockups tint the description --text-subtle and the date
					     --text-faint; at 11px on --surface-1 that is 3.6:1 and 2.6:1,
					     under the WCAG AA 4.5:1 floor for small text, so every platform
					     uses --text-muted (~6.4:1) here. Deliberate deviation from the
					     mockups, approved 2026-08-23. -->
					<div
						class="flex items-center justify-between bg-surface-1 gap-3 {isAndroid
							? 'rounded-m3-sm px-3.25 py-2.75'
							: IS_DESKTOP
								? 'rounded-md px-3.5 py-2.5'
								: 'rounded p-3'}"
					>
						<div class="min-w-0 flex-1">
							<div
								class={isAndroid
									? 'text-danger-600 text-label font-mono'
									: IS_DESKTOP
										? 'text-danger-700 text-label font-mono'
										: 'text-danger-600 font-medium'}
							>
								{#if currencyFirst}
									−{giftCard.currency}
									{transaction.amount.toFixed(2)}
								{:else}
									-{transaction.amount.toFixed(2)}
									{giftCard.currency}
								{/if}
							</div>
							{#if transaction.description}
								<div
									class={isAndroid
										? 'text-text-muted text-mono-sm truncate font-sans tracking-normal'
										: IS_DESKTOP
											? 'text-text-muted text-body-sm truncate'
											: 'text-sm text-text-muted'}
								>
									{transaction.description}
								</div>
							{/if}
						</div>
						<div class="flex items-center gap-3">
							<div
								class={isAndroid
									? 'text-text-muted text-mono-sm font-mono whitespace-nowrap'
									: IS_DESKTOP
										? 'text-text-muted text-mono-sm font-mono whitespace-nowrap'
										: 'text-xs text-text-subtle'}
							>
								{formatDisplayDate(transaction.transaction_date, currentLocale)}
							</div>
							{#if giftCard.permissions?.can_edit_transactions}
								<button
									onclick={() => promptDeleteTransaction(transaction.id)}
									disabled={isOffline}
									class="btn-text-danger text-base flex items-center {isOffline
										? 'opacity-50 cursor-not-allowed'
										: ''}"
									title={tr('giftCards.transactions.deleteButton')}
								>
									{#if isOffline}
										<svg
											class="w-4 h-4"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d={ICON_LOCK}
											></path>
										</svg>
									{:else}
										×
									{/if}
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div
				class="text-center py-8 {isAndroid
					? 'bg-m3-card-chip rounded-m3-sm'
					: 'bg-surface-1 rounded'}"
			>
				{#if isOffline}
					<p class="text-warning-600 text-sm">
						{tr('giftCards.transactions.notCachedOffline')}
					</p>
				{:else}
					<p class="text-text-subtle text-sm">
						{tr('giftCards.transactions.noTransactions')}
					</p>
				{/if}
			</div>
		{/if}
	</div>
</div>

<ConfirmModal
	isOpen={showDeleteTransactionModal}
	title={tr('giftCards.transactions.deleteConfirm')}
	message={tr('giftCards.transactions.deleteConfirmMessage')}
	confirmText={tr('common.delete')}
	cancelText={tr('common.cancel')}
	variant="danger"
	onconfirm={confirmDeleteTransaction}
	oncancel={() => (showDeleteTransactionModal = false)}
/>
