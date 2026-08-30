/**
 * Visual baseline: one screenshot per route × platform.
 *
 * The matrix is platform-based, not viewport-based: `platform` in
 * src/lib/utils/platform.ts reads the User-Agent at module load, so a narrow
 * viewport alone still renders the desktop branch. Viewports vary alongside
 * the UA only to match each platform's real form factor.
 *
 * Record: npx playwright test tests/structure/baseline.spec.ts --update-snapshots
 * Verify: npx playwright test tests/structure/baseline.spec.ts
 *
 * Never bulk-update these. Each changed image needs a stated reason.
 */
import { expect, test } from '@playwright/test';
import { ALL_PLATFORMS, STATIC_ROUTES, platformsFor } from './routes';
import {
	PLATFORM_UA,
	PLATFORM_VIEWPORT,
	maskLocators,
	snapshotName,
	stabilise
} from './helpers';

for (const platform of ALL_PLATFORMS) {
	test.describe(`baseline / ${platform}`, () => {
		// Redirect-only routes render no markup — nothing to baseline.
		// `noScreenshot` routes render content whose height moves between runs
		// for non-layout reasons; each states why. They stay covered by the
		// structure rules.
		const routes = STATIC_ROUTES.filter(
			(r) =>
				r.kind !== 'redirect' &&
				!r.noScreenshot &&
				platformsFor(r).includes(platform)
		);

		for (const route of routes) {
			test(route.id, async ({ browser }) => {
				const context = await browser.newContext({
					userAgent: PLATFORM_UA[platform],
					viewport: PLATFORM_VIEWPORT[platform],
					...(route.auth
						? { storageState: 'tests/structure/.auth/admin.json' }
						: {})
				});
				const page = await context.newPage();
				try {
					await page.goto(route.path);
					await stabilise(page);
					await expect(page).toHaveScreenshot(snapshotName(route, platform), {
						fullPage: true,
						mask: maskLocators(page),
						// Sub-pixel text rendering differs slightly per run; anything
						// larger than this is a real layout change.
						maxDiffPixelRatio: 0.002
					});
				} finally {
					await context.close();
				}
			});
		}

		// Dynamic routes carry no screenshot baseline at all.
		//
		// Their concrete path is resolved by taking the first item off the list
		// page, and that list is not ordered stably — a run resolves to a
		// different record than the one the baseline was recorded from, so the
		// captured content differs for reasons that have nothing to do with
		// layout. Pinning them would need fixed seed UUIDs (cmd/seed/main.go
		// assigns random ones), which is a production-code change.
		//
		// They stay covered by structure.spec.ts, which asserts on the layout
		// rather than on which record happens to be shown.
	});
}
