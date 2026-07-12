import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class DashboardPage extends BasePage {
	// Dashboard h1 is now the "Deine Favoriten" / "Your favorites" title
	// (Direction B redesign); the greeting moved to the eyebrow line.
	readonly welcomeHeading = this.page
		.locator('h1')
		.filter({ hasText: /Favoriten|Favorites/i });
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

	readonly statTotalBalance = this.page.getByTestId('dashboard-stat-balance');
	readonly statEntries = this.page.getByTestId('dashboard-stat-entries');

	async expectStatsVisible() {
		await expect(this.statTotalBalance).toBeVisible();
		await expect(this.statEntries).toBeVisible();
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
