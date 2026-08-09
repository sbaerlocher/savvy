<script lang="ts">
	import { platform } from '$lib/utils/platform';

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
		groupKey = '',
		openGroup = $bindable(undefined)
	}: {
		label: string;
		value: string;
		options: Option[];
		idPrefix?: string;
		/** iOS only: this group's identity within an accordion set. */
		groupKey?: string;
		/** iOS only: the key of the single open group, shared across siblings so
		 *  expanding one collapses the others. Left undefined by call sites that
		 *  do not want accordion behaviour — then every group stays collapsed
		 *  until tapped and opening one does not affect the rest. */
		openGroup?: string;
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

{#if platform === 'ios'}
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
		<span
			id={groupId}
			class="block uppercase text-text-subtle {platform === 'android'
				? 'mb-2.5 text-eyebrow'
				: 'mb-2 text-xs font-medium tracking-wider'}"
		>
			{label}
		</span>

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
