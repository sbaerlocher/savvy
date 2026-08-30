import type { Page, TestInfo } from '@playwright/test';
import type { Platform, RouteSpec } from './routes';

/**
 * User agents per platform. `platform` in src/lib/utils/platform.ts is derived
 * from the UA at module load — NOT from a breakpoint — so a narrow viewport
 * alone still renders the desktop branch. Baselining the iOS and Android
 * branches requires these UAs.
 */
export const PLATFORM_UA: Record<Platform, string> = {
	ios: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1',
	android:
		'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36',
	desktop:
		'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
};

export const PLATFORM_VIEWPORT: Record<
	Platform,
	{ width: number; height: number }
> = {
	ios: { width: 390, height: 844 },
	android: { width: 412, height: 915 },
	desktop: { width: 1440, height: 900 }
};

/**
 * Freeze everything that would otherwise differ between two identical runs.
 * Without this the screenshot diffs are noise and prove nothing.
 */
export async function stabilise(page: Page): Promise<void> {
	// Client-side redirects (/settings, auth screens when already signed in)
	// keep navigating after load. Measuring or capturing mid-navigation tears
	// down the page's execution context, which surfaces as a confusing failure
	// about the tool rather than about the redirect.
	await settleUrl(page);
	await page.addStyleTag({
		content: `
			*, *::before, *::after {
				animation-duration: 0s !important;
				animation-delay: 0s !important;
				transition-duration: 0s !important;
				transition-delay: 0s !important;
				scroll-behavior: auto !important;
			}
			/* The caret blinks on its own timer and lands mid-blink in captures. */
			* { caret-color: transparent !important; }
		`
	});
	// Under parallel load the dev server occasionally never reaches a full
	// network-idle window. That is a busy machine, not a broken page, so a
	// timeout here falls through to the height check rather than failing the
	// test with a misleading "layout" error.
	await page
		.waitForLoadState('networkidle', { timeout: 15000 })
		.catch(() => undefined);
	// Webfonts swap in after first paint; a capture before that shows fallbacks.
	await page.evaluate(() => document.fonts.ready);
	await settleHeight(page);
}

/**
 * Wait until the URL stops changing — i.e. client-side redirects are done.
 *
 * Starts from the current URL rather than an empty string, so a route that
 * never redirects settles on the first comparison instead of paying a poll
 * interval it does not need.
 */
async function settleUrl(page: Page, tries = 12): Promise<void> {
	let last = page.url();
	for (let i = 0; i < tries; i++) {
		await page.waitForTimeout(120);
		const url = page.url();
		if (url === last) return;
		last = url;
	}
}

/**
 * Wait until the document stops growing.
 *
 * `networkidle` is not enough here: these pages hydrate from IndexedDB and
 * re-render lists after the network has already gone quiet, so two runs of the
 * same route captured at "idle" differ by hundreds of pixels in height —
 * a diff that says nothing about the layout. Polling the scroll height until
 * it repeats catches that tail.
 */
async function settleHeight(page: Page, tries = 40): Promise<void> {
	let last = -1;
	let stable = 0;
	for (let i = 0; i < tries; i++) {
		const h = await page.evaluate(
			() => document.documentElement.scrollHeight
		);
		stable = h === last ? stable + 1 : 0;
		// Three identical readings ~150ms apart: the list is done re-rendering.
		if (stable >= 2) return;
		last = h;
		await page.waitForTimeout(150);
	}
}

/**
 * Selectors whose content changes between runs (timestamps, relative times,
 * generated ids). Masked in screenshots rather than removed, so layout — the
 * thing the refactor actually touches — still shows.
 */
export const VOLATILE = [
	'time',
	'[datetime]',
	'[data-testid="last-sync"]',
	'[data-testid="relative-time"]',
	'.font-mono' // version string in the footer, uuids in admin tables
];

export function maskLocators(page: Page) {
	return VOLATILE.map((sel) => page.locator(sel));
}

/** Snapshot name: one file per route × platform. */
export function snapshotName(route: RouteSpec, platform: Platform): string {
	return `${route.id}--${platform}.png`;
}

/**
 * Resolve a `:id` route against live data. The seed assigns random UUIDs
 * (cmd/seed/main.go has no fixed ids), so the concrete path can only be read
 * off the running app. Returns null when no seeded item exists — the caller
 * skips rather than fails, and reports the gap.
 */
export async function resolveDynamicRoute(
	page: Page,
	template: string
): Promise<string | null> {
	const listFor: Record<string, string> = {
		'/cards/:id': '/wallet?type=cards',
		'/vouchers/:id': '/wallet?type=vouchers',
		'/gift-cards/:id': '/wallet?type=gift-cards',
		'/merchants/:id': '/merchants',
		'/admin/users/:id/edit': '/admin/users',
		'/admin/merchants/:id/edit': '/admin/merchants'
	};
	const listPath = listFor[template];
	if (!listPath) return null;

	await page.goto(listPath);
	await page.waitForLoadState('networkidle');

	const prefix = template.split('/:')[0];
	const href = await page
		.locator(`a[href^="${prefix}/"]`)
		.first()
		.getAttribute('href')
		.catch(() => null);
	if (!href) return null;

	// `/admin/users/:id/edit` needs the trailing segment the list link omits.
	return template.endsWith('/edit') && !href.endsWith('/edit')
		? `${href}/edit`
		: href;
}

/**
 * The box the page's content actually gets — not just `<main>`'s own style.
 *
 * Asserting on `<main>` alone proves nothing: it is shell-provided and
 * therefore identical everywhere by construction. What differs between pages
 * is the padding and width they stack *inside* it. So this walks from `<main>`
 * down through single-child container wrappers and returns the innermost
 * content edge: the width available to the page, and the total horizontal
 * inset from the viewport.
 */
export async function contentContainerBox(page: Page): Promise<{
	maxWidth: string;
	contentWidth: number;
	insetLeft: number;
	insetRight: number;
} | null> {
	return page.evaluate(() => {
		const main = document.querySelector('main');
		if (!main) return null;

		// Descend while the only child is a plain layout wrapper — that is the
		// page re-stating the container rather than rendering content.
		let node: Element = main;
		for (let depth = 0; depth < 4; depth++) {
			const kids = Array.from(node.children).filter(
				(el) => el.tagName === 'DIV'
			);
			if (kids.length !== 1) break;
			const kid = kids[0];
			const cs = getComputedStyle(kid);
			const wrapperish =
				cs.maxWidth !== 'none' ||
				parseFloat(cs.paddingLeft) > 0 ||
				parseFloat(cs.paddingRight) > 0;
			if (!wrapperish) break;
			node = kid;
		}

		const rect = node.getBoundingClientRect();
		const cs = getComputedStyle(node);
		const padL = parseFloat(cs.paddingLeft) || 0;
		const padR = parseFloat(cs.paddingRight) || 0;
		// Insets are measured against the body box, not the viewport: a visible
		// scrollbar makes the body narrower than the viewport, and measuring the
		// right inset against the viewport then reports that scrollbar width as
		// layout asymmetry on every symmetric page.
		const body = document.body.getBoundingClientRect();
		return {
			maxWidth: getComputedStyle(main).maxWidth,
			contentWidth: Math.round(rect.width - padL - padR),
			insetLeft: Math.round(rect.left - body.left + padL),
			insetRight: Math.round(body.right - rect.right + padR)
		};
	});
}

export function attachNote(testInfo: TestInfo, name: string, body: string) {
	return testInfo.attach(name, { body, contentType: 'text/plain' });
}
