import { describe, it, expect } from 'vitest';
import { checkSymbologySuitability } from './barcode';

describe('checkSymbologySuitability', () => {
	it('returns undefined for empty/whitespace content (nothing typed yet)', () => {
		expect(checkSymbologySuitability('', 'EAN13')).toBeUndefined();
		expect(checkSymbologySuitability('   ', 'EAN13')).toBeUndefined();
	});

	it('never warns for arbitrary-content symbologies', () => {
		for (const sym of [
			'CODE128',
			'QR',
			'PDF417',
			'DATAMATRIX',
			'AZTEC',
			'CODE93'
		]) {
			expect(
				checkSymbologySuitability('anything-!@# goes', sym)
			).toBeUndefined();
		}
	});

	it('never warns for unknown / arbitrary-content symbology values', () => {
		expect(checkSymbologySuitability('abc', 'MAXICODE')).toBeUndefined();
		expect(checkSymbologySuitability('abc', 'PDF417')).toBeUndefined();
		expect(checkSymbologySuitability('abc', 'CODE93')).toBeUndefined();
	});

	it('accepts well-formed fixed-length numeric codes', () => {
		expect(checkSymbologySuitability('4006381333931', 'EAN13')).toBeUndefined();
		expect(checkSymbologySuitability('96385074', 'EAN8')).toBeUndefined();
		expect(checkSymbologySuitability('036000291452', 'UPCA')).toBeUndefined();
		expect(checkSymbologySuitability('1234567', 'UPCE')).toBeUndefined();
		expect(
			checkSymbologySuitability('12345678901231', 'ITF14')
		).toBeUndefined();
		expect(
			checkSymbologySuitability('9781234567897', 'ISBN13')
		).toBeUndefined();
		expect(checkSymbologySuitability('012345678X', 'ISBN10')).toBeUndefined();
		expect(checkSymbologySuitability('0378-5955', 'ISSN')).toBeUndefined();
		expect(checkSymbologySuitability('03785955', 'ISSN')).toBeUndefined();
	});

	it('warns on malformed ISBN-10 / ISSN', () => {
		expect(checkSymbologySuitability('12345', 'ISBN10')).toBe(
			'common.symbologyWarningIsbn10'
		);
		expect(checkSymbologySuitability('037859', 'ISSN')).toBe(
			'common.symbologyWarningIssn'
		);
	});

	it('warns on wrong length / non-digit for numeric symbologies', () => {
		expect(checkSymbologySuitability('12345', 'EAN13')).toBe(
			'common.symbologyWarningEan13'
		);
		expect(checkSymbologySuitability('ABCDEFGHIJKLM', 'EAN13')).toBe(
			'common.symbologyWarningEan13'
		);
		expect(checkSymbologySuitability('1234567', 'EAN8')).toBe(
			'common.symbologyWarningEan8'
		);
		expect(checkSymbologySuitability('12345', 'UPCA')).toBe(
			'common.symbologyWarningUpca'
		);
		expect(checkSymbologySuitability('12', 'UPCE')).toBe(
			'common.symbologyWarningUpce'
		);
		expect(checkSymbologySuitability('123', 'ITF14')).toBe(
			'common.symbologyWarningItf14'
		);
		expect(checkSymbologySuitability('1231234567897', 'ISBN13')).toBe(
			'common.symbologyWarningIsbn13'
		);
	});

	it('ITF requires an even number of digits', () => {
		expect(checkSymbologySuitability('123456', 'ITF')).toBeUndefined();
		expect(checkSymbologySuitability('12345', 'ITF')).toBe(
			'common.symbologyWarningItf'
		);
		expect(checkSymbologySuitability('12a4', 'ITF')).toBe(
			'common.symbologyWarningItf'
		);
	});

	it('CODE39 accepts its charset, warns on lowercase/illegal chars', () => {
		expect(
			checkSymbologySuitability('ABC-123 $/+%.', 'CODE39')
		).toBeUndefined();
		expect(checkSymbologySuitability('lowercase', 'CODE39')).toBe(
			'common.symbologyWarningCode39'
		);
		expect(checkSymbologySuitability('A@B', 'CODE39')).toBe(
			'common.symbologyWarningCode39'
		);
	});

	it('CODABAR accepts digits + optional A-D delimiters, warns on letters', () => {
		expect(checkSymbologySuitability('A1234B', 'CODABAR')).toBeUndefined();
		expect(checkSymbologySuitability('1234', 'CODABAR')).toBeUndefined();
		expect(checkSymbologySuitability('12XY34', 'CODABAR')).toBe(
			'common.symbologyWarningCodabar'
		);
	});

	it('trims surrounding whitespace before checking', () => {
		expect(
			checkSymbologySuitability('  4006381333931  ', 'EAN13')
		).toBeUndefined();
	});
});
