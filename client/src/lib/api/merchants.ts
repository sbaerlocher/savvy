import { api } from './client';
import type { MerchantDTO } from '$lib/types/api';
import { offlineDB } from '$lib/stores/offline-db';
import { browser } from '$app/environment';
import { logger } from '$lib/utils/logger';

const log = logger.child('MerchantsAPI');

export interface MerchantCreateInput {
	name: string;
	color?: string;
	logo_url?: string;
	website?: string;
}

export interface MerchantUpdateInput {
	name?: string;
	color?: string;
	logo_url?: string;
	website?: string;
}

export const merchantsApi = {
	/**
	 * List all merchants — IndexedDB-first with network refresh
	 */
	async list(search?: string): Promise<{ merchants: MerchantDTO[] }> {
		if (browser) {
			const cached = await offlineDB.getAllMerchants();

			if (!navigator.onLine) {
				// Offline: filter cached merchants client-side if search is provided
				if (search) {
					const lower = search.toLowerCase();
					return {
						merchants: cached.filter((m) =>
							m.name.toLowerCase().includes(lower)
						)
					};
				}
				return { merchants: cached };
			}

			try {
				const query = search ? `?search=${encodeURIComponent(search)}` : '';
				const result = await api.get<{ merchants: MerchantDTO[] }>(
					`/merchants${query}`
				);

				// Only replace full cache on non-search requests
				if (!search) {
					await offlineDB.replaceAllMerchants(result.merchants);
				}

				return result;
			} catch (error) {
				log.warn(
					'[merchants.list] Network failed, returning cached data:',
					error
				);
				if (search) {
					const lower = search.toLowerCase();
					return {
						merchants: cached.filter((m) =>
							m.name.toLowerCase().includes(lower)
						)
					};
				}
				return { merchants: cached };
			}
		}

		// SSR fallback
		const query = search ? `?search=${encodeURIComponent(search)}` : '';
		return api.get<{ merchants: MerchantDTO[] }>(`/merchants${query}`);
	},

	/**
	 * Get single merchant — Network-first with IndexedDB fallback
	 */
	async get(id: string): Promise<{ merchant: MerchantDTO }> {
		// Offline: IndexedDB only
		if (browser && !navigator.onLine) {
			const merchant = await offlineDB.getMerchant(id);
			if (merchant) {
				return { merchant };
			}
			throw new Error('Merchant not available offline');
		}

		// Online: network first
		try {
			const result = await api.get<{ merchant: MerchantDTO }>(
				`/merchants/${id}`
			);

			if (browser) {
				await offlineDB.saveMerchant(result.merchant);
			}

			return result;
		} catch (error) {
			// Network failed — try IndexedDB fallback
			if (browser) {
				const merchant = await offlineDB.getMerchant(id);
				if (merchant) {
					return { merchant };
				}
			}
			throw error;
		}
	},

	/**
	 * Create new merchant (Admin only)
	 */
	async create(
		input: MerchantCreateInput
	): Promise<{ message: string; merchant: MerchantDTO }> {
		const result = await api.post<{ message: string; merchant: MerchantDTO }>(
			'/admin/merchants',
			input
		);

		if (browser) {
			await offlineDB.saveMerchant(result.merchant);
		}

		return result;
	},

	/**
	 * Update merchant (Admin only)
	 */
	async update(
		id: string,
		input: MerchantUpdateInput
	): Promise<{ message: string; merchant: MerchantDTO }> {
		const result = await api.patch<{ message: string; merchant: MerchantDTO }>(
			`/admin/merchants/${id}`,
			input
		);

		if (browser) {
			await offlineDB.saveMerchant(result.merchant);
		}

		return result;
	},

	/**
	 * Delete merchant (Admin only)
	 */
	async delete(id: string): Promise<{ message: string }> {
		const result = await api.delete<{ message: string }>(
			`/admin/merchants/${id}`
		);

		if (browser) {
			await offlineDB.deleteMerchant(id);
		}

		return result;
	}
};
