import type { CardDTO, GiftCardDTO } from '$lib/types/api';

export type GiftCardStatus = 'active' | 'depleted' | 'expired';

/**
 * Computes the effective status of a gift card from its balance and expiry.
 * `depleted` takes precedence over `expired` (a used-up card is depleted
 * regardless of its expiry date).
 */
export function getGiftCardStatus(giftCard: GiftCardDTO): GiftCardStatus {
	if (giftCard.current_balance === 0) return 'depleted';
	if (giftCard.expires_at && new Date(giftCard.expires_at) < new Date())
		return 'expired';
	return 'active';
}

/**
 * Whether a gift card is currently usable (has balance and is not expired).
 * Used to decide if it should surface in quick-access views like the dashboard.
 */
export function isGiftCardActive(giftCard: GiftCardDTO): boolean {
	return getGiftCardStatus(giftCard) === 'active';
}

/**
 * Whether a card is currently usable. Cards carry a manual status
 * (active | inactive | expired | lost | blocked); only active cards should
 * surface in quick-access views like the dashboard.
 */
export function isCardActive(card: CardDTO): boolean {
	return card.status === 'active';
}

/**
 * Whether a voucher is currently valid. Vouchers carry a pre-computed status
 * (`valid` | `expired` | `inactive` | `used`) from the API mapper; only
 * `valid` vouchers should surface in quick-access views like the dashboard.
 */
export function isVoucherValid(voucher: { status?: string }): boolean {
	return voucher.status === 'valid';
}
