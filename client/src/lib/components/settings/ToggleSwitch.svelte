<script lang="ts">
	import { platform } from '$lib/utils/platform';

	interface Props {
		checked: boolean;
		label: string;
		description?: string;
		disabled?: boolean;
		isSaving?: boolean;
		onToggle: () => void;
		/**
		 * Renders the switch alone, without the label/description block. The iOS
		 * grouped-inset rows and the Android M3 list items draw their own label,
		 * so they own the row layout and only borrow the control. Other platforms
		 * keep the full row.
		 */
		bare?: boolean;
	}

	let {
		checked,
		label,
		description,
		disabled = false,
		isSaving = false,
		onToggle,
		bare = false
	}: Props = $props();

	// `platform` is a module constant, so these are plain consts, not $derived.
	// iOS uses the 51x31 UISwitch from the mockup, Android the 52x32 M3 switch
	// with a 2px outline, and the other platforms keep the existing 44x24
	// control. The iOS knob reuses shadow-sm: the mockup's own --shadow-toggle
	// has no counterpart in tokens.css, and shadow-sm is the closest existing
	// step.
	const IS_IOS = platform === 'ios';
	const IS_ANDROID = platform === 'android';

	const trackClass = IS_IOS
		? 'relative inline-flex h-7.75 w-12.75 shrink-0 cursor-pointer rounded-full transition-colors duration-280 ease-in-out focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed'
		: IS_ANDROID
			? 'relative box-border inline-flex h-8 w-13 shrink-0 cursor-pointer rounded-m3-full border-2 transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed'
			: 'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';

	// Android tints the outline as well, so its track carries the state on both
	// background and border; the other platforms only swap the fill.
	const trackStateClass = $derived(
		IS_ANDROID
			? checked
				? 'border-accent bg-accent'
				: 'border-text-placeholder bg-border'
			: checked
				? 'bg-accent'
				: 'bg-border'
	);

	const knobClass = IS_IOS
		? 'pointer-events-none absolute top-0.5 left-0.5 inline-block h-6.75 w-6.75 transform rounded-full bg-surface shadow-sm transition-transform duration-280 ease-in-out'
		: IS_ANDROID
			? 'pointer-events-none absolute top-1/2 left-0 -translate-y-1/2 rounded-m3-full transition-all duration-200 ease-in-out'
			: 'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out';

	// The Android knob grows and recolours with the state, not just shifts. Its
	// offsets sit 2px inside the others because the outline counts towards the
	// border box.
	const knobStateClass = $derived(
		IS_ANDROID
			? checked
				? 'h-5.5 w-5.5 translate-x-6 bg-white'
				: 'h-4 w-4 translate-x-0.75 bg-text-subtle'
			: checked
				? 'translate-x-5'
				: 'translate-x-0'
	);
</script>

{#snippet control()}
	<button
		type="button"
		role="switch"
		aria-checked={checked}
		aria-label={label}
		onclick={onToggle}
		disabled={disabled || isSaving}
		class="{trackClass} {trackStateClass}"
	>
		<span class="{knobClass} {knobStateClass}"></span>
	</button>
{/snippet}

{#if bare}
	{@render control()}
{:else}
	<div class="flex items-center justify-between">
		<div class="flex-1 mr-4">
			<p class="text-sm font-medium text-text-ink2">{label}</p>
			{#if description}
				<p class="text-xs text-text-subtle mt-1">{description}</p>
			{/if}
		</div>
		{@render control()}
	</div>
{/if}
