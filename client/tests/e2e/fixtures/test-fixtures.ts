import { test as base, type Page } from '@playwright/test';
import { AdminPage } from '../pages/admin.page';
import { DashboardPage } from '../pages/dashboard.page';
import { LoginPage } from '../pages/login.page';
import { ResourceDetailPage } from '../pages/resource-detail.page';
import { ResourceFormPage } from '../pages/resource-form.page';
import {
	ResourceListPage,
	type ResourceType
} from '../pages/resource-list.page';
import { ProfilePage } from '../pages/profile.page';
import { SecurityPage } from '../pages/security.page';
import { SettingsPage } from '../pages/settings.page';

const TEST_USERS = {
	regular: {
		email: 'anna.mueller@example.com',
		password: 'test123',
		name: 'Anna Müller'
	},
	admin: {
		email: 'admin@example.com',
		password: 'test123',
		name: 'Admin User'
	},
	shared: {
		email: 'thomas.schmidt@example.com',
		password: 'test123',
		name: 'Thomas Schmidt'
	},
	another: {
		email: 'maria.garcia@example.com',
		password: 'test123',
		name: 'Maria Garcia'
	}
} as const;

export type TestFixtures = {
	loginPage: LoginPage;
	dashboardPage: DashboardPage;
	adminPage: AdminPage;
	profilePage: ProfilePage;
	securityPage: SecurityPage;
	settingsPage: SettingsPage;
	authenticatedPage: Page;
	adminAuthenticatedPage: Page;
	authenticatedProfilePage: Page;
	authenticatedSecurityPage: Page;
	authenticatedSettingsPage: Page;
	cardsListPage: ResourceListPage;
	vouchersListPage: ResourceListPage;
	giftCardsListPage: ResourceListPage;
	cardDetailPage: ResourceDetailPage;
	voucherDetailPage: ResourceDetailPage;
	giftCardDetailPage: ResourceDetailPage;
	cardFormPage: ResourceFormPage;
	voucherFormPage: ResourceFormPage;
	giftCardFormPage: ResourceFormPage;
};

export const test = base.extend<TestFixtures>({
	loginPage: async ({ page }, use) => {
		await use(new LoginPage(page));
	},

	dashboardPage: async ({ page }, use) => {
		await use(new DashboardPage(page));
	},

	adminPage: async ({ page }, use) => {
		await use(new AdminPage(page));
	},

	profilePage: async ({ page }, use) => {
		await use(new ProfilePage(page));
	},

	securityPage: async ({ page }, use) => {
		await use(new SecurityPage(page));
	},

	settingsPage: async ({ page }, use) => {
		await use(new SettingsPage(page));
	},

	authenticatedPage: async ({ page }, use) => {
		await page.context().clearCookies();
		const loginPage = new LoginPage(page);
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await use(page);
	},

	adminAuthenticatedPage: async ({ page }, use) => {
		await page.context().clearCookies();
		const loginPage = new LoginPage(page);
		await loginPage.login(TEST_USERS.admin.email, TEST_USERS.admin.password);
		await use(page);
	},

	authenticatedProfilePage: async ({ page }, use) => {
		await page.context().clearCookies();
		const loginPage = new LoginPage(page);
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await page.goto('/profile');
		await page.waitForURL(/\/profile/, { timeout: 10000 });
		await use(page);
	},

	authenticatedSecurityPage: async ({ page }, use) => {
		await page.context().clearCookies();
		const loginPage = new LoginPage(page);
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await page.goto('/security');
		await page.waitForURL(/\/security/, { timeout: 10000 });
		await use(page);
	},

	authenticatedSettingsPage: async ({ page }, use) => {
		await page.context().clearCookies();
		const loginPage = new LoginPage(page);
		await loginPage.login(
			TEST_USERS.regular.email,
			TEST_USERS.regular.password
		);
		await page.goto('/notifications');
		await page.waitForURL(/\/notifications/, { timeout: 10000 });
		await use(page);
	},

	cardsListPage: async ({ page }, use) => {
		await use(new ResourceListPage(page, 'cards'));
	},

	vouchersListPage: async ({ page }, use) => {
		await use(new ResourceListPage(page, 'vouchers'));
	},

	giftCardsListPage: async ({ page }, use) => {
		await use(new ResourceListPage(page, 'gift-cards'));
	},

	cardDetailPage: async ({ page }, use) => {
		await use(new ResourceDetailPage(page, 'cards'));
	},

	voucherDetailPage: async ({ page }, use) => {
		await use(new ResourceDetailPage(page, 'vouchers'));
	},

	giftCardDetailPage: async ({ page }, use) => {
		await use(new ResourceDetailPage(page, 'gift-cards'));
	},

	cardFormPage: async ({ page }, use) => {
		await use(new ResourceFormPage(page, 'cards'));
	},

	voucherFormPage: async ({ page }, use) => {
		await use(new ResourceFormPage(page, 'vouchers'));
	},

	giftCardFormPage: async ({ page }, use) => {
		await use(new ResourceFormPage(page, 'gift-cards'));
	}
});

export { expect } from '@playwright/test';
export { TEST_USERS };
export type { ResourceType };
