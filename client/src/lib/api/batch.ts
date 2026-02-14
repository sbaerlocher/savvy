import { api, getCSRFToken, ApiError } from './client';
import type {
	BatchDeleteRequest,
	BatchShareRequest,
	BatchTransferRequest,
	BatchResponse
} from '$lib/types/api';

type ResourceType = 'cards' | 'vouchers' | 'gift-cards';

function batchDelete(
	resource: ResourceType,
	ids: string[]
): Promise<BatchResponse> {
	return api.post<BatchResponse>(`/${resource}/batch/delete`, {
		ids
	} satisfies BatchDeleteRequest);
}

function batchShare(
	resource: ResourceType,
	req: BatchShareRequest
): Promise<BatchResponse> {
	return api.post<BatchResponse>(`/${resource}/batch/share`, req);
}

function batchTransfer(
	resource: ResourceType,
	ids: string[],
	newOwnerEmail: string
): Promise<BatchResponse> {
	return api.post<BatchResponse>(`/${resource}/batch/transfer`, {
		ids,
		new_owner_email: newOwnerEmail
	} satisfies BatchTransferRequest);
}

const batchErrorMap: Record<string, string> = {
	// Batch item-level errors (from failed[] array)
	'already shared with this user': 'batch.errors.alreadyShared',
	'Not the owner': 'batch.errors.notOwner',
	// HTTP 400 errors (from ErrorResponse.message)
	'Cannot share with yourself': 'batch.errors.selfShare',
	'Cannot transfer to yourself': 'batch.errors.selfTransfer',
	'Could not share with this email address': 'batch.errors.shareFailed',
	'Could not transfer to this email address': 'batch.errors.transferFailed'
};

/**
 * Translates known backend batch error strings to i18n keys.
 * Falls back to the raw error string if no translation is found.
 */
export function translateBatchError(
	error: string,
	t: (key: string) => string
): string {
	const key = batchErrorMap[error];
	return key ? t(key) : error;
}

async function batchExport(
	resource: ResourceType,
	ids: string[]
): Promise<{ blob: Blob; filename: string }> {
	const response = await fetch(`/api/v1/${resource}/batch/export`, {
		method: 'POST',
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			'X-CSRF-Token': getCSRFToken() || ''
		},
		body: JSON.stringify({ ids })
	});

	if (!response.ok) {
		let error = 'export_failed';
		let message = 'Export failed';
		try {
			const data = await response.json();
			error = data.error || error;
			message = data.message || message;
		} catch {
			// non-JSON error response
		}
		throw new ApiError(response.status, error, message);
	}

	const blob = await response.blob();
	const filename =
		response.headers
			.get('Content-Disposition')
			?.match(/filename="?([^"]+)"?/)?.[1] || `savvy-export-${resource}.json`;

	return { blob, filename };
}

export const batchApi = {
	deleteCards: (ids: string[]) => batchDelete('cards', ids),
	deleteVouchers: (ids: string[]) => batchDelete('vouchers', ids),
	deleteGiftCards: (ids: string[]) => batchDelete('gift-cards', ids),

	shareCards: (req: BatchShareRequest) => batchShare('cards', req),
	shareVouchers: (req: BatchShareRequest) => batchShare('vouchers', req),
	shareGiftCards: (req: BatchShareRequest) => batchShare('gift-cards', req),

	transferCards: (ids: string[], email: string) =>
		batchTransfer('cards', ids, email),
	transferVouchers: (ids: string[], email: string) =>
		batchTransfer('vouchers', ids, email),
	transferGiftCards: (ids: string[], email: string) =>
		batchTransfer('gift-cards', ids, email),

	exportCards: (ids: string[]) => batchExport('cards', ids),
	exportVouchers: (ids: string[]) => batchExport('vouchers', ids),
	exportGiftCards: (ids: string[]) => batchExport('gift-cards', ids)
};
