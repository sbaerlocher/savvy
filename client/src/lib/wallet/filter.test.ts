import { describe, expect, it } from 'vitest';
import {
	applyCommonFilters,
	expiresWithinDays,
	matchesCardStatus,
	searchMerchant,
	sortItems
} from './filter';

const iso = (offsetDays: number): string =>
	new Date(Date.now() + offsetDays * 24 * 60 * 60 * 1000).toISOString();

describe('expiresWithinDays', () => {
	it('returns false for undefined', () => {
		expect(expiresWithinDays(undefined, 7)).toBe(false);
	});

	it('returns true inside the window', () => {
		expect(expiresWithinDays(iso(3), 7)).toBe(true);
	});

	it('returns false outside the window', () => {
		expect(expiresWithinDays(iso(10), 7)).toBe(false);
	});

	it('returns false for a past date (already expired)', () => {
		expect(expiresWithinDays(iso(-1), 7)).toBe(false);
	});

	it('returns false at the far boundary (exactly now = diff 0)', () => {
		// diff must be strictly > 0, so a date "now" is excluded.
		expect(expiresWithinDays(new Date().toISOString(), 7)).toBe(false);
	});
});

type Item = {
	id: string;
	owner?: { id?: string };
	is_favorite: boolean;
	expiry?: string;
};

const items: Item[] = [
	{ id: 'a', owner: { id: 'me' }, is_favorite: true, expiry: iso(3) },
	{ id: 'b', owner: { id: 'other' }, is_favorite: false, expiry: iso(20) },
	{ id: 'c', is_favorite: true, expiry: undefined },
	{ id: 'd', owner: { id: 'me' }, is_favorite: false, expiry: iso(40) }
];

const getExpiry = (i: Item) => i.expiry;

describe('applyCommonFilters', () => {
	it('passes everything through with default filters', () => {
		const out = applyCommonFilters(items, getExpiry, 'all', false, 'all', 'me');
		expect(out.map((i) => i.id)).toEqual(['a', 'b', 'c', 'd']);
	});

	it('filters to owner=mine (owned or unowned)', () => {
		const out = applyCommonFilters(
			items,
			getExpiry,
			'mine',
			false,
			'all',
			'me'
		);
		expect(out.map((i) => i.id)).toEqual(['a', 'c', 'd']);
	});

	it('filters to owner=shared', () => {
		const out = applyCommonFilters(
			items,
			getExpiry,
			'shared',
			false,
			'all',
			'me'
		);
		expect(out.map((i) => i.id)).toEqual(['b']);
	});

	it('filters to favorites only', () => {
		const out = applyCommonFilters(items, getExpiry, 'all', true, 'all', 'me');
		expect(out.map((i) => i.id)).toEqual(['a', 'c']);
	});

	it('filters expiring within 7 days', () => {
		const out = applyCommonFilters(items, getExpiry, 'all', false, '7', 'me');
		expect(out.map((i) => i.id)).toEqual(['a']);
	});

	it('filters expiring within 30 days', () => {
		const out = applyCommonFilters(items, getExpiry, 'all', false, '30', 'me');
		expect(out.map((i) => i.id)).toEqual(['a', 'b']);
	});

	it('combines owner + favorites', () => {
		const out = applyCommonFilters(items, getExpiry, 'mine', true, 'all', 'me');
		expect(out.map((i) => i.id)).toEqual(['a', 'c']);
	});
});

type SortItem = { id: string; created: string; value: number; expiry?: string };

const sortFixtures: SortItem[] = [
	{ id: 'x', created: iso(-3), value: 30, expiry: iso(10) },
	{ id: 'y', created: iso(-1), value: 10, expiry: undefined },
	{ id: 'z', created: iso(-2), value: 20, expiry: iso(5) }
];

const gd = (i: SortItem) => i.created;
const gv = (i: SortItem) => i.value;
const ge = (i: SortItem) => i.expiry;

describe('sortItems', () => {
	it('sorts newest first', () => {
		const out = sortItems(sortFixtures, gd, gv, ge, 'newest');
		expect(out.map((i) => i.id)).toEqual(['y', 'z', 'x']);
	});

	it('sorts oldest first', () => {
		const out = sortItems(sortFixtures, gd, gv, ge, 'oldest');
		expect(out.map((i) => i.id)).toEqual(['x', 'z', 'y']);
	});

	it('sorts value descending', () => {
		const out = sortItems(sortFixtures, gd, gv, ge, 'value-desc');
		expect(out.map((i) => i.id)).toEqual(['x', 'z', 'y']);
	});

	it('sorts value ascending', () => {
		const out = sortItems(sortFixtures, gd, gv, ge, 'value-asc');
		expect(out.map((i) => i.id)).toEqual(['y', 'z', 'x']);
	});

	it('sorts by expiry ascending with undefined pushed last', () => {
		const out = sortItems(sortFixtures, gd, gv, ge, 'expiry-asc');
		expect(out.map((i) => i.id)).toEqual(['z', 'x', 'y']);
	});

	it('leaves order unchanged for an unknown sort key', () => {
		const out = sortItems(sortFixtures, gd, gv, ge, 'unknown');
		expect(out.map((i) => i.id)).toEqual(['x', 'y', 'z']);
	});

	it('does not mutate the input array', () => {
		const input = [...sortFixtures];
		sortItems(input, gd, gv, ge, 'newest');
		expect(input.map((i) => i.id)).toEqual(['x', 'y', 'z']);
	});
});

describe('matchesCardStatus', () => {
	it("'active' matches only exactly active", () => {
		expect(matchesCardStatus('active', 'active')).toBe(true);
		expect(matchesCardStatus('inactive', 'active')).toBe(false);
		// Regression: expired/lost/blocked cards leaked into the default view
		// because the filter only excluded 'inactive'.
		expect(matchesCardStatus('expired', 'active')).toBe(false);
		expect(matchesCardStatus('lost', 'active')).toBe(false);
		expect(matchesCardStatus('blocked', 'active')).toBe(false);
	});

	it("'inactive' groups every non-active status", () => {
		expect(matchesCardStatus('inactive', 'inactive')).toBe(true);
		expect(matchesCardStatus('expired', 'inactive')).toBe(true);
		expect(matchesCardStatus('lost', 'inactive')).toBe(true);
		expect(matchesCardStatus('active', 'inactive')).toBe(false);
	});

	it("'all' matches everything", () => {
		expect(matchesCardStatus('active', 'all')).toBe(true);
		expect(matchesCardStatus('expired', 'all')).toBe(true);
		expect(matchesCardStatus(undefined, 'all')).toBe(true);
	});
});

describe('searchMerchant', () => {
	it('matches a substring case-insensitively', () => {
		expect(searchMerchant('Migros', 'gro', true)).toBe(true);
	});

	it('does not match when substring is absent', () => {
		expect(searchMerchant('Migros', 'coop', true)).toBe(false);
	});

	it('returns false for an undefined name', () => {
		expect(searchMerchant(undefined, 'gro', true)).toBe(false);
	});

	it('returns false when merchant matching is disabled', () => {
		expect(searchMerchant('Migros', 'gro', false)).toBe(false);
	});
});
