<script lang="ts">
	import type { DuplicateWarning } from '$lib/types/api';
	import { t } from '$lib/stores/i18n';
	import { get } from 'svelte/store';

	const tr = (key: string, params?: Record<string, string | number>) =>
		get(t)(key, params);

	interface Props {
		warning: DuplicateWarning | null | undefined;
		resourceType: 'card' | 'voucher' | 'gift_card';
		onNavigate?: (id: string) => void;
		onrestore?: () => void;
	}

	let { warning, resourceType, onNavigate, onrestore }: Props = $props();

	function getResourceTypeLabel(type: string): string {
		switch (type) {
			case 'card':
				return tr('cards.title');
			case 'voucher':
				return tr('vouchers.title');
			case 'gift_card':
				return tr('giftCards.title');
			default:
				return type;
		}
	}

	function handleNavigate() {
		if (warning?.existing_id && onNavigate) {
			onNavigate(warning.existing_id);
		}
	}
</script>

{#if warning?.has_duplicate}
	<div class="duplicate-warning" role="alert">
		<div class="warning-icon">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="1.5"
				stroke="currentColor"
				class="icon"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
				/>
			</svg>
		</div>
		<div class="warning-content">
			<h3 class="warning-title">{tr('duplicate.warning_title')}</h3>
			<p class="warning-message">
				{#if warning.deleted}
					{tr('duplicate.deleted_message')}
				{:else}
					{tr('duplicate.warning_message', {
						resourceType: getResourceTypeLabel(resourceType),
						merchantName: warning.merchant_name || tr('common.unknown'),
						resourceNumber: warning.resource_number || ''
					})}
				{/if}
			</p>
			<div class="warning-actions">
				{#if warning.deleted && onrestore}
					<button
						type="button"
						class="view-existing-btn"
						data-testid="restore-duplicate"
						onclick={onrestore}
					>
						{tr('duplicate.restore')}
					</button>
				{/if}
				{#if warning.existing_id && onNavigate && !warning.deleted}
					<button
						type="button"
						class="view-existing-btn"
						onclick={handleNavigate}
					>
						{tr('duplicate.view_existing')}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.duplicate-warning {
		display: flex;
		gap: 1rem;
		padding: 1rem;
		margin-bottom: 1.5rem;
		background-color: #fef3cd;
		border: 1px solid #ffc107;
		border-radius: 0.5rem;
		color: #856404;
	}

	.warning-icon {
		flex-shrink: 0;
	}

	.icon {
		width: 1.5rem;
		height: 1.5rem;
		color: #ffc107;
	}

	.warning-content {
		flex: 1;
	}

	.warning-title {
		margin: 0 0 0.5rem 0;
		font-size: 1rem;
		font-weight: 600;
		color: #856404;
	}

	.warning-message {
		margin: 0 0 0.75rem 0;
		font-size: 0.875rem;
		line-height: 1.5;
		color: #856404;
	}

	.warning-actions {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.view-existing-btn {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #856404;
		background-color: transparent;
		border: 1px solid #856404;
		border-radius: 0.25rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.view-existing-btn:hover {
		background-color: #856404;
		color: #fff;
	}

	.view-existing-btn:focus {
		outline: 2px solid #ffc107;
		outline-offset: 2px;
	}
</style>
