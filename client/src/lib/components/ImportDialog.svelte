<script lang="ts">
	import { importApi } from '$lib/api';
	import Modal from '$lib/components/ui/Modal.svelte';
	import { t } from '$lib/stores/i18n';
	import { toastStore } from '$lib/stores/toast';
	import type { ImportPreview, ImportResult } from '$lib/types/api';
	import { get } from 'svelte/store';
	import { platform } from '$lib/utils/platform';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	function translateImportError(message: string): string {
		return message
			.replace(
				/already exists \(duplicate\)/g,
				tr('settings.import.errorDuplicate')
			)
			.replace(/database error/g, tr('settings.import.errorDatabase'));
	}

	// Desktop renders the ImportDesktop mockup: --surface panel on --radius-modal
	// with --shadow-modal, mockup type steps mapped onto the nearest global token
	// (13.5px -> --text-body / --text-label, 26px -> --text-title, 20px ->
	// --text-heading) and taller token-radius buttons. The iOS and Android arms
	// keep their current sizing, so every delta below is platform-gated.
	const isDesktop = platform === 'other';
	const descClass = isDesktop ? 'text-body' : 'text-sm';
	const copyClass = isDesktop ? 'text-body' : 'text-sm';
	const helperClass = isDesktop ? 'text-body-sm' : 'text-xs';
	const promptClass = isDesktop
		? 'text-body font-semibold'
		: 'text-sm font-medium';
	const copyGapClass = isDesktop ? 'mb-0.75' : 'mb-1';
	const helperGapClass = isDesktop ? 'mb-3.5' : 'mb-3';
	const hintGapClass = isDesktop ? 'gap-3.5' : 'gap-4';
	const radioRowPadClass = isDesktop ? 'p-3.25' : 'p-3';
	const radioClass = isDesktop ? 'size-4.5 accent-accent-600' : '';
	const tileLabelClass = isDesktop ? 'text-xs mt-0.75' : 'text-xs';
	const errorListClass = isDesktop ? 'space-y-1.25' : 'space-y-1';
	const errorItemClass = isDesktop ? 'text-xs leading-snug' : 'text-xs';
	const importingPadClass = isDesktop ? 'py-5' : 'py-8';
	const bodyPadClass = isDesktop ? 'py-5' : 'py-4';
	const tilePadClass = isDesktop ? 'p-3.5' : 'p-3';
	const previewValueClass = isDesktop ? 'text-title' : 'text-2xl font-bold';
	const resultValueClass = isDesktop ? 'text-heading' : 'text-lg font-bold';
	const footerGapClass = isDesktop ? 'gap-2.5' : 'gap-3';
	const ctaClass = isDesktop
		? 'text-label h-10 rounded-lg shadow-accent'
		: 'text-sm';
	const footerGhostClass = isDesktop ? 'text-label h-11 rounded-lg' : '';
	const footerPrimaryClass = isDesktop
		? 'text-label h-11 rounded-lg shadow-accent'
		: '';

	type Step = 'select' | 'csv-type' | 'preview' | 'importing' | 'result';
	type CSVType = 'cards' | 'vouchers' | 'gift-cards';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		onImported: () => void;
		defaultResourceType?: CSVType;
	}

	let { isOpen, onClose, onImported, defaultResourceType }: Props = $props();

	let step = $state<Step>('select');
	let file = $state<File | null>(null);
	let fileType = $state<'json' | 'csv' | null>(null);
	let csvType = $state<CSVType>('cards');
	let preview = $state<ImportPreview | null>(null);
	let result = $state<ImportResult | null>(null);
	let jsonData = $state<unknown>(null);
	let isDragging = $state(false);

	let fileInput: HTMLInputElement | undefined = $state();

	function reset() {
		step = 'select';
		file = null;
		fileType = null;
		csvType = defaultResourceType ?? 'cards';
		preview = null;
		result = null;
		jsonData = null;
		isDragging = false;
	}

	function handleClose() {
		if (result) {
			onImported();
		}
		reset();
		onClose();
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		isDragging = true;
	}

	function handleDragLeave() {
		isDragging = false;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
		const droppedFile = e.dataTransfer?.files[0];
		if (droppedFile) {
			processFile(droppedFile);
		}
	}

	function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		const selectedFile = input.files?.[0];
		if (selectedFile) {
			processFile(selectedFile);
		}
	}

	async function processFile(selectedFile: File) {
		file = selectedFile;
		const ext = selectedFile.name.toLowerCase().split('.').pop();

		if (ext === 'json') {
			fileType = 'json';
			try {
				const text = await selectedFile.text();
				jsonData = JSON.parse(text);
				const previewResult = await importApi.previewJSON(jsonData);
				preview = previewResult;

				if (
					previewResult.cards === 0 &&
					previewResult.vouchers === 0 &&
					previewResult.gift_cards === 0
				) {
					toastStore.error(tr('settings.import.noData'));
					reset();
					return;
				}

				step = 'preview';
			} catch {
				toastStore.error(tr('settings.import.error'));
				reset();
			}
		} else if (ext === 'csv') {
			fileType = 'csv';
			if (defaultResourceType) {
				csvType = defaultResourceType;
				await startImport();
				return;
			}
			step = 'csv-type';
		} else {
			toastStore.error(tr('settings.import.error'));
		}
	}

	async function startImport() {
		step = 'importing';

		try {
			if (fileType === 'json' && jsonData) {
				result = await importApi.importJSON(jsonData);
			} else if (fileType === 'csv' && file) {
				switch (csvType) {
					case 'cards':
						result = await importApi.importCardsCSV(file);
						break;
					case 'vouchers':
						result = await importApi.importVouchersCSV(file);
						break;
					case 'gift-cards':
						result = await importApi.importGiftCardsCSV(file);
						break;
				}
			}

			if (result) {
				const totalImported =
					result.cards_imported +
					result.vouchers_imported +
					result.gift_cards_imported;
				if (result.skipped > 0) {
					toastStore.warning(tr('settings.import.partialSuccess'));
				} else if (totalImported > 0) {
					toastStore.success(tr('settings.import.success'));
				}
			}

			step = 'result';
		} catch {
			toastStore.error(tr('settings.import.error'));
			step = 'select';
		}
	}

	$effect(() => {
		if (isOpen) {
			reset();
		}
	});
