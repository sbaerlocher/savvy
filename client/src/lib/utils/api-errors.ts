import { ApiError } from '$lib/api/client';
import type { DuplicateWarning } from '$lib/types/api';

/**
 * Extracts duplicate warning data from a 409 duplicate_barcode API error.
 * Returns null for any other error type.
 */
export function extractDuplicate(err: unknown): DuplicateWarning | null {
	if (err instanceof ApiError && err.error === 'duplicate_barcode' && err.data?.duplicate) {
		return err.data.duplicate as DuplicateWarning;
	}
	return null;
}
