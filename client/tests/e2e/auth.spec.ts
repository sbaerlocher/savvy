import { expect, test, TEST_USERS } from './fixtures/test-fixtures';

test.describe('Authentication', () => {
	test.beforeEach(async ({ page }) => {
		await page.context().clearCookies();
	});

	test('should display login page', async ({ loginPage }) => {
		await loginPage.goto();
		await loginPage.expectLoginPage();
		await expect(
			loginPage.submitButton.filter({ hasText: /Login|Anmelden/i })
		).toBeVisible();
	});

	test('should login with valid credentials', async ({ loginPage, page }) => {
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await expect(page).toHaveURL('/dashboard', { timeout: 10000 });
		await expect(
			page.locator('[data-testid="favorites-section"]')
		).toBeVisible({
			timeout: 5000
		});
	});

	test('should show error with invalid credentials', async ({ loginPage }) => {
		await loginPage.goto();
		await loginPage.emailInput.fill('invalid@example.com');
		await loginPage.passwordInput.fill('wrongpassword');
		await loginPage.submitButton.click();
		await loginPage.expectError();
	});

	test('should logout successfully', async ({ loginPage, page }) => {
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await expect(page).toHaveURL('/dashboard', { timeout: 10000 });

		// Logout sits in DesktopNav's user menu, which is `hidden sm:block`. Below
		// `sm` that menu does not exist and the action lives on /profile instead,
		// the account destination of the bottom nav — so take whichever route the
		// current layout offers.
		const userMenuButton = page
			.locator(
				'button[aria-label*="User" i], button[aria-label*="Benutzer" i]'
			)
			.filter({ visible: true })
			.first();
		if (await userMenuButton.isVisible().catch(() => false)) {
			await userMenuButton.click();
		} else {
			await page.goto('/profile');
		}

		// Match on the accessible name, not on `.text-danger-600`: the profile page
		// styles account deletion with that same class and renders it before the
		// logout, so a class-based `.first()` would click "delete account" there.
		const logoutButton = page
			.getByRole('button', { name: /Logout|Abmelden|Déconnexion/i })
			.filter({ visible: true })
			.first();
		await logoutButton.waitFor({ state: 'visible', timeout: 10000 });
		await logoutButton.click();

		await expect(page).toHaveURL('/login', { timeout: 10000 });
		await page.goto('/dashboard');
		await expect(page).toHaveURL('/login');
	});

	test('should redirect to login when accessing protected route without auth', async ({
		page
	}) => {
		await page.goto('/dashboard');
		await expect(page).toHaveURL('/login', { timeout: 5000 });
	});

	test('should redirect to login when accessing cards without auth', async ({
		page
	}) => {
		await page.goto('/cards');
		await expect(page).toHaveURL('/login', { timeout: 5000 });
	});

	test('should persist login across page navigation', async ({
		loginPage,
		page
	}) => {
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await expect(page).toHaveURL('/dashboard', { timeout: 10000 });

		// Legacy list routes redirect into the unified /wallet overview
		// (type-filtered); /dashboard stays. Assert the effective URL after redirect.
		const navigations: Array<[string, string]> = [
			['/cards', '/wallet?type=cards'],
			['/vouchers', '/wallet?type=vouchers'],
			['/gift-cards', '/wallet?type=gift-cards'],
			['/dashboard', '/dashboard']
		];
		for (const [path, expectedUrl] of navigations) {
			await page.goto(path);
			await expect(page).toHaveURL(expectedUrl);
		}
	});
});

test.describe('Registration', () => {
	test.beforeEach(async ({ page }) => {
		await page.context().clearCookies();
	});

	test('should display registration page if enabled', async ({ page }) => {
		await page.goto('/register');
		await page.waitForLoadState('networkidle');

		if (page.url().includes('/register')) {
			await expect(
				page
					.locator('h1, h2')
					.filter({ hasText: /Register|Registrieren|Sign up|Konto erstellen/i })
			).toBeVisible({ timeout: 5000 });
			await expect(
				page.locator('input[type="email"], input[name="email"]').first()
			).toBeVisible();
			await expect(
				page.locator('input[type="password"], input[name="password"]').first()
			).toBeVisible();
			await expect(page.locator('button[type="submit"]')).toBeVisible();
		} else {
			await expect(page).toHaveURL('/login');
		}
	});

	test('should validate required fields on registration', async ({ page }) => {
		await page.goto('/register');
		await page.waitForLoadState('networkidle');

		if (page.url().includes('/register')) {
			await page.click('button[type="submit"]');
			await expect(page).toHaveURL(/\/register/);
		}
	});
});

test.describe('Admin Authentication', () => {
	test.beforeEach(async ({ page }) => {
		await page.context().clearCookies();
	});

	test('should deny access to admin panel for regular users', async ({
		loginPage,
		page
	}) => {
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await expect(page).toHaveURL('/dashboard', { timeout: 10000 });

		await page.goto('/admin/users');
		await expect(page).not.toHaveURL(/\/admin\/users\/?$/);
	});

	test('should allow access to admin panel for admin users', async ({
		loginPage,
		page
	}) => {
		await loginPage.login(TEST_USERS.admin.email, TEST_USERS.admin.password);
		await expect(page).toHaveURL('/dashboard', { timeout: 10000 });

		await page.goto('/admin/users');
		await page.waitForLoadState('networkidle');
		await expect(page).toHaveURL('/admin/users');
		await expect(
			page.locator('h1').filter({ hasText: /Users|Benutzer/i })
		).toBeVisible({
			timeout: 5000
		});
	});
});
