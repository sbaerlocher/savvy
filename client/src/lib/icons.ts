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
