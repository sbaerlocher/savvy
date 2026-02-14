import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class DashboardPage extends BasePage {
	readonly welcomeHeading = this.page
		.locator('h1')
		.filter({ hasText: /Willkommen|Welcome/i });
	readonly statCards = this.page.getByTestId('dashboard-stat-cards');
	readonly statVouchers = this.page.getByTestId('dashboard-stat-vouchers');
	readonly statGiftCards = this.page.getByTestId('dashboard-stat-gift-cards');
	readonly favoritesHeading = this.page
		.locator('text=/Favorites|Favoriten/i')
		.first();
	readonly favoritesList = this.page
		.locator('[data-testid="favorites-list"], .favorites-section')
		.first();

	constructor(page: Page) {
		super(page);
	}

	async goto() {
		await this.page.goto('/dashboard');
		await this.waitForPageReady();
	}

	async waitForDashboardApi() {
		await this.page.waitForResponse(
			(resp) =>
				resp.url().includes('/api/v1/dashboard') && resp.status() === 200,
			{ timeout: 10000 }
		);
	}

	async expectStatsVisible() {
		await expect(this.statCards).toBeVisible();
		await expect(this.statVouchers).toBeVisible();
		await expect(this.statGiftCards).toBeVisible();
	}

	get favoriteItems(): Locator {
		return this.favoritesList.locator(
			'a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]'
		);
	}

	get recentItemLinks(): Locator {
		return this.page
			.locator(
				'a[href^="/cards/"][href*="-"]:not([href="/cards/new"]):not([data-testid]),' +
					'a[href^="/vouchers/"][href*="-"]:not([href="/vouchers/new"]):not([data-testid]),' +
					'a[href^="/gift-cards/"][href*="-"]:not([href="/gift-cards/new"]):not([data-testid])'
			)
			.first();
	}
}
