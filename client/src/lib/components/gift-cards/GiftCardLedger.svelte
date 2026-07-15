<script lang="ts">
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import { t, locale } from '$lib/stores/i18n';
	import { giftCardsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { MERCHANT_DEFAULT_COLOR } from '$lib/utils/merchant-color';
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
<div class="rounded-xl border border-border bg-white p-6">
	<!-- Balance Display -->
	<div class="mb-6">
		<p class="text-sm text-text-ink2 mb-1">
			{tr('giftCards.balance.current')}
		</p>
		<p
			class="text-3xl font-bold"
			style="color: {giftCard.merchant?.color || MERCHANT_DEFAULT_COLOR}"
		>
			{giftCard.current_balance.toFixed(2)}
			{giftCard.currency}
		</p>
		<!-- Progress Bar -->
		<div class="mt-3 bg-border rounded-full h-3">
			<div
				class="h-3 rounded-full transition-all {percentageRemaining > 50
					? 'bg-success-500'
					: percentageRemaining > 20
						? 'bg-warning-500'
						: 'bg-danger-600'}"
				style="width: {percentageRemaining}%"
			></div>
		</div>
		<p class="text-xs text-text-muted mt-2">
			{tr('giftCards.balance.initial')}: {giftCard.initial_balance.toFixed(2)}
			{giftCard.currency}
		</p>
	</div>

	<!-- Transactions Section -->
	<div class="border-t pt-6">
		<div class="flex justify-between items-center mb-4">
			<h3 class="text-sm font-medium text-text-ink2">
				{tr('giftCards.transactions.title')}
			</h3>
			{#if giftCard.permissions?.can_edit_transactions && !showTransactionForm}
				<button
					data-testid="add-transaction"
					onclick={() => (showTransactionForm = true)}
					disabled={isOffline}
					class="btn btn-xs btn-danger whitespace-nowrap flex items-center gap-1.5 {isOffline
						? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
						: ''}"
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
								d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
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
				class="border border-danger-200 bg-danger-50 rounded-lg p-4 space-y-4 mb-4"
			>
				<div>
					<label
						for="transactionDate-input"
						class="block text-sm font-medium text-text-ink2 mb-1"
					>
						{tr('giftCards.transactions.date')} *
					</label>
					<input
						id="transactionDate-input"
						type="date"
						required
						bind:value={transactionDate}
						max={todayInputValue()}
						class="input w-full text-base bg-white"
					/>
				</div>

				<div>
					<label
						for="transactionAmount-input"
						class="block text-sm font-medium text-text-ink2 mb-1"
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
						class="input bg-white"
						placeholder="10.00"
					/>
				</div>

				<div>
					<label
						for="transactionDescription-input"
						class="block text-sm font-medium text-text-ink2 mb-1"
					>
						{tr('giftCards.transactions.description')}
					</label>
					<input
						id="transactionDescription-input"
						type="text"
						bind:value={transactionDescription}
						class="input bg-white"
						placeholder={tr('giftCards.transactions.descriptionPlaceholder')}
					/>
				</div>

				<div class="flex gap-2">
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
				</div>
			</div>
		{/if}

		{#if transactions.length > 0}
			<div class="space-y-2">
				{#each transactions as transaction (transaction.id)}
					<div
						class="flex items-center justify-between p-3 bg-surface-1 rounded gap-3"
					>
						<div class="flex-1">
							<div class="font-medium text-danger-600">
								-{transaction.amount.toFixed(2)}
								{giftCard.currency}
							</div>
							{#if transaction.description}
								<div class="text-sm text-text-muted">
									{transaction.description}
								</div>
							{/if}
						</div>
						<div class="flex items-center gap-3">
							<div class="text-xs text-text-subtle">
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
												d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
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
			<div class="text-center py-8 bg-surface-1 rounded">
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
