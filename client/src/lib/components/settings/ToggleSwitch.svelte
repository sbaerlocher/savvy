<script lang="ts">
	interface Props {
		checked: boolean;
		label: string;
		description?: string;
		disabled?: boolean;
		isSaving?: boolean;
		onToggle: () => void;
	}

	let {
		checked,
		label,
		description,
		disabled = false,
		isSaving = false,
		onToggle
	}: Props = $props();
</script>

<div class="flex items-center justify-between">
	<div class="flex-1 mr-4">
		<p class="text-sm font-medium text-text-ink2">{label}</p>
		{#if description}
			<p class="text-xs text-text-subtle mt-1">{description}</p>
		{/if}
	</div>
	<button
		type="button"
		role="switch"
		aria-checked={checked}
		aria-label={label}
		onclick={onToggle}
		disabled={disabled || isSaving}
		class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed {checked
			? 'bg-accent'
			: 'bg-border'}"
	>
		<span
			class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out {checked
				? 'translate-x-5'
				: 'translate-x-0'}"
		></span>
	</button>
</div>
