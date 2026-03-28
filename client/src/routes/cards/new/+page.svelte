<script lang="ts">
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';
	import { cardsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import type { DuplicateWarning } from '$lib/types/api';
	import { extractDuplicate } from '$lib/utils/api-errors';

	import { logger } from '$lib/utils/logger';
	import CardForm from '$lib/components/cards/CardForm.svelte';
	import SharedInfoBox from '$lib/components/SharedInfoBox.svelte';
	import EmailAutocomplete from '$lib/components/EmailAutocomplete.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('CardsNewPage');

	let cardNumber = $state('');
	let merchantId = $state('');
	let program = $state('');
	let barcodeType = $state('CODE128');
	let notes = $state('');
	let isLoading = $state(false);
	let duplicateWarning = $state<DuplicateWarning | null>(null);

	// Sharing state
	let shareEmail = $state('');
	let canEdit = $state(false);
	let canDelete = $state(false);

	async function handleSubmit() {
		isLoading = true;
		duplicateWarning = null;
		try {
			const response = await cardsApi.create({
				card_number: cardNumber,
				merchant_id: merchantId || undefined,
				program: program || undefined,
				barcode_type: barcodeType || undefined,
				notes: notes || undefined,
				share_with_email: shareEmail || undefined,
				share_can_edit: shareEmail ? canEdit : undefined,
				share_can_delete: shareEmail ? canDelete : undefined
			});
			toastStore.success(tr('cards.createSuccess'));
			// Force full page reload to ensure fresh data in lists
			window.location.href = `/cards/${response.card.id}`;
		} catch (err: any) {
			const duplicate = extractDuplicate(err);
			if (duplicate) {
				duplicateWarning = duplicate;
			} else {
				toastStore.error(err.message || tr('cards.createError'));
			}
		} finally {
			isLoading = false;
		}
	}

	function handleCancel() {
		goto('/cards');
	}
</script>

<svelte:head>
	<title>{tr('cards.newCard')} - {tr('common.appName')}</title>
</svelte:head>

<div class="mb-6">
	<a href="/cards" class="text-cyan-600 hover:text-cyan-700"
		>{tr('common.backToOverview')}</a
	>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Left column: Form (2/3 width) -->
	<div class="lg:col-span-2">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h1 class="text-3xl font-bold text-gray-900 mb-6">
				{tr('cards.newCard')}
			</h1>
			<DuplicateWarningBanner
				warning={duplicateWarning}
				resourceType="card"
				onNavigate={(id) => goto(`/cards/${id}`)}
			/>
			<CardForm
				bind:cardNumber
				bind:merchantId
				bind:program
				bind:barcodeType
				bind:notes
				onSubmit={handleSubmit}
				onCancel={handleCancel}
				{isLoading}
				submitLabel={tr('cards.createButton')}
			/>
		</div>
	</div>

	<!-- Right column: Sharing (1/3 width) -->
	<div class="lg:col-span-1">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-gray-900 mb-4">
				{tr('cards.sharing.title')}
			</h2>
			<p class="text-sm text-gray-600 mb-4">
				{tr('cards.sharing.shareOnCreate')}
			</p>

			<div class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4">
				<!-- Email Input with Autocomplete -->
				<EmailAutocomplete
					bind:value={shareEmail}
					label={tr('cards.sharing.userEmail')}
					hint={tr('forms.userMustBeRegistered')}
					inputId="share_email"
				/>

				<!-- Permissions -->
				<div class="space-y-2">
					<label class="flex items-start">
						<input
							type="checkbox"
							bind:checked={canEdit}
							class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
						/>
						<div class="ml-2">
							<span class="block text-sm font-medium text-gray-900"
								>{tr('cards.sharing.canEdit')}</span
							>
							<span class="text-xs text-gray-500">
								{tr('cards.sharing.canEditDesc')}
							</span>
						</div>
					</label>
					<label class="flex items-start">
						<input
							type="checkbox"
							bind:checked={canDelete}
							class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
						/>
						<div class="ml-2">
							<span class="block text-sm font-medium text-gray-900"
								>{tr('cards.sharing.canDelete')}</span
							>
							<span class="text-xs text-gray-500"
								>{tr('cards.sharing.canDeleteDesc')}</span
							>
						</div>
					</label>
				</div>

				<!-- Info Box -->
				<SharedInfoBox
					title={tr('cards.sharing.whatIsShared')}
					items={[
						tr('cards.sharing.sharedItemCardNumber'),
						tr('cards.sharing.sharedItemDetails'),
						tr('cards.sharing.sharedItemNotes')
					]}
				/>
			</div>
		</div>
	</div>
</div>
