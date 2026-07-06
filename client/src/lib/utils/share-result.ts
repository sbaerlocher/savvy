import type { ShareCreateResponse } from '$lib/types/api';
import { ApiError } from '$lib/api/client';

type Tr = (key: string, params?: Record<string, string | number>) => string;

/**
 * Turns a multi-recipient share response into a user-facing toast.
 * Full success → success toast; partial or full failure → error toast.
 * `tr` is passed in so this stays decoupled from the i18n store.
 */
export function formatShareResult(
	res: ShareCreateResponse,
	tr: Tr
): { message: string; isError: boolean } {
	const failedCount = res.failed?.length ?? 0;
	const total = res.success_count + failedCount;

	if (failedCount === 0) {
		return {
			message: tr('common.shareResultAll', { count: res.success_count }),
			isError: false
		};
	}
	if (res.success_count === 0) {
		return {
			message: tr('common.shareResultNone', { count: failedCount }),
			isError: true
		};
	}
	return {
		message: tr('common.shareResultPartial', {
			success: res.success_count,
			total,
			failed: failedCount
		}),
		isError: true
	};
}

/**
 * Extracts a ShareCreateResponse from a rejected share call. When every
 * recipient fails the backend returns 422 with the same body shape, which the
 * client surfaces as ApiError.data — so the caller can still show a precise
 * "N failed" toast instead of a generic error. Returns null for other errors.
 */
export function shareResponseFromError(
	err: unknown
): ShareCreateResponse | null {
	if (
		err instanceof ApiError &&
		err.status === 422 &&
		err.data &&
		typeof err.data.success_count === 'number' &&
		Array.isArray(err.data.failed)
	) {
		return err.data as ShareCreateResponse;
	}
	return null;
}
