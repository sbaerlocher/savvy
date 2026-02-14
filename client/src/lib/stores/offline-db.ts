/**
 * IndexedDB Offline Store for Savvy PWA
 *
 * Provides local storage for Cards, Vouchers, Gift Cards, and Dashboard.
 * Single source of truth for offline data — no TTL, no pending operations.
 *
 * Stores:
 * - cards: CardDTO objects
 * - vouchers: VoucherDTO objects
 * - gift_cards: GiftCardDTO objects
 * - dashboard: Cached DashboardResponse (single entry, key: "main")
 */

import type {
	CardDTO,
	VoucherDTO,
	GiftCardDTO,
	MerchantDTO,
	DashboardResponse
} from '$lib/types/api';

const DB_NAME = 'savvy-offline';
const DB_VERSION = 7; // v7: added merchants store

/**
 * Type-safe store names
 */
export type StoreName =
	| 'cards'
	| 'vouchers'
	| 'gift_cards'
	| 'dashboard'
	| 'merchants';

/**
 * Custom error classes
 */
export class OfflineDBError extends Error {
	constructor(
		message: string,
		public readonly cause?: unknown
	) {
		super(message);
		this.name = 'OfflineDBError';
	}
}

export class DBNotInitializedError extends OfflineDBError {
	constructor() {
		super('Database not initialized');
		this.name = 'DBNotInitializedError';
	}
}

export class DBTransactionError extends OfflineDBError {
	constructor(operation: string, cause?: unknown) {
		super(`Transaction failed during ${operation}`, cause);
		this.name = 'DBTransactionError';
	}
}

export class OfflineDB {
	private db: IDBDatabase | null = null;
	private initPromise: Promise<void> | null = null;

	/**
	 * Ensure database is initialized (prevents race conditions)
	 */
	private async ensureDB(): Promise<void> {
		if (this.db) return;
		if (this.initPromise) return this.initPromise;
		this.initPromise = this.init();
		await this.initPromise;
	}

	/**
	 * Initialize the database
	 */
	private async init(): Promise<void> {
		return new Promise((resolve, reject) => {
			const request = indexedDB.open(DB_NAME, DB_VERSION);

			request.onerror = () => reject(request.error);
			request.onsuccess = () => {
				this.db = request.result;
				resolve();
			};

			request.onupgradeneeded = (event) => {
				const db = (event.target as IDBOpenDBRequest).result;
				const oldVersion = event.oldVersion;

				console.log(
					`[OfflineDB] Upgrading database from v${oldVersion} to v${DB_VERSION}`
				);

				// Migration v0 → v1: Initial schema
				if (oldVersion < 1) {
					if (!db.objectStoreNames.contains('cards')) {
						db.createObjectStore('cards', { keyPath: 'id' });
					}
					if (!db.objectStoreNames.contains('vouchers')) {
						db.createObjectStore('vouchers', { keyPath: 'id' });
					}
					if (!db.objectStoreNames.contains('gift_cards')) {
						db.createObjectStore('gift_cards', { keyPath: 'id' });
					}
					if (!db.objectStoreNames.contains('pending_operations')) {
						const pendingStore = db.createObjectStore('pending_operations', {
							keyPath: 'id'
						});
						pendingStore.createIndex('timestamp', 'timestamp', {
							unique: false
						});
					}
				}

				// Migration v3 → v4: Remove unused stores
				if (oldVersion >= 1 && oldVersion < 4) {
					if (db.objectStoreNames.contains('pending_operations')) {
						db.deleteObjectStore('pending_operations');
					}
					if (db.objectStoreNames.contains('cache_metadata')) {
						db.deleteObjectStore('cache_metadata');
					}
					console.log(
						'[OfflineDB] Migration v3→v4: removed pending_operations and cache_metadata stores'
					);
				}

				// Migration v4 → v5: Add dashboard store
				if (oldVersion < 5) {
					if (!db.objectStoreNames.contains('dashboard')) {
						db.createObjectStore('dashboard', { keyPath: 'key' });
					}
					console.log('[OfflineDB] Migration v4→v5: added dashboard store');
				}

				// Migration v6 → v7: Add merchants store
				if (oldVersion < 7) {
					if (!db.objectStoreNames.contains('merchants')) {
						db.createObjectStore('merchants', { keyPath: 'id' });
					}
					console.log('[OfflineDB] Migration v6→v7: added merchants store');
				}
			};
		});
	}

	// ==========================================================================
	// Generic CRUD methods
	// ==========================================================================

	private async get<T>(storeName: StoreName, key: string): Promise<T | null> {
		await this.ensureDB();
		if (!this.db) throw new DBNotInitializedError();

		return new Promise((resolve, reject) => {
			const transaction = this.db!.transaction(storeName, 'readonly');
			const store = transaction.objectStore(storeName);
			const request = store.get(key);

			request.onsuccess = () => resolve(request.result || null);
			request.onerror = () =>
				reject(new DBTransactionError(`get ${storeName}`, request.error));
		});
	}

