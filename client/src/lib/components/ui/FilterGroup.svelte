<script lang="ts">
	import { platform } from '$lib/utils/platform';

	interface Option {
		value: string;
		label: string;
	}

	// Move a node to <body> so position:fixed is relative to the viewport. An
	// ancestor with backdrop-filter (the glass cards) otherwise re-anchors fixed
	// positioning to itself, trapping/clipping the menu.
	function portal(node: HTMLElement) {
		document.body.appendChild(node);
		return {
			destroy() {
				node.remove();
			}
		};
	}

	// Single-select filter group.
	//   iOS      → collapsed row (label + current value + chevron); tapping opens
	//              a pull-down menu of options — nothing is expanded until tapped.
	//   Android  → M3 filter chips.
	//   desktop  → neutral chips.
	let {
		label,
		value = $bindable(),
		options,
		idPrefix = 'filter'
	}: {
		label: string;
		value: string;
		options: Option[];
		idPrefix?: string;
	} = $props();

	const groupId = $derived(`${idPrefix}-group`);
	const currentLabel = $derived(
		options.find((o) => o.value === value)?.label ?? ''
	);

	let menuOpen = $state(false);
	let triggerEl = $state<HTMLButtonElement | null>(null);
	// Fixed menu position, computed from the trigger on open. Fixed positioning
	// escapes the per-card backdrop-filter stacking contexts, which would
	// otherwise trap the menu under the following filter card.
	let menuStyle = $state('');

	function toggleMenu() {
		if (menuOpen) {
			menuOpen = false;
			return;
		}
		if (triggerEl) {
			const r = triggerEl.getBoundingClientRect();
			// Anchor top-right of the menu to the trigger's right edge, below it.
			const right = Math.round(window.innerWidth - r.right);
			const top = Math.round(r.bottom + 4);
			menuStyle = `position:fixed; top:${top}px; right:${right}px;`;
		}
		menuOpen = true;
	}

	function pick(v: string) {
		value = v;
		menuOpen = false;
	}
</script>

{#if platform === 'ios'}
	<!-- iOS: one collapsed row per filter; a pull-down menu reveals the options.
	     Raise the whole group above sibling cards when open — each glass card is
	     its own stacking context, so the menu would otherwise sit under the next
	     card. -->
	<div class="relative">
		<button
			bind:this={triggerEl}
			type="button"
			onclick={toggleMenu}
			aria-haspopup="menu"
			aria-expanded={menuOpen}
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
				{currentLabel}
				<svg
					class="h-4 w-4 text-text-faint transition-transform {menuOpen
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

		{#if menuOpen}
			<!-- Portaled to <body> so fixed positioning clears the glass cards'
			     backdrop-filter stacking contexts. -->
			<div use:portal>
				<!-- Backdrop to close on outside tap. -->
				<button
					type="button"
					class="fixed inset-0 z-[70] cursor-default"
					aria-label="Close menu"
					onclick={() => (menuOpen = false)}
				></button>
				<!-- Pull-down menu (iOS context-menu style). -->
				<div
					role="menu"
					aria-label={label}
					style={menuStyle}
					class="liquid-glass-menu z-[71] min-w-[12rem] overflow-hidden rounded-2xl py-1"
				>
					{#each options as opt (opt.value)}
						{@const selected = value === opt.value}
						<button
							type="button"
							role="menuitemradio"
							aria-checked={selected}
							onclick={() => pick(opt.value)}
							class="flex w-full items-center justify-between gap-4 px-4 py-2.5 text-left text-[length:var(--text-code)] hover:bg-black/5 {selected
								? 'text-text'
								: 'text-text-muted'}"
						>
							<span>{opt.label}</span>
							{#if selected}
								<svg
									class="h-4 w-4 shrink-0 text-accent"
									fill="none"
									stroke="currentColor"
									stroke-width="2.5"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M20 6L9 17l-5-5"
									/>
								</svg>
							{/if}
						</button>
					{/each}
				</div>
			</div>
		{/if}
	</div>
{:else}
	<div class="py-4">
		<span
			id={groupId}
			class="mb-2 block text-xs font-medium uppercase tracking-wider text-text-subtle"
		>
			{label}
		</span>

		<!-- Android M3 filter chips / neutral chips on desktop. -->
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
					class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {selected
						? 'bg-accent-100 text-accent-850'
						: 'border border-border bg-white text-text-muted hover:bg-surface-1'}"
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
