<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { merchantsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';

	// Form state
	let name = $state('');
	let color = $state('#3B82F6');
	let colorText = $derived(color);
	let logoUrl = $state('');
	let website = $state('');
	let isSubmitting = $state(false);
	let errors = $state<{ [key: string]: string }>({});

	const isOffline = $derived(!$isOnline);

	// Server-side guard in +layout.server.ts ensures user is admin
	// No client-side check needed (would be bypassable)

	function isValidHexColor(value: string): boolean {
		return /^#[0-9A-Fa-f]{6}$/.test(value);
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

			await merchantsApi.create(input);
			toastStore.success($t('admin.merchants.createSuccess'));
			goto(resolve('/merchants'));
		} catch (err) {
			const message =
				err instanceof Error ? err.message : $t('admin.merchants.createError');
			toastStore.error(message);
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>{$t('admin.merchants.createMerchant')} - {$t('common.appName')}</title>
</svelte:head>

<div class="px-4 pb-20 md:pb-4">
	<!-- Back Button -->
	<div class="mb-6">
		<a
			href={resolve('/merchants')}
			class="text-cyan-600 hover:text-cyan-700 transition-colors"
		>
			{$t('common.backToOverview')}
		</a>
	</div>

	<!-- 2-Column Layout -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Left column: Form (2/3 width) -->
		<div class="lg:col-span-2">
			<div class="bg-white rounded-lg shadow-lg p-6">
				<h1 class="text-3xl font-bold text-gray-900 mb-6">
					{$t('admin.merchants.createMerchant')}
				</h1>

				<!-- Form -->
				<form
					onsubmit={(e) => {
						e.preventDefault();
						handleSubmit();
					}}
					class="space-y-6"
				>
					<!-- Name -->
					<div>
						<label
							for="name"
							class="block text-sm font-medium text-gray-700 mb-2"
						>
							{$t('admin.merchants.name')} <span class="text-red-500">*</span>
						</label>
						<input
							id="name"
							type="text"
							bind:value={name}
							placeholder={$t('admin.merchants.namePlaceholder')}
							class="input {errors.name ? 'border-red-500' : ''}"
							required
						/>
						{#if errors.name}
							<p class="mt-1 text-sm text-red-600">{errors.name}</p>
						{/if}
					</div>

					<!-- Color -->
					<div>
						<label
							for="color"
							class="block text-sm font-medium text-gray-700 mb-2"
						>
							{$t('admin.merchants.color')}
						</label>
						<div class="flex items-center gap-3">
							<input
								id="color"
								type="color"
								bind:value={color}
								class="h-10 w-20 rounded border border-gray-300 cursor-pointer"
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
								placeholder="#3B82F6"
								class="input flex-1"
							/>
						</div>
						<p class="mt-1 text-sm text-gray-500">
							{$t('admin.merchants.colorHint')}
						</p>
					</div>

					<!-- Logo URL -->
					<div>
						<label
							for="logoUrl"
							class="block text-sm font-medium text-gray-700 mb-2"
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
						<p class="mt-1 text-sm text-gray-500">
							{$t('admin.merchants.logoUrlHint')}
						</p>
					</div>

					<!-- Website -->
					<div>
						<label
							for="website"
							class="block text-sm font-medium text-gray-700 mb-2"
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
						<p class="mt-1 text-sm text-gray-500">
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
							{isSubmitting ? $t('common.saving') : $t('common.create')}
						</button>
						<a href={resolve('/merchants')} class="btn btn-sm btn-ghost">
							{$t('common.cancel')}
						</a>
					</div>

					{#if isOffline}
						<div class="text-center text-sm text-red-600">
							{$t('common.offlineWarning')}
						</div>
					{/if}
				</form>
			</div>
		</div>

		<!-- Right column: Info (1/3 width) -->
		<div class="lg:col-span-1">
			<div class="bg-white rounded-lg shadow-lg p-6">
				<h2 class="text-xl font-bold text-gray-900 mb-4">
					{$t('admin.merchants.title')}
				</h2>
				<p class="text-sm text-gray-600 mb-4">
					{$t('admin.merchants.createSubtitle')}
				</p>

				<div class="space-y-4 text-sm">
					<div class="border-l-4 border-cyan-500 pl-4">
						<h3 class="font-semibold text-gray-900 mb-1">
							{$t('admin.merchants.name')}
						</h3>
						<p class="text-gray-600">{$t('admin.merchants.namePlaceholder')}</p>
					</div>

					<div class="border-l-4 border-purple-500 pl-4">
						<h3 class="font-semibold text-gray-900 mb-1">
							{$t('admin.merchants.color')}
						</h3>
						<p class="text-gray-600">{$t('admin.merchants.colorHint')}</p>
					</div>

					<div class="border-l-4 border-green-500 pl-4">
						<h3 class="font-semibold text-gray-900 mb-1">
							{$t('admin.merchants.logoUrl')}
						</h3>
						<p class="text-gray-600">{$t('admin.merchants.logoUrlHint')}</p>
					</div>

					<div class="border-l-4 border-orange-500 pl-4">
						<h3 class="font-semibold text-gray-900 mb-1">
							{$t('admin.merchants.website')}
						</h3>
						<p class="text-gray-600">{$t('admin.merchants.websiteHint')}</p>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
