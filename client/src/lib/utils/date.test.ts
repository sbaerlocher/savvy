import { describe, it, expect } from 'vitest';
import { formatDisplayDate, todayInputValue } from './date';

describe('formatDisplayDate', () => {
	// Assert equality against the same primitive the helper wraps so the test is
	// robust to whatever ICU data the runtime ships.
	it('matches toLocaleDateString of the date-only part', () => {
		const iso = '2026-03-15';
		expect(formatDisplayDate(iso, 'de-CH')).toBe(
			new Date('2026-03-15').toLocaleDateString('de-CH')
		);
	});

	it('strips the time component before formatting', () => {
		const withTime = '2026-03-15T23:59:59Z';
		expect(formatDisplayDate(withTime, 'en-US')).toBe(
			formatDisplayDate('2026-03-15', 'en-US')
		);
	});

	it('passes the locale through verbatim (no remapping)', () => {
		const iso = '2026-03-15';
		// Raw i18n code, as BarcodeDisplay/GiftCardLedger pass it.
		expect(formatDisplayDate(iso, 'de')).toBe(
			new Date('2026-03-15').toLocaleDateString('de')
		);
	});
});

describe('todayInputValue', () => {
	it('returns a YYYY-MM-DD string', () => {
		expect(todayInputValue()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
	});

	it('equals the date-only part of the current ISO timestamp', () => {
		expect(todayInputValue()).toBe(new Date().toISOString().split('T')[0]);
	});
});
