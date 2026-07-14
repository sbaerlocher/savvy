import { describe, it, expect } from 'vitest';
import {
	getGiftCardStatus,
	isGiftCardActive,
	isVoucherValid
} from './resource-status';
import type { GiftCardDTO } from '$lib/types/api';

const past = '2000-01-01T00:00:00Z';
const future = '2999-01-01T00:00:00Z';

function giftCard(overrides: Partial<GiftCardDTO>): GiftCardDTO {
	return {
		current_balance: 50,
		expires_at: future,
		...overrides
	} as GiftCardDTO;
}

describe('getGiftCardStatus', () => {
	it('is active when it has balance and is not expired', () => {
		expect(getGiftCardStatus(giftCard({}))).toBe('active');
	});

	it('is depleted when balance is zero', () => {
		expect(getGiftCardStatus(giftCard({ current_balance: 0 }))).toBe(
			'depleted'
		);
	});

	it('is expired when the expiry date is in the past', () => {
		expect(getGiftCardStatus(giftCard({ expires_at: past }))).toBe('expired');
	});

	it('prefers depleted over expired for a used-up, expired card', () => {
		expect(
			getGiftCardStatus(giftCard({ current_balance: 0, expires_at: past }))
		).toBe('depleted');
	});

	it('treats a card without expiry as active', () => {
		expect(getGiftCardStatus(giftCard({ expires_at: undefined }))).toBe(
			'active'
		);
	});
});

describe('isGiftCardActive', () => {
	it('hides depleted and expired gift cards from quick-access', () => {
		expect(isGiftCardActive(giftCard({}))).toBe(true);
		expect(isGiftCardActive(giftCard({ current_balance: 0 }))).toBe(false);
		expect(isGiftCardActive(giftCard({ expires_at: past }))).toBe(false);
	});
});

describe('isVoucherValid', () => {
	it('keeps only valid vouchers', () => {
		expect(isVoucherValid({ status: 'valid' })).toBe(true);
	});

	it('hides expired, inactive and used vouchers', () => {
		expect(isVoucherValid({ status: 'expired' })).toBe(false);
		expect(isVoucherValid({ status: 'inactive' })).toBe(false);
		expect(isVoucherValid({ status: 'used' })).toBe(false);
	});

	it('hides vouchers with no status', () => {
		expect(isVoucherValid({})).toBe(false);
	});
});
