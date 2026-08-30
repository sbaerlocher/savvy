<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { isOnline } from '$lib/stores/offline';
	import { t } from '$lib/stores/i18n';
	import { merchantsApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { MERCHANT_DEFAULT_COLOR } from '$lib/utils/merchant-color';
	import { platform } from '$lib/utils/platform';

	// Android renders the form M3-flavoured (m3-filled-form restyles the
	// shared .input/.label/.btn classes); `platform` is a module constant.
	const IS_ANDROID = platform === 'android';

	// Form state
	let name = $state('');
	let color = $state(MERCHANT_DEFAULT_COLOR);
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

<!-- 2-Column Layout -->
<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Left column: Form (2/3 width) -->
	<div class="lg:col-span-2">
		<div
			class={IS_ANDROID
				? 'm3-filled-form rounded-m3-lg bg-m3-card border border-border p-4'
				: 'bg-white rounded-lg shadow-lg p-6'}
		>
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
					<label for="name" class="label">
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
					<label for="color" class="label">
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
					<label for="logoUrl" class="label">
						{$t('admin.merchants.logoUrl')}
					</label>
					<input
						id="logoUrl"
						type="url"
						bind:value={logoUrl}
						placeholder="https://example.com/logo.png"
						class="input"
					/>
					<p class="mt-1 text-sm text-text-subtle">
						{$t('admin.merchants.logoUrlHint')}
					</p>
				</div>

				<!-- Website -->
				<div>
					<label for="website" class="label">
						{$t('admin.merchants.website')}
					</label>
					<input
						id="website"
						type="url"
						bind:value={website}
						placeholder="https://example.com"
						class="input"
					/>
					<p class="mt-1 text-sm text-text-subtle">
						{$t('admin.merchants.websiteHint')}
					</p>
				</div>

				<!-- Actions: same divided row as the other create/edit forms
				     (paired layout — compact primary + ghost cancel). On Android
				     the m3-filled-form CSS re-lays this last row out as an M3
				     text button + filled pill, right-aligned. -->
				<div
					class={IS_ANDROID
						? 'flex gap-2 pt-4'
						: 'mt-6 flex items-center gap-2.5 border-t border-border-soft pt-5'}
				>
					<button
						type="submit"
						disabled={isSubmitting || isOffline}
						class="btn btn-primary px-6 {isSubmitting || isOffline
							? 'opacity-50 cursor-not-allowed'
							: ''}"
					>
						{isSubmitting ? $t('common.saving') : $t('common.create')}
					</button>
					<a href={resolve('/merchants')} class="btn btn-ghost">
						{$t('common.cancel')}
					</a>
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
		<div
			class={IS_ANDROID
				? 'rounded-m3-lg bg-m3-card border border-border p-5'
				: 'bg-white rounded-lg shadow-lg p-6'}
		>
			<h2 class="text-xl font-bold text-text mb-4">
				{$t('admin.merchants.title')}
			</h2>
			<p class="text-sm text-text-muted mb-4">
				{$t('admin.merchants.createSubtitle')}
			</p>

			<div class="space-y-4 text-sm">
				<div class="border-l-4 border-accent pl-4">
					<h3 class="font-semibold text-text mb-1">
						{$t('admin.merchants.name')}
					</h3>
					<p class="text-text-muted">
						{$t('admin.merchants.namePlaceholder')}
					</p>
				</div>

				<div class="border-l-4 border-purple-500 pl-4">
					<h3 class="font-semibold text-text mb-1">
						{$t('admin.merchants.color')}
					</h3>
					<p class="text-text-muted">{$t('admin.merchants.colorHint')}</p>
				</div>

				<div class="border-l-4 border-success-500 pl-4">
					<h3 class="font-semibold text-text mb-1">
						{$t('admin.merchants.logoUrl')}
					</h3>
					<p class="text-text-muted">{$t('admin.merchants.logoUrlHint')}</p>
				</div>

				<div class="border-l-4 border-warning-500 pl-4">
					<h3 class="font-semibold text-text mb-1">
						{$t('admin.merchants.website')}
					</h3>
					<p class="text-text-muted">{$t('admin.merchants.websiteHint')}</p>
				</div>
			</div>
		</div>
	</div>
</div>
