<script lang="ts">
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
		labelEditTransactionsDesc
	}: Props = $props();
</script>

<div class="space-y-2">
	<label class="flex items-start">
		<input
			type="checkbox"
			bind:checked={canEdit}
			class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
		/>
		<div class="ml-2">
			<span class="block text-sm font-medium text-gray-900">{labelEdit}</span>
			{#if labelEditDesc}
				<span class="text-xs text-gray-500">{labelEditDesc}</span>
			{/if}
		</div>
	</label>
	<label class="flex items-start">
		<input
			type="checkbox"
			bind:checked={canDelete}
			class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
		/>
		<div class="ml-2">
			<span class="block text-sm font-medium text-gray-900">{labelDelete}</span>
			{#if labelDeleteDesc}
				<span class="text-xs text-gray-500">{labelDeleteDesc}</span>
			{/if}
		</div>
	</label>
	{#if showEditTransactions && labelEditTransactions}
		<label class="flex items-start">
			<input
				type="checkbox"
				bind:checked={canEditTransactions}
				class="mt-0.5 h-4 w-4 text-cyan-600 focus:ring-cyan-500 border-gray-300 rounded"
			/>
			<div class="ml-2">
				<span class="block text-sm font-medium text-gray-900"
					>{labelEditTransactions}</span
				>
				{#if labelEditTransactionsDesc}
					<span class="text-xs text-gray-500">{labelEditTransactionsDesc}</span>
				{/if}
			</div>
		</label>
	{/if}
</div>
