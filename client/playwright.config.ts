import { defineConfig, devices } from '@playwright/test';

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
	// Both suites live under tests/: e2e/ (behaviour) and structure/ (the
	// layout-refactor baseline). Projects below scope each one.
	testDir: './tests',
	fullyParallel: true, // Tests run in parallel within same browser
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	timeout: process.env.CI ? 60000 : 30000, // Increased timeout for CI (60s vs 30s)
	// workers: Auto-determined (browsers run in parallel)
	reporter: 'html',
	globalSetup: './tests/global.setup.ts',
	use: {
		baseURL: process.env.BASE_URL || 'https://e2e.savvy.test',
		trace: 'on-first-retry',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure',
		ignoreHTTPSErrors: true, // mkcert-signed dde traefik cert + WebKit Mixed Content (SVL-E2E-001)
		actionTimeout: process.env.CI ? 15000 : 10000, // Increased action timeout for CI
		navigationTimeout: process.env.CI ? 15000 : 10000 // Increased navigation timeout for CI
	},

	// Structural baseline snapshots live next to their spec, one file per
	// route × platform, so a route's baseline is findable by name.
	// `{testDir}` resolves per project — the structure projects already point
	// at ./tests/structure, so no second `structure/` segment here.
	snapshotPathTemplate: '{testDir}/__screenshots__/{arg}{ext}',

	projects: [
		// ========================================================================
		// BROWSER PROJECTS (run in parallel)
		// ========================================================================
		// globalSetup creates fresh DB once, then all browsers run in parallel
		// No need for per-browser db-reset or project dependencies
		// ========================================================================
		{
			name: 'chromium',
			testDir: './tests/e2e',
			use: { ...devices['Desktop Chrome'] }
		},
		{
			name: 'firefox',
			testDir: './tests/e2e',
			use: { ...devices['Desktop Firefox'] }
		},
		{
			name: 'Mobile Chrome',
			testDir: './tests/e2e',
			use: { ...devices['Pixel 5'] }
		},

		// ========================================================================
		// STRUCTURE PROJECTS (layout refactor baseline)
		// ========================================================================
		// `structure:auth` logs in once and freezes the session; the other two
		// depend on it so ~100 captures don't each replay the login flow.
		//
		// Platform variants come from the User-Agent, NOT the viewport:
		// src/lib/utils/platform.ts derives `platform` from navigator.userAgent
		// at module load. The specs set both per platform.
		// ========================================================================
		{
			name: 'structure:auth',
			testDir: './tests/structure',
			testMatch: /auth\.setup\.ts/,
			use: {
				...devices['Desktop Chrome'],
				// Same resolver bypass as the structure project below.
				launchOptions: {
					args: ['--host-resolver-rules=MAP *.savvy.test 127.0.0.1']
				}
			}
		},
		{
			name: 'structure',
			testDir: './tests/structure',
			testIgnore: /auth\.setup\.ts/,
			dependencies: ['structure:auth'],
			// The structure suite waits for redirects, fonts and list re-renders
			// to settle before it measures anything. That budget adds up past the
			// suite-wide 30s per-test cap, which then surfaces as a timeout rather
			// than as the layout result the test was meant to report.
			timeout: 90000,
			// Every test navigates the app for real. At full parallelism the Vite
			// dev server falls behind and tests fail on a missing `<main>` or a
			// navigation timeout — a report about server load dressed up as a
			// layout regression. Four workers keep the suite honest.
			workers: 4,
			// One retry for the same reason: a dev-server stall shows up as a
			// wandering timeout on a different route each run, never as a repeated
			// failure. A real layout regression fails both attempts.
			retries: 1,
			use: {
				...devices['Desktop Chrome'],
				// Full-page captures wait for lists to settle, and dynamic routes
				// navigate twice (list → detail). The suite-wide 10s default cuts
				// those off mid-render and reports a timeout as a layout failure.
				actionTimeout: 30000,
				navigationTimeout: 30000,
				// *.savvy.test always points at the local dde traefik. Resolving it
				// in-browser sidesteps macOS's mDNSResponder, which under sustained
				// load stalls on /etc/resolver lookups and fails whole runs with
				// navigation timeouts while the app itself answers in ~50ms.
				launchOptions: {
					args: ['--host-resolver-rules=MAP *.savvy.test 127.0.0.1']
				}
			}
		}

		// ========================================================================
		// WEBKIT/SAFARI DISABLED (SVL-E2E-001)
		// ========================================================================
		// WebKit has a bug where it upgrades http:// modulepreload links to https://
		// causing TLS errors with the local dev server (localhost:8080 http only).
		//
		// This is a Playwright WebKit issue, NOT a real Safari issue.
		// Real Mobile Safari on iPhone works perfectly (manual testing confirmed).
		//
		// Chromium + Firefox + Mobile Chrome provide sufficient E2E coverage.
		// WebKit/Safari should be tested manually on real devices.
		// ========================================================================
		// {
		//   name: "webkit",
		//   use: { ...devices["Desktop Safari"] }
		// }
	]

	/*
	 * E2E tests require the production build running on port 8080 via Docker.
	 * Use the dde e2e plugins (`.dde/plugins/e2e.*.sh`) for proper setup:
	 * - Starts PostgreSQL + app-e2e (production build with embedded SvelteKit)
	 * - Auto-seeds test data
	 * - Runs tests against http://localhost:8080
	 *
	 * webServer is disabled because we use Docker Compose for E2E tests.
	 */
});
