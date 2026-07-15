/**
 * Formats the date-only part of an ISO string for display.
 *
 * Strips any time component (`split('T')[0]`) before constructing the Date so
 * the local timezone can't shift the calendar day, then delegates to
 * `toLocaleDateString`. The `locale` is passed through verbatim — callers decide
 * whether they hand in a raw i18n code (`'de'`/`'en'`/`'fr'`, with their own
 * fallback) or a mapped BCP-47 tag (`'de-CH'`). This helper never remaps it.
 */
export function formatDisplayDate(iso: string, locale: string): string {
	return new Date(iso.split('T')[0]).toLocaleDateString(locale);
}

/** Today as a `YYYY-MM-DD` string for `<input type="date">` values/max. */
export function todayInputValue(): string {
	return new Date().toISOString().split('T')[0];
}
