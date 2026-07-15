/**
 * Fallback color used when a merchant has no color set.
 * Literal hex mirror of the --color-merchant-default token, because
 * <input type="color">, color-mix() interpolation and canvas/data layers
 * need a real hex value, not a CSS var().
 * TODO: keep in sync with --color-merchant-default in tokens.css.
 */
export const MERCHANT_DEFAULT_COLOR = '#8a8378';
