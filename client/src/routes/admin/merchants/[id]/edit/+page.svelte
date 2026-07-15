<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { merchantsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { MERCHANT_DEFAULT_COLOR } from '$lib/utils/merchant-color';
	import type { MerchantDTO } from '$lib/types/api';

	// Form state
	let merchant = $state<MerchantDTO | null>(null);
	let name = $state('');
	let color = $state(MERCHANT_DEFAULT_COLOR);
	// Writable derived: mirrors `color` but can be temporarily overridden by
	// the text input while the user types an (possibly invalid) value.
	let colorText = $derived(color);
	let logoUrl = $state('');
	let website = $state('');
	let isLoading = $state(true);
	let isSubmitting = $state(false);
	let showDeleteModal = $state(false);
	let errors = $state<{ [key: string]: string }>({});

	const merchantId = $derived($page.params.id!);
	const isOffline = $derived(!$isOnline);

	// Server-side guard in +layout.server.ts ensures user is admin
	// No client-side check needed (would be bypassable)
	onMount(async () => {
		await loadMerchant();
	});

	function isValidHexColor(value: string): boolean {
		return /^#[0-9A-Fa-f]{6}$/.test(value);
	}

	async function loadMerchant() {
		isLoading = true;
		try {
			const response = await merchantsApi.get(merchantId);
			merchant = response.merchant;

			// Populate form
			name = merchant.name;
			// Validate and sanitize color value
			const rawColor = (merchant.color || '').trim();
			const validColor =
				rawColor && isValidHexColor(rawColor)
					? rawColor
					: MERCHANT_DEFAULT_COLOR;
			color = validColor;
			logoUrl = merchant.logo_url || '';
			website = merchant.website || '';
		} catch {
			toastStore.error($t('admin.merchants.loadError'));
			goto(resolve('/merchants'));
		} finally {
			isLoading = false;
		}
	}

	function validate(): boolean {
		errors = {};

		if (!name.trim()) {
			errors.name = $t('admin.merchants.errors.nameRequired');
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validate()) {
			return;
		}

		isSubmitting = true;

		try {
			const input = {
				name: name.trim(),
				color: color || undefined,
				logo_url: logoUrl.trim() || undefined,
				website: website.trim() || undefined
			};

			await merchantsApi.update(merchantId, input);
			toastStore.success($t('admin.merchants.updateSuccess'));
			goto(resolve(`/merchants/${merchantId}`));
		} catch (err: unknown) {
			const message =
				err instanceof Error ? err.message : $t('admin.merchants.updateError');
			toastStore.error(message || $t('admin.merchants.updateError'));
		} finally {
			isSubmitting = false;
		}
	}

	async function handleDelete() {
		try {
			await merchantsApi.delete(merchantId);
			toastStore.success(
				$t('admin.merchants.deleteSuccess', { name: merchant?.name ?? '' })
			);
			goto(resolve('/merchants'));
		} catch {
			toastStore.error($t('admin.merchants.deleteError'));
		}
		showDeleteModal = false;
	}
</script>

<svelte:head>
	<title>{$t('admin.merchants.editMerchant')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Back Button -->
	<div class="mb-6">
		<a
			href={resolve('/merchants')}
			class="text-accent hover:text-accent-hover transition-colors"
		>
			{$t('common.backToOverview')}
		</a>
	</div>

	{#if isLoading}
		<LoadingSpinner />
	{:else if merchant}
		<!-- 2-Column Layout -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Left column: Form (2/3 width) -->
			<div class="lg:col-span-2">
				<div class="bg-white rounded-lg shadow-lg p-6">
					<h1 class="text-3xl font-bold text-text mb-6">
						{$t('admin.merchants.editMerchant')}
					</h1>

					<!-- Form -->
					<form
						onsubmit={(e) => {
							e.preventDefault();
							handleSubmit();
						}}
						class="space-y-6"
					>
						<!-- Merchant ID (read-only) -->
						<div>
							<label
								for="merchant-id"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.merchants.id')}
							</label>
							<input
								id="merchant-id"
								type="text"
								value={merchant.id}
								readonly
								class="input bg-surface-1 cursor-not-allowed font-mono text-sm"
							/>
						</div>

						<!-- Name -->
						<div>
							<label
								for="name"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.merchants.name')}
								<span class="text-danger-500">*</span>
							</label>
							<input
								id="name"
								type="text"
								bind:value={name}
								placeholder={$t('admin.merchants.namePlaceholder')}
								class="input {errors.name ? 'border-danger-500' : ''}"
								required
							/>
							{#if errors.name}
								<p class="mt-1 text-sm text-danger-600">{errors.name}</p>
							{/if}
						</div>

						<!-- Color -->
						<div>
							<label
								for="color"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.merchants.color')}
							</label>
							<div class="flex items-center gap-3">
								<input
									id="color"
									type="color"
									bind:value={color}
									class="h-10 w-20 rounded border border-border-field cursor-pointer"
								/>
								<input
									type="text"
									name="color-text"
									bind:value={colorText}
									onblur={() => {
										const val = colorText.trim();
										if (val && isValidHexColor(val)) {
											color = val;
										} else {
											colorText = color;
										}
									}}
									placeholder={MERCHANT_DEFAULT_COLOR}
									class="input flex-1"
								/>
							</div>
							<p class="mt-1 text-sm text-text-subtle">
								{$t('admin.merchants.colorHint')}
							</p>
						</div>

						<!-- Logo URL -->
						<div>
							<label
								for="logoUrl"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.merchants.logoUrl')}
							</label>
							<input
								id="logoUrl"
								type="url"
								bind:value={logoUrl}
								placeholder="https://example.com/logo.png"
								class="input bg-white"
							/>
							<p class="mt-1 text-sm text-text-subtle">
								{$t('admin.merchants.logoUrlHint')}
							</p>
							{#if logoUrl}
								<div class="mt-2">
									<img
										src={logoUrl}
										alt="Logo preview"
										class="h-12 w-auto"
										onerror={(e: Event) => {
											(e.currentTarget as HTMLImageElement).style.display =
												'none';
										}}
									/>
								</div>
							{/if}
						</div>

						<!-- Website -->
						<div>
							<label
								for="website"
								class="block text-sm font-medium text-text-ink2 mb-2"
							>
								{$t('admin.merchants.website')}
							</label>
							<input
								id="website"
								type="url"
								bind:value={website}
								placeholder="https://example.com"
								class="input bg-white"
							/>
							<p class="mt-1 text-sm text-text-subtle">
								{$t('admin.merchants.websiteHint')}
							</p>
						</div>

						<!-- Actions -->
						<div class="flex gap-3 pt-4">
							<button
								type="submit"
								disabled={isSubmitting || isOffline}
								class="btn btn-sm btn-primary flex-1 {isSubmitting || isOffline
									? 'opacity-50 cursor-not-allowed'
									: ''}"
							>
								{isSubmitting ? $t('common.saving') : $t('common.save')}
							</button>
							<a href={resolve('/merchants')} class="btn btn-sm btn-ghost">
								{$t('common.cancel')}
							</a>
						</div>

						<!-- Delete -->
						<div class="pt-4 border-t border-border">
							<button
								type="button"
								onclick={() => (showDeleteModal = true)}
								disabled={isOffline}
								class="btn btn-text-danger w-full flex items-center justify-center gap-1.5 {isOffline
									? 'pointer-events-none blur-[0.5px]'
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
								{/if}
								{$t('common.delete')}
							</button>
						</div>

						{#if isOffline}
							<div class="text-center text-sm text-danger-600">
								{$t('common.offlineWarning')}
							</div>
						{/if}
					</form>
				</div>
			</div>

			<!-- Right column: Info (1/3 width) -->
			<div class="lg:col-span-1">
				<div class="bg-white rounded-lg shadow-lg p-6">
					<h2 class="text-xl font-bold text-text mb-4">
						{$t('admin.merchants.title')}
					</h2>

					<div class="space-y-4 text-sm">
						<!-- Timestamps -->
						<div class="border border-border rounded-lg p-4 space-y-3">
							<div>
								<span
									class="text-xs font-medium text-text-subtle uppercase tracking-wider"
								>
									{$t('admin.merchants.createdAt')}
								</span>
								<p class="mt-1 text-text">
									{new Date(merchant.created_at).toLocaleString()}
								</p>
							</div>
							<div class="border-t border-border pt-3">
								<span
									class="text-xs font-medium text-text-subtle uppercase tracking-wider"
								>
									{$t('admin.merchants.updatedAt')}
								</span>
								<p class="mt-1 text-text">
									{new Date(merchant.updated_at).toLocaleString()}
								</p>
							</div>
						</div>

						<!-- Preview -->
						<div class="border border-border rounded-lg p-4">
							<h3
								class="text-xs font-medium text-text-subtle uppercase tracking-wider mb-3"
							>
								Preview
							</h3>
							<div class="flex items-center gap-3">
								{#if color}
									<div
										class="w-8 h-8 rounded border border-border-field"
										style="background-color: {color}"
									></div>
								{/if}
								<div class="flex-1 min-w-0">
									<p class="font-medium text-text truncate">
										{name || 'Merchant Name'}
									</p>
									{#if website}
										<a
											href={website}
											target="_blank"
											rel="noopener noreferrer external"
											class="text-xs text-accent hover:text-accent-800 truncate block"
										>
											{website}
										</a>
									{/if}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
	<div
		class="fixed inset-0 bg-black bg-opacity-50 z-40"
		onclick={() => (showDeleteModal = false)}
		role="presentation"
	></div>
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		role="dialog"
		aria-modal="true"
		aria-labelledby="delete-merchant-dialog-title"
	>
		<div class="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
			<div class="flex items-start mb-4">
				<div class="flex-shrink-0 mr-3">
					<svg
						class="h-6 w-6 text-danger-600"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
						/>
					</svg>
				</div>
				<div class="flex-1">
					<h3
						id="delete-merchant-dialog-title"
						class="text-lg font-semibold text-danger-600"
					>
						{$t('admin.merchants.confirmDeleteTitle')}
					</h3>
				</div>
			</div>
			<p class="text-text-muted mb-6 ml-9">
				{$t('admin.merchants.confirmDeleteMessage', {
					name: merchant?.name ?? ''
				})}
			</p>
			<div class="flex gap-3 justify-end">
				<button
					type="button"
					class="px-4 py-2 rounded-md border border-border-field hover:bg-surface-1 transition-colors text-text-ink2"
					onclick={() => (showDeleteModal = false)}
				>
					{$t('common.cancel')}
				</button>
				<button type="button" class="btn btn-danger" onclick={handleDelete}>
					{$t('common.delete')}
				</button>
			</div>
		</div>
	</div>
{/if}
