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
