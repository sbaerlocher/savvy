import { api } from './client';
import { createShareWithPermissionsApi } from './shares';
import type {
	CardDTO,
	CardCreateRequest,
	CardUpdateRequest,
	PermissionDTO,
	ShareDTO
} from '$lib/types/api';
import { offlineDB } from '$lib/stores/offline-db';
import { browser } from '$app/environment';
import { logger } from '$lib/utils/logger';

const log = logger.child('CardsAPI');

// Create share API operations for cards
const shareOps = createShareWithPermissionsApi('cards');

const READ_ONLY_PERMISSIONS: PermissionDTO = {
	can_view: true,
	can_edit: false,
	can_delete: false,
	is_owner: false
};

export const cardsApi = {
	/**
	 * List all cards — IndexedDB-first with network refresh
	 *
	 * Online: fetch from network → update IndexedDB → return fresh data
	 * Offline: return cached data from IndexedDB
	 * Network error: return cached data as fallback
	 *
	 * @param page - Page number (1-indexed, optional for progressive loading)
	 * @param perPage - Items per page (optional, default 25)
	 */
	async list(
		page?: number,
		perPage?: number
	): Promise<{
		cards: CardDTO[];
		pagination?: import('$lib/types/api').PaginationMeta;
	}> {
		if (browser) {
			// Read all cards (owned + shared) from IndexedDB
			const cached = await offlineDB.getAllCards();

			if (!navigator.onLine) {
				return { cards: cached };
			}

			try {
				// Build query params for pagination
				const queryParams = new URLSearchParams();
				if (page !== undefined) queryParams.set('page', String(page));
				if (perPage !== undefined) queryParams.set('per_page', String(perPage));
				const queryString = queryParams.toString();
				const url = queryString ? `/cards?${queryString}` : '/cards';

				const result = await api.get<{
					cards: CardDTO[];
					pagination?: import('$lib/types/api').PaginationMeta;
				}>(url);

				// Store all cards together (owned + shared) in IndexedDB
				// Backend already returns UNION of owned + shared
				// permissions.is_owner flag distinguishes them
				if (!page) {
					// Non-paginated: replace all (sync with server)
					await offlineDB.replaceAllCards(result.cards);
				} else {
					// Paginated: append to IndexedDB
					await offlineDB.saveManyCards(result.cards);
				}

				return result;
			} catch (error) {
				log.warn('[cards.list] Network failed, returning cached data:', error);
				return { cards: cached };
			}
		}

		// SSR fallback
		const queryParams = new URLSearchParams();
		if (page !== undefined) queryParams.set('page', String(page));
		if (perPage !== undefined) queryParams.set('per_page', String(perPage));
		const queryString = queryParams.toString();
		const url = queryString ? `/cards?${queryString}` : '/cards';
		return api.get<{
			cards: CardDTO[];
			pagination?: import('$lib/types/api').PaginationMeta;
		}>(url);
	},

	/**
	 * Get single card from IndexedDB cache only (no network)
	 */
	async getCached(id: string): Promise<{
		card: CardDTO;
		permissions: PermissionDTO;
		shares: ShareDTO[];
	} | null> {
		if (!browser) return null;
		const card = await offlineDB.getCard(id);
		if (!card) return null;
		return { card, permissions: READ_ONLY_PERMISSIONS, shares: [] };
	},

	/**
	 * Get single card — Network-first with IndexedDB fallback
	 */
	async get(id: string): Promise<{
		card: CardDTO;
		permissions: PermissionDTO;
		shares: ShareDTO[];
	}> {
		// Offline: IndexedDB only
		if (browser && !navigator.onLine) {
			const card = await offlineDB.getCard(id);
			if (card) {
				return { card, permissions: READ_ONLY_PERMISSIONS, shares: [] };
			}
			throw new Error('Card not available offline');
		}

		// Online: network first
		try {
			const result = await api.get<{
				card: CardDTO;
				permissions: PermissionDTO;
				shares: ShareDTO[];
			}>(`/cards/${id}`);

			if (browser) {
				await offlineDB.saveCard(result.card);
			}

			return result;
		} catch (error) {
			// Network failed — try IndexedDB fallback
			if (browser) {
				const card = await offlineDB.getCard(id);
				if (card) {
					return { card, permissions: READ_ONLY_PERMISSIONS, shares: [] };
				}
			}
			throw error;
		}
	},

	/**
	 * Create new card
	 */
	async create(
		data: CardCreateRequest
	): Promise<{ card: CardDTO; permissions: PermissionDTO }> {
		const result = await api.post<{
			card: CardDTO;
			permissions: PermissionDTO;
		}>('/cards', data);

		if (browser) {
			await offlineDB.saveCard(result.card);
		}

		return result;
	},

	/**
	 * Update card — Optimistic Update Pattern
	 */
	async update(
		id: string,
		data: CardUpdateRequest
	): Promise<{
		card: CardDTO;
		permissions: PermissionDTO;
		shares: ShareDTO[];
	}> {
		const originalCard = browser ? await offlineDB.getCard(id) : null;

		// Optimistic update
		if (browser && navigator.onLine && originalCard) {
			await offlineDB.saveCard({ ...originalCard, ...data });
		}

		try {
			const result = await api.patch<{
				card: CardDTO;
				permissions: PermissionDTO;
				shares: ShareDTO[];
			}>(`/cards/${id}`, data);

			if (browser) {
				await offlineDB.saveCard(result.card);
			}

			return result;
		} catch (error) {
			// Rollback on failure
			if (browser && originalCard) {
				await offlineDB.saveCard(originalCard);
			}
			throw error;
		}
	},

	/**
	 * Delete card
	 */
	async delete(id: string): Promise<{ message: string }> {
		const result = await api.delete<{ message: string }>(`/cards/${id}`);

		if (browser) {
			await offlineDB.deleteCard(id);
		}

		return result;
	},

	/**
	 * Restore soft-deleted card
	 */
	async restore(
		id: string
	): Promise<{ card: CardDTO; permissions: PermissionDTO }> {
		return api.post<{ card: CardDTO; permissions: PermissionDTO }>(
			`/cards/${id}/restore`
		);
	},

	/**
	 * Toggle favorite — Optimistic Update Pattern
	 */
	async toggleFavorite(id: string): Promise<{ is_favorite: boolean }> {
		const originalCard = browser ? await offlineDB.getCard(id) : null;
		const originalIsFavorite = originalCard?.is_favorite ?? false;

		if (browser && navigator.onLine && originalCard) {
			originalCard.is_favorite = !originalIsFavorite;
			await offlineDB.saveCard(originalCard);
		}

		try {
			const result = await api.post<{ is_favorite: boolean }>(
				`/cards/${id}/favorite`
			);

			if (browser) {
				const cachedCard = await offlineDB.getCard(id);
				if (cachedCard) {
					cachedCard.is_favorite = result.is_favorite;
					await offlineDB.saveCard(cachedCard);
				}
			}

			return result;
		} catch (error) {
			// Rollback on failure
			if (browser && originalCard) {
				originalCard.is_favorite = originalIsFavorite;
				await offlineDB.saveCard(originalCard);
			}
			throw error;
		}
	},

	// Share operations (delegated to shared utility)
	...shareOps
};
