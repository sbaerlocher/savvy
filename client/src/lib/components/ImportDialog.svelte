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
	// --text-heading) and taller token-radius buttons. Android renders the
	// ImportAndroid mockup: a tonal M3 high-container panel on the M3 shape
	// scale, pill actions and a text button. The iOS arm keeps its current
	// sizing, so every delta below is platform-gated. Both flags are module
	// constants because the platform is fixed for the session.
	const isDesktop = platform === 'other';
	const isAndroid = platform === 'android';
	// iOS renders the ImportIOS mockup as a Liquid Glass sheet: wider corner
	// radius, modal shadow, tighter chrome padding and 44pt controls. Module-level
	// const, not $derived — `platform` is resolved once at module load.
	const isIOS = platform === 'ios';
	const descClass = isDesktop
		? 'text-body'
		: isAndroid
			? 'text-label font-normal'
			: 'text-sm';
	const copyClass = isDesktop ? 'text-body' : 'text-sm';
	const helperClass = isDesktop ? 'text-body-sm' : 'text-xs';
	const promptClass = isDesktop
		? 'text-body font-semibold'
		: 'text-sm font-medium';
	const titleClass = isAndroid
		? 'text-heading font-semibold'
		: 'text-lg font-semibold';
	const descGapClass = isAndroid ? 'mt-1.25' : 'mt-1';
	const copyGapClass = isDesktop ? 'mb-0.75' : 'mb-1';
	const helperGapClass = isDesktop ? 'mb-3.5' : 'mb-3';
	const hintGapClass = isDesktop ? 'gap-3.5' : isAndroid ? 'gap-2.5' : 'gap-4';
	const dropZoneClass = isAndroid
		? 'rounded-m3-md px-4.5 py-6.5'
		: 'rounded-lg p-8';
	const radioRowPadClass = isDesktop ? 'p-3.25' : 'p-3';
	const radioRowRadiusClass = isAndroid ? 'rounded-m3-md' : 'rounded-lg';
	const radioClass = isDesktop ? 'size-4.5' : '';
	const androidRadioClass = (selected: boolean) =>
		`rounded-m3-full h-4.5 w-4.5 flex-none appearance-none border-2 bg-clip-content p-0.75 ${
			selected ? 'border-accent-600 bg-accent-600' : 'border-border-field'
		}`;
	const tileLabelClass = isDesktop ? 'text-xs mt-0.75' : 'text-xs';
	const errorListClass = isDesktop ? 'space-y-1.25' : 'space-y-1';
	const errorItemClass = isDesktop ? 'text-xs leading-snug' : 'text-xs';
	const importingPadClass = isDesktop ? 'py-5' : isAndroid ? 'py-8.5' : 'py-8';
	const bodyPadClass = isDesktop
		? 'py-5'
		: isAndroid || isIOS
			? 'py-4.5'
			: 'py-4';
	const tilePadClass = isDesktop ? 'p-3.5' : 'p-3';
	// The preview tiles sit three to a row, so Android trims the horizontal
	// padding the mockup gives them to keep the widest label on one line.
	const previewTilePadClass = isDesktop
		? 'p-3.5'
		: isAndroid
			? 'px-1 py-3'
			: 'p-3';
	const tileRadiusClass = isAndroid ? 'rounded-m3-md' : 'rounded-lg';
	const tileGapClass = isAndroid || isIOS ? 'gap-2.5' : 'gap-3';
	const previewValueClass = isDesktop
		? 'text-title'
		: `text-2xl font-bold${isIOS ? ' leading-tight' : ''}`;
	const resultValueClass = isDesktop
		? 'text-heading'
		: `text-lg font-bold${isIOS ? ' leading-tight' : ''}`;
	const footerGapClass = isDesktop
		? 'gap-2.5'
		: isAndroid
			? 'gap-1.5'
			: isIOS
				? 'gap-2.5'
				: 'gap-3';
	const footerPadClass =
		isAndroid || isIOS ? 'px-5 pt-3.5 pb-5' : 'px-6 pb-6 pt-4';
	// M3 dialog actions: a text button and a filled pill, matching the
	// convention already used by the gift-card ledger sheet on Android.
	const androidGhostClass =
		'text-label text-accent-700 rounded-m3-full inline-flex h-10 items-center px-3.5';
	const androidPrimaryClass =
		'bg-accent-600 text-on-accent text-label rounded-m3-full inline-flex h-10 items-center px-5';
	const ctaClass = isDesktop
		? 'text-label h-10 rounded-lg shadow-accent'
		: 'text-sm';
	const footerGhostClass = isDesktop ? 'text-label h-11 rounded-lg' : '';
	const footerPrimaryClass = isDesktop
		? 'text-label h-11 rounded-lg shadow-accent'
		: '';
	const headerPadClass = isIOS ? 'px-5 pt-5.5 pb-3.5' : 'px-6 pt-6 pb-4';
	const bodyPadXClass = isIOS ? 'px-5' : 'px-6';
	// 44pt hit target with --radius-lg corners, per the mockup. Kept local to the
	// import sheet: the global .btn is unchanged, so every other screen keeps its
	// current button shape. The focus ring and disabled states come from
	// .btn/.btn-primary/.btn-ghost, which the iOS arm replaces wholesale — so they
	// are restated here.
	const iosBtnBase =
		'inline-flex h-11 items-center justify-center rounded-[var(--radius-lg)] px-4.5 text-sm font-semibold transition-colors focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';
	const iosPrimaryClass = `${iosBtnBase} bg-accent text-on-accent shadow-[var(--shadow-accent)] hover:bg-accent-hover focus:ring-accent`;
	// Ghost stays transparent on glass, hovering to a translucent tint — an opaque
	// fill would cancel the backdrop blur underneath.
	const iosGhostClass = `${iosBtnBase} border border-border-field bg-transparent text-text-muted hover:bg-[var(--color-glass-hollow)] focus:ring-text-faint`;
	const ghostButtonClass = isIOS
		? iosGhostClass
		: isAndroid
			? androidGhostClass
			: `btn btn-ghost ${footerGhostClass}`;
	const primaryButtonClass = isIOS
		? iosPrimaryClass
		: isAndroid
			? androidPrimaryClass
			: `btn btn-primary ${footerPrimaryClass}`;
	// The select-step CTA sits inside the drop zone, not the footer, so off iOS it
	// keeps plain .btn-primary sizing from ctaClass.
	const ctaButtonClass = isIOS ? iosPrimaryClass : 'btn btn-primary';

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
	clearBottomNav={!isAndroid}
	labelledby="import-dialog-title"
