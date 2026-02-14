import type {
	CardDTO,
	VoucherDTO,
	GiftCardDTO,
	MerchantDTO
} from '$lib/types/api';

/** Aggregated merchant data derived client-side from cached items. */
export interface MerchantOverview {
	id: string;
	name: string;
	color: string;
	website?: string;
	cards_count: number;
	cards_inactive_count: number;
	vouchers_count: number;
	vouchers_inactive_count: number;
	gift_cards_count: number;
	gift_cards_inactive_count: number;
	total_gift_card_balance: number;
}

interface MerchantEntry {
	merchant: MerchantDTO;
	cards_count: number;
	cards_inactive_count: number;
	vouchers_count: number;
	vouchers_inactive_count: number;
	gift_cards_count: number;
	gift_cards_inactive_count: number;
	total_gift_card_balance: number;
}

/**
 * Derives MerchantOverview[] from all merchants and cached items.
 * Seeds the map with all merchants first (so merchants without items appear),
 * then groups items by merchant_id and computes aggregated counts and balances.
 */
export function deriveMerchantOverview(
	cards: CardDTO[],
	vouchers: VoucherDTO[],
	giftCards: GiftCardDTO[],
	allMerchants: MerchantDTO[] = []
): MerchantOverview[] {
	const merchantMap = new Map<string, MerchantEntry>();

	function newEntry(merchant: MerchantDTO): MerchantEntry {
		return {
			merchant,
			cards_count: 0,
			cards_inactive_count: 0,
			vouchers_count: 0,
			vouchers_inactive_count: 0,
			gift_cards_count: 0,
			gift_cards_inactive_count: 0,
			total_gift_card_balance: 0
		};
	}

	// Seed with all known merchants so they appear even without items
	for (const m of allMerchants) {
		merchantMap.set(m.id, newEntry(m));
	}

	function ensureMerchant(
		merchantId: string,
		merchantData?: MerchantDTO
	): MerchantEntry | undefined {
		if (!merchantMap.has(merchantId) && merchantData) {
			merchantMap.set(merchantId, newEntry(merchantData));
		}
		return merchantMap.get(merchantId);
	}

	for (const card of cards) {
		if (card.merchant_id && card.merchant) {
			const entry = ensureMerchant(card.merchant_id, card.merchant);
			if (entry) {
				if (card.status === 'active' || !card.status) {
					entry.cards_count++;
				} else {
					entry.cards_inactive_count++;
				}
			}
		}
	}

	for (const voucher of vouchers) {
		if (voucher.merchant_id && voucher.merchant) {
			const entry = ensureMerchant(voucher.merchant_id, voucher.merchant);
			if (entry) {
				if (voucher.status === 'valid') {
					entry.vouchers_count++;
				} else {
					entry.vouchers_inactive_count++;
				}
			}
		}
	}

	for (const giftCard of giftCards) {
		if (giftCard.merchant_id && giftCard.merchant) {
			const entry = ensureMerchant(giftCard.merchant_id, giftCard.merchant);
			if (entry) {
				const isActive =
					!giftCard.expires_at || new Date(giftCard.expires_at) >= new Date();
				if (isActive && (giftCard.current_balance ?? 0) > 0) {
					entry.gift_cards_count++;
				} else {
					entry.gift_cards_inactive_count++;
				}
				entry.total_gift_card_balance += giftCard.current_balance ?? 0;
			}
		}
	}

	return Array.from(merchantMap.values()).map(({ merchant, ...counts }) => ({
		id: merchant.id,
		name: merchant.name,
		color: merchant.color ?? '#3B82F6',
		website: merchant.website,
		...counts
	}));
}

/**
 * Extracts MerchantDTO from items sharing the same merchant_id.
 * Returns the merchant from the first item that has a non-null merchant field.
 */
export function extractMerchantFromItems(
	merchantId: string,
	cards: CardDTO[],
	vouchers: VoucherDTO[],
	giftCards: GiftCardDTO[]
): MerchantDTO | null {
	for (const card of cards) {
		if (card.merchant_id === merchantId && card.merchant) return card.merchant;
	}
	for (const voucher of vouchers) {
		if (voucher.merchant_id === merchantId && voucher.merchant)
			return voucher.merchant;
	}
	for (const giftCard of giftCards) {
		if (giftCard.merchant_id === merchantId && giftCard.merchant)
			return giftCard.merchant;
	}
	return null;
}
