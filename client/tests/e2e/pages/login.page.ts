import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class LoginPage extends BasePage {
	readonly emailInput = this.page.locator('input[name="email"]');
	readonly passwordInput = this.page.locator('input[name="password"]');
	readonly submitButton = this.page.locator('button[type="submit"]');
	readonly heading = this.page
		.locator('h1, h2')
		.filter({ hasText: /Login|Anmelden/i });
	readonly registerLink = this.page.locator(
		'a[href="/register"], a:has-text("Register"), a:has-text("Registrieren")'
	);
	readonly errorMessage = this.page
		.getByText(/Invalid.*password|Ungültig/i)
		.first();

	constructor(page: Page) {
		super(page);
	}

	async goto() {
		await this.page.goto('/login');
		await this.waitForPageReady();
	}

	async login(email: string, password: string) {
		const maxRetries = 3;

		for (let attempt = 1; attempt <= maxRetries; attempt++) {
			await this.goto();

			const errorRetryButton = this.page.locator(
				'button:has-text("Erneut versuchen"), button:has-text("Try again")'
			);
			if (await errorRetryButton.isVisible().catch(() => false)) {
				await errorRetryButton.click();
				await this.page.waitForLoadState('networkidle');
			}

			await this.emailInput.waitFor({ state: 'visible', timeout: 10000 });
			await this.emailInput.fill(email);
			await this.passwordInput.fill(password);
			await this.submitButton.click();

			try {
				await this.page.waitForURL((url) => !url.pathname.includes('/login'), {
					timeout: 10000
				});
				return;
			} catch {
				if (attempt === maxRetries) {
					throw new Error(
						`Login failed after ${maxRetries} attempts for ${email}`
					);
				}
				await this.page.waitForTimeout(1000);
			}
		}
	}

	async expectLoginPage() {
		await expect(this.heading).toBeVisible({ timeout: 5000 });
		await expect(this.emailInput).toBeVisible();
		await expect(this.passwordInput).toBeVisible();
	}

	async expectError() {
		await expect(this.errorMessage).toBeVisible({ timeout: 5000 });
		await expect(this.page).toHaveURL(/\/login/);
	}
}