>
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="pointer-events-auto w-full max-h-[90vh] overflow-y-auto {platform ===
		'ios'
			? 'liquid-glass-surface rounded-[var(--radius-modal)] shadow-[var(--shadow-modal)] max-w-lg'
			: isAndroid
				? 'bg-m3-surface-container-high rounded-m3-xl shadow-m3-dialog max-w-sm'
				: 'bg-surface rounded-modal shadow-modal max-w-lg'}"
		onclick={(e) => e.stopPropagation()}
		onkeydown={(e) => e.stopPropagation()}
		role="document"
	>
		<!-- Header -->
		<div class="{headerPadClass} border-b border-border-soft">
			<h3 id="import-dialog-title" class="{titleClass} text-text">
				{tr('settings.import.title')}
			</h3>
			<p class="{descGapClass} {descClass} text-text-subtle">
				{tr('settings.import.description')}
			</p>
		</div>

		<!-- Body -->
		<div class="{bodyPadXClass} {bodyPadClass}">
			{#if step === 'select'}
				<!-- File Selection -->
				<div
					class="border-2 border-dashed {dropZoneClass} text-center transition-colors {isDragging
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
						class={isAndroid
							? androidPrimaryClass
							: isIOS
								? ctaButtonClass
								: `btn btn-primary ${ctaClass}`}
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
								class="flex items-center gap-3 {radioRowPadClass} {radioRowRadiusClass} border cursor-pointer transition-colors {csvType ===
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
									class={isAndroid
										? androidRadioClass(csvType === option.value)
										: `text-accent focus:ring-accent accent-accent-600 ${radioClass}`}
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
						<div class="grid grid-cols-3 {tileGapClass}">
							{#if preview.cards > 0}
								<div
									class="bg-accent-50 {tileRadiusClass} {previewTilePadClass} text-center"
								>
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
									class="bg-success-50 {tileRadiusClass} {previewTilePadClass} text-center"
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
								<div
									class="bg-purple-50 {tileRadiusClass} {previewTilePadClass} text-center"
								>
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
						<div class="grid grid-cols-2 {tileGapClass}">
							{#if result.cards_imported > 0}
								<div class="bg-accent-50 {tileRadiusClass} {tilePadClass}">
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
								<div class="bg-success-50 {tileRadiusClass} {tilePadClass}">
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
								<div class="bg-purple-50 {tileRadiusClass} {tilePadClass}">
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
								<div class="bg-warning-50 {tileRadiusClass} {tilePadClass}">
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
							<div
								class="border border-danger-200 {tileRadiusClass} {tilePadClass}"
							>
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
			class="{footerPadClass} flex justify-end {footerGapClass} border-t border-border-soft"
		>
			{#if step === 'select' || step === 'result'}
				<button type="button" onclick={handleClose} class={ghostButtonClass}>
					{tr('settings.import.close')}
				</button>
			{:else if step === 'csv-type'}
				<button type="button" onclick={reset} class={ghostButtonClass}>
					{tr('common.back')}
				</button>
				<button type="button" onclick={startImport} class={primaryButtonClass}>
					{tr('settings.import.button')}
				</button>
			{:else if step === 'preview'}
				<button type="button" onclick={reset} class={ghostButtonClass}>
					{tr('common.back')}
				</button>
				<button type="button" onclick={startImport} class={primaryButtonClass}>
					{tr('settings.import.button')}
				</button>
			{/if}
		</div>
	</div>
</Modal>
