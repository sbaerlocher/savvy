<script lang="ts">
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import { ICON_TRANSFER } from '$lib/icons';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	// Overflow menu behind the ••• on the Android detail title row: share and
	// transfer (owner only) plus delete. Each action is optional — the caller
	// passes only what the user's permissions allow, and an entry without a
	// handler is simply not rendered.
	let {
		open,
		isOffline,
		shareLabel,
		deleteLabel,
		onClose,
		onshare,
		ontransfer,
		ondelete
	}: {
		open: boolean;
		isOffline: boolean;
		shareLabel: string;
		deleteLabel: string;
		onClose: () => void;
		onshare?: () => void;
		ontransfer?: () => void;
		ondelete?: () => void;
	} = $props();

	const tr = (key: string) => get(t)(key);

	function run(action: () => void) {
		onClose();
		action();
	}

	const ENTRY =
		'text-label flex h-14 w-full items-center gap-4 rounded-m3-md px-4 text-left disabled:opacity-50';
</script>

<BottomSheet
	{open}
	{onClose}
	tonalAndroid
	allowWide
	ariaLabel={tr('common.more')}
>
	<div class="px-2 pb-2">
		{#if onshare}
			<button
				type="button"
				disabled={isOffline}
				onclick={() => run(onshare)}
				class="{ENTRY} text-text hover:bg-surface-1"
			>
				<svg
					class="h-4.5 w-4.5 shrink-0"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					viewBox="0 0 24 24"
				>
					<circle cx="18" cy="5" r="3" />
					<circle cx="6" cy="12" r="3" />
					<circle cx="18" cy="19" r="3" />
					<path d="M8.6 13.5l6.8 4M15.4 6.5l-6.8 4" />
				</svg>
				{shareLabel}
			</button>
		{/if}
		{#if ontransfer}
			<button
				type="button"
				disabled={isOffline}
				onclick={() => run(ontransfer)}
				class="{ENTRY} text-text hover:bg-surface-1"
			>
				<svg
					class="h-4.5 w-4.5 shrink-0"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					viewBox="0 0 24 24"
				>
					<path d={ICON_TRANSFER} />
				</svg>
				{tr('common.transfer')}
			</button>
		{/if}
		{#if ondelete}
			<button
				type="button"
				disabled={isOffline}
				onclick={() => run(ondelete)}
				class="{ENTRY} text-danger-600 hover:bg-danger-50"
			>
				<svg
					class="h-4.5 w-4.5 shrink-0"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					viewBox="0 0 24 24"
				>
					<path d="M4 7h16" />
					<path d="M9 7V5a1 1 0 011-1h4a1 1 0 011 1v2" />
					<path d="M6 7l1 13a1 1 0 001 1h8a1 1 0 001-1l1-13" />
				</svg>
				{deleteLabel}
			</button>
		{/if}
	</div>
</BottomSheet>
