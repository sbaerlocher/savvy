/** Barcode formats supported by the scanner */
export const SCANNER_FORMATS = [
	'aztec',
	'code_128',
	'code_39',
	'code_93',
	'codabar',
	'data_matrix',
	'ean_13',
	'ean_8',
	'itf',
	'pdf417',
	'qr_code',
	'upc_a',
	'upc_e'
] as const;

const FORMAT_MAP: Record<string, string> = {
	QRCODE: 'QR',
	QR: 'QR',
	CODE128: 'CODE128',
	CODE39: 'CODE39',
	CODE93: 'CODE93',
	CODABAR: 'CODABAR',
	EAN8: 'EAN8',
	EAN13: 'EAN13',
	UPCA: 'UPCA',
	UPCE: 'UPCE',
	ITF: 'ITF',
	PDF417: 'PDF417',
	DATAMATRIX: 'DATAMATRIX',
	AZTEC: 'AZTEC'
};

/** Normalize detector format string to application format (e.g. 'qr_code' → 'QR') */
export function mapBarcodeFormat(format: string): string {
	if (!format) return 'CODE128';
	const normalized = format.replace(/[-_/\s]/g, '').toUpperCase();
	return FORMAT_MAP[normalized] ?? normalized;
}

/** Validate barcode value against expected format constraints.
 *  Returns an i18n key (e.g. 'common.scanFormatWarningEan13') as warning. */
export function validateBarcodeFormat(
	barcode: string,
	format: string
): { valid: boolean; warningKey?: string } {
	if (format.includes('ean_13') && !/^\d{13}$/.test(barcode)) {
		return { valid: false, warningKey: 'common.scanFormatWarningEan13' };
	}
	if (format.includes('ean_8') && !/^\d{8}$/.test(barcode)) {
		return { valid: false, warningKey: 'common.scanFormatWarningEan8' };
	}
	if (format.includes('upc_a') && !/^\d{12}$/.test(barcode)) {
		return { valid: false, warningKey: 'common.scanFormatWarningUpca' };
	}
	return { valid: true };
}

/** Content-vs-symbology suitability for the manual barcode-type picker in the
 *  forms. Returns an i18n warning key when the entered content cannot be
 *  encoded by the chosen symbology, undefined when it fits (or the symbology
 *  accepts arbitrary content, e.g. CODE128/QR). Keyed by the form's type
 *  values (CODE128, EAN13, …), NOT the scanner's snake_case formats. */
const SYMBOLOGY_RULES: Record<string, { test: RegExp; warningKey: string }> = {
	// Fixed-length numeric symbologies.
	EAN13: { test: /^\d{13}$/, warningKey: 'common.symbologyWarningEan13' },
	EAN8: { test: /^\d{8}$/, warningKey: 'common.symbologyWarningEan8' },
	UPCA: { test: /^\d{12}$/, warningKey: 'common.symbologyWarningUpca' },
	UPCE: { test: /^\d{6,8}$/, warningKey: 'common.symbologyWarningUpce' },
	ITF14: { test: /^\d{14}$/, warningKey: 'common.symbologyWarningItf14' },
	ISBN13: {
		test: /^(978|979)\d{10}$/,
		warningKey: 'common.symbologyWarningIsbn13'
	},
	// ISBN-10: 9 digits + check digit (last may be X).
	ISBN10: {
		test: /^\d{9}[\dX]$/i,
		warningKey: 'common.symbologyWarningIsbn10'
	},
	// ISSN: 7 digits + check digit (last may be X), optional NNNN-NNNC hyphen.
	ISSN: {
		test: /^\d{4}-?\d{3}[\dX]$/i,
		warningKey: 'common.symbologyWarningIssn'
	},
	// Variable-length but constrained character sets.
	// ponytail: CODE93 encodes full ASCII via shift characters → effectively
	// arbitrary content like CODE128/QR, so it deliberately gets no rule.
	// ITF (Interleaved 2 of 5) encodes digit PAIRS → even count required.
	ITF: { test: /^(\d{2})+$/, warningKey: 'common.symbologyWarningItf' },
	// CODE39 standard charset: digits, uppercase, and - . space $ / + %
	CODE39: {
		test: /^[0-9A-Z\-. $/+%]*$/,
		warningKey: 'common.symbologyWarningCode39'
	},
	// CODABAR: digits + a few symbols, optional A-D start/stop.
	CODABAR: {
		test: /^[A-D]?[0-9\-$:/.+]*[A-D]?$/i,
		warningKey: 'common.symbologyWarningCodabar'
	}
};

