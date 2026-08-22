<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	// Material 3 outlined text field (Android auth screens): floating label that
	// sits on the white card surface (--color-m3-card), accent 2px outline while focused, neutral 1px
	// otherwise. Android-only by construction — call sites gate on `platform`.
	let {
		id,
		name,
		type = 'text',
		label,
		value = $bindable(''),
		autocomplete,
		required = false,
		disabled = false
	}: {
		id: string;
		name: string;
		type?: 'text' | 'email' | 'password';
		label: string;
		value?: string;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		required?: boolean;
		disabled?: boolean;
	} = $props();

	let focused = $state(false);

	// Svelte forbids a dynamic `type` alongside `bind:value`, so the three
	// variants are spelled out and share this class string.
	const inputClass = $derived(
		`text-subheading text-text h-14 w-full rounded-m3-xs bg-transparent px-4 disabled:opacity-50 ${
			focused ? 'border-2 border-accent-600' : 'border border-border-field'
		}`
	);
</script>

<div class="relative">
	{#if type === 'email'}
		<input
			{id}
			{name}
			type="email"
			{autocomplete}
			{required}
			{disabled}
			bind:value
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
			class={inputClass}
		/>
	{:else if type === 'password'}
		<input
			{id}
			{name}
			type="password"
			{autocomplete}
			{required}
			{disabled}
			bind:value
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
			class={inputClass}
		/>
	{:else}
		<input
			{id}
			{name}
			type="text"
			{autocomplete}
			{required}
			{disabled}
			bind:value
			onfocus={() => (focused = true)}
			onblur={() => (focused = false)}
			class={inputClass}
		/>
	{/if}
	<label
		for={id}
		class="bg-m3-card text-body-sm absolute -top-2 left-3 px-1.5 font-semibold {focused
			? 'text-accent-700'
			: 'text-text-muted'}"
	>
		{label}
	</label>
</div>
