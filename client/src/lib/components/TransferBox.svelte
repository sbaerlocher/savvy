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
		/**
		 * 'sheet' renders the Android M3 bottom-sheet body (mockup frame 8): the
		 * form is always open — the sheet itself is the disclosure — and the
		 * surrounding card chrome drops away.
		 */
		variant?: 'card' | 'sheet';
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
		ontransfer,
		variant = 'card'
	}: Props = $props();

	const isSheet = $derived(variant === 'sheet');

	// The shared transfer strings carry their own decorative glyphs ("✓ …",
	// "⚠️ Achtung"). The M3 sheet draws real icons, so the glyph is stripped at
	// render — changing the strings would alter the other platforms' copy.
	// U+FE0F is a variation selector, so the warning sign is matched as the
	// sequence it actually is rather than as a member of a character class.
	const stripGlyph = (s: string) =>
		s.replace(/^(?:\s|[\u2713\u2714!]|\u26A0\uFE0F?)+/, '');

	let showForm = $state(false);
</script>

{#if isSheet}
	<!-- Android M3 bottom sheet (screen-ResourceDetailAndroid, frame 8). -->
	<h2 class="text-heading text-transfer-900 mb-4 font-semibold tracking-tight">
		{title ?? $t('common.transferOwnership')}
	</h2>

	<div
		class="bg-danger-50 border-danger-200 mb-4 flex items-start gap-2.5 rounded-m3-md border px-3.5 py-3"
	>
		<svg
			class="text-danger-700 mt-px h-4.5 w-4.5 shrink-0"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			viewBox="0 0 24 24"
		>
			<path d="M12 9v4M12 17h.01" />
			<path
				d="M10.3 3.9L2 18a2 2 0 001.7 3h16.6a2 2 0 001.7-3L13.7 3.9a2 2 0 00-3.4 0z"
			/>
		</svg>
		<p class="text-body-sm text-danger-800">
			<b>{stripGlyph(warningTitle)}</b>
			{stripGlyph(warningDetails)}
		</p>
	</div>

	<div class="mb-4">
		<EmailAutocomplete
			bind:value={email}
			label={emailLabel}
			hint={emailHint}
			inputId="transfer-email-input"
			disabled={isOffline}
		/>
	</div>

	<p class="text-text-subtle mb-2.5 text-eyebrow font-bold uppercase">
		{whatHappensLabel}
	</p>
	<ul class="mb-5 flex flex-col gap-2.5">
		{#each details as detail (detail)}
			<li class="flex items-start gap-2.5">
				<svg
					class="text-transfer-600 mt-px h-4 w-4 shrink-0"
					fill="none"
					stroke="currentColor"
					stroke-width="2.2"
					stroke-linecap="round"
					stroke-linejoin="round"
					viewBox="0 0 24 24"
				>
					<path d="M20 6L9 17l-5-5" />
				</svg>
				<span class="text-label text-text-ink2 font-normal"
					>{stripGlyph(detail)}</span
				>
			</li>
		{/each}
	</ul>

	<button
		type="button"
		onclick={ontransfer}
		disabled={isOffline || !email.trim()}
		class="bg-transfer-600 text-on-accent text-label flex h-12 w-full items-center justify-center rounded-m3-full disabled:opacity-50"
	>
		{transferButtonLabel}
	</button>
{:else}
	<div class="bg-white rounded-lg shadow-lg p-6 border-2 border-transfer-200">
		<div class="flex justify-between items-center mb-4">
			<h3 class="text-lg font-semibold text-transfer-900">
				{title ?? $t('common.transferOwnership')}
			</h3>
			{#if !showForm}
				<button
					onclick={() => (showForm = true)}
					disabled={isOffline}
					class="btn btn-xs btn-transfer whitespace-nowrap flex items-center gap-1.5 {isOffline
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
			<div
				class="border border-transfer-200 bg-transfer-50 rounded-lg p-4 space-y-4"
			>
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
						class="btn btn-transfer flex-1 {isOffline || !email.trim()
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
{/if}
