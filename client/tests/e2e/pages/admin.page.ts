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
		// The admin page owns a search field that renders on every viewport
		// (routes/admin/users/+page.svelte). The global chrome adds more fields
		// with the same shape — DesktopNav's is `hidden sm:flex`, so on mobile a
		// blind .first() picks that hidden one. Scope to the page content (the
		// navs are siblings of <main>) and require visibility.
		return this.page
			.locator('main')
			.locator(
				'input[type="search"], input[placeholder*="Search" i], input[placeholder*="Suchen" i]'
			)
			.filter({ visible: true })
			.first();
	}

	get tableRows(): Locator {
		// The users list is a <table> on the native platforms but a CSS grid on
		// desktop (mockup screen-AdminDesktop), so the rows there are tagged
		// instead of being <tr>. Audit log is still a plain table.
		return this.page.locator('[data-testid="admin-user-row"], tbody tr');
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
