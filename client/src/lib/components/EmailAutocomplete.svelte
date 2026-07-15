<script lang="ts">
	import { onDestroy } from 'svelte';
	import { sharedUsersApi } from '$lib/api';
	import type { UserDTO } from '$lib/types/api';
	import { logger } from '$lib/utils/logger';
	import { t } from '$lib/stores/i18n';
	import { formatUserName } from '$lib/utils/user';

	interface Props {
		value?: string;
		/** When multiple=true, selected recipients are bound here as a list of emails. */
		values?: string[];
		multiple?: boolean;
		label: string;
		hint?: string;
		inputId?: string;
		disabled?: boolean;
		required?: boolean;
	}

	let {
		value = $bindable(''),
		values = $bindable([]),
		multiple = false,
		label,
		hint,
		inputId = 'email-autocomplete',
		disabled = false,
		required = true
	}: Props = $props();

	function addEmail(email: string) {
		const trimmed = email.trim().toLowerCase();
		if (!trimmed || values.includes(trimmed)) return;
		values = [...values, trimmed];
	}

	function removeEmail(email: string) {
		values = values.filter((e) => e !== email);
	}

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
		if (multiple) {
			addEmail(user.email);
			value = '';
		} else {
			value = user.email;
		}
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
		// Commit a fully-typed email synchronously so clicking "Share" (which
		// blurs the input) sees the recipient on the *first* click. Skip while a
		// suggestion dropdown is open — then the user is picking from the list and
		// selectUser handles the commit, avoiding a duplicate chip.
		if (
			multiple &&
			!showSuggestions &&
			highlightedIndex < 0 &&
			value.includes('@')
		) {
			addEmail(value);
			value = '';
		}
		// Close the dropdown after the click on a suggestion button registers.
		setTimeout(() => {
			showSuggestions = false;
			highlightedIndex = -1;
		}, 200);
	}

	function onKeydown(event: KeyboardEvent) {
		// In multiple mode, Enter commits the typed email to a chip even when no
		// autocomplete suggestion is open — must run before the suggestion guard.
		if (
			multiple &&
			event.key === 'Enter' &&
			highlightedIndex < 0 &&
			value.trim()
		) {
			event.preventDefault();
			addEmail(value);
			value = '';
			showSuggestions = false;
			highlightedIndex = -1;
			return;
		}
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
	<label for={inputId} class="block text-sm font-medium text-text-ink2 mb-1">
		{label}{#if required}
			*{/if}
	</label>
	{#if multiple && values.length > 0}
		<div class="flex flex-wrap gap-1 mb-2">
			{#each values as email (email)}
				<span
					class="inline-flex items-center gap-1 rounded-full bg-accent-100 text-accent-800 text-xs px-2 py-1"
				>
					{email}
					<button
						type="button"
						onclick={() => removeEmail(email)}
						{disabled}
						aria-label={$t('common.removeRecipient', { email })}
						class="text-accent hover:text-accent-900 leading-none"
					>
						&times;
					</button>
				</span>
			{/each}
		</div>
	{/if}
	<input
		id={inputId}
		type="email"
		{value}
		role="combobox"
		aria-expanded={showSuggestions}
		aria-controls="{inputId}-listbox"
		aria-activedescendant={highlightedIndex >= 0
			? `${inputId}-option-${highlightedIndex}`
			: undefined}
		oninput={onInput}
		onfocus={onFocus}
		onblur={onBlur}
		onkeydown={onKeydown}
		required={required && !multiple}
		name="share-recipient"
		placeholder={$t('giftCards.sharing.emailPlaceholder')}
		autocomplete="new-password"
		data-1p-ignore
		data-lpignore="true"
		{disabled}
		class="input bg-white"
	/>

	{#if showSuggestions && suggestedUsers.length > 0}
		<div
			id="{inputId}-listbox"
			role="listbox"
			class="absolute z-10 w-full mt-1 bg-white border border-border-field rounded-md shadow-lg max-h-48 overflow-y-auto"
		>
			{#each suggestedUsers as user, index (user.id)}
				<button
					id="{inputId}-option-{index}"
					type="button"
					role="option"
					onclick={() => selectUser(user)}
					aria-selected={index === highlightedIndex}
					class="w-full text-left px-3 py-2 hover:bg-border-soft focus:bg-border-soft focus:outline-none {index ===
					highlightedIndex
						? 'bg-border-soft'
						: ''}"
				>
					<div class="font-medium text-sm text-text">
						{formatUserName(user)}
					</div>
					<div class="text-xs text-text-subtle">{user.email}</div>
				</button>
			{/each}
		</div>
	{/if}

	{#if hint}
		<p class="text-xs text-text-subtle mt-1">{hint}</p>
	{/if}
</div>
