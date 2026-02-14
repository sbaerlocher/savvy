import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class AdminPage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	get heading(): Locator {
		return this.page.locator('h1').filter({ hasText: /Users|Benutzer/i });
	}

	get usersSection(): Locator {
		return this.page.locator('text=/Benutzer|Users/i').first();
	}

	get searchInput(): Locator {
		return this.page
			.locator(
				'input[type="search"], input[placeholder*="Search" i], input[placeholder*="Suchen" i]'
			)
			.first();
	}

	get tableRows(): Locator {
		return this.page.locator('tbody tr');
	}

	async goto(section: 'users' | 'audit-log' = 'users') {
		const path = `/admin/${section}`;
		const apiEndpoint =
			section === 'users' ? '/api/v1/admin/users' : '/api/v1/admin/audit-log';

		const apiResponse = this.page
			.waitForResponse((resp) => resp.url().includes(apiEndpoint), {
				timeout: 15000
			})
			.catch(() => null);
		await this.page.goto(path);
		await apiResponse;
		await this.waitForPageReady();
	}

	async navigateToAuditLog() {
		await this.goto('audit-log');
	}

	async search(term: string) {
		await expect(this.searchInput).toBeVisible({ timeout: 5000 });
		await this.searchInput.fill(term);
		// Wait for debounce to settle
		await this.page.waitForFunction(() => true, null, { timeout: 600 });
	}
}
