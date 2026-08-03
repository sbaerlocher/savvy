import { defineConfig, devices } from '@playwright/test';

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
	testDir: './tests/e2e',
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

	projects: [
		// ========================================================================
		// BROWSER PROJECTS (run in parallel)
		// ========================================================================
		// globalSetup creates fresh DB once, then all browsers run in parallel
		// No need for per-browser db-reset or project dependencies
		// ========================================================================
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		},
		{
			name: 'firefox',
			use: { ...devices['Desktop Firefox'] }
		},
		{
			name: 'Mobile Chrome',
			use: { ...devices['Pixel 5'] }
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