/** Check whether `content` is encodable by the chosen `symbology` (a form
 *  barcode-type value). Empty content never warns (user hasn't typed yet).
 *  Unknown / arbitrary-content symbologies (CODE128, QR, PDF417, DATAMATRIX,
 *  AZTEC, …) never warn. Returns an i18n key or undefined. */
export function checkSymbologySuitability(
	content: string,
	symbology: string
): string | undefined {
	const trimmed = content.trim();
	if (!trimmed) return undefined;
	const rule = SYMBOLOGY_RULES[symbology];
	if (!rule) return undefined;
	return rule.test.test(trimmed) ? undefined : rule.warningKey;
}

/**
 * Validates an EAN/UPC check digit using the standard algorithm.
 * Works for EAN-13, EAN-8, UPC-A, and ITF-14.
 */
function isValidEanCheckDigit(digits: string): boolean {
	const nums = digits.split('').map(Number);
	const last = nums.pop()!;
	const sum = nums.reduce((acc, d, i) => acc + d * (i % 2 === 0 ? 1 : 3), 0);
	return (10 - (sum % 10)) % 10 === last;
}

/**
 * Detects the barcode type based on the scanned content
 * @param barcode - The scanned barcode string
 * @returns The detected barcode type
 */
export function detectBarcodeType(barcode: string): string {
	// Remove whitespace
	const cleaned = barcode.trim();

	// EAN-13: 13 digits with valid check digit
	if (/^\d{13}$/.test(cleaned) && isValidEanCheckDigit(cleaned)) {
		return 'EAN13';
	}

	// EAN-8: 8 digits with valid check digit
	if (/^\d{8}$/.test(cleaned) && isValidEanCheckDigit(cleaned)) {
		return 'EAN8';
	}

	// UPC-A: 12 digits with valid check digit
	if (/^\d{12}$/.test(cleaned) && isValidEanCheckDigit(cleaned)) {
		return 'UPCA';
	}

	// UPC-E: 6-8 digits (typically 6 or 8)
	if (/^\d{6}$/.test(cleaned)) {
		return 'UPCE';
	}

	// ITF-14: 14 digits with valid check digit
	if (/^\d{14}$/.test(cleaned) && isValidEanCheckDigit(cleaned)) {
		return 'ITF14';
	}

	// ISBN-13: 13 digits starting with 978 or 979 with valid check digit
	if (/^(978|979)\d{10}$/.test(cleaned) && isValidEanCheckDigit(cleaned)) {
		return 'ISBN13';
	}

	// QR Code: Usually longer and can contain special characters
	// If it contains URLs, special formatting, or is very long, likely QR
	if (
		cleaned.length > 50 ||
		cleaned.includes('http') ||
		cleaned.includes('\n') ||
		/[{}[\]]/.test(cleaned)
	) {
		return 'QR';
	}

	// CODE39: Alphanumeric, typically uppercase with allowed special chars
	if (/^[A-Z0-9\-. $/+%]+$/.test(cleaned)) {
		return 'CODE39';
	}

	// CODE93: Similar to CODE39 but more compact
	if (/^[A-Z0-9\-. $/+%*]+$/.test(cleaned) && cleaned.length < 20) {
		return 'CODE93';
	}

	// CODABAR: Digits with start/stop characters (A-D)
	if (/^[A-D]\d+[A-D]$/i.test(cleaned)) {
		return 'CODABAR';
	}

	// ITF (Interleaved 2 of 5): Even number of digits
	if (
		/^\d+$/.test(cleaned) &&
		cleaned.length % 2 === 0 &&
		cleaned.length < 14
	) {
		return 'ITF';
	}

	// Default to CODE128 for alphanumeric content
	// CODE128 is the most versatile and common for mixed content
	return 'CODE128';
}
