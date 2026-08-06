<script lang="ts">
	import { ICON_LOCK } from '$lib/icons';
	import { t } from '$lib/stores/i18n';
	import EmailAutocomplete from './EmailAutocomplete.svelte';

	interface Props {
		isOffline: boolean;
		title?: string;
		/** Button label when online */
		openButtonLabel: string;
		/** Button label for the confirm/transfer action */
		transferButtonLabel: string;
		warningTitle: string;
		warningDetails: string;
		emailLabel: string;
		emailHint?: string;
		whatHappensLabel: string;
		details: string[];
		email?: string;
		/** Called when user clicks the Transfer button (open confirm modal in parent) */
		ontransfer?: () => void;
	}

	const LOCK_ICON_PATH = ICON_LOCK;

	let {
		isOffline,
		title,
		openButtonLabel,
		transferButtonLabel,
		warningTitle,
		warningDetails,
		emailLabel,
		emailHint,
		whatHappensLabel,
		details,
		email = $bindable(''),
		ontransfer
	}: Props = $props();

	let showForm = $state(false);
</script>

<div class="bg-white rounded-lg shadow-lg p-6 border-2 border-[var(--color-transfer-200)]">
	<div class="flex justify-between items-center mb-4">
		<h3 class="text-lg font-semibold text-[var(--color-transfer-900)]">
			{title ?? $t('common.transferOwnership')}
		</h3>
		{#if !showForm}
			<button
				onclick={() => (showForm = true)}
				disabled={isOffline}
				class="btn btn-xs btn-purple whitespace-nowrap flex items-center gap-1.5 {isOffline
					? 'opacity-50 cursor-not-allowed pointer-events-none blur-[0.5px]'
					: ''}"
			>
				{#if isOffline}
					<svg
						class="w-3.5 h-3.5"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={LOCK_ICON_PATH}
						></path>
					</svg>
					{transferButtonLabel}
				{:else}
					{openButtonLabel}
				{/if}
			</button>
		{/if}
	</div>

	{#if showForm}
		<div class="border border-[var(--color-transfer-200)] bg-[var(--color-transfer-50)] rounded-lg p-4 space-y-4">
			<div class="bg-warning-50 border border-warning-200 rounded-lg p-3">
				<p class="text-sm font-medium text-warning-800">
					<strong>{warningTitle}</strong>
				</p>
				<p class="text-xs text-warning-700 mt-1">{warningDetails}</p>
			</div>

			<EmailAutocomplete
				bind:value={email}
				label={emailLabel}
				hint={emailHint}
				inputId="transfer-email-input"
				disabled={isOffline}
			/>

			<div>
				<p class="text-sm font-medium text-text-ink2 mb-2">
					{whatHappensLabel}
				</p>
				<ul class="text-xs text-text-muted space-y-1">
					{#each details as detail (detail)}
						<li>{detail}</li>
					{/each}
				</ul>
			</div>

			<div class="flex gap-2">
				<button
					onclick={ontransfer}
					disabled={isOffline || !email.trim()}
					class="btn btn-purple flex-1 {isOffline || !email.trim()
						? 'opacity-50 cursor-not-allowed'
						: ''}"
				>
					{transferButtonLabel}
				</button>
				<button onclick={() => (showForm = false)} class="btn btn-ghost">
					{$t('common.cancel')}
				</button>
			</div>
		</div>
	{/if}
</div>
