<script lang="ts">
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	// Overflow menu behind the ••• of the Android detail app bar. The mockup
	// keeps sharing and transfer on the button row under the card, so the menu
	// carries the remaining owner action: delete.
	let {
		open,
		isOffline,
		deleteLabel,
		onClose,
		ondelete
	}: {
		open: boolean;
		isOffline: boolean;
		deleteLabel: string;
		onClose: () => void;
		ondelete: () => void;
	} = $props();

	const tr = (key: string) => get(t)(key);
</script>

<BottomSheet {open} {onClose} tonalAndroid ariaLabel={tr('common.more')}>
	<div class="px-2 pb-2">
		<button
			type="button"
			disabled={isOffline}
			onclick={() => {
				onClose();
				ondelete();
			}}
			class="text-label text-danger-600 hover:bg-danger-50 flex h-14 w-full items-center gap-4 rounded-m3-md px-4 text-left disabled:opacity-50"
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
	</div>
</BottomSheet>
