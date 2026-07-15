import { describe, it, expect } from 'vitest';
import { formatUserName } from './user';

describe('formatUserName', () => {
	it('joins first and last name when both present', () => {
		expect(formatUserName({ first_name: 'Ada', last_name: 'Lovelace' })).toBe(
			'Ada Lovelace'
		);
	});

	it('returns first name only when last name missing', () => {
		expect(formatUserName({ first_name: 'Ada' })).toBe('Ada');
	});

	it('falls back to email when no first name', () => {
		expect(
			formatUserName({ last_name: 'Lovelace', email: 'ada@example.com' })
		).toBe('ada@example.com');
	});

	it('falls back to email when only email is present', () => {
		expect(formatUserName({ email: 'ada@example.com' })).toBe(
			'ada@example.com'
		);
	});

	it('returns fallback for an empty user object', () => {
		expect(formatUserName({}, 'Unknown')).toBe('Unknown');
	});

	it('returns fallback for null user', () => {
		expect(formatUserName(null, 'Unknown')).toBe('Unknown');
	});

	it('returns fallback for undefined user', () => {
		expect(formatUserName(undefined, 'Unknown')).toBe('Unknown');
	});

	it('defaults fallback to empty string', () => {
		expect(formatUserName(null)).toBe('');
		expect(formatUserName({})).toBe('');
	});

	it('treats empty-string first/last as absent and uses email', () => {
		expect(
			formatUserName({
				first_name: '',
				last_name: '',
				email: 'ada@example.com'
			})
		).toBe('ada@example.com');
	});

	it('uses first name when last name is empty string', () => {
		expect(formatUserName({ first_name: 'Ada', last_name: '' })).toBe('Ada');
	});

	it('ignores last name when first name is empty string', () => {
		expect(
			formatUserName({
				first_name: '',
				last_name: 'Lovelace',
				email: 'ada@example.com'
			})
		).toBe('ada@example.com');
	});

	it('handles null fields gracefully', () => {
		expect(
			formatUserName({
				first_name: null,
				last_name: null,
				email: null
			})
		).toBe('');
	});
});
