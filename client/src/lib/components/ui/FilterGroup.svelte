<script lang="ts">
	import { platform } from '$lib/utils/platform';
	import SectionLabel from '$lib/components/ui/SectionLabel.svelte';

	interface Option {
		value: string;
		label: string;
	}

	// Single-select filter group.
	//   iOS      → inline accordion row: label + current value + chevron; tapping
	//              expands the options in place, picking one collapses it again.
	//   Android  → M3 filter chips.
	//   desktop  → neutral chips.
	let {
		label,
		value = $bindable(),
		options,
		idPrefix = 'filter',
		groupKey = idPrefix,
		openGroup = $bindable(''),
		flat = false,
		variant = 'list',
		plainLabel = false
	}: {
		label: string;
		value: string;
		options: Option[];
		idPrefix?: string;
		/** iOS only: this group's identity within an accordion set. Defaults to
		 *  `idPrefix`, so a standalone group still toggles on its own. */
		groupKey?: string;
		/** iOS only: the key of the single open group, shared across siblings so
		 *  expanding one collapses the others. A call site that does not bind it
		 *  gets a self-contained row that opens and closes on its own. */
		openGroup?: string;
		/** iOS only: skip the accordion — render the caption above the options,
		 *  permanently expanded (mockup screen-MerchantsIOS filter sheet). */
		flat?: boolean;
		/** iOS flat groups only: 'list' keeps the checkmark rows, 'chips' lays the
		 *  options out as a wrapping pill row. */
		variant?: 'list' | 'chips';
		/** Android/desktop: set the caption in mixed case at the label step
		 *  instead of the uppercase eyebrow. The admin filter sheet does (mockup
		 *  screen-AdminAndroid); the wallet and merchant sheets keep the kicker. */
		plainLabel?: boolean;
	} = $props();

	const groupId = $derived(`${idPrefix}-group`);
	const currentLabel = $derived(
		options.find((o) => o.value === value)?.label ?? ''
	);

	// Android and Desktop render the same chip markup; only the chip class
	// differs (M3 tonal/outlined vs. neutral). Keep one branch, vary the class.
	const chipShape = platform === 'android' ? 'rounded-m3-sm' : 'rounded-lg';
	// Android selected chip is the accent tint, not the secondary container —
	// the sheet itself is already tonal, so a tonal chip would not read as
	// selected against it (wallet mockup).
	const chipSelected = 'bg-accent-100 text-accent-850';
	const chipUnselected =
		platform === 'android'
			? 'border border-border-chip bg-transparent text-text-muted hover:bg-surface-1'
			: 'border border-border bg-white text-text-muted hover:bg-surface-1';
	// Android: 8px/14px inset with a 13px semibold label (M3 small chip).
	const chipSize =
		platform === 'android'
			? 'px-3.5 py-2 text-label'
			: 'px-3 py-1.5 text-sm font-medium';

	// Accordion state lives in the parent (openGroup), so opening one group
	// collapses its siblings. Only one row is ever expanded.
	const open = $derived(!!groupKey && openGroup === groupKey);

	function toggle() {
		openGroup = open ? '' : groupKey;
	}

	function pick(v: string) {
		value = v;
		openGroup = '';
	}
</script>

