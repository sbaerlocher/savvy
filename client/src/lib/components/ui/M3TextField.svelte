<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { t } from '$lib/stores/i18n';

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
		disabled = false,
		trailingIcon = false,
		hint
	}: {
		id: string;
		name: string;
		type?: 'text' | 'email' | 'password';
		label: string;
		value?: string;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		required?: boolean;
		disabled?: boolean;
		/** Android login mockup shows a visibility toggle inside the password field. */
		trailingIcon?: boolean;
		/** Helper copy below the field (Android register mockup). */
		hint?: string;
	} = $props();

	let revealed = $state(false);

	let focused = $state(false);

	// A dynamic `type` alongside `bind:value` is the pattern Svelte's
	// `attribute_invalid_type` rule targets, so every variant — including the
	// revealed password — is spelled out and shares this class string.
	const inputClass = $derived(
		`text-subheading text-text h-14 w-full rounded-m3-xs bg-transparent px-4 disabled:opacity-50 ${
			trailingIcon ? 'pr-11.5' : ''
		} ${focused ? 'border-2 border-accent-600' : 'border border-border-field'}`
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
	{:else if type === 'password' && revealed}
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
	{#if trailingIcon}
		<button
			type="button"
			{disabled}
			onclick={() => (revealed = !revealed)}
			aria-label={revealed
				? $t('common.hidePassword')
				: $t('common.showPassword')}
			aria-pressed={revealed}
			class="text-text-faint absolute top-1/2 right-3.5 -translate-y-1/2 disabled:opacity-50"
		>
			<svg
				class="h-4.5 w-4.5"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z" />
				<circle cx="12" cy="12" r="3" />
				{#if revealed}
					<path d="M4 4l16 16" />
				{/if}
			</svg>
		</button>
	{/if}
	<label
		for={id}
		class="bg-m3-card text-body-sm absolute -top-2 left-3 px-1.5 font-semibold {focused
			? 'text-accent-700'
			: 'text-text-muted'}"
	>
		{label}
	</label>
	{#if hint}
		<!-- In flow, not absolute: the copy can wrap to a second line and would
		     otherwise run under whatever follows the field. -->
		<span class="text-text-subtle text-mono-sm mt-1 block pl-3.5 font-sans">
			{hint}
		</span>
	{/if}
</div>
