<script lang="ts">
	import type { NotificationDTO } from '$lib/types/api';

	// Desktop notification icon set (mirrors ui-NotifIcon from the design
	// project). Uses its own glyphs rather than $lib/icons because the mockup
	// draws share/transfer/expiry/validity with a distinct, lighter stroke.
	let {
		type,
		size = 20
	}: {
		type: NotificationDTO['type'];
		size?: number;
	} = $props();
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width="1.8"
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>
	{#if type === 'transfer_received'}
		<path d="M4 8.5h13" />
		<path d="M14 5.5l3 3-3 3" />
		<path d="M20 15.5H7" />
		<path d="M10 12.5l-3 3 3 3" />
	{:else if type === 'expiry_reminder'}
		<circle cx="12" cy="13" r="7.2" />
		<path d="M12 9.4V13l2.4 1.6" />
		<path d="M5 3.2L2.4 5.6M19 3.2l2.6 2.4" />
	{:else if type === 'validity_start'}
		<rect x="3" y="5" width="18" height="16" rx="2.4" />
		<path d="M3 9.5h18M8 3v4M16 3v4" />
		<path d="M8.8 15l2.2 2.2 4.2-4.4" />
	{:else}
		<circle cx="6" cy="12" r="2.6" />
		<circle cx="18" cy="6" r="2.6" />
		<circle cx="18" cy="18" r="2.6" />
		<path d="M8.4 10.8l7.2-3.4M8.4 13.2l7.2 3.4" />
	{/if}
</svg>
