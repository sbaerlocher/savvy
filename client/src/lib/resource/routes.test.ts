import { describe, it, expect } from 'vitest';
import { resourceDetailPath } from './routes';

describe('resourceDetailPath', () => {
	it('maps card to the card detail route', () => {
		expect(resourceDetailPath('card', 'abc')).toBe('/cards/abc');
	});

	it('maps voucher to the voucher detail route', () => {
		expect(resourceDetailPath('voucher', 'abc')).toBe('/vouchers/abc');
	});

	it('maps gift_card to the gift-card detail route', () => {
		expect(resourceDetailPath('gift_card', 'abc')).toBe('/gift-cards/abc');
	});

	it('falls back to the dashboard for an unknown type', () => {
		expect(resourceDetailPath('mystery', 'abc')).toBe('/dashboard');
	});
});
