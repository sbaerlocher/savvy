/**
 * Preload and cache all important data for offline use
 */

import { cardsApi } from '$lib/api/cards';
import { vouchersApi } from '$lib/api/vouchers';
import { giftCardsApi } from '$lib/api/gift-cards';
import { dashboardApi } from '$lib/api/dashboard';
import { merchantsApi } from '$lib/api/merchants';
import { logger } from '$lib/utils/logger';

// Create child logger for preload
const preloadLogger = logger.child('Preload');

/** Max concurrent detail requests to stay within rate limits */
const PRELOAD_CONCURRENCY = 3;

/** Delay between batches in ms to avoid hitting rate limits */
const PRELOAD_BATCH_DELAY_MS = 1000;

/** Helper to sleep for a given number of milliseconds */
const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

/**
 * Preload all data in the background
 * This fills the cache for offline use.
 * List requests run sequentially in two phases to avoid exceeding rate limits.
 */
export async function preloadAllData() {
	preloadLogger.info('Preloading data for offline use...');

	// Phase 1: Load list views sequentially to stay within burst limit
	const listApis = [
		{ name: 'cards', fn: () => cardsApi.list() },
		{ name: 'vouchers', fn: () => vouchersApi.list() },
		{ name: 'dashboard', fn: () => dashboardApi.get() },
		{ name: 'merchants', fn: () => merchantsApi.list() }
	];

	for (const api of listApis) {
		try {
			await api.fn();
		} catch (e) {
			preloadLogger.warn(`Failed to preload ${api.name}:`, e);
		}
	}

	// Phase 2: Load gift cards list, then throttled detail requests
	try {
		const result = await giftCardsApi.list();
		await preloadGiftCardDetails(result.gift_cards);
	} catch (e) {
		preloadLogger.warn('Failed to preload gift cards:', e);
	}

	preloadLogger.info('Preload completed!');
}

/**
 * Preload gift card details (including transactions) for offline use.
 * The list endpoint doesn't include transactions, so we fetch each detail.
 * Uses concurrency limit and delays between batches to avoid 429 rate limit errors.
 */
async function preloadGiftCardDetails(giftCards: { id: string }[]) {
	if (giftCards.length === 0) return;

	preloadLogger.info(
		`Preloading ${giftCards.length} gift card details (with transactions)...`
	);

	// Process in batches with delay between them to respect rate limits
	for (let i = 0; i < giftCards.length; i += PRELOAD_CONCURRENCY) {
		const batch = giftCards.slice(i, i + PRELOAD_CONCURRENCY);
		const results = await Promise.allSettled(
			batch.map((gc) =>
				giftCardsApi
					.get(gc.id)
					.catch((e) =>
						preloadLogger.warn(`Failed to preload gift card ${gc.id}:`, e)
					)
			)
		);

		// Stop preloading if we hit rate limits
		const hitRateLimit = results.some(
			(r) => r.status === 'rejected' && r.reason?.status === 429
		);
		if (hitRateLimit) {
			preloadLogger.warn(
				'Rate limit hit during preload, stopping gift card detail preload'
			);
			break;
		}

		// Delay between batches to stay within rate limits
		if (i + PRELOAD_CONCURRENCY < giftCards.length) {
			await sleep(PRELOAD_BATCH_DELAY_MS);
		}
	}
}

/**
 * Full preload: loads all API data for offline use
 *
 * NOTE: Static assets (JS/CSS/HTML) are precached by the Service Worker on install.
 */
export async function preloadEverything() {
	preloadLogger.info('Full preload started...');
	await preloadAllData();
	preloadLogger.info('Full preload completed!');
}
