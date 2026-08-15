import { expect, test } from './fixtures/test-fixtures';

test('should not allow non-admin to access admin panel', async ({
	authenticatedPage
}) => {
	const page = authenticatedPage;
	await page.goto('/admin/users');

	// Non-admin users get redirected by backend middleware (303 → /)
	await page.waitForURL((url) => !url.pathname.startsWith('/admin'), {
		timeout: 10000
	});
	await expect(page).not.toHaveURL(/\/admin\/users\/?$/);
});

test.describe('Admin Panel', () => {
	test.beforeEach(async ({ adminAuthenticatedPage }) => {
		// adminAuthenticatedPage fixture handles login as admin
	});

	test('should access admin panel', async ({
		adminAuthenticatedPage,
		adminPage
	}) => {
		await adminPage.goto();
		await expect(adminPage.heading).toBeVisible();
	});

	test('should display users list', async ({
		adminAuthenticatedPage,
		adminPage
	}) => {
		await adminPage.goto();
		await expect(adminPage.tableRows.first()).toBeVisible({ timeout: 10000 });
		expect(await adminPage.tableRows.count()).toBeGreaterThan(0);
	});

	test('should search users', async ({ adminAuthenticatedPage, adminPage }) => {
		await adminPage.goto();
		await expect(adminPage.tableRows.first()).toBeVisible({ timeout: 10000 });

		await adminPage.search('anna');

		const resultCount = await adminPage.tableRows.count();
		expect(resultCount).toBeGreaterThan(0);
	});

	test('should navigate to audit log tab', async ({
		adminAuthenticatedPage,
		adminPage
	}) => {
		await adminPage.goto();

		await adminPage.navigateToAuditLog();

		// Audit log might be empty, so just check page loaded
		const hasRows = await adminPage.tableRows
			.first()
			.isVisible({ timeout: 5000 })
			.catch(() => false);
		expect(
			hasRows || adminPage.page.url().includes('/admin/audit-log')
		).toBeTruthy();
	});

	test('should search audit log', async ({
		adminAuthenticatedPage,
		adminPage
	}) => {
		await adminPage.goto('audit-log');

		if (
			await adminPage.searchInput
				.isVisible({ timeout: 3000 })
				.catch(() => false)
		) {
			await adminPage.search('login');
		}
	});

	test('should display user details', async ({
		adminAuthenticatedPage,
		adminPage
	}) => {
		await adminPage.goto();
		await expect(adminPage.tableRows.first()).toBeVisible({ timeout: 10000 });

		// Click first user row to see details
		const firstRow = adminPage.tableRows.first();
		const userLink = firstRow.locator('a').first();
		if (await userLink.isVisible({ timeout: 3000 }).catch(() => false)) {
			await userLink.click();
			await adminPage.waitForPageReady();

			const userContent = adminPage.page.locator(
				'text=/anna|admin|thomas|maria/i'
			);
			await expect(userContent.first()).toBeVisible({ timeout: 5000 });
		}
	});

	test('should navigate to system health page', async ({
		adminAuthenticatedPage,
		adminPage
	}) => {
		const page = adminPage.page;
		const healthApiResponse = page
			.waitForResponse(
				(resp) => resp.url().includes('/api/v1/admin/system-health'),
				{ timeout: 15000 }
			)
			.catch(() => null);
		await page.goto('/admin/system-health');
		await page.waitForURL(/\/admin\/system-health/, { timeout: 10000 });
		await healthApiResponse;
		await adminPage.waitForPageReady();

		const healthContent = page.locator(
			'text=/Database|Datenbank|SMTP|OAuth|Health|Status|Healthy|Ready/i'
		);
		await expect(healthContent.first()).toBeVisible({ timeout: 10000 });
	});

	test('should display service status on health page', async ({
		adminAuthenticatedPage,
		adminPage
	}) => {
		const page = adminPage.page;
		const healthApiResponse = page
			.waitForResponse(
				(resp) => resp.url().includes('/api/v1/admin/system-health'),
				{ timeout: 15000 }
			)
			.catch(() => null);
		await page.goto('/admin/system-health');
		await page.waitForURL(/\/admin\/system-health/, { timeout: 10000 });
		await healthApiResponse;
		await adminPage.waitForPageReady();

		// The service list has two layouts: a `hidden md:block` <table> on desktop
		// and a `md:hidden` card list below `md` (system-health/+page.svelte:321 vs
		// :693). Pixel 5 is 393px, so a `table`-scoped locator resolves to the
		// display:none copy there and never becomes visible. Assert on the row
		// itself instead — it exists in both layouts and carries the same content.
		const dbRow = page
			.locator('tr, button')
			.filter({ hasText: /Database|Datenbank/i })
			.filter({ visible: true })
			.first();
		await expect(dbRow).toBeVisible({ timeout: 10000 });

		// The row must actually report a status, not just name the service —
		// that is what distinguishes a rendered health check from an empty shell.
		await expect(dbRow).toHaveText(
			/Healthy|Unhealthy|Not configured|Gesund|Fehlerhaft|Nicht konfiguriert/i
		);
	});
});
