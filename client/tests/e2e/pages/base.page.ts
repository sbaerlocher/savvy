import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';

export class BasePage {
	constructor(readonly page: Page) {}

	get toast(): Locator {
		return this.page.locator('[role="status"], [role="alert"], .toast').first();
	}

	get errorToast(): Locator {
		return this.page.locator('[role="alert"], .toast').first();
	}

	get confirmModal(): Locator {
		return this.page.locator('[role="dialog"]');
	}

	get confirmButton(): Locator {
		return this.page.locator('[data-testid="modal-confirm"]');
	}

	get offlineIndicator(): Locator {
		return this.page.getByText(/offline|keine verbindung/i);
	}

	async waitForPageReady() {
		try {
			await this.page.waitForSelector('text=/Loading|Laden|Lädt|Chargement/i', {
				state: 'hidden',
				timeout: 10000
			});
		} catch {
			// Loading might already be hidden
		}
	}

	async waitForApi(endpoint: string, method?: string, timeout = 15000) {
		return this.page.waitForResponse(
			(response) => {
				const urlMatches = response.url().includes(endpoint);
				const methodMatches = method
					? response.request().method() === method
					: true;
				return urlMatches && methodMatches;
			},
			{ timeout }
		);
	}

	async expectToast(message?: string) {
		await expect(this.toast).toBeVisible({ timeout: 5000 });
		if (message) {
			await expect(this.toast).toContainText(message);
		}
	}

	async confirmDeletion(apiEndpoint: string) {
		await expect(this.confirmButton).toBeVisible({ timeout: 5000 });

		const deleteResponse = this.page.waitForResponse(
			(resp) =>
				resp.url().includes(apiEndpoint) &&
				resp.request().method() === 'DELETE',
			{ timeout: 10000 }
		);

		await this.confirmButton.click();
		await deleteResponse;
	}

	async confirmAction() {
		await expect(this.confirmButton).toBeVisible({ timeout: 5000 });
		await this.confirmButton.click();
	}

	get languageMenuButton(): Locator {
		return this.page
			.locator('button[aria-haspopup="true"]')
			.filter({ has: this.page.locator('svg') })
			.first();
	}

	get languageMenu(): Locator {
		return this.page.locator('[role="menu"]');
	}

	async switchLanguage(langName: string) {
		// Click globe icon button in navbar to open language menu
		const globeButton = this.page
			.locator(
				'button[aria-label*="Language" i], button[aria-label*="Sprache" i], button[aria-label*="selectLanguage"]'
			)
			.first();
		if (!(await globeButton.isVisible({ timeout: 3000 }).catch(() => false))) {
			return false;
		}
		await globeButton.click();

		// Wait for dropdown menu and click language option
		const menuItem = this.page
			.locator('[role="menuitem"]')
			.filter({ hasText: langName });
		if (!(await menuItem.isVisible({ timeout: 3000 }).catch(() => false))) {
			// Try pressing Escape to close if menu didn't open
			await this.page.keyboard.press('Escape');
			return false;
		}
		await menuItem.click();
		await this.waitForPageReady();
		return true;
	}
}
