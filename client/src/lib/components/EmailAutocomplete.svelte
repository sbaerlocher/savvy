<script lang="ts">
	import { onDestroy } from 'svelte';
	import { sharedUsersApi } from '$lib/api';
	import { t } from '$lib/stores/i18n';
	import type { UserDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';

	interface Props {
		value?: string;
		label: string;
		hint?: string;
		inputId?: string;
		disabled?: boolean;
		required?: boolean;
	}

	let {
		value = $bindable(''),
		label,
		hint,
		inputId = 'email-autocomplete',
		disabled = false,
		required = true
	}: Props = $props();

	const componentLogger = logger.child('EmailAutocomplete');
	let suggestedUsers = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
	let highlightedIndex = $state(-1);
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	onDestroy(() => {
		if (searchTimeout) clearTimeout(searchTimeout);
	});

	async function searchUsers(query: string) {
		if (searchTimeout) clearTimeout(searchTimeout);
		if (query.length < 2) {
			suggestedUsers = [];
			showSuggestions = false;
			return;
		}
		searchTimeout = setTimeout(async () => {
			try {
				const response = await sharedUsersApi.search(query);
				suggestedUsers = response.users;
				showSuggestions = suggestedUsers.length > 0;
				highlightedIndex = -1;
			} catch (err) {
				componentLogger.error('Failed to search users:', err);
				suggestedUsers = [];
			}
		}, 300);
	}

	function selectUser(user: UserDTO) {
		value = user.email;
		showSuggestions = false;
		suggestedUsers = [];
		highlightedIndex = -1;
	}

	function onInput(event: Event) {
		const input = event.target as HTMLInputElement;
		value = input.value;
		searchUsers(input.value);
	}

	function onFocus() {
		if (value.length >= 2) {
			searchUsers(value);
		}
	}

	function onBlur() {
		setTimeout(() => {
			showSuggestions = false;
			highlightedIndex = -1;
		}, 200);
	}

	function onKeydown(event: KeyboardEvent) {
		if (!showSuggestions || suggestedUsers.length === 0) return;
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			highlightedIndex = Math.min(
				highlightedIndex + 1,
				suggestedUsers.length - 1
			);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			highlightedIndex = Math.max(highlightedIndex - 1, -1);
		} else if (event.key === 'Enter' && highlightedIndex >= 0) {
			event.preventDefault();
			selectUser(suggestedUsers[highlightedIndex]);
		} else if (event.key === 'Escape') {
			showSuggestions = false;
			highlightedIndex = -1;
		}
	}
</script>

<div class="relative">
	<label for={inputId} class="block text-sm font-medium text-gray-700 mb-1">
		{label}{#if required}
			*{/if}
	</label>
	<input
		id={inputId}
		type="email"
		{value}
		role="combobox"
		aria-expanded={showSuggestions}
		aria-controls="{inputId}-listbox"
		aria-activedescendant={highlightedIndex >= 0 ? `{inputId}-option-${highlightedIndex}` : undefined}
		oninput={onInput}
		onfocus={onFocus}
		onblur={onBlur}
		onkeydown={onKeydown}
		{required}
		placeholder="benutzer@example.com"
		autocomplete="off"
		{disabled}
		class="input bg-white"
	/>

	{#if showSuggestions && suggestedUsers.length > 0}
		<div
			id="{inputId}-listbox"
			role="listbox"
			class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
		>
			{#each suggestedUsers as user, index}
				<button
					id="{inputId}-option-{index}"
					type="button"
					role="option"
					onclick={() => selectUser(user)}
					aria-selected={index === highlightedIndex}
					class="w-full text-left px-3 py-2 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none {index ===
					highlightedIndex
						? 'bg-gray-100'
						: ''}"
				>
					<div class="font-medium text-sm text-gray-900">
						{#if user.first_name && user.last_name}
							{user.first_name}
							{user.last_name}
						{:else if user.first_name}
							{user.first_name}
						{:else}
							{user.email}
						{/if}
					</div>
					<div class="text-xs text-gray-500">{user.email}</div>
				</button>
			{/each}
		</div>
	{/if}

	{#if hint}
		<p class="text-xs text-gray-500 mt-1">{hint}</p>
	{/if}
</div>
