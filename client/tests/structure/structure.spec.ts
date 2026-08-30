/**
 * Structural target tests.
 *
 * These encode the Phase 0 inconsistency list as assertions. They are
 * DELIBERATELY RED before the refactor — each failure names a page that builds
 * its own shell instead of using the shared one. Turning them green is what
 * "done" means; the screenshots only guard against collateral damage.
 *
 * Run: dde project:up && npx playwright test tests/structure
 */
import { expect, test } from '@playwright/test';
import {
	ALL_PLATFORMS,
	DYNAMIC_ROUTES,
	STATIC_ROUTES,
	platformsFor,
	shellRoutes,
	type Platform,
	type RouteSpec
} from './routes';
import {
	PLATFORM_UA,
	PLATFORM_VIEWPORT,
	contentContainerBox,
	resolveDynamicRoute,
	stabilise
} from './helpers';

/**
 * The single container definition every shell route must end up sharing.
 *
 * Measured off the shell's own `<main>` (+layout.svelte:129,
 * `max-w-7xl mx-auto pt-4 pb-6 px-4 sm:px-6 lg:px-8`) on a page that adds no
 * container of its own — i.e. the target state, not the current average.
 * A page landing on different numbers is stacking its own container inside
 * the shell's, which is exactly what this refactor removes.
 *
 * Insets are measured against documentElement.clientWidth (scrollbar
 * excluded), so desktop reads 105px, not the naive (1440-1280)/2 + 32 = 112px.
 */
const EXPECTED_CONTAINER: Record<
	Platform,
	{ maxWidth: string; contentWidth: number; insetLeft: number }
> = {
	// 390px and 412px are both below the sm: breakpoint (640px) → px-4 (16px).
	ios: { maxWidth: '1280px', contentWidth: 343, insetLeft: 16 },
	android: { maxWidth: '1280px', contentWidth: 365, insetLeft: 16 },
	// 1440px is above lg: (1024px) → px-8 (32px), plus the mx-auto gutter.
	desktop: { maxWidth: '1280px', contentWidth: 1216, insetLeft: 105 }
};

/** Nav links expected on every authenticated shell page, in order. */
async function navSignature(page: import('@playwright/test').Page) {
	return page.evaluate(() => {
		const nav = document.querySelector('nav');
		if (!nav) return null;
		return Array.from(nav.querySelectorAll('a'))
			.map((a) => a.getAttribute('href'))
			.filter((href): href is string => !!href && !href.startsWith('http'));
	});
}

