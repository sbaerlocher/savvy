<script lang="ts">
	import { ICON_CHECK } from '$lib/icons';

	interface Props {
		canEdit?: boolean;
		canDelete?: boolean;
		canEditTransactions?: boolean;
		showEditTransactions?: boolean;
		labelEdit: string;
		labelEditDesc?: string;
		labelDelete: string;
		labelDeleteDesc?: string;
		labelEditTransactions?: string;
		labelEditTransactionsDesc?: string;
		/**
		 * iOS mockup checkbox chrome: a 24px rounded tick box filled with the
		 * accent instead of the native control. Opt-in per call site so the
		 * Android/desktop share forms keep the native checkbox; a call site that
		 * passes it should already be gated on `platform === 'ios'`.
		 */
		iosBoxes?: boolean;
	}

	let {
		canEdit = $bindable(false),
		canDelete = $bindable(false),
		canEditTransactions = $bindable(false),
		showEditTransactions = false,
		labelEdit,
		labelEditDesc,
		labelDelete,
		labelDeleteDesc,
		labelEditTransactions,
		labelEditTransactionsDesc,
		iosBoxes = false
	}: Props = $props();
</script>

<!-- One iOS permission row: the real checkbox stays in the DOM (sr-only) so
     label association, keyboard focus and form semantics are unchanged; the
     visible box is drawn next to it. -->
{#snippet iosRow(
	checked: boolean,
	onchange: (v: boolean) => void,
	label: string,
	desc: string | undefined
)}
	<label class="flex cursor-pointer items-start gap-2.75">
		<input
			type="checkbox"
			{checked}
			onchange={(e) => onchange(e.currentTarget.checked)}
			class="peer sr-only"
		/>
		<span
			class="mt-0.25 flex h-6 w-6 flex-none items-center justify-center rounded-sm border-2 peer-focus-visible:outline peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-accent {checked
				? 'border-accent bg-accent'
				: 'border-border-field bg-white'}"
		>
			{#if checked}
				<svg
					class="h-3.25 w-3.25"
					fill="none"
					stroke="var(--color-on-accent)"
					stroke-width="3"
					viewBox="0 0 24 24"
					aria-hidden="true"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d={ICON_CHECK} />
				</svg>
			{/if}
		</span>
		<span>
			<span class="block text-body font-semibold text-text">{label}</span>
			{#if desc}
				<span class="mt-0.25 block text-body-sm text-text-subtle">{desc}</span>
			{/if}
		</span>
	</label>
{/snippet}

{#if iosBoxes}
	<div class="flex flex-col gap-2.75">
		{@render iosRow(canEdit, (v) => (canEdit = v), labelEdit, labelEditDesc)}
		{@render iosRow(
			canDelete,
			(v) => (canDelete = v),
			labelDelete,
			labelDeleteDesc
		)}
		{#if showEditTransactions && labelEditTransactions}
			{@render iosRow(
				canEditTransactions,
				(v) => (canEditTransactions = v),
				labelEditTransactions,
				labelEditTransactionsDesc
			)}
		{/if}
	</div>
{:else}
	<div class="space-y-2">
		<label class="flex items-start">
			<input
				type="checkbox"
				bind:checked={canEdit}
				class="mt-0.5 h-4 w-4 text-accent focus:ring-accent border-border-field rounded"
			/>
			<div class="ml-2">
				<span class="block text-sm font-medium text-text">{labelEdit}</span>
				{#if labelEditDesc}
					<span class="text-xs text-text-subtle">{labelEditDesc}</span>
				{/if}
			</div>
		</label>
		<label class="flex items-start">
			<input
				type="checkbox"
				bind:checked={canDelete}
				class="mt-0.5 h-4 w-4 text-accent focus:ring-accent border-border-field rounded"
			/>
			<div class="ml-2">
				<span class="block text-sm font-medium text-text">{labelDelete}</span>
				{#if labelDeleteDesc}
					<span class="text-xs text-text-subtle">{labelDeleteDesc}</span>
				{/if}
			</div>
		</label>
		{#if showEditTransactions && labelEditTransactions}
			<label class="flex items-start">
				<input
					type="checkbox"
					bind:checked={canEditTransactions}
					class="mt-0.5 h-4 w-4 text-accent focus:ring-accent border-border-field rounded"
				/>
				<div class="ml-2">
					<span class="block text-sm font-medium text-text"
						>{labelEditTransactions}</span
					>
					{#if labelEditTransactionsDesc}
						<span class="text-xs text-text-subtle"
							>{labelEditTransactionsDesc}</span
						>
					{/if}
				</div>
			</label>
		{/if}
	</div>
{/if}
