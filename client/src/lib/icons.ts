/**
 * Heroicons-style 24x24 stroke/fill `d` path data reused across components.
 *
 * These are path strings only — the surrounding `<svg>` element (viewBox,
 * stroke/fill, size, classes, aria-*) stays inline at each call site. Only the
 * `d="..."` literal is replaced with the matching constant.
 */

/** X / close mark. */
export const ICON_CLOSE = 'M6 18L18 6M6 6l12 12';

/** Exclamation triangle (warning). */
export const ICON_WARNING =
	'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z';

/** Horizontal transfer arrows. */
export const ICON_TRANSFER = 'M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4';

/** Magnifying glass (search). */
export const ICON_SEARCH = 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z';

/** Clipboard with check mark. */
export const ICON_CLIPBOARD_CHECK =
	'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4';

/** Barcode bars (wallet barcode toggle). */
export const ICON_BARCODE_TOGGLE =
	'M4 5h1v14H4V5zm3 0h1v14H7V5zm3 0h2v14h-2V5zm4 0h1v14h-1V5zm3 0h2v14h-2V5z';

/** Share nodes. */
export const ICON_SHARE =
	'M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z';

/** Spinner arc (paired with animate-spin on the svg). */
export const ICON_SPINNER =
	'M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z';

/** Closed padlock. */
export const ICON_LOCK =
	'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z';

/** Bare check mark (no circle) — M3 selected-chip leading icon. */
export const ICON_CHECK = 'M5 12.5l4.5 4.5L19 7';

/** Three stacked lines — M3 "select all" / overflow glyph. */
export const ICON_LINES = 'M4 6h16M4 12h16M4 18h16';

/** Three tapered lines — M3 filter glyph (Android merchants mockup). Distinct
 *  from ICON_FUNNEL, which the non-Android chrome uses. */
export const ICON_FILTER_LINES = 'M4 5h16M7 12h10M10 19h4';

/** Pencil over a baseline (edit). */
export const ICON_PENCIL = 'M12 20h9M16.5 3.5a2.1 2.1 0 013 3L7 19l-4 1 1-4z';

/** Download arrow into a tray (export). */
export const ICON_EXPORT =
	'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1M12 16V4m0 12l-4-4m4 4l4-4';

/** Check mark inside a circle. */
export const ICON_CHECK_CIRCLE =
	'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z';

/** Camera body (paired with ICON_CAMERA_LENS). */
export const ICON_CAMERA =
	'M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z';

/** Camera lens circle (paired with ICON_CAMERA). */
export const ICON_CAMERA_LENS = 'M15 13a3 3 0 11-6 0 3 3 0 016 0z';

/** Funnel (filter). */
export const ICON_FUNNEL =
	'M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z';

/** Trash can (delete). */
export const ICON_TRASH =
	'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16';

/** Bell (notifications). */
export const ICON_BELL =
	'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9';

/** Info "i" inside a circle. */
export const ICON_INFO_CIRCLE =
	'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z';

/** Storefront (merchant empty states). */
export const ICON_STOREFRONT =
	'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';

/** Chevron pointing left (back navigation). */
export const ICON_CHEVRON_LEFT = 'M15 19l-7-7 7-7';

/** Chevron pointing right (grouped-inset row disclosure). */
export const ICON_CHEVRON_RIGHT = 'M1 1l5.5 6L1 13';

/** Chevron pointing down (accordion disclosure). */
export const ICON_CHEVRON_DOWN = 'M19 9l-7 7-7-7';

/** Shield (admin mode). */
export const ICON_SHIELD = 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z';

/** Group of users (admin user management). */
export const ICON_USERS =
	'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z';

/** Shopping cart (merchant management). */
export const ICON_CART =
	'M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z';

/** Clipboard document (audit log). */
export const ICON_DOCUMENT =
	'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01';

/** Bar chart (system health). */
export const ICON_CHART =
	'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z';

/** Plus (create). */
export const ICON_PLUS = 'M12 5v14M5 12h14';

/** Circular refresh arrow (health re-check). */
export const ICON_REFRESH_CIRCLE = 'M21 12a9 9 0 11-3.2-6.9M21 3v5h-5';

/** Circular arrows (auto-refresh indicator). */
export const ICON_REFRESH =
	'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15';

/** X mark inside a circle (failed state). */
export const ICON_X_CIRCLE =
	'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z';
