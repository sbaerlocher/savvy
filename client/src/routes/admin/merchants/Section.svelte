<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { merchantsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';
	import { platform } from '$lib/utils/platform';
	import {
		ICON_CHECK,
		ICON_CHEVRON_DOWN,
		ICON_CHEVRON_RIGHT,
		ICON_CLOSE,
		ICON_FILTER_LINES,
		ICON_PLUS,
		ICON_SEARCH
	} from '$lib/icons';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { MERCHANT_DEFAULT_COLOR } from '$lib/utils/merchant-color';
	import type { MerchantDTO } from '$lib/types/api';

	const pageLogger = logger.child('AdminMerchantsPage');

	const DESKTOP = platform === 'other';

	// Android renders the M3 chrome shared with /admin/users: search pill plus
	// round filter button, tonal accordion cards and a create FAB. `platform`
	// is a module constant, so a plain const, not $derived.
	const IS_ANDROID = platform === 'android';

	// iOS renders liquid-glass chrome (same pattern as /admin/users): glass
	// search pill + filter button, glass accordion cards, create in the title
	// row via the page's `iosCreate` snippet.
	const IS_IOS = platform === 'ios';

	// State
	let merchants = $state<MerchantDTO[]>([]);
	let filteredMerchants = $state<MerchantDTO[]>([]);
	let isLoading = $state(true);
	let search = $state('');
	let sortBy = $state('name-asc');
	let showFilterMenu = $state(false);
	// iOS filter sheet: whether the sort disclosure row is expanded.
	let sortExpanded = $state(false);
	let expandedMerchantId = $state<string | null>(null);
	let deleteTarget = $state<MerchantDTO | null>(null);

	const isOffline = $derived(!$isOnline);
	const hasActiveFilters = $derived(search !== '' || sortBy !== 'name-asc');

	// Android filter sheet renders sort as an M3 radio list; the other
	// platforms keep their <select>s below.
	const sortOptions = $derived([
		{ value: 'name-asc', label: $t('admin.merchants.sortNameAsc') },
		{ value: 'name-desc', label: $t('admin.merchants.sortNameDesc') },
		{ value: 'newest', label: $t('admin.merchants.sortNewest') },
		{ value: 'oldest', label: $t('admin.merchants.sortOldest') }
	]);

	// ✅ Server-side admin check in +layout.server.ts (same as /admin/users)
	onMount(async () => {
		pageLogger.debug(
			'Admin access granted (server-side check), loading merchants'
		);
		await loadMerchants();
	});

	async function loadMerchants() {
		isLoading = true;
		try {
			const response = await merchantsApi.list();
			// $effect (below) re-runs applyFilters() reactively on merchants change.
			merchants = response.merchants;
		} catch {
			toastStore.error($t('admin.merchants.loadError'));
		} finally {
			isLoading = false;
		}
	}

	function applyFilters() {
		let result = merchants;

		// Search filter (by name, client-side like users page)
		if (search) {
			const query = search.toLowerCase();
			result = result.filter((m) => m.name.toLowerCase().includes(query));
		}

		// Sort
		result = [...result].sort((a, b) => {
			switch (sortBy) {
				case 'name-asc':
					return a.name.localeCompare(b.name);
				case 'name-desc':
					return b.name.localeCompare(a.name);
				case 'newest':
					return (
						new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
					);
				case 'oldest':
					return (
						new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
					);
				default:
					return 0;
			}
		});

		filteredMerchants = result;
	}

	function toggleExpandMerchant(id: string) {
		expandedMerchantId = expandedMerchantId === id ? null : id;
	}

	async function handleDelete() {
		if (!deleteTarget) return;
		const merchant = deleteTarget;
		deleteTarget = null;
		try {
			await merchantsApi.delete(merchant.id);
			toastStore.success(
				$t('admin.merchants.deleteSuccess', { name: merchant.name })
			);
			await loadMerchants();
		} catch {
			toastStore.error($t('admin.merchants.deleteError'));
		}
	}

	$effect(() => {
		applyFilters();
	});
</script>

{#if DESKTOP}
	<!-- Desktop: count line, toolbar and table share one elevated panel
	     (same pattern as /admin/users). -->
	<div class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-4' : ''} gap-6">
		<div class={showFilterMenu ? 'lg:col-span-3' : ''}>
			<div
				class="overflow-hidden rounded-4xl border border-border bg-surface shadow-panel"
			>
				<div class="px-7.5 pt-6 pb-4.5">
					<div class="mb-4.5">
						<p class="mt-0.5 text-label font-normal text-text-subtle">
							{merchants.length}
							{$t('admin.merchants.title')}
						</p>
					</div>

					<div class="flex gap-2.5">
						<label
							class="flex h-10 flex-1 items-center gap-2.25 rounded-lg border border-border-field bg-surface-1 px-3.5 focus-within:border-accent focus-within:ring-2 focus-within:ring-accent/20"
						>
							<svg
								class="h-4.25 w-4.25 shrink-0 text-text-subtle"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								viewBox="0 0 24 24"
								aria-hidden="true"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d={ICON_SEARCH}
								/>
							</svg>
							<input
								type="search"
								bind:value={search}
								placeholder={$t('common.search')}
								aria-label={$t('common.search')}
								class="min-w-0 flex-1 bg-transparent text-body text-text placeholder:text-text-placeholder focus:outline-none"
							/>
						</label>

						<button
							type="button"
							onclick={(e: MouseEvent) => {
								e.stopPropagation();
								showFilterMenu = !showFilterMenu;
							}}
							class="relative inline-flex h-10 items-center gap-2 rounded-lg border border-border-field bg-surface px-4 text-label text-text-ink2 transition-colors hover:bg-surface-1"
							aria-label={$t('common.filter')}
							aria-expanded={showFilterMenu}
						>
							<svg
								class="h-4.25 w-4.25"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								viewBox="0 0 24 24"
								aria-hidden="true"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d={ICON_FILTER_LINES}
								/>
							</svg>
							{$t('common.filter')}
							{#if hasActiveFilters}
								<span
									class="absolute -top-0.75 -right-0.75 h-2.5 w-2.5 rounded-full bg-accent"
								></span>
							{/if}
						</button>

						<a
							href={resolve('/admin/merchants/new')}
							class="inline-flex h-10 items-center rounded-lg bg-accent px-4.5 text-label whitespace-nowrap text-white shadow-sm transition-colors hover:bg-accent-hover {isOffline
								? 'pointer-events-none cursor-not-allowed opacity-50 blur-[0.5px]'
								: ''}"
						>
							{$t('admin.merchants.createMerchant')}
						</a>
					</div>
				</div>

				{#if isLoading}
					<div class="border-t border-border-soft"><LoadingSpinner /></div>
				{:else if filteredMerchants.length === 0}
					<div
						class="border-t border-border-soft py-12 text-center text-text-subtle"
					>
						{$t('admin.merchants.noMerchants')}
					</div>
				{:else}
					<div class="overflow-x-auto border-t border-border-soft">
						<table class="min-w-full">
							<thead>
								<tr class="border-b border-border-soft bg-surface-1">
									<th
										class="px-7.5 py-3 text-left text-section-eyebrow uppercase text-text-subtle"
									>
										{$t('admin.merchants.name')}
									</th>
									<th
										class="px-7.5 py-3 text-left text-section-eyebrow uppercase text-text-subtle"
									>
										{$t('admin.merchants.website')}
									</th>
									<th
										class="px-7.5 py-3 text-left text-section-eyebrow uppercase text-text-subtle"
									>
										{$t('admin.merchants.createdAt')}
									</th>
									<th
										class="px-7.5 py-3 text-right text-section-eyebrow uppercase text-text-subtle"
									>
										{$t('admin.merchants.actions')}
									</th>
								</tr>
							</thead>
							<tbody>
								{#each filteredMerchants as merchant (merchant.id)}
									<tr
										class="border-b border-border-soft transition-colors hover:bg-surface-1"
									>
										<td class="px-7.5 py-3.5 whitespace-nowrap">
											<span class="flex min-w-0 items-center gap-3">
												<span
													class="h-4 w-4 flex-shrink-0 rounded-full border border-border"
													style="background-color: {merchant.color ||
														MERCHANT_DEFAULT_COLOR}"
												></span>
												<span class="truncate text-body font-medium text-text"
													>{merchant.name}</span
												>
											</span>
										</td>
										<td
											class="max-w-xs truncate px-7.5 py-3.5 whitespace-nowrap text-body text-text-ink2"
										>
											{#if merchant.website}
												<!-- eslint-disable svelte/no-navigation-without-resolve -- external merchant website URL, not an app route -->
												<a
													href={merchant.website}
													target="_blank"
													rel="noopener noreferrer"
													class="text-accent hover:text-accent-900 transition-colors"
												>
													{merchant.website}
												</a>
												<!-- eslint-enable svelte/no-navigation-without-resolve -->
											{:else}
												—
											{/if}
										</td>
										<td
											class="px-7.5 py-3.5 whitespace-nowrap text-body text-text-ink2"
										>
											{new Date(merchant.created_at).toLocaleDateString()}
										</td>
										<td class="px-7.5 py-3.5 text-right whitespace-nowrap">
											<button
												onclick={() => toggleExpandMerchant(merchant.id)}
												class="text-label text-accent hover:text-accent-900 transition-colors"
											>
												{expandedMerchantId === merchant.id
													? $t('common.close')
													: $t('admin.merchants.showDetails')}
											</button>
										</td>
									</tr>
									{#if expandedMerchantId === merchant.id}
										<tr class="border-b border-border-soft bg-surface-1">
											<td colspan="4" class="px-7.5 py-4">
												<div class="space-y-3">
													<!-- Merchant Details -->
													<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
														{#if merchant.description}
															<div class="md:col-span-2">
																<span
																	class="text-xs font-medium text-text-subtle"
																	>{$t('common.description')}:</span
																>
																<p class="text-sm text-text">
																	{merchant.description}
																</p>
															</div>
														{/if}
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.id')}:</span
															>
															<p class="font-mono text-sm break-all text-text">
																{merchant.id}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.color')}:</span
															>
															<p
																class="flex items-center gap-2 font-mono text-sm text-text"
															>
																<span
																	class="inline-block h-4 w-4 rounded-full border border-border"
																	style="background-color: {merchant.color ||
																		MERCHANT_DEFAULT_COLOR}"
																></span>
																{merchant.color || '—'}
															</p>
														</div>
													</div>

													<!-- Actions -->
													<div
														class="flex flex-wrap gap-2 border-t border-border pt-2"
													>
														<a
															href={resolve(
																`/admin/merchants/${merchant.id}/edit`
															)}
															class="btn btn-sm btn-primary {isOffline
																? 'opacity-50 cursor-not-allowed pointer-events-none'
																: ''}"
														>
															{$t('common.edit')}
														</a>
														<button
															onclick={() => (deleteTarget = merchant)}
															disabled={isOffline}
															class="btn btn-sm btn-ghost text-danger-600 hover:text-danger-700 {isOffline
																? 'opacity-50 cursor-not-allowed'
																: ''}"
														>
															{$t('common.delete')}
														</button>
													</div>
												</div>
											</td>
										</tr>
									{/if}
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>

		<!-- Filter Side-Panel (Desktop only, 1/4 width, sticky) -->
		{#if showFilterMenu}
			<div class="hidden lg:block lg:col-span-1">
				<div class="bg-white rounded-lg shadow-lg p-6 sticky top-4">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-lg font-semibold text-text">
							{$t('common.filter')}
						</h3>
						<button
							type="button"
							onclick={() => (showFilterMenu = false)}
							class="text-text-faint hover:text-text-muted transition-colors"
							aria-label={$t('common.close')}
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
									d="M6 18L18 6M6 6l12 12"
								></path>
							</svg>
						</button>
					</div>

					<div class="space-y-4">
						<div>
							<label
								for="sortBy"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('common.sort')}
							</label>
							<select id="sortBy" bind:value={sortBy} class="input text-sm">
								<option value="name-asc"
									>{$t('admin.merchants.sortNameAsc')}</option
								>
								<option value="name-desc"
									>{$t('admin.merchants.sortNameDesc')}</option
								>
								<option value="newest"
									>{$t('admin.merchants.sortNewest')}</option
								>
								<option value="oldest"
									>{$t('admin.merchants.sortOldest')}</option
								>
							</select>
						</div>

						{#if hasActiveFilters}
							<button
								type="button"
								onclick={() => {
									search = '';
									sortBy = 'name-asc';
								}}
								class="w-full btn btn-sm btn-ghost text-sm"
							>
								{$t('common.resetFilters')}
							</button>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>
{:else if IS_IOS}
	<!-- Glass search pill + filter button carrying the active-filter dot. -->
	<div class="mb-3.5 flex gap-2.25">
		<label
			class="liquid-glass-surface flex h-10.5 flex-1 items-center gap-2.25 rounded-xl px-3.5"
		>
			<svg
				class="h-4.25 w-4.25 shrink-0 text-text-subtle"
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
				bind:value={search}
				placeholder={$t('common.search')}
				aria-label={$t('common.search')}
				class="min-w-0 flex-1 bg-transparent text-body text-text placeholder:text-text-placeholder focus:outline-none"
			/>
		</label>
		<button
			type="button"
			onclick={(e: MouseEvent) => {
				e.stopPropagation();
				showFilterMenu = !showFilterMenu;
			}}
			aria-label={$t('common.filter')}
			aria-expanded={showFilterMenu}
			class="liquid-glass-surface relative flex h-10.5 w-10.5 shrink-0 items-center justify-center rounded-xl text-text-muted"
		>
			<svg
				class="h-4.5 w-4.5"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				viewBox="0 0 24 24"
				aria-hidden="true"
			>
				<path d={ICON_FILTER_LINES} />
			</svg>
			{#if hasActiveFilters}
				<span
					class="absolute top-2.25 right-2.25 h-1.75 w-1.75 rounded-full border-2 border-surface bg-accent-600"
				></span>
			{/if}
		</button>
	</div>

	{#if isLoading}
		<LoadingSpinner />
	{:else if filteredMerchants.length === 0}
		<div class="text-center py-12 text-text-subtle">
			{$t('admin.merchants.noMerchants')}
		</div>
	{:else}
		<!-- iOS: each merchant is its own grouped-inset glass card that expands
		     in place (same chrome as /admin/users). -->
		<div class="flex flex-col gap-2.5">
			{#each filteredMerchants as merchant (merchant.id)}
				{@const isExpanded = expandedMerchantId === merchant.id}
				<div
					class="liquid-glass-card overflow-hidden rounded-[var(--radius-inset)]"
				>
					<button
						type="button"
						onclick={() => toggleExpandMerchant(merchant.id)}
						aria-expanded={isExpanded}
						class="flex w-full items-center gap-3 px-3.75 py-3.25 text-left transition-colors active:bg-surface-1"
					>
						<span
							aria-hidden="true"
							class="h-9.5 w-9.5 shrink-0 rounded-full border border-border"
							style="background-color: {merchant.color ||
								MERCHANT_DEFAULT_COLOR}"
						></span>
						<span class="min-w-0 flex-1">
							<span class="block truncate text-body font-semibold text-text"
								>{merchant.name}</span
							>
							<span
								class="mt-px block truncate text-chip font-normal text-text-subtle"
								>{merchant.website ||
									new Date(merchant.created_at).toLocaleDateString()}</span
							>
						</span>
						<svg
							class="h-4 w-4 shrink-0 text-text-faint transition-transform {isExpanded
								? 'rotate-180'
								: ''}"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path d={ICON_CHEVRON_DOWN} />
						</svg>
					</button>

					{#if isExpanded}
						<div
							class="border-t border-border-soft bg-surface-2 px-3.75 py-3.5"
						>
							<div class="mb-3.5 grid grid-cols-2 gap-x-3.5 gap-y-3">
								{#if merchant.description}
									<div class="col-span-2">
										<div class="mb-0.75 text-tag text-text-faint uppercase">
											{$t('common.description')}
										</div>
										<div class="text-label font-normal text-text">
											{merchant.description}
										</div>
									</div>
								{/if}
								{#if merchant.website}
									<div class="col-span-2">
										<div class="mb-0.75 text-tag text-text-faint uppercase">
											{$t('admin.merchants.website')}
										</div>
										<!-- eslint-disable svelte/no-navigation-without-resolve -- external merchant website URL, not an app route -->
										<a
											href={merchant.website}
											target="_blank"
											rel="noopener noreferrer"
											class="text-label break-all text-accent"
										>
											{merchant.website}
										</a>
										<!-- eslint-enable svelte/no-navigation-without-resolve -->
									</div>
								{/if}
								<div class="col-span-2">
									<div class="mb-0.75 text-tag text-text-faint uppercase">
										{$t('admin.merchants.id')}
									</div>
									<div class="font-mono text-chip font-normal text-text">
										{merchant.id}
									</div>
								</div>
								<div>
									<div class="mb-0.75 text-tag text-text-faint uppercase">
										{$t('admin.merchants.color')}
									</div>
									<div
										class="flex items-center gap-2 font-mono text-label font-normal text-text"
									>
										<span
											class="inline-block h-4 w-4 rounded-full border border-border"
											style="background-color: {merchant.color ||
												MERCHANT_DEFAULT_COLOR}"
										></span>
										{merchant.color || '—'}
									</div>
								</div>
								<div>
									<div class="mb-0.75 text-tag text-text-faint uppercase">
										{$t('admin.merchants.createdAt')}
									</div>
									<div class="text-label font-normal text-text">
										{new Date(merchant.created_at).toLocaleDateString()}
									</div>
								</div>
							</div>

							<div class="flex flex-wrap gap-2">
								<a
									href={resolve(`/admin/merchants/${merchant.id}/edit`)}
									class="inline-flex h-9 items-center rounded-full bg-accent px-4 text-chip text-on-accent {isOffline
										? 'pointer-events-none opacity-45'
										: ''}"
								>
									{$t('common.edit')}
								</a>
								<button
									type="button"
									onclick={() => (deleteTarget = merchant)}
									disabled={isOffline}
									class="inline-flex h-9 items-center rounded-full border border-border-field bg-surface px-4 text-chip text-danger-600 disabled:opacity-45"
								>
									{$t('common.delete')}
								</button>
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
{:else if IS_ANDROID}
	<!-- Create merchant as the screen's FAB (M3: one FAB per screen). MobileNav
	     drops its "New" FAB on admin routes, so this one takes the slot above
	     the bottom nav. -->
	<a
		href={resolve('/admin/merchants/new')}
		aria-label={$t('admin.merchants.createMerchant')}
		class="mobile-nav-fab-android bg-accent text-on-accent fixed right-4.5 z-50 flex h-14 w-14 items-center justify-center rounded-m3-lg shadow-[var(--shadow-fab)] {isOffline
			? 'pointer-events-none opacity-50'
			: ''}"
	>
		<svg
			class="h-6 w-6"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path d={ICON_PLUS} />
		</svg>
	</a>

	<!-- Docked search pill plus the round filter button carrying the active
	     dot. -->
	<div class="mb-3.5 flex gap-2">
		<div
			class="bg-m3-surface-container-high rounded-m3-full flex h-11.5 flex-1 items-center gap-2.5 px-4"
		>
			<svg
				class="text-text-muted h-4.5 w-4.5 shrink-0"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path stroke-linecap="round" stroke-linejoin="round" d={ICON_SEARCH} />
			</svg>
			<input
				type="search"
				bind:value={search}
				placeholder={$t('common.search')}
				aria-label={$t('common.search')}
				class="text-body text-text placeholder:text-text-subtle min-w-0 flex-1 bg-transparent focus:outline-none"
			/>
		</div>
		<button
			type="button"
			onclick={() => (showFilterMenu = true)}
			aria-label={$t('common.filter')}
			aria-expanded={showFilterMenu}
			class="bg-m3-surface-container-high text-text-ink2 rounded-m3-full relative inline-flex h-11.5 w-11.5 shrink-0 items-center justify-center"
		>
			<svg
				class="h-4.75 w-4.75"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d={ICON_FILTER_LINES}
				/>
			</svg>
			{#if hasActiveFilters}
				<span
					class="bg-accent border-m3-surface-container-high absolute top-2.5 right-2.75 h-1.75 w-1.75 rounded-m3-sm border-[1.5px]"
				></span>
			{/if}
		</button>
	</div>

	<!-- One tonal card per merchant, expanding into the detail block. -->
	<div class="flex flex-col gap-2.5">
		{#if isLoading}
			<LoadingSpinner />
		{:else if filteredMerchants.length === 0}
			<div class="text-text-subtle py-12 text-center">
				{$t('admin.merchants.noMerchants')}
			</div>
		{:else}
			{#each filteredMerchants as merchant (merchant.id)}
				{@const expanded = expandedMerchantId === merchant.id}
				<div
					class="rounded-m3-lg bg-m3-card border-border overflow-hidden border"
				>
					<button
						type="button"
						onclick={() => toggleExpandMerchant(merchant.id)}
						aria-expanded={expanded}
						class="hover:bg-ground-active flex w-full items-center gap-3.5 px-4 py-3.25 text-left transition-colors"
					>
						<span
							aria-hidden="true"
							class="border-border rounded-m3-full h-10 w-10 shrink-0 border"
							style="background-color: {merchant.color ||
								MERCHANT_DEFAULT_COLOR}"
						></span>
						<span class="min-w-0 flex-1">
							<span class="text-body text-text block truncate font-semibold"
								>{merchant.name}</span
							>
							<span class="text-body-sm text-text-muted mt-px block truncate"
								>{merchant.website ||
									new Date(merchant.created_at).toLocaleDateString()}</span
							>
						</span>
						<svg
							class="text-text-subtle h-4 w-4 shrink-0 transition-transform {expanded
								? 'rotate-180'
								: ''}"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d={ICON_CHEVRON_DOWN}
							/>
						</svg>
					</button>

					{#if expanded}
						<div class="bg-surface-1 px-4 pt-0.5 pb-4">
							<div class="border-accent-100 border-l-2 pt-3 pl-3.5">
								<div class="mb-3.5 grid grid-cols-2 gap-x-3.5 gap-y-3">
									{#if merchant.description}
										<div class="col-span-2">
											<div class="text-tag text-text-faint mb-0.75 uppercase">
												{$t('common.description')}
											</div>
											<div class="text-label text-text font-normal">
												{merchant.description}
											</div>
										</div>
									{/if}
									{#if merchant.website}
										<div class="col-span-2">
											<div class="text-tag text-text-faint mb-0.75 uppercase">
												{$t('admin.merchants.website')}
											</div>
											<!-- eslint-disable svelte/no-navigation-without-resolve -- external merchant website URL, not an app route -->
											<a
												href={merchant.website}
												target="_blank"
												rel="noopener noreferrer"
												class="text-accent text-label break-all"
											>
												{merchant.website}
											</a>
											<!-- eslint-enable svelte/no-navigation-without-resolve -->
										</div>
									{/if}
									<div class="col-span-2">
										<div class="text-tag text-text-faint mb-0.75 uppercase">
											{$t('admin.merchants.id')}
										</div>
										<div class="font-mono text-body-sm text-text truncate">
											{merchant.id}
										</div>
									</div>
									<div>
										<div class="text-tag text-text-faint mb-0.75 uppercase">
											{$t('admin.merchants.color')}
										</div>
										<div
											class="text-label text-text flex items-center gap-2 font-mono font-normal"
										>
											<span
												class="border-border inline-block h-4 w-4 rounded-full border"
												style="background-color: {merchant.color ||
													MERCHANT_DEFAULT_COLOR}"
											></span>
											{merchant.color || '—'}
										</div>
									</div>
									<div>
										<div class="text-tag text-text-faint mb-0.75 uppercase">
											{$t('admin.merchants.createdAt')}
										</div>
										<div class="text-label text-text font-normal">
											{new Date(merchant.created_at).toLocaleDateString()}
										</div>
									</div>
								</div>

								<div class="flex flex-wrap gap-2">
									<a
										href={resolve(`/admin/merchants/${merchant.id}/edit`)}
										class="bg-accent text-on-accent rounded-m3-full text-body-sm inline-flex h-9 items-center px-4 font-semibold {isOffline
											? 'pointer-events-none opacity-45'
											: ''}"
									>
										{$t('common.edit')}
									</a>
									<button
										type="button"
										onclick={() => (deleteTarget = merchant)}
										disabled={isOffline}
										class="border-border-field bg-m3-card text-danger-600 rounded-m3-full text-body-sm inline-flex h-9 items-center border px-4 font-semibold disabled:opacity-45"
									>
										{$t('common.delete')}
									</button>
								</div>
							</div>
						</div>
					{/if}
				</div>
			{/each}
		{/if}
	</div>
{:else}
	<!-- Search bar and action buttons -->
	<div class="mb-6 flex flex-col sm:flex-row gap-3">
		<!-- Search Bar -->
		<div class="flex-1">
			<input
				type="text"
				bind:value={search}
				placeholder={$t('common.search')}
				class="input bg-white"
			/>
		</div>

		<!-- Action Buttons (Desktop) -->
		<div class="hidden sm:flex gap-3">
			<!-- Filter Button -->
			<button
				type="button"
				onclick={(e: MouseEvent) => {
					e.stopPropagation();
					showFilterMenu = !showFilterMenu;
				}}
				class="flex items-center justify-center gap-2 control px-4 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
				title={$t('common.filter')}
				aria-label={$t('common.filter')}
				aria-expanded={showFilterMenu}
			>
				<svg
					class="w-5 h-5 text-text-muted"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
					/>
				</svg>
				{#if hasActiveFilters}
					<span class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
					></span>
				{/if}
			</button>

			<!-- Create Merchant Button -->
			<a
				href={resolve('/admin/merchants/new')}
				class="btn btn-primary whitespace-nowrap {isOffline
					? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
					: ''}"
			>
				{$t('admin.merchants.createMerchant')}
			</a>
		</div>

		<!-- Action Buttons (Mobile) -->
		<div class="flex sm:hidden gap-3">
			<!-- Filter Button (Mobile) -->
			<button
				type="button"
				onclick={(e: MouseEvent) => {
					e.stopPropagation();
					showFilterMenu = !showFilterMenu;
				}}
				class="flex-1 flex items-center justify-center control px-3 bg-white border border-border-field rounded-md hover:bg-surface-1 transition-colors relative"
				aria-label={$t('common.filter')}
				aria-expanded={showFilterMenu}
			>
				<svg
					class="w-5 h-5 text-text-muted"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
					/>
				</svg>
				{#if hasActiveFilters}
					<span class="absolute -top-1 -right-1 w-3 h-3 bg-accent rounded-full"
					></span>
				{/if}
			</button>

			<!-- Create Merchant Button (Mobile) -->
			<a
				href={resolve('/admin/merchants/new')}
				class="btn btn-sm btn-primary flex-1 text-center {isOffline
					? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
					: ''}"
			>
				{$t('admin.merchants.createMerchant')}
			</a>
		</div>
	</div>

	<!-- Grid with optional Side-Panel -->
	<div class="grid grid-cols-1 {showFilterMenu ? 'lg:grid-cols-4' : ''} gap-6">
		<!-- Merchants Table (3/4 when filter is open on desktop) -->
		<div class={showFilterMenu ? 'lg:col-span-3' : ''}>
			{#if isLoading}
				<LoadingSpinner />
			{:else if filteredMerchants.length === 0}
				<div class="text-center py-12 text-text-subtle">
					{$t('admin.merchants.noMerchants')}
				</div>
			{:else}
				<div class="bg-white shadow rounded-lg overflow-hidden">
					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-border">
							<thead class="bg-surface-1">
								<tr>
									<th
										class="px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.merchants.name')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.merchants.website')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-left text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.merchants.createdAt')}
									</th>
									<th
										class="hidden md:table-cell px-6 py-3 text-right text-xs font-medium text-text-subtle uppercase tracking-wider"
									>
										{$t('admin.merchants.actions')}
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-border">
								{#each filteredMerchants as merchant (merchant.id)}
									<tr
										class="hover:bg-surface-1 transition-colors md:cursor-default cursor-pointer"
										onclick={() => {
											if (window.innerWidth < 768)
												toggleExpandMerchant(merchant.id);
										}}
									>
										<td
											class="px-6 py-4 whitespace-nowrap text-sm font-medium text-text"
										>
											<div class="flex items-center justify-between">
												<span class="flex items-center gap-3 min-w-0">
													<span
														class="w-4 h-4 rounded-full flex-shrink-0 border border-border"
														style="background-color: {merchant.color ||
															MERCHANT_DEFAULT_COLOR}"
													></span>
													<span class="truncate">{merchant.name}</span>
												</span>
												<svg
													class="w-4 h-4 text-text-faint ml-2 md:hidden flex-shrink-0 transition-transform {expandedMerchantId ===
													merchant.id
														? 'rotate-180'
														: ''}"
													fill="none"
													stroke="currentColor"
													viewBox="0 0 24 24"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M19 9l-7 7-7-7"
													/>
												</svg>
											</div>
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-text-subtle max-w-xs truncate"
										>
											{#if merchant.website}
												<!-- eslint-disable svelte/no-navigation-without-resolve -- external merchant website URL, not an app route -->
												<a
													href={merchant.website}
													target="_blank"
													rel="noopener noreferrer"
													class="text-accent hover:text-accent-900 transition-colors"
													onclick={(e) => e.stopPropagation()}
												>
													{merchant.website}
												</a>
												<!-- eslint-enable svelte/no-navigation-without-resolve -->
											{:else}
												—
											{/if}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-sm text-text-subtle"
										>
											{new Date(merchant.created_at).toLocaleDateString()}
										</td>
										<td
											class="hidden md:table-cell px-6 py-4 whitespace-nowrap text-right text-sm"
										>
											<button
												onclick={() => toggleExpandMerchant(merchant.id)}
												class="text-accent hover:text-accent-900 font-medium transition-colors"
											>
												{expandedMerchantId === merchant.id
													? $t('common.close')
													: $t('admin.merchants.showDetails')}
											</button>
										</td>
									</tr>
									{#if expandedMerchantId === merchant.id}
										<tr class="bg-surface-1">
											<td colspan="4" class="px-6 py-4">
												<div class="space-y-3">
													<!-- Merchant Details -->
													<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
														{#if merchant.description}
															<div class="md:col-span-2">
																<span
																	class="text-xs font-medium text-text-subtle"
																	>{$t('common.description')}:</span
																>
																<p class="text-sm text-text">
																	{merchant.description}
																</p>
															</div>
														{/if}
														<div class="md:hidden">
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.website')}:</span
															>
															<p class="text-sm text-text">
																{#if merchant.website}
																	<!-- eslint-disable svelte/no-navigation-without-resolve -- external merchant website URL, not an app route -->
																	<a
																		href={merchant.website}
																		target="_blank"
																		rel="noopener noreferrer"
																		class="text-accent hover:text-accent-900 transition-colors break-all"
																	>
																		{merchant.website}
																	</a>
																	<!-- eslint-enable svelte/no-navigation-without-resolve -->
																{:else}
																	—
																{/if}
															</p>
														</div>
														<div class="md:hidden">
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.createdAt')}:</span
															>
															<p class="text-sm text-text">
																{new Date(merchant.created_at).toLocaleString()}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.id')}:</span
															>
															<p class="text-sm text-text font-mono break-all">
																{merchant.id}
															</p>
														</div>
														<div>
															<span class="text-xs font-medium text-text-subtle"
																>{$t('admin.merchants.color')}:</span
															>
															<p
																class="text-sm text-text font-mono flex items-center gap-2"
															>
																<span
																	class="w-4 h-4 rounded-full inline-block border border-border"
																	style="background-color: {merchant.color ||
																		MERCHANT_DEFAULT_COLOR}"
																></span>
																{merchant.color || '—'}
															</p>
														</div>
													</div>

													<!-- Actions -->
													<div
														class="flex flex-wrap gap-2 pt-2 border-t border-border"
													>
														<a
															href={resolve(
																`/admin/merchants/${merchant.id}/edit`
															)}
															class="btn btn-sm btn-primary {isOffline
																? 'opacity-50 cursor-not-allowed pointer-events-none'
																: ''}"
														>
															{$t('common.edit')}
														</a>
														<button
															onclick={() => (deleteTarget = merchant)}
															disabled={isOffline}
															class="btn btn-sm btn-ghost text-danger-600 hover:text-danger-700 {isOffline
																? 'opacity-50 cursor-not-allowed'
																: ''}"
														>
															{$t('common.delete')}
														</button>
													</div>
												</div>
											</td>
										</tr>
									{/if}
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		</div>

		<!-- Filter Side-Panel (Desktop only, 1/4 width, sticky) -->
		{#if showFilterMenu}
			<div class="hidden lg:block lg:col-span-1">
				<div class="bg-white rounded-lg shadow-lg p-6 sticky top-4">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-lg font-semibold text-text">
							{$t('common.filter')}
						</h3>
						<button
							type="button"
							onclick={() => (showFilterMenu = false)}
							class="text-text-faint hover:text-text-muted transition-colors"
							aria-label={$t('common.close')}
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
									d="M6 18L18 6M6 6l12 12"
								></path>
							</svg>
						</button>
					</div>

					<div class="space-y-4">
						<div>
							<label
								for="sortBy"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('common.sort')}
							</label>
							<select id="sortBy" bind:value={sortBy} class="input text-sm">
								<option value="name-asc"
									>{$t('admin.merchants.sortNameAsc')}</option
								>
								<option value="name-desc"
									>{$t('admin.merchants.sortNameDesc')}</option
								>
								<option value="newest"
									>{$t('admin.merchants.sortNewest')}</option
								>
								<option value="oldest"
									>{$t('admin.merchants.sortOldest')}</option
								>
							</select>
						</div>

						{#if hasActiveFilters}
							<button
								type="button"
								onclick={() => {
									search = '';
									sortBy = 'name-asc';
								}}
								class="w-full btn btn-sm btn-ghost text-sm"
							>
								{$t('common.resetFilters')}
							</button>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>
{/if}

<!-- Mobile Filter Bottom Sheet -->
<BottomSheet
	open={showFilterMenu}
	onClose={() => (showFilterMenu = false)}
	maxHeight={IS_IOS ? '88%' : '80vh'}
	ariaLabel={$t('common.filter')}
	tonalAndroid
>
	{#if IS_IOS}
		<!-- iOS filter sheet (same chrome as /admin/users): "Done" text action in
		     the header, sort as a disclosure row opening an inset list. -->
		<div class="flex flex-col">
			<div
				class="flex items-center justify-between border-b border-border-soft px-5 pt-1.5 pb-3"
			>
				<span class="text-heading text-text">{$t('common.filter')}</span>
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="text-[length:var(--text-code)] font-semibold text-accent-700 transition-opacity active:opacity-60"
				>
					{$t('common.done')}
				</button>
			</div>

			<div class="px-5 pt-4 pb-7">
				<div class="mb-2.25 text-tag text-text-subtle uppercase">
					{$t('common.sort')}
				</div>
				<button
					type="button"
					onclick={() => (sortExpanded = !sortExpanded)}
					aria-expanded={sortExpanded}
					class="flex h-12 w-full items-center justify-between rounded-xl border border-border bg-surface px-4 text-left"
				>
					<span class="text-[length:var(--text-code)] font-normal text-text">
						{sortOptions.find((o) => o.value === sortBy)?.label ?? ''}
					</span>
					<svg
						class="h-3.5 w-2 shrink-0 text-text-faint transition-transform {sortExpanded
							? 'rotate-90'
							: ''}"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						viewBox="0 0 8 14"
						aria-hidden="true"
					>
						<path d={ICON_CHEVRON_RIGHT} />
					</svg>
				</button>

				{#if sortExpanded}
					<div
						role="radiogroup"
						aria-label={$t('common.sort')}
						class="mt-1.5 overflow-hidden rounded-xl border border-border bg-surface"
					>
						{#each sortOptions as opt, i (opt.value)}
							{@const selected = sortBy === opt.value}
							<button
								type="button"
								role="radio"
								aria-checked={selected}
								onclick={() => {
									sortBy = opt.value;
									sortExpanded = false;
								}}
								class="flex w-full items-center justify-between px-4 py-2.75 text-left text-[length:var(--text-code)] font-normal {selected
									? 'text-text'
									: 'text-text-muted'} {i < sortOptions.length - 1
									? 'border-b border-border-soft'
									: ''}"
							>
								{opt.label}
								{#if selected}
									<svg
										class="h-4 w-4 shrink-0 text-accent"
										fill="none"
										stroke="currentColor"
										stroke-width="2.4"
										stroke-linecap="round"
										stroke-linejoin="round"
										viewBox="0 0 24 24"
										aria-hidden="true"
									>
										<path d={ICON_CHECK} />
									</svg>
								{/if}
							</button>
						{/each}
					</div>
				{/if}

				{#if hasActiveFilters}
					<div class="flex justify-center pt-4.5 pb-0.5">
						<button
							type="button"
							onclick={() => {
								search = '';
								sortBy = 'name-asc';
							}}
							class="text-[length:var(--text-label)] font-medium text-text-muted"
						>
							{$t('common.resetFilters')}
						</button>
					</div>
				{/if}
			</div>
		</div>
	{:else if IS_ANDROID}
		<!-- Android M3 filter sheet (same chrome as /admin/users): sort as a radio
	     list, reset + apply in the footer. -->
		<div class="px-5.5 pt-1 pb-2">
			<div class="mb-5 flex items-center justify-between">
				<h3 class="text-heading text-text">{$t('common.filter')}</h3>
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					aria-label={$t('common.close')}
					class="text-text-subtle hover:text-text-muted flex transition-colors"
				>
					<svg
						class="h-5.5 w-5.5"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						viewBox="0 0 24 24"
					>
						<path stroke-linecap="round" d={ICON_CLOSE} />
					</svg>
				</button>
			</div>

			<div class="text-label text-text-ink2 mb-2.5 block">
				{$t('common.sort')}
			</div>
			<div class="mb-6" role="radiogroup" aria-label={$t('common.sort')}>
				{#each sortOptions as option (option.value)}
					<button
						type="button"
						role="radio"
						aria-checked={sortBy === option.value}
						onclick={() => (sortBy = option.value)}
						class="text-body flex w-full items-center gap-3 py-2.5 text-left {sortBy ===
						option.value
							? 'text-text'
							: 'text-text-muted'}"
					>
						<span class="flex h-4 w-4 shrink-0 items-center justify-center">
							{#if sortBy === option.value}
								<svg
									class="text-accent h-4 w-4"
									fill="none"
									stroke="currentColor"
									stroke-width="2.4"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d={ICON_CHECK}
									/>
								</svg>
							{/if}
						</span>
						{option.label}
					</button>
				{/each}
			</div>

			<div class="flex justify-end gap-2">
				<button
					type="button"
					onclick={() => {
						search = '';
						sortBy = 'name-asc';
					}}
					class="text-accent rounded-m3-full text-label inline-flex h-10.5 items-center px-5"
				>
					{$t('common.reset')}
				</button>
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="bg-accent text-on-accent rounded-m3-full text-label inline-flex h-10.5 items-center px-6"
				>
					{$t('common.apply')}
				</button>
			</div>
		</div>
	{:else}
		<div class="p-6">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-lg font-semibold text-text">
					{$t('common.filter')}
				</h3>
				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="text-text-faint hover:text-text-muted transition-colors"
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
						></path>
					</svg>
				</button>
			</div>

			<div class="space-y-4">
				<div>
					<label
						for="sortByMobile"
						class="block text-sm font-medium text-text-ink2 mb-2"
					>
						{$t('common.sort')}
					</label>
					<select id="sortByMobile" bind:value={sortBy} class="input bg-white">
						<option value="name-asc">{$t('admin.merchants.sortNameAsc')}</option
						>
						<option value="name-desc"
							>{$t('admin.merchants.sortNameDesc')}</option
						>
						<option value="newest">{$t('admin.merchants.sortNewest')}</option>
						<option value="oldest">{$t('admin.merchants.sortOldest')}</option>
					</select>
				</div>

				{#if hasActiveFilters}
					<button
						type="button"
						onclick={() => {
							search = '';
							sortBy = 'name-asc';
						}}
						class="w-full btn btn-ghost"
					>
						{$t('common.resetFilters')}
					</button>
				{/if}

				<button
					type="button"
					onclick={() => (showFilterMenu = false)}
					class="w-full btn btn-primary"
				>
					{$t('common.done')}
				</button>
			</div>
		</div>
	{/if}
</BottomSheet>

<!-- Delete confirmation -->
<ConfirmModal
	isOpen={deleteTarget !== null}
	title={$t('admin.merchants.confirmDeleteTitle')}
	message={$t('admin.merchants.confirmDeleteMessage', {
		name: deleteTarget?.name ?? ''
	})}
	confirmText={$t('common.delete')}
	cancelText={$t('common.cancel')}
	variant="danger"
	onconfirm={handleDelete}
	oncancel={() => (deleteTarget = null)}
/>
