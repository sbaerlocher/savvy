import { api } from './client';
import { createShareWithPermissionsApi } from './shares';
import type {
	GiftCardDTO,
	GiftCardCreateRequest,
	GiftCardUpdateRequest,
	TransactionDTO,
	TransactionCreateRequest,
	GiftCardResponse
} from '$lib/types/api';
import { offlineDB } from '$lib/stores/offline-db';
import { browser } from '$app/environment';
import { logger } from '$lib/utils/logger';

const log = logger.child('GiftCardsAPI');

// Create share API operations for gift cards
const shareOps = createShareWithPermissionsApi('gift-cards');

const READ_ONLY_PERMISSIONS = {
	can_view: true,
	can_edit: false,
	can_delete: false,
	can_edit_transactions: false,
	is_owner: false
};

export const giftCardsApi = {
	/**
	 * List all gift cards — IndexedDB-first with network refresh
	 *
	 * @param page - Page number (1-indexed, optional for progressive loading)
	 * @param perPage - Items per page (optional, default 25)
	 */
	async list(
		page?: number,
		perPage?: number
	): Promise<{
		gift_cards: GiftCardDTO[];
		pagination?: import('$lib/types/api').PaginationMeta;
	}> {
		if (browser) {
			// Read all gift cards (owned + shared) from IndexedDB
			const cached = await offlineDB.getAllGiftCards();

			if (!navigator.onLine) {
				return { gift_cards: cached };
			}

			try {
				// Build query params for pagination
				const queryParams = new URLSearchParams();
				if (page !== undefined) queryParams.set('page', String(page));
				if (perPage !== undefined) queryParams.set('per_page', String(perPage));
				const queryString = queryParams.toString();
				const url = queryString ? `/gift-cards?${queryString}` : '/gift-cards';

				const result = await api.get<{
					gift_cards: GiftCardDTO[];
					pagination?: import('$lib/types/api').PaginationMeta;
				}>(url);

				// Store all gift cards together (owned + shared) in IndexedDB
				// Backend already returns UNION of owned + shared
				// permissions.is_owner flag distinguishes them
				if (!page) {
					// Non-paginated: replace all (sync with server)
					await offlineDB.replaceAllGiftCards(result.gift_cards);
				} else {
					// Paginated: append to IndexedDB
					await offlineDB.saveManyGiftCards(result.gift_cards);
				}

				return result;
			} catch (error) {
				log.warn(
					'[gift-cards.list] Network failed, returning cached data:',
					error
				);
				return { gift_cards: cached };
			}
		}

		// SSR fallback
		const queryParams = new URLSearchParams();
		if (page !== undefined) queryParams.set('page', String(page));
		if (perPage !== undefined) queryParams.set('per_page', String(perPage));
		const queryString = queryParams.toString();
		const url = queryString ? `/gift-cards?${queryString}` : '/gift-cards';
		return api.get<{
			gift_cards: GiftCardDTO[];
			pagination?: import('$lib/types/api').PaginationMeta;
		}>(url);
	},

	/**
	 * Get single gift card from IndexedDB cache only (no network)
	 */
	async getCached(id: string): Promise<GiftCardResponse | null> {
		if (!browser) return null;
		const giftCard = await offlineDB.getGiftCard(id);
		if (!giftCard) return null;
		return {
			gift_card: giftCard,
			permissions: READ_ONLY_PERMISSIONS,
			transactions: giftCard.transactions || [],
			shares: []
		};
	},

	/**
	 * Get single gift card — Network-first with IndexedDB fallback
	 */
	async get(id: string): Promise<GiftCardResponse> {
		if (browser && !navigator.onLine) {
			const giftCard = await offlineDB.getGiftCard(id);
			if (giftCard) {
				return {
					gift_card: giftCard,
					permissions: READ_ONLY_PERMISSIONS,
					transactions: giftCard.transactions || [],
					shares: []
				};
			}
			throw new Error('Gift card not available offline');
		}

		try {
			const result = await api.get<GiftCardResponse>(`/gift-cards/${id}`);

			if (browser) {
				// Embed transactions in gift card for offline cache
				const giftCardWithTransactions = {
					...result.gift_card,
					transactions: result.transactions || []
				};
				await offlineDB.saveGiftCard(giftCardWithTransactions);
			}

			return result;
		} catch (error) {
			if (browser) {
				const giftCard = await offlineDB.getGiftCard(id);
				if (giftCard) {
					return {
						gift_card: giftCard,
						permissions: READ_ONLY_PERMISSIONS,
						transactions: giftCard.transactions || [],
						shares: []
					};
				}
			}
			throw error;
		}
	},

	/**
	 * Create new gift card
	 */
	async create(data: GiftCardCreateRequest): Promise<GiftCardResponse> {
		const result = await api.post<GiftCardResponse>('/gift-cards', data);

		if (browser) {
			await offlineDB.saveGiftCard(result.gift_card);
		}

		return result;
	},

	/**
	 * Update gift card — Optimistic Update Pattern
	 */
	async update(
		id: string,
		data: GiftCardUpdateRequest
	): Promise<GiftCardResponse> {
		const originalGiftCard = browser ? await offlineDB.getGiftCard(id) : null;

		if (browser && navigator.onLine && originalGiftCard) {
			await offlineDB.saveGiftCard({ ...originalGiftCard, ...data });
		}

		try {
			const result = await api.patch<GiftCardResponse>(
				`/gift-cards/${id}`,
				data
			);

			if (browser) {
				await offlineDB.saveGiftCard(result.gift_card);
			}

			return result;
		} catch (error) {
			if (browser && originalGiftCard) {
				await offlineDB.saveGiftCard(originalGiftCard);
			}
			throw error;
		}
	},

	/**
	 * Delete gift card
	 */
	async delete(id: string): Promise<{ message: string }> {
		const result = await api.delete<{ message: string }>(`/gift-cards/${id}`);

		if (browser) {
			await offlineDB.deleteGiftCard(id);
		}

		return result;
	},

	/**
	 * Toggle favorite — Optimistic Update Pattern
	 */
	async toggleFavorite(id: string): Promise<{ is_favorite: boolean }> {
		const originalGiftCard = browser ? await offlineDB.getGiftCard(id) : null;
		const originalIsFavorite = originalGiftCard?.is_favorite ?? false;

		if (browser && navigator.onLine && originalGiftCard) {
			originalGiftCard.is_favorite = !originalIsFavorite;
			await offlineDB.saveGiftCard(originalGiftCard);
		}

		try {
			const result = await api.post<{ is_favorite: boolean }>(
				`/gift-cards/${id}/favorite`
			);

			if (browser) {
				const cached = await offlineDB.getGiftCard(id);
				if (cached) {
					cached.is_favorite = result.is_favorite;
					await offlineDB.saveGiftCard(cached);
				}
			}

			return result;
		} catch (error) {
			if (browser && originalGiftCard) {
				originalGiftCard.is_favorite = originalIsFavorite;
				await offlineDB.saveGiftCard(originalGiftCard);
			}
			throw error;
		}
	},

	/**
	 * List transactions for gift card — with offline fallback from cached gift card
	 */
	async listTransactions(
		id: string
	): Promise<{ transactions: TransactionDTO[] }> {
		if (browser && !navigator.onLine) {
			const giftCard = await offlineDB.getGiftCard(id);
			return { transactions: giftCard?.transactions || [] };
		}

		try {
			return await api.get<{ transactions: TransactionDTO[] }>(
				`/gift-cards/${id}/transactions`
			);
		} catch (error) {
			if (browser) {
				const giftCard = await offlineDB.getGiftCard(id);
				if (giftCard?.transactions) {
					log.warn(
						'[gift-cards.listTransactions] Network failed, returning cached transactions'
					);
					return { transactions: giftCard.transactions };
				}
			}
			throw error;
		}
	},

	/**
	 * Create transaction
	 */
	async createTransaction(
		id: string,
		data: TransactionCreateRequest
	): Promise<{ transaction: TransactionDTO }> {
		const result = await api.post<{ transaction: TransactionDTO }>(
			`/gift-cards/${id}/transactions`,
			data
		);

		// Refresh gift card cache to update balance and transactions
		if (browser) {
			try {
				const updated = await api.get<GiftCardResponse>(`/gift-cards/${id}`);
				await offlineDB.saveGiftCard({
					...updated.gift_card,
					transactions: updated.transactions || []
				});
			} catch (error) {
				log.warn(
					'[gift-cards] Failed to refresh cache after transaction:',
					error
				);
			}
		}

		return result;
	},

	/**
	 * Update transaction
	 */
	async updateTransaction(
		giftCardId: string,
		transactionId: string,
		data: Partial<TransactionCreateRequest>
	): Promise<{ transaction: TransactionDTO }> {
		const result = await api.patch<{ transaction: TransactionDTO }>(
			`/gift-cards/${giftCardId}/transactions/${transactionId}`,
			data
		);

		if (browser) {
			try {
				const updated = await api.get<GiftCardResponse>(
					`/gift-cards/${giftCardId}`
				);
				await offlineDB.saveGiftCard({
					...updated.gift_card,
					transactions: updated.transactions || []
				});
			} catch (error) {
				log.warn(
					'[gift-cards] Failed to refresh cache after transaction:',
					error
				);
			}
		}

		return result;
	},

	/**
	 * Delete transaction
	 */
	async deleteTransaction(
		giftCardId: string,
		transactionId: string
	): Promise<{ message: string }> {
		const result = await api.delete<{ message: string }>(
			`/gift-cards/${giftCardId}/transactions/${transactionId}`
		);

		if (browser) {
			try {
				const updated = await api.get<GiftCardResponse>(
					`/gift-cards/${giftCardId}`
				);
				await offlineDB.saveGiftCard({
					...updated.gift_card,
					transactions: updated.transactions || []
				});
			} catch (error) {
				log.warn(
					'[gift-cards] Failed to refresh cache after transaction:',
					error
				);
			}
		}

		return result;
	},

	// Share operations (delegated to shared utility)
	...shareOps
};
