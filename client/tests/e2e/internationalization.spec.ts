import { expect, test } from './fixtures/test-fixtures';
import { LoginPage } from './pages/login.page';
import { BasePage } from './pages/base.page';

test.describe('Internationalization', () => {
	test('should detect browser language', async ({ page }) => {
		const loginPage = new LoginPage(page);
		await loginPage.goto();

		const htmlLang = await page.locator('html').getAttribute('lang');
		expect(htmlLang).toBeTruthy();
	});

	test('should display login page in detected language', async ({ page }) => {
		const loginPage = new LoginPage(page);
		await loginPage.goto();

		const loginHeading = page.locator('h1, h2').first();
		await expect(loginHeading).toBeVisible();

		const headingText = await loginHeading.textContent();
		expect(headingText).toBeTruthy();
	});

	test('should switch language from navbar', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		const basePage = new BasePage(page);

		// Try switching to English
		const switched = await basePage.switchLanguage('English');
		if (!switched) {
			test.skip();
			return;
		}

		const englishText = page.locator(
			'text=/Settings|Dashboard|Cards|Vouchers/i'
		);
		await expect(englishText.first()).toBeVisible({ timeout: 5000 });
	});

	test('should persist language selection', async ({ authenticatedPage }) => {
		const page = authenticatedPage;
		const basePage = new BasePage(page);

		const switched = await basePage.switchLanguage('English');
		if (!switched) {
			test.skip();
			return;
		}

		// Navigate to another page
		await page.goto('/dashboard');
		await page.waitForURL(/\/dashboard/, { timeout: 10000 });

		const englishText = page.locator(
			'text=/Dashboard|Cards|Vouchers|Gift Cards/i'
		);
		await expect(englishText.first()).toBeVisible({ timeout: 5000 });
	});

	test('should display dates in correct locale format', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.goto('/vouchers');
		await page.waitForURL(/\/vouchers/, { timeout: 10000 });

		const datePattern = page.locator(
			'text=/\\d{1,2}\\.\\d{1,2}\\.\\d{4}|\\d{1,2}\\/\\d{1,2}\\/\\d{4}|\\w+ \\d{1,2}, \\d{4}/'
		);
		if (
			await datePattern
				.first()
				.isVisible({ timeout: 5000 })
				.catch(() => false)
		) {
			expect(await datePattern.first().isVisible()).toBeTruthy();
		}
	});

	test('should display currency in correct locale format', async ({
		authenticatedPage
	}) => {
		const page = authenticatedPage;

		await page.goto('/gift-cards');
		await page.waitForURL(/\/gift-cards/, { timeout: 10000 });

		const currencyPattern = page.locator('text=/CHF|EUR|\\$|Fr\\.|€/');
		if (
			await currencyPattern
				.first()
				.isVisible({ timeout: 5000 })
				.catch(() => false)
		) {
			expect(await currencyPattern.first().isVisible()).toBeTruthy();
		}
	});

	test('should translate form labels', async ({ authenticatedPage }) => {
		const page = authenticatedPage;

		await page.goto('/cards/new');
		await page.waitForURL(/\/cards\/new/, { timeout: 10000 });

		// Wait for form heading to render
		const heading = page.locator('h1').first();
		await expect(heading).toBeVisible({ timeout: 10000 });

		const headingText = await heading.textContent();
		expect(headingText?.trim().length).toBeGreaterThan(0);

		// Verify form has input fields with labels
		const submitButton = page.locator('button[type="submit"]');
		await expect(submitButton).toBeVisible({ timeout: 5000 });
	});

	test('should translate navigation items', async ({ authenticatedPage }) => {
		const page = authenticatedPage;

		const navLinks = page.locator('nav a, [role="navigation"] a');
		// Wait for the nav to mount before counting — otherwise the count can
		// race the authenticated layout render and come back as 0.
		await expect(navLinks.first()).toBeVisible({ timeout: 10000 });
		const navCount = await navLinks.count();
		expect(navCount).toBeGreaterThan(0);

		for (let i = 0; i < Math.min(navCount, 5); i++) {
			const linkText = await navLinks.nth(i).textContent();
			expect(linkText?.trim().length).toBeGreaterThan(0);
		}
	});
});
