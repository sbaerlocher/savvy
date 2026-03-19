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
	}

	let {
		value = $bindable(''),
		label,
		hint,
		inputId = 'email-autocomplete',
		disabled = false
	}: Props = $props();

	const componentLogger = logger.child('EmailAutocomplete');
	let suggestedUsers = $state<UserDTO[]>([]);
	let showSuggestions = $state(false);
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
				showSuggestions = true;
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
		}, 200);
	}
</script>

<div class="relative">
	<label for={inputId} class="block text-sm font-medium text-gray-700 mb-1">
		{label} *
	</label>
	<input
		id={inputId}
		type="email"
		{value}
		oninput={onInput}
		onfocus={onFocus}
		onblur={onBlur}
		required
		placeholder="benutzer@example.com"
		autocomplete="off"
		{disabled}
		class="input bg-white"
	/>

	{#if showSuggestions && suggestedUsers.length > 0}
		<div
			class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
		>
			{#each suggestedUsers as user}
				<button
					type="button"
					onclick={() => selectUser(user)}
					class="w-full text-left px-3 py-2 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
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