{#if platform === 'ios' && flat}
	<!-- iOS flat group: an uppercase caption above permanently expanded options
	     (mockup screen-MerchantsIOS filter sheet). -->
	<div>
		<SectionLabel id={groupId}>{label}</SectionLabel>
		{#if variant === 'chips'}
			<div
				role="radiogroup"
				aria-labelledby={groupId}
				class="flex flex-wrap gap-2"
			>
				{#each options as opt (opt.value)}
					{@const selected = value === opt.value}
					<button
						type="button"
						role="radio"
						aria-checked={selected}
						onclick={() => (value = opt.value)}
						class="inline-flex items-center gap-1.25 rounded-full px-3.25 py-1.75 text-label transition-colors {selected
							? 'bg-accent text-on-accent'
							: 'border border-border bg-surface text-text-muted'}"
					>
						{#if selected}
							<svg
								class="h-3.25 w-3.25"
								fill="none"
								stroke="currentColor"
								stroke-width="2.6"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M20 6L9 17l-5-5"
								/>
							</svg>
						{/if}
						{opt.label}
					</button>
				{/each}
			</div>
		{:else}
			<div role="radiogroup" aria-labelledby={groupId} class="flex flex-col">
				{#each options as opt (opt.value)}
					{@const selected = value === opt.value}
					<button
						type="button"
						role="radio"
						aria-checked={selected}
						onclick={() => (value = opt.value)}
						class="flex w-full items-center gap-3 py-2.25 text-left text-[length:var(--text-code)] font-normal {selected
							? 'text-text'
							: 'text-text-muted'}"
					>
						<span class="flex h-4 w-4 shrink-0 items-center justify-center">
							{#if selected}
								<svg
									class="h-4 w-4 text-accent"
									fill="none"
									stroke="currentColor"
									stroke-width="2.4"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M20 6L9 17l-5-5"
									/>
								</svg>
							{/if}
						</span>
						{opt.label}
					</button>
				{/each}
			</div>
		{/if}
	</div>
{:else if platform === 'ios'}
	<!-- iOS: an inline accordion row. Collapsed it shows label + current value +
	     chevron; tapping expands the options in place (checkmark on the picked
	     one) and picking collapses it again (mockup screen-WalletIOS). Only one
	     row in the set is open at a time — see openGroup. -->
	<div>
		<button
			type="button"
			onclick={toggle}
			aria-expanded={open}
			aria-controls={open ? groupId : undefined}
			class="flex w-full items-center justify-between py-3 text-left"
		>
			<span
				class="text-xs font-medium uppercase tracking-wider text-text-subtle"
			>
				{label}
			</span>
			<span
				class="flex items-center gap-1 text-[length:var(--text-code)] text-text"
			>
				{#if !open}{currentLabel}{/if}
				<svg
					class="h-4 w-4 text-text-faint transition-transform {open
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
						d="M19 9l-7 7-7-7"
					/>
				</svg>
			</span>
		</button>

		{#if open}
			<div id={groupId} role="radiogroup" aria-label={label} class="pb-2 pl-2">
				{#each options as opt (opt.value)}
					{@const selected = value === opt.value}
					<button
						type="button"
						role="radio"
						aria-checked={selected}
						onclick={() => pick(opt.value)}
						class="flex w-full items-center gap-3 py-2.25 text-left text-[length:var(--text-code)] {selected
							? 'text-text'
							: 'text-text-muted'}"
					>
						<span class="flex h-4 w-4 shrink-0 items-center justify-center">
							{#if selected}
								<svg
									class="h-4 w-4 text-accent"
									fill="none"
									stroke="currentColor"
									stroke-width="2.4"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M20 6L9 17l-5-5"
									/>
								</svg>
							{/if}
						</span>
						{opt.label}
					</button>
				{/each}
			</div>
		{/if}
	</div>
{:else}
	<!-- Android (M3 tonal/outlined chips) and Desktop (neutral chips) share
	     this markup; only chipShape/chipSelected/chipUnselected differ. -->
	<div class={platform === 'android' ? '' : 'py-4'}>
		{#if plainLabel}
			<span id={groupId} class="mb-2.5 block text-label text-text-ink2">
				{label}
			</span>
		{:else}
			<SectionLabel id={groupId}>{label}</SectionLabel>
		{/if}

		<div
			role="radiogroup"
			aria-labelledby={groupId}
			class="flex flex-wrap gap-2"
		>
			{#each options as opt (opt.value)}
				{@const selected = value === opt.value}
				<button
					type="button"
					role="radio"
					aria-checked={selected}
					onclick={() => (value = opt.value)}
					class="inline-flex items-center gap-1.5 {chipShape} {chipSize} transition-colors {selected
						? chipSelected
						: chipUnselected}"
				>
					{#if selected}
						<svg
							class="h-3.5 w-3.5"
							fill="none"
							stroke="currentColor"
							stroke-width="2.4"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M20 6L9 17l-5-5"
							/>
						</svg>
					{/if}
					{opt.label}
				</button>
			{/each}
		</div>
	</div>
{/if}