</script>

<Modal
	open={isOpen}
	onclose={handleClose}
	layer="elevated"
	mobileLayout="center"
	labelledby="import-dialog-title"
>
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="pointer-events-auto max-w-lg w-full max-h-[90vh] overflow-y-auto {platform ===
		'ios'
			? 'liquid-glass-surface rounded-2xl'
			: platform === 'android'
				? 'bg-m3-surface-container rounded-m3-xl shadow-m3-dialog'
				: 'bg-surface rounded-modal shadow-modal'}"
		onclick={(e) => e.stopPropagation()}
		onkeydown={(e) => e.stopPropagation()}
		role="document"
	>
		<!-- Header -->
		<div class="px-6 pt-6 pb-4 border-b border-border-soft">
			<h3 id="import-dialog-title" class="text-lg font-semibold text-text">
				{tr('settings.import.title')}
			</h3>
			<p class="mt-1 {descClass} text-text-subtle">
				{tr('settings.import.description')}
			</p>
		</div>

		<!-- Body -->
		<div class="px-6 {bodyPadClass}">
			{#if step === 'select'}
				<!-- File Selection -->
				<div
					class="border-2 border-dashed rounded-lg p-8 text-center transition-colors {isDragging
						? 'border-accent bg-accent-50'
						: 'border-border-field hover:border-text-faint'}"
					ondragover={handleDragOver}
					ondragleave={handleDragLeave}
					ondrop={handleDrop}
					role="region"
				>
					<svg
						class="mx-auto w-12 h-12 mb-3 {isDesktop && isDragging
							? 'text-accent'
							: 'text-text-faint'}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
						/>
					</svg>
					<p class="{copyClass} text-text-muted {copyGapClass}">
						{tr('settings.import.dragDrop')}
					</p>
					<p class="{helperClass} text-text-faint {helperGapClass}">
						{tr('settings.import.orClickToSelect')}
					</p>
					<button
						type="button"
						onclick={() => fileInput?.click()}
						class="btn btn-primary {ctaClass}"
					>
						{tr('settings.import.selectFile')}
					</button>
					<input
						bind:this={fileInput}
						type="file"
						accept=".json,.csv"
						onchange={handleFileSelect}
						class="hidden"
					/>
					<div
						class="mt-4 flex justify-center {hintGapClass} {helperClass} text-text-faint"
					>
						<span>{tr('settings.import.jsonFormat')}</span>
						<span>|</span>
						<span>{tr('settings.import.csvFormat')}</span>
					</div>
				</div>
			{:else if step === 'csv-type'}
				<!-- CSV Resource Type Selection -->
				<div class="space-y-4">
					<p class="{promptClass} text-text-ink2">
						{tr('settings.import.csvResourceType')}
					</p>
					<div class="space-y-2">
						{#each [{ value: 'cards' as CSVType, label: tr('settings.import.csvCards') }, { value: 'vouchers' as CSVType, label: tr('settings.import.csvVouchers') }, { value: 'gift-cards' as CSVType, label: tr('settings.import.csvGiftCards') }] as option (option.value)}
							<label
								class="flex items-center gap-3 {radioRowPadClass} rounded-lg border cursor-pointer transition-colors {csvType ===
								option.value
									? 'border-accent bg-accent-50'
									: 'border-border hover:border-border-field'}"
							>
								<input
									type="radio"
									name="csvType"
									value={option.value}
									checked={csvType === option.value}
									onchange={() => (csvType = option.value)}
									class="text-accent focus:ring-accent {radioClass}"
								/>
								<span class="{copyClass} text-text-ink2">{option.label}</span>
							</label>
						{/each}
					</div>
					{#if file}
						<p class="text-xs text-text-faint truncate">
							{file.name}
						</p>
					{/if}
				</div>
			{:else if step === 'preview'}
				<!-- JSON Preview -->
				<div class="space-y-4">
					<h4 class="{promptClass} text-text-ink2">
						{tr('settings.import.preview')}
					</h4>
					{#if preview}
						<div class="grid grid-cols-3 gap-3">
							{#if preview.cards > 0}
								<div class="bg-accent-50 rounded-lg {tilePadClass} text-center">
									<div class="{previewValueClass} text-accent">
										{preview.cards}
									</div>
									<div class="{tileLabelClass} text-accent">
										{tr('settings.import.previewCards')}
									</div>
								</div>
							{/if}
							{#if preview.vouchers > 0}
								<div
									class="bg-success-50 rounded-lg {tilePadClass} text-center"
								>
									<div class="{previewValueClass} text-success-600">
										{preview.vouchers}
									</div>
									<div class="{tileLabelClass} text-success-500">
										{tr('settings.import.previewVouchers')}
									</div>
								</div>
							{/if}
							{#if preview.gift_cards > 0}
								<div class="bg-purple-50 rounded-lg {tilePadClass} text-center">
									<div class="{previewValueClass} text-purple-600">
										{preview.gift_cards}
									</div>
									<div class="{tileLabelClass} text-purple-500">
										{tr('settings.import.previewGiftCards')}
									</div>
								</div>
							{/if}
						</div>
					{/if}
					{#if file}
						<p class="text-xs text-text-faint truncate">
							{file.name}
						</p>
					{/if}
				</div>
			{:else if step === 'importing'}
				<!-- Importing Progress -->
				<div class="flex flex-col items-center {importingPadClass}">
					<span class="relative inline-flex h-6 w-6 mb-4"
						><span
							class="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-400 opacity-75"
						></span><span
							class="relative inline-flex rounded-full h-6 w-6 bg-accent"
						></span></span
					>
					<p class="{copyClass} text-text-muted">
						{tr('settings.import.importing')}
					</p>
				</div>
			{:else if step === 'result'}
				<!-- Import Result -->
				{#if result}
					<div class="space-y-4">
						<div class="grid grid-cols-2 gap-3">
							{#if result.cards_imported > 0}
								<div class="bg-accent-50 rounded-lg {tilePadClass}">
									<div class="{resultValueClass} text-accent">
										{result.cards_imported}
									</div>
									<div class="{tileLabelClass} text-accent">
										{tr('settings.import.previewCards')}
										{tr('settings.import.imported').toLowerCase()}
									</div>
								</div>
							{/if}
							{#if result.vouchers_imported > 0}
								<div class="bg-success-50 rounded-lg {tilePadClass}">
									<div class="{resultValueClass} text-success-600">
										{result.vouchers_imported}
									</div>
									<div class="{tileLabelClass} text-success-500">
										{tr('settings.import.previewVouchers')}
										{tr('settings.import.imported').toLowerCase()}
									</div>
								</div>
							{/if}
							{#if result.gift_cards_imported > 0}
								<div class="bg-purple-50 rounded-lg {tilePadClass}">
									<div class="{resultValueClass} text-purple-600">
										{result.gift_cards_imported}
									</div>
									<div class="{tileLabelClass} text-purple-500">
										{tr('settings.import.previewGiftCards')}
										{tr('settings.import.imported').toLowerCase()}
									</div>
								</div>
							{/if}
							{#if result.skipped > 0}
								<div class="bg-warning-50 rounded-lg {tilePadClass}">
									<div class="{resultValueClass} text-warning-600">
										{result.skipped}
									</div>
									<div class="{tileLabelClass} text-warning-500">
										{tr('settings.import.skipped')}
									</div>
								</div>
							{/if}
						</div>

						{#if result.errors && result.errors.length > 0}
							<div class="border border-danger-200 rounded-lg {tilePadClass}">
								<h4 class="{promptClass} text-danger-700 mb-2">
									{tr('settings.import.errors')}
								</h4>
								<ul class="{errorListClass} max-h-32 overflow-y-auto">
									{#each result.errors as error (`${error.row ?? ''}-${error.field ?? ''}-${error.message}`)}
										<li class="{errorItemClass} text-danger-600">
											{#if error.row}{tr('settings.import.row', {
													row: error.row
												})}:
											{/if}{translateImportError(error.message)}
										</li>
									{/each}
								</ul>
							</div>
						{/if}
					</div>
				{/if}
			{/if}
		</div>

		<!-- Footer -->
		<div
			class="px-6 pb-6 flex justify-end {footerGapClass} border-t border-border-soft pt-4"
		>
			{#if step === 'select' || step === 'result'}
				<button
					type="button"
					onclick={handleClose}
					class="btn btn-ghost {footerGhostClass}"
				>
					{tr('settings.import.close')}
				</button>
			{:else if step === 'csv-type'}
				<button
					type="button"
					onclick={reset}
					class="btn btn-ghost {footerGhostClass}"
				>
					{tr('common.back')}
				</button>
				<button
					type="button"
					onclick={startImport}
					class="btn btn-primary {footerPrimaryClass}"
				>
					{tr('settings.import.button')}
				</button>
			{:else if step === 'preview'}
				<button
					type="button"
					onclick={reset}
					class="btn btn-ghost {footerGhostClass}"
				>
					{tr('common.back')}
				</button>
				<button
					type="button"
					onclick={startImport}
					class="btn btn-primary {footerPrimaryClass}"
				>
					{tr('settings.import.button')}
				</button>
			{/if}
		</div>
	</div>
</Modal>
