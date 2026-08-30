/**
 * Accessibility baseline.
 *
 * Counts violations per route; it does not force them to zero. The refactor
 * must not increase the count — that is the whole contract. The recorded
 * counts live in axe-baseline.json next to this file.
 *
 * Record: AXE_RECORD=1 npx playwright test tests/structure/axe.spec.ts
 * Verify: npx playwright test tests/structure/axe.spec.ts
 */
import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';
import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { STATIC_ROUTES, platformsFor } from './routes';
import { PLATFORM_UA, PLATFORM_VIEWPORT, stabilise } from './helpers';

const BASELINE_FILE = fileURLToPath(new URL('./axe-baseline.json', import.meta.url));
const RECORDING = process.env.AXE_RECORD === '1';

function readBaseline(): Record<string, number> {
	try {
		return JSON.parse(readFileSync(BASELINE_FILE, 'utf8'));
	} catch {
		return {};
	}
}

const recorded: Record<string, number> = {};

test.describe('axe / desktop', () => {
	test.use({
		userAgent: PLATFORM_UA.desktop,
		viewport: PLATFORM_VIEWPORT.desktop
	});

	const routes = STATIC_ROUTES.filter(
		(r) => r.kind !== 'redirect' && platformsFor(r).includes('desktop')
	);

	for (const route of routes) {
		test(route.id, async ({ browser }) => {
			const context = await browser.newContext({
				userAgent: PLATFORM_UA.desktop,
				viewport: PLATFORM_VIEWPORT.desktop,
				...(route.auth
					? { storageState: 'tests/structure/.auth/admin.json' }
					: {})
			});
			const page = await context.newPage();
			try {
				await page.goto(route.path);
				// stabilise() waits for client-side redirects to finish first —
				// AxeBuilder evaluates inside the page, and a navigation mid-analysis
				// destroys its execution context.
				await stabilise(page);

				const results = await new AxeBuilder({ page })
					.withTags(['wcag2a', 'wcag2aa'])
					.analyze();
				const count = results.violations.length;

				if (RECORDING) {
					recorded[route.id] = count;
					test.info().annotations.push({
						type: 'axe-recorded',
						description: `${route.id}=${count}`
					});
					return;
				}

				const baseline = readBaseline();
				const before = baseline[route.id];
				expect(
					before,
					`no axe baseline for ${route.id} — run AXE_RECORD=1 first`
				).toBeDefined();
				expect(
					count,
					`axe violations rose from ${before} to ${count}: ` +
						results.violations.map((v) => v.id).join(', ')
				).toBeLessThanOrEqual(before);
			} finally {
				await context.close();
			}
		});
	}

	test.afterAll(() => {
		if (RECORDING && Object.keys(recorded).length > 0) {
			// Merge, so recording a subset of routes does not drop the rest.
			const merged = { ...readBaseline(), ...recorded };
			writeFileSync(BASELINE_FILE, JSON.stringify(merged, null, 2) + '\n');
		}
	});
});
