import { api } from './client';
import { createShareApi } from './shares';
import type {
	VoucherDTO,
	VoucherCreateRequest,
	VoucherUpdateRequest,
	PermissionDTO,
	ShareDTO
} from '$lib/types/api';
import { offlineDB } from '$lib/stores/offline-db';
import { browser } from '$app/environment';
import { logger } from '$lib/utils/logger';

const log = logger.child('VouchersAPI');

// Create share API operations for vouchers (read-only shares, no permissions)
const shareOps = createShareApi('vouchers');

const READ_ONLY_PERMISSIONS: PermissionDTO = {
	can_view: true,
	can_edit: false,
	can_delete: false,
	is_owner: false
};

export const vouchersApi = {
	/**
	 * List all vouchers — IndexedDB-first with network refresh
	 *
	 * @param page - Page number (1-indexed, optional for progressive loading)
	 * @param perPage - Items per page (optional, default 25)
	 */
	async list(
		page?: number,
		perPage?: number
	): Promise<{
		vouchers: VoucherDTO[];
		pagination?: import('$lib/types/api').PaginationMeta;
	}> {
		if (browser) {
			// Read all vouchers (owned + shared) from IndexedDB
			const cached = await offlineDB.getAllVouchers();

			if (!navigator.onLine) {
				return { vouchers: cached };
			}

			try {
				// Build query params for pagination
				const queryParams = new URLSearchParams();
				if (page !== undefined) queryParams.set('page', String(page));
				if (perPage !== undefined) queryParams.set('per_page', String(perPage));
				const queryString = queryParams.toString();
				const url = queryString ? `/vouchers?${queryString}` : '/vouchers';

				const result = await api.get<{
					vouchers: VoucherDTO[];
					pagination?: import('$lib/types/api').PaginationMeta;
				}>(url);

				// Store all vouchers together (owned + shared) in IndexedDB
				// Backend already returns UNION of owned + shared
				// permissions.is_owner flag distinguishes them
				if (!page) {
					// Non-paginated: replace all (sync with server)
					await offlineDB.replaceAllVouchers(result.vouchers);
				} else {
					// Paginated: append to IndexedDB
					await offlineDB.saveManyVouchers(result.vouchers);
				}

				return result;
			} catch (error) {
				log.warn(
					'[vouchers.list] Network failed, returning cached data:',
					error
				);
				return { vouchers: cached };
			}
		}

		// SSR fallback
		const queryParams = new URLSearchParams();
		if (page !== undefined) queryParams.set('page', String(page));
		if (perPage !== undefined) queryParams.set('per_page', String(perPage));
		const queryString = queryParams.toString();
		const url = queryString ? `/vouchers?${queryString}` : '/vouchers';
		return api.get<{
			vouchers: VoucherDTO[];
			pagination?: import('$lib/types/api').PaginationMeta;
		}>(url);
	},

	/**
	 * Get single voucher from IndexedDB cache only (no network)
	 */
	async getCached(id: string): Promise<{
		voucher: VoucherDTO;
		permissions: PermissionDTO;
		shares: ShareDTO[];
	} | null> {
		if (!browser) return null;
		const voucher = await offlineDB.getVoucher(id);
		if (!voucher) return null;
		return { voucher, permissions: READ_ONLY_PERMISSIONS, shares: [] };
	},

	/**
	 * Get single voucher — Network-first with IndexedDB fallback
	 */
	async get(id: string): Promise<{
		voucher: VoucherDTO;
		permissions: PermissionDTO;
		shares: ShareDTO[];
	}> {
		if (browser && !navigator.onLine) {
			const voucher = await offlineDB.getVoucher(id);
			if (voucher) {
				return { voucher, permissions: READ_ONLY_PERMISSIONS, shares: [] };
			}
			throw new Error('Voucher not available offline');
		}

		try {
			const result = await api.get<{
				voucher: VoucherDTO;
				permissions: PermissionDTO;
				shares: ShareDTO[];
			}>(`/vouchers/${id}`);

			if (browser) {
				await offlineDB.saveVoucher(result.voucher);
			}

			return result;
		} catch (error) {
			if (browser) {
				const voucher = await offlineDB.getVoucher(id);
				if (voucher) {
					return { voucher, permissions: READ_ONLY_PERMISSIONS, shares: [] };
				}
			}
			throw error;
		}
	},

	/**
	 * Create new voucher
	 */
	async create(data: VoucherCreateRequest): Promise<{ voucher: VoucherDTO }> {
		const result = await api.post<{ voucher: VoucherDTO }>('/vouchers', data);

		if (browser) {
			await offlineDB.saveVoucher(result.voucher);
		}

		return result;
	},

	/**
	 * Update voucher — Optimistic Update Pattern
	 */
	async update(
		id: string,
		data: VoucherUpdateRequest
	): Promise<{
		voucher: VoucherDTO;
		permissions?: PermissionDTO;
		shares?: ShareDTO[];
	}> {
		const originalVoucher = browser ? await offlineDB.getVoucher(id) : null;

		if (browser && navigator.onLine && originalVoucher) {
			await offlineDB.saveVoucher({ ...originalVoucher, ...data });
		}

		try {
			const result = await api.patch<{
				voucher: VoucherDTO;
				permissions?: PermissionDTO;
				shares?: ShareDTO[];
			}>(`/vouchers/${id}`, data);

			if (browser) {
				await offlineDB.saveVoucher(result.voucher);
			}

			return result;
		} catch (error) {
			if (browser && originalVoucher) {
				await offlineDB.saveVoucher(originalVoucher);
			}
			throw error;
		}
	},

	/**
	 * Delete voucher
	 */
	async delete(id: string): Promise<{ message: string }> {
		const result = await api.delete<{ message: string }>(`/vouchers/${id}`);

		if (browser) {
			await offlineDB.deleteVoucher(id);
		}

		return result;
	},

	/**
	 * Toggle favorite — Optimistic Update Pattern
	 */
	async toggleFavorite(id: string): Promise<{ is_favorite: boolean }> {
		const originalVoucher = browser ? await offlineDB.getVoucher(id) : null;
		const originalIsFavorite = originalVoucher?.is_favorite ?? false;

		if (browser && navigator.onLine && originalVoucher) {
			originalVoucher.is_favorite = !originalIsFavorite;
			await offlineDB.saveVoucher(originalVoucher);
		}

		try {
			const result = await api.post<{ is_favorite: boolean }>(
				`/vouchers/${id}/favorite`
			);

			if (browser) {
				const cached = await offlineDB.getVoucher(id);
				if (cached) {
					cached.is_favorite = result.is_favorite;
					await offlineDB.saveVoucher(cached);
				}
			}

			return result;
		} catch (error) {
			if (browser && originalVoucher) {
				originalVoucher.is_favorite = originalIsFavorite;
				await offlineDB.saveVoucher(originalVoucher);
			}
			throw error;
		}
	},

	// Share operations (delegated to shared utility)
	...shareOps
};
