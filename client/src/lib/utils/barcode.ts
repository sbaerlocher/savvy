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

/** Validate barcode value against expected format constraints */
export function validateBarcodeFormat(
	barcode: string,
	format: string
): { valid: boolean; warning?: string } {
	if (format.includes('ean_13') && !/^\d{13}$/.test(barcode)) {
		return { valid: false, warning: 'EAN-13 should have 13 digits' };
	}
	if (format.includes('ean_8') && !/^\d{8}$/.test(barcode)) {
		return { valid: false, warning: 'EAN-8 should have 8 digits' };
	}
	if (format.includes('upc_a') && !/^\d{12}$/.test(barcode)) {
		return { valid: false, warning: 'UPC-A should have 12 digits' };
	}
	return { valid: true };
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
		/[{}\[\]]/.test(cleaned)
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