	private async getAll<T>(storeName: StoreName): Promise<T[]> {
		await this.ensureDB();
		if (!this.db) throw new DBNotInitializedError();

		return new Promise((resolve, reject) => {
			const transaction = this.db!.transaction(storeName, 'readonly');
			const store = transaction.objectStore(storeName);
			const request = store.getAll();

			request.onsuccess = () => resolve(request.result || []);
			request.onerror = () =>
				reject(new DBTransactionError(`getAll ${storeName}`, request.error));
		});
	}

	/**
	 * NOTE: Waits for transaction.oncomplete to ensure data is fully committed.
	 */
	private async put<T>(storeName: StoreName, value: T): Promise<void> {
		await this.ensureDB();
		if (!this.db) throw new DBNotInitializedError();

		return new Promise((resolve, reject) => {
			const transaction = this.db!.transaction(storeName, 'readwrite');
			const store = transaction.objectStore(storeName);
			const request = store.put(value);

			transaction.oncomplete = () => resolve();
			transaction.onerror = () =>
				reject(new DBTransactionError(`put ${storeName}`, transaction.error));
			request.onerror = () =>
				reject(new DBTransactionError(`put ${storeName}`, request.error));
		});
	}

	/**
	 * NOTE: Waits for transaction.oncomplete to ensure deletion is committed.
	 */
	private async deleteItem(storeName: StoreName, key: string): Promise<void> {
		await this.ensureDB();
		if (!this.db) throw new DBNotInitializedError();

		return new Promise((resolve, reject) => {
			const transaction = this.db!.transaction(storeName, 'readwrite');
			const store = transaction.objectStore(storeName);
			const request = store.delete(key);

			transaction.oncomplete = () => resolve();
			transaction.onerror = () =>
				reject(
					new DBTransactionError(`delete ${storeName}`, transaction.error)
				);
			request.onerror = () =>
				reject(new DBTransactionError(`delete ${storeName}`, request.error));
		});
	}

	private async clear(storeName: StoreName): Promise<void> {
		await this.ensureDB();
		if (!this.db) throw new DBNotInitializedError();

		return new Promise((resolve, reject) => {
			const transaction = this.db!.transaction(storeName, 'readwrite');
			const store = transaction.objectStore(storeName);
			const request = store.clear();

			transaction.oncomplete = () => resolve();
			transaction.onerror = () =>
				reject(new DBTransactionError(`clear ${storeName}`, transaction.error));
			request.onerror = () =>
				reject(new DBTransactionError(`clear ${storeName}`, request.error));
		});
	}

	/**
	 * Bulk-save items in a single transaction (~10x faster than individual saves)
	 */
	private async bulkSave<T extends { id: string }>(
		storeName: StoreName,
		items: T[]
	): Promise<void> {
		await this.ensureDB();
		if (!this.db) throw new DBNotInitializedError();

		await new Promise<void>((resolve, reject) => {
			const transaction = this.db!.transaction(storeName, 'readwrite');
			const store = transaction.objectStore(storeName);

			for (const item of items) {
				store.put(item);
			}

			transaction.oncomplete = () => resolve();
			transaction.onerror = () =>
				reject(
					new DBTransactionError(`bulkSave ${storeName}`, transaction.error)
				);
		});
	}

	/**
	 * Replace all items: clears store then bulk-saves in a single transaction.
	 * Ensures IndexedDB matches the server state exactly (removes stale items).
	 *
	 * SAFETY: Validates input and aborts transaction on error to prevent
	 * partial commits (store cleared but no items written).
	 */
	private async replaceAll<T extends { id: string }>(
		storeName: StoreName,
		items: T[]
	): Promise<void> {
		if (!Array.isArray(items)) {
			console.error(
				`[OfflineDB] replaceAll: items is not an array for ${storeName}, skipping to preserve cache`
			);
			return;
		}

		await this.ensureDB();
		if (!this.db) throw new DBNotInitializedError();

		await new Promise<void>((resolve, reject) => {
			const transaction = this.db!.transaction(storeName, 'readwrite');
			const store = transaction.objectStore(storeName);

			// Set up handlers BEFORE operations to ensure they fire even if the loop throws
			transaction.oncomplete = () => resolve();
			transaction.onerror = () =>
				reject(
					new DBTransactionError(`replaceAll ${storeName}`, transaction.error)
				);
			transaction.onabort = () =>
				reject(
					new DBTransactionError(
						`replaceAll ${storeName} aborted`,
						transaction.error
					)
				);

			try {
				store.clear();
				for (const item of items) {
					store.put(item);
				}
			} catch (error) {
				// Abort transaction to roll back the clear() and prevent data loss
				try {
					transaction.abort();
				} catch {
					/* already aborted */
				}
			}
		});
	}

	// ==========================================================================
	// Cards
	// ==========================================================================

