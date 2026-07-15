import { describe, it, expect } from 'vitest';
import {
	checkSymbologySuitability,
	barcodeBcid,
	BARCODE_BCID_MAP,
	is2DBcid,
	is2DType,
	sanitizeBarcodeValue,
	isValidBarcodeLength,
	cameraErrorKey
} from './barcode';

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

describe('barcodeBcid / BARCODE_BCID_MAP', () => {
	it('maps known types (case-insensitive) to their BCID', () => {
		expect(barcodeBcid('CODE128')).toBe('code128');
		expect(barcodeBcid('qr')).toBe('qrcode');
		expect(barcodeBcid('ItF')).toBe('interleaved2of5');
		expect(barcodeBcid('ISBN13')).toBe('ean13');
	});

	it('falls back to code128 for unknown types', () => {
		expect(barcodeBcid('NOPE')).toBe('code128');
		expect(barcodeBcid('')).toBe('code128');
	});

	it('exposes the underlying map', () => {
		expect(BARCODE_BCID_MAP.AZTEC).toBe('azteccode');
	});
});

describe('is2DBcid', () => {
	it('classifies 2D BCIDs as true', () => {
		for (const bcid of [
			'qrcode',
			'pdf417',
			'datamatrix',
			'azteccode',
			'maxicode'
		]) {
			expect(is2DBcid(bcid)).toBe(true);
		}
	});

	it('classifies 1D / unknown BCIDs as false', () => {
		for (const bcid of ['code128', 'ean13', 'interleaved2of5', 'QRCODE', '']) {
			expect(is2DBcid(bcid)).toBe(false);
		}
	});
});

describe('is2DType', () => {
	it('classifies 2D types (case-insensitive) as true', () => {
		for (const type of [
			'QR',
			'qrcode',
			'PDF417',
			'DataMatrix',
			'AZTEC',
			'MAXICODE'
		]) {
			expect(is2DType(type)).toBe(true);
		}
	});

	it('classifies 1D / unknown types as false', () => {
		for (const type of ['CODE128', 'EAN13', 'ITF', 'azteccode', '']) {
			expect(is2DType(type)).toBe(false);
		}
	});
});

describe('sanitizeBarcodeValue', () => {
	it('strips C0/DEL/C1 control characters', () => {
		expect(sanitizeBarcodeValue('ab\x00c\x1Fd\x7Fe\x9Ff')).toBe('abcdef');
	});

	it('leaves printable content untouched', () => {
		expect(sanitizeBarcodeValue('Hello-123 $/+%')).toBe('Hello-123 $/+%');
	});
});

describe('isValidBarcodeLength', () => {
	it('accepts 1..255 chars', () => {
		expect(isValidBarcodeLength('a')).toBe(true);
		expect(isValidBarcodeLength('a'.repeat(255))).toBe(true);
	});

	it('rejects empty and >255 chars', () => {
		expect(isValidBarcodeLength('')).toBe(false);
		expect(isValidBarcodeLength('a'.repeat(256))).toBe(false);
	});
});

describe('cameraErrorKey', () => {
	it('maps each known error name to its i18n key', () => {
		expect(cameraErrorKey('HttpsRequiredError')).toBe(
			'common.scanHttpsRequired'
		);
		expect(cameraErrorKey('NotAllowedError')).toBe(
			'common.scanCameraPermissionDenied'
		);
		expect(cameraErrorKey('PermissionDeniedError')).toBe(
			'common.scanCameraPermissionDenied'
		);
		expect(cameraErrorKey('NotFoundError')).toBe('common.scanNoCameraFound');
		expect(cameraErrorKey('NotReadableError')).toBe(
			'common.scanCameraNotAvailable'
		);
		expect(cameraErrorKey('AbortError')).toBe('common.scanCameraNotAvailable');
		expect(cameraErrorKey('OverconstrainedError')).toBe(
			'common.scanCameraConstraintsError'
		);
		expect(cameraErrorKey('SecurityError')).toBe(
			'common.scanCameraSecurityBlocked'
		);
		expect(cameraErrorKey('NotSupportedError')).toBe(
			'common.scanCameraNotSupported'
		);
		expect(cameraErrorKey('TypeError')).toBe('common.scanCameraNotSupported');
		expect(cameraErrorKey('TimeoutError')).toBe('common.scanCameraTimeout');
	});

	it('falls back to scanNoCameraFound for unknown / empty names', () => {
		expect(cameraErrorKey('WeirdError')).toBe('common.scanNoCameraFound');
		expect(cameraErrorKey('')).toBe('common.scanNoCameraFound');
	});
});
