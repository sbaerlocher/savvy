/**
 * Currency formatting utilities using Intl.NumberFormat
 */

import type { Language } from '$lib/stores/i18n';

/**
 * Maps language codes to locale strings for Intl.NumberFormat
 */
const localeMap: Record<Language, string> = {
	de: 'de-CH', // German (Switzerland) - default for CHF
	en: 'en-US', // English (United States)
	fr: 'fr-FR' // French (France)
};

/**
 * Formats a currency value according to the user's language and the voucher's currency.
 *
 * @param value - The numeric value to format
 * @param currency - ISO 4217 currency code (CHF, EUR, USD, GBP)
 * @param language - User's language preference (de, en, fr)
 * @returns Formatted currency string (e.g., "CHF 50.00", "50,00 €", "$50.00")
 *
 * @example
 * formatCurrency(50, 'CHF', 'de') // "CHF 50.00"
 * formatCurrency(50, 'EUR', 'de') // "50,00 €"
 * formatCurrency(50, 'USD', 'en') // "$50.00"
 */
export function formatCurrency(
	value: number,
	currency: string,
	language: Language | string
): string {
	// Extract language code from locale string (e.g., 'en-US' → 'en')
	const langCode = language.split('-')[0] as Language;
	const locale = localeMap[langCode] || 'de-CH';

	try {
		return new Intl.NumberFormat(locale, {
			style: 'currency',
			currency: currency,
			currencyDisplay: 'symbol' // Use symbols ($, €, £) instead of codes (USD, EUR, GBP)
		}).format(value);
	} catch (error) {
		// Fallback if currency code is invalid
		console.warn(`Invalid currency code: ${currency}`, error);
		return `${currency} ${value.toFixed(2)}`;
	}
}