	async getCard(id: string): Promise<CardDTO | null> {
		return this.get<CardDTO>('cards', id);
	}

	async getAllCards(): Promise<CardDTO[]> {
		return this.getAll<CardDTO>('cards');
	}

	async saveCard(card: CardDTO): Promise<void> {
		await this.put('cards', card);
	}

	async saveManyCards(cards: CardDTO[]): Promise<void> {
		return this.bulkSave('cards', cards);
	}

	async replaceAllCards(cards: CardDTO[]): Promise<void> {
		return this.replaceAll('cards', cards);
	}

	async deleteCard(id: string): Promise<void> {
		await this.deleteItem('cards', id);
	}

	async clearCards(): Promise<void> {
		return this.clear('cards');
	}

	// ==========================================================================
	// Vouchers
	// ==========================================================================

	async getVoucher(id: string): Promise<VoucherDTO | null> {
		return this.get<VoucherDTO>('vouchers', id);
	}

	async getAllVouchers(): Promise<VoucherDTO[]> {
		return this.getAll<VoucherDTO>('vouchers');
	}

	async saveVoucher(voucher: VoucherDTO): Promise<void> {
		await this.put('vouchers', voucher);
	}

	async saveManyVouchers(vouchers: VoucherDTO[]): Promise<void> {
		return this.bulkSave('vouchers', vouchers);
	}

	async replaceAllVouchers(vouchers: VoucherDTO[]): Promise<void> {
		return this.replaceAll('vouchers', vouchers);
	}

	async deleteVoucher(id: string): Promise<void> {
		await this.deleteItem('vouchers', id);
	}

	async clearVouchers(): Promise<void> {
		return this.clear('vouchers');
	}

	// ==========================================================================
	// Gift Cards
	// ==========================================================================

	async getGiftCard(id: string): Promise<GiftCardDTO | null> {
		return this.get<GiftCardDTO>('gift_cards', id);
	}

	async getAllGiftCards(): Promise<GiftCardDTO[]> {
		return this.getAll<GiftCardDTO>('gift_cards');
	}

	async saveGiftCard(giftCard: GiftCardDTO): Promise<void> {
		await this.put('gift_cards', giftCard);
	}

	async saveManyGiftCards(giftCards: GiftCardDTO[]): Promise<void> {
		return this.bulkSave('gift_cards', giftCards);
	}

	async replaceAllGiftCards(giftCards: GiftCardDTO[]): Promise<void> {
		return this.replaceAll('gift_cards', giftCards);
	}

	async deleteGiftCard(id: string): Promise<void> {
		await this.deleteItem('gift_cards', id);
	}

	async clearGiftCards(): Promise<void> {
		return this.clear('gift_cards');
	}

	// ==========================================================================
	// Dashboard
	// ==========================================================================

	async getDashboard(): Promise<DashboardResponse | null> {
		const result = await this.get<{ key: string; data: DashboardResponse }>(
			'dashboard',
			'main'
		);
		return result?.data ?? null;
	}

	async saveDashboard(data: DashboardResponse): Promise<void> {
		await this.put('dashboard', { key: 'main', data });
	}

	async clearDashboard(): Promise<void> {
		return this.clear('dashboard');
	}

	// ==========================================================================
	// Merchants
	// ==========================================================================

	async getMerchant(id: string): Promise<MerchantDTO | null> {
		return this.get<MerchantDTO>('merchants', id);
	}

	async getAllMerchants(): Promise<MerchantDTO[]> {
		return this.getAll<MerchantDTO>('merchants');
	}

	async saveMerchant(merchant: MerchantDTO): Promise<void> {
		await this.put('merchants', merchant);
	}

	async saveManyMerchants(merchants: MerchantDTO[]): Promise<void> {
		return this.bulkSave('merchants', merchants);
	}

	async replaceAllMerchants(merchants: MerchantDTO[]): Promise<void> {
		return this.replaceAll('merchants', merchants);
	}

	async deleteMerchant(id: string): Promise<void> {
		await this.deleteItem('merchants', id);
	}

	async clearMerchants(): Promise<void> {
		return this.clear('merchants');
	}

	// ==========================================================================
	// Utility
	// ==========================================================================

	async clearAll(): Promise<void> {
		await this.clearCards();
		await this.clearVouchers();
		await this.clearGiftCards();
		await this.clearDashboard();
		await this.clearMerchants();
	}

	/**
	 * Reset database to clean state (for testing)
	 */
	async reset(): Promise<void> {
		await this.clearAll();
		this.initPromise = null;
	}

	/**
	 * Close the database connection
	 */
	close(): void {
		if (this.db) {
			this.db.close();
			this.db = null;
		}
		this.initPromise = null;
	}
}

/**
 * Factory function for testing with dependency injection
 */
export const createOfflineDB = (): OfflineDB => new OfflineDB();

/**
 * Default singleton instance for application use
 */
export const offlineDB = createOfflineDB();