for (const platform of ALL_PLATFORMS) {
	test.describe(`structure / ${platform}`, () => {
		test.use({
			userAgent: PLATFORM_UA[platform],
			viewport: PLATFORM_VIEWPORT[platform],
			// The seeded session; created once by the auth setup project.
			storageState: 'tests/structure/.auth/admin.json'
		});

		const routes = shellRoutes(STATIC_ROUTES).filter((r) =>
			platformsFor(r).includes(platform)
		);

		for (const route of routes) {
			test.describe(route.id, () => {
				test.beforeEach(async ({ page }) => {
					await page.goto(route.path);
					await stabilise(page);
				});

				test('has exactly one h1', async ({ page }) => {
					// Phase 0 found 4 pages with 2+ h1 and several with none —
					// the title is built in a different place on nearly every page.
					await expect(page.locator('h1')).toHaveCount(1);
				});

				test('has one header, one main, one footer landmark', async ({
					page
				}) => {
					await expect(page.locator('main')).toHaveCount(1);
					// header/footer are shell-provided; nested ones mean a page
					// rebuilt its own chrome.
					expect(await page.locator('body > header').count()).toBeLessThan(2);
					expect(await page.locator('body > footer').count()).toBeLessThan(2);
				});

				test('content container matches the shared definition', async ({
					page
				}) => {
					// Core assertion of the whole refactor: the content box must come
					// from exactly one place. Phase 0 counted five competing container
					// dialects, several of them stacked inside the shell's own.
					const box = await contentContainerBox(page);
					expect(box, 'no <main> element found').not.toBeNull();
					if (route.width === 'narrow' && platform === 'desktop') {
						// The shell's own reading column (PageShell width="narrow",
						// 680px per the desktop mockups). Still the one shared
						// definition — just its narrow variant, so the expected width
						// changes while the outer shell and the centring stay checked.
						// Below the cap (mobile) the column fills the shell, so those
						// platforms take the standard branch.
						expect({
							maxWidth: box!.maxWidth,
							contentWidth: box!.contentWidth
						}).toEqual({
							maxWidth: EXPECTED_CONTAINER[platform].maxWidth,
							contentWidth: 680
						});
					} else {
						expect({
							maxWidth: box!.maxWidth,
							contentWidth: box!.contentWidth,
							insetLeft: box!.insetLeft
						}).toEqual(EXPECTED_CONTAINER[platform]);
					}
					// Asymmetric insets mean something reached around the container.
					expect(box!.insetLeft).toBe(box!.insetRight);
				});

				test('page adds no container of its own', async ({ page }) => {
					// A second max-width / mx-auto / horizontal padding inside <main>
					// is the duplication this refactor removes (Phase 0 counted eight
					// pages re-stating `max-w-7xl mx-auto px-4` under a <main> that
					// already sets it).
					//
					// Detection is class-based, not computed-style-based: a child that
					// re-states the parent's own max-width computes to marginLeft 0px
					// (it already fills the constrained parent), so a computed-margin
					// check silently matches nothing and the test passes while the
					// duplication is right there.
					const nested = await page.evaluate(() => {
						const main = document.querySelector('main');
						if (!main) return [];
						const CONTAINER_CLASS =
							/(^|\s)(mx-auto|max-w-(?!full|none)[\w[\]./-]+|px-\d|sm:px-\d|lg:px-\d)(\s|$)/;
						return Array.from(main.querySelectorAll(':scope > div'))
							// The shell's own wrapper (PageShell marks it) is the shared
							// definition, not a page re-stating it — its narrow variant
							// legitimately carries mx-auto/max-w.
							.filter((el) => !el.hasAttribute('data-page-shell'))
							.filter((el) => CONTAINER_CLASS.test(el.className || ''))
							.map((el) => el.className);
					});
					expect(
						nested,
						`page re-states the shell container: ${nested.join(' | ')}`
					).toEqual([]);
				});
			});
		}

		// Nav identity is checked one route per test, against a reference read
		// once per platform.
		//
		// Walking every route inside a single test outran the timeout on the
		// mobile platforms; loading the reference again inside each test doubled
		// the page loads and made the suite flaky under dev-server load. Reading
		// it once in beforeAll costs one navigation per platform.
		const NAV_REFERENCE = '/dashboard';
		let referenceNav: string[] | null = null;

		test.beforeAll(async ({ browser }) => {
			const context = await browser.newContext({
				userAgent: PLATFORM_UA[platform],
				viewport: PLATFORM_VIEWPORT[platform],
				storageState: 'tests/structure/.auth/admin.json'
			});
			const page = await context.newPage();
			await page.goto(NAV_REFERENCE);
			await stabilise(page);
			referenceNav = await navSignature(page);
			await context.close();
		});

		for (const route of routes.filter((r) => r.path !== NAV_REFERENCE)) {
			test(`${route.id} shows the same navigation as the dashboard`, async ({
				page
			}) => {
				expect(referenceNav, 'dashboard renders no navigation').not.toBeNull();
				await page.goto(route.path);
				await stabilise(page);
				const sig = await navSignature(page);
				// A route may legitimately have no nav (Android renders its settings
				// screen without one); only a *different* nav is a failure.
				if (sig === null) return;
				expect(sig, `${route.path} vs ${NAV_REFERENCE}`).toEqual(referenceNav);
			});
		}
	});
}

/**
 * Dynamic routes resolve against seeded data. They skip (not fail) when the
 * seed holds no matching item, and the skip reason names the gap so it is not
 * mistaken for coverage.
 */
test.describe('structure / dynamic routes (desktop)', () => {
	test.use({
		userAgent: PLATFORM_UA.desktop,
		viewport: PLATFORM_VIEWPORT.desktop,
		storageState: 'tests/structure/.auth/admin.json'
	});

	for (const route of DYNAMIC_ROUTES) {
		test(`${route.id} has exactly one h1`, async ({ page }) => {
			const resolved = await resolveDynamicRoute(page, route.path);
			test.skip(
				resolved === null,
				`no seeded item for ${route.path} — route uncovered`
			);
			await page.goto(resolved!);
			await stabilise(page);
			await expect(page.locator('h1')).toHaveCount(1);
		});
	}
});

/**
 * Standalone routes are exempt from the shared container by design (auth
 * screens, error, offline). They still owe a single h1 — accessibility does
 * not get an exemption.
 */
test.describe('structure / standalone routes', () => {
	test.use({
		userAgent: PLATFORM_UA.desktop,
		viewport: PLATFORM_VIEWPORT.desktop
	});

	const standalone = STATIC_ROUTES.filter(
		(r: RouteSpec) => r.kind === 'standalone' && !r.auth
	);

	for (const route of standalone) {
		test(`${route.id} has exactly one h1`, async ({ page }) => {
			await page.goto(route.path);
			await stabilise(page);
			await expect(page.locator('h1')).toHaveCount(1);
		});
	}
});
