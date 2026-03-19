<script lang="ts">
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';
	import { cardsApi, sharedUsersApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import type { UserDTO } from '$lib/types/api';

	import { logger } from '$lib/utils/logger';
	import CardForm from '$lib/components/cards/CardForm.svelte';
	import SharedInfoBox from '$lib/components/SharedInfoBox.svelte';

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

	// Sharing state
	let shareEmail = $state('');
	let canEdit = $state(false);
	let canDelete = $state(false);

	// Email autocomplete state
	let suggestions = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
	let selectedIndex = $state(-1);

	async function fetchSuggestions() {
		if (shareEmail.length < 2) {
			suggestions = [];
			showSuggestions = false;
			return;
		}

		try {
			const response = await sharedUsersApi.search(shareEmail);
			suggestions = response.users || [];
			showSuggestions = suggestions.length > 0;
			selectedIndex = -1;
		} catch (err) {
			pageLogger.error('Failed to fetch user suggestions:', err);
			suggestions = [];
			showSuggestions = false;
		}
	}

	function selectSuggestion(user: UserDTO) {
		shareEmail = user.email;
		showSuggestions = false;
		selectedIndex = -1;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!showSuggestions) return;

		if (event.key === 'ArrowDown') {
			event.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, suggestions.length - 1);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, -1);
		} else if (event.key === 'Enter' && selectedIndex >= 0) {
			event.preventDefault();
			selectSuggestion(suggestions[selectedIndex]);
		} else if (event.key === 'Escape') {
			showSuggestions = false;
			selectedIndex = -1;
		}
	}

	function hideSuggestions() {
		setTimeout(() => {
			showSuggestions = false;
		}, 200);
	}

	async function handleSubmit() {
		isLoading = true;
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
			toastStore.error(err.message || tr('cards.createError'));
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
				<div class="relative">
					<label
						for="share_email"
						class="block text-sm font-medium text-gray-700 mb-1"
					>
						{tr('cards.sharing.userEmail')} *
					</label>
					<input
						type="email"
						id="share_email"
						bind:value={shareEmail}
						oninput={() => fetchSuggestions()}
						onkeydown={handleKeydown}
						onblur={hideSuggestions}
						required
						class="w-full px-3 py-2 bg-white border border-gray-300 rounded-md focus:ring-cyan-500 focus:border-cyan-500"
						placeholder="benutzer@example.com"
						autocomplete="off"
					/>
					<p class="text-xs text-gray-500 mt-1">
						{tr('forms.userMustBeRegistered')}
					</p>

					<!-- Autocomplete Dropdown -->
					{#if showSuggestions}
						<div
							class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto"
						>
							{#each suggestions as suggestion, index}
								<button
									type="button"
									onclick={() => selectSuggestion(suggestion)}
									class="w-full text-left px-4 py-2 hover:bg-cyan-50 border-b border-gray-100 last:border-b-0 {index ===
									selectedIndex
										? 'bg-cyan-50'
										: ''}"
								>
									<div class="flex flex-col">
										<span class="text-sm font-medium text-gray-900"
											>{suggestion.first_name}
											{suggestion.last_name}</span
										>
										<span class="text-xs text-gray-500">{suggestion.email}</span
										>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				</div>

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
