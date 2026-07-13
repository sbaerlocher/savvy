<script lang="ts">
	import type { MerchantDTO } from '$lib/types/api';
	import { t } from '$lib/stores/i18n';

	interface Props {
		merchants: MerchantDTO[];
		value?: string;
		required?: boolean;
		id?: string;
		onchange?: () => void;
	}

	let {
		merchants,
		value = $bindable(''),
		required = false,
		id = 'merchant',
		onchange
	}: Props = $props();

	let searchQuery = $state('');
	let open = $state(false);
	let highlightedIndex = $state(-1);
	let containerEl: HTMLDivElement;
	let inputEl: HTMLInputElement;

	$effect(() => {
		if (required && inputEl) {
			inputEl.setCustomValidity(value ? '' : 'Please select a merchant');
		}
	});

	const selectedMerchant = $derived(merchants.find((m) => m.id === value));

	const displayValue = $derived(
		open ? searchQuery : (selectedMerchant?.name ?? '')
	);

	const filtered = $derived(
		searchQuery.trim() === ''
			? merchants
			: merchants.filter((m) =>
					m.name.toLowerCase().includes(searchQuery.toLowerCase())
				)
	);

	function onFocus() {
		searchQuery = '';
		highlightedIndex = -1;
		open = true;
	}

	function onInput(e: Event) {
		searchQuery = (e.target as HTMLInputElement).value;
		highlightedIndex = -1;
		open = true;
	}

	function selectMerchant(merchant: MerchantDTO) {
		value = merchant.id;
		searchQuery = '';
		open = false;
		onchange?.();
	}

	function clearSelection(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		value = '';
		searchQuery = '';
		open = false;
		onchange?.();
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
		} else if (e.key === 'ArrowDown') {
			e.preventDefault();
			open = true;
			highlightedIndex = Math.min(highlightedIndex + 1, filtered.length - 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			open = true;
			highlightedIndex = Math.max(highlightedIndex - 1, 0);
		} else if (
			e.key === 'Enter' &&
			highlightedIndex >= 0 &&
			filtered[highlightedIndex]
		) {
			e.preventDefault();
			selectMerchant(filtered[highlightedIndex]);
		}
	}

	function onBlur(_e: FocusEvent) {
		// Delay so mousedown on list items fires first
		setTimeout(() => {
			if (containerEl && !containerEl.contains(document.activeElement)) {
				open = false;
				searchQuery = '';
			}
		}, 150);
	}
</script>

<div bind:this={containerEl} class="relative">
	<!-- Search / display input -->
	<div class="relative">
		<input
			{id}
			type="text"
			role="combobox"
			aria-expanded={open}
			aria-controls="{id}-listbox"
			aria-autocomplete="list"
			aria-required={required}
			class="input pr-10"
			style="font-size: 16px;"
			placeholder={$t('merchants.searchPlaceholder')}
			value={displayValue}
			onfocus={onFocus}
			oninput={onInput}
			onkeydown={onKeydown}
			onblur={onBlur}
			autocomplete="off"
			bind:this={inputEl}
		/>

		<!-- Clear button or chevron -->
		<div class="absolute inset-y-0 right-0 flex items-center pr-2">
			{#if value}
				<button
					type="button"
					class="rounded p-0.5 text-text-faint hover:text-text-muted focus:outline-none"
					onmousedown={clearSelection}
					tabindex="-1"
					aria-label="Clear"
				>
					<svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
						<path
							d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z"
						/>
					</svg>
				</button>
			{:else}
				<svg
					class="h-4 w-4 text-text-faint pointer-events-none"
					viewBox="0 0 20 20"
					fill="currentColor"
				>
					<path
						fill-rule="evenodd"
						d="M10 3a.75.75 0 01.55.24l3.25 3.5a.75.75 0 11-1.1 1.02L10 4.852 7.3 7.76a.75.75 0 01-1.1-1.02l3.25-3.5A.75.75 0 0110 3zm-3.76 9.2a.75.75 0 011.06.04l2.7 2.908 2.7-2.908a.75.75 0 111.1 1.02l-3.25 3.5a.75.75 0 01-1.1 0l-3.25-3.5a.75.75 0 01.04-1.06z"
						clip-rule="evenodd"
					/>
				</svg>
			{/if}
		</div>
	</div>

	<!-- Dropdown -->
	{#if open}
		<ul
			id="{id}-listbox"
			role="listbox"
			class="absolute z-50 mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
		>
			{#if filtered.length === 0}
				<li class="relative cursor-default select-none px-4 py-2 text-text-subtle">
					{$t('merchants.noResults')}
				</li>
			{:else}
				{#each filtered as merchant, i (merchant.id)}
					<li
						id="{id}-option-{i}"
						role="option"
						aria-selected={merchant.id === value}
						class="relative cursor-default select-none py-2 pl-3 pr-9 {i ===
						highlightedIndex
							? 'bg-accent text-white'
							: 'text-text'}"
						onmousedown={() => selectMerchant(merchant)}
						onmouseenter={() => (highlightedIndex = i)}
					>
						<span
							class="block truncate {merchant.id === value
								? 'font-semibold'
								: 'font-normal'}"
						>
							{merchant.name}
						</span>

						{#if merchant.id === value}
							<span
								class="absolute inset-y-0 right-0 flex items-center pr-4 {i ===
								highlightedIndex
									? 'text-white'
									: 'text-accent'}"
							>
								<svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
									<path
										fill-rule="evenodd"
										d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z"
										clip-rule="evenodd"
									/>
								</svg>
							</span>
						{/if}
					</li>
				{/each}
			{/if}
		</ul>
	{/if}
</div>
