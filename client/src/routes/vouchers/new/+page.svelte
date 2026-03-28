<script lang="ts">
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';
	import { t } from '$lib/stores/i18n';
	import { vouchersApi } from '$lib/api';
	import { toastStore } from '$lib/stores/toast';
	import { logger } from '$lib/utils/logger';

	import type { DuplicateWarning } from '$lib/types/api';
	import VoucherForm from '$lib/components/vouchers/VoucherForm.svelte';
	import SharedInfoBox from '$lib/components/SharedInfoBox.svelte';
	import EmailAutocomplete from '$lib/components/EmailAutocomplete.svelte';
	import DuplicateWarningBanner from '$lib/components/DuplicateWarningBanner.svelte';

	// Svelte 5 compatible translation wrapper
	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);
	const pageLogger = logger.child('VoucherNewPage');

	let code = $state('');
	let merchantId = $state('');
	let type = $state('percentage');
	let value = $state(0);
	let currency = $state('CHF');
	let barcodeType = $state('CODE128');
	let usageLimitType = $state('single_use');
	let validFrom = $state('');
	let validUntil = $state('');
	let description = $state('');
	let isLoading = $state(false);

	// Validation errors
	let errors = $state<{
		merchant?: string;
		value?: string;
		validUntil?: string;
	}>({});

	// Sharing state
	let shareEmail = $state('');
	let duplicateWarning = $state<DuplicateWarning | null>(null);

	async function handleSubmit() {
		isLoading = true;
		duplicateWarning = null;

		// Reset errors
		errors = {};

		// Validate required fields
		let hasErrors = false;

		if (!validUntil) {
			errors = { ...errors, validUntil: tr('vouchers.validUntilRequired') };
			hasErrors = true;
		}

		if (!merchantId) {
			errors = { ...errors, merchant: tr('vouchers.errors.merchantRequired') };
			hasErrors = true;
		}

		if (value <= 0) {
			errors = { ...errors, value: tr('vouchers.errors.valueRequired') };
			hasErrors = true;
		}

		if (hasErrors) {
			isLoading = false;
			return;
		}

		try {
			// Use today for validFrom if not provided
			const today = new Date().toISOString().split('T')[0];

			const response = await vouchersApi.create({
				code,
				merchant_id: merchantId || undefined,
				type,
				value,
				currency: currency || undefined,
				barcode_type: barcodeType || undefined,
				usage_limit_type: usageLimitType || undefined,
				valid_from: validFrom ? `${validFrom}T00:00:00Z` : `${today}T00:00:00Z`,
				valid_until: `${validUntil}T23:59:59Z`,
				description: description || undefined,
				share_with_email: shareEmail || undefined
			});
			toastStore.success(tr('vouchers.createSuccess'));
			// Force full page reload to ensure fresh data in lists
			window.location.href = `/vouchers/${response.voucher.id}`;
		} catch (err: any) {
			if (err.error === 'duplicate_barcode' && err.data?.duplicate) {
				duplicateWarning = err.data.duplicate as DuplicateWarning;
			} else {
				toastStore.error(err.message || tr('vouchers.createError'));
			}
		} finally {
			isLoading = false;
		}
	}

	function handleCancel() {
		goto('/vouchers');
	}
</script>

<svelte:head>
	<title>{tr('vouchers.newVoucher')} - {tr('common.appName')}</title>
</svelte:head>

<div class="mb-6">
	<a href="/vouchers" class="text-cyan-600 hover:text-cyan-700"
		>{tr('common.backToOverview')}</a
	>
</div>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Left column: Form (2/3 width) -->
	<div class="lg:col-span-2">
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h1 class="text-3xl font-bold text-gray-900 mb-6">
				{tr('vouchers.newVoucher')}
			</h1>
			<DuplicateWarningBanner
				warning={duplicateWarning}
				resourceType="voucher"
				onNavigate={(id) => goto(`/vouchers/${id}`)}
			/>
			<VoucherForm
				bind:code
				bind:merchantId
				bind:type
				bind:value
				bind:currency
				bind:barcodeType
				bind:validFrom
				bind:validUntil
				bind:usageLimitType
				bind:description
				bind:errors
				onSubmit={handleSubmit}
				onCancel={handleCancel}
				{isLoading}
				submitLabel={tr('vouchers.createButton')}
			/>
		</div>
	</div>

	<!-- Right column: Sharing (1/3 width) -->
	<div class="lg:col-span-1 space-y-4">
		<!-- Sharing Box -->
		<div class="bg-white rounded-lg shadow-lg p-6">
			<h2 class="text-xl font-bold text-gray-900 mb-4">{tr('common.share')}</h2>
			<p class="text-sm text-gray-600 mb-4">
				{tr('vouchers.sharing.shareOnCreate')}
			</p>

			<div class="border border-cyan-200 bg-cyan-50 rounded-lg p-4 space-y-4">
				<!-- Email Input with Autocomplete -->
				<EmailAutocomplete
					bind:value={shareEmail}
					label={tr('vouchers.sharing.userEmail')}
					hint={tr('vouchers.sharing.hint')}
					inputId="share_email"
				/>

				<!-- Info Box -->
				<SharedInfoBox
					title={tr('vouchers.sharing.whatIsShared')}
					items={[
						tr('vouchers.sharing.sharedCode'),
						tr('vouchers.sharing.sharedDetails'),
						tr('vouchers.sharing.sharedDescription')
					]}
					note={tr('vouchers.sharing.readOnlyNote')}
				/>
			</div>
		</div>
	</div>
</div>
