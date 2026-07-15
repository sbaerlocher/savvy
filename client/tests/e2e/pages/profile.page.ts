import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class ProfilePage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	// --- Page heading ---
	get heading(): Locator {
		return this.page
			.locator('h1, h2')
			.filter({ hasText: /Mein Profil|My Profile|Mon profil/i })
			.first();
	}

	// --- Profile section ---
	get firstNameInput(): Locator {
		return this.page.locator('#firstName');
	}

	get lastNameInput(): Locator {
		return this.page.locator('#lastName');
	}

	get emailInput(): Locator {
		return this.page.locator('#email');
	}

	get saveProfileButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Speichern"), button:has-text("Save Profile"), button:has-text("Save")'
			)
			.first();
	}

	// --- Email verification ---
	get sendVerificationButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Bestätigungsmail"), button:has-text("Verification"), button:has-text("Verifizierung")'
			)
			.first();
	}

	get emailVerifiedBadge(): Locator {
		return this.page
			.locator('.rounded-full')
			.filter({ hasText: /✓|verifiziert|verified/i })
			.first();
	}

	// --- Export section ---
	get exportButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Daten exportieren"), button:has-text("Export Data"), button:has-text("Export")'
			)
			.first();
	}

	// --- Danger zone ---
	get deleteAccountButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Konto endgültig löschen"), button:has-text("Konto löschen"), button:has-text("Delete Account")'
			)
			.first();
	}

	get deleteConfirmationInput(): Locator {
		return this.page.locator('#deleteConfirmation');
	}

	get deletePasswordInput(): Locator {
		return this.page.locator('#deletePassword');
	}

	get deleteConfirmButton(): Locator {
		return this.page
			.locator(
				'button.bg-danger-600, button:has-text("Konto löschen"), button:has-text("Delete Account")'
			)
			.last();
	}

	// --- Navigation ---
	async goto() {
		const profileResponse = this.page
			.waitForResponse((resp) => resp.url().includes('/api/v1/profile'), {
				timeout: 15000
			})
			.catch(() => null);
		await this.page.goto('/profile');
		await this.page.waitForURL(/\/profile/, { timeout: 10000 });
		await profileResponse;
		await this.waitForPageReady();
	}

	async expectProfilePage() {
		await expect(this.heading).toBeVisible({ timeout: 5000 });
	}

	async updateProfile(firstName: string, lastName: string) {
		await expect(this.firstNameInput).toBeVisible({ timeout: 5000 });
		await this.firstNameInput.clear();
		await this.firstNameInput.fill(firstName);
		await this.lastNameInput.clear();
		await this.lastNameInput.fill(lastName);

		const updateResponse = this.page
			.waitForResponse(
				(resp) =>
					resp.url().includes('/api/v1/profile') &&
					resp.request().method() === 'PATCH' &&
					resp.status() < 400,
				{ timeout: 10000 }
			)
			.catch(() => null);
		await this.saveProfileButton.click();
		await updateResponse;
	}

	async triggerExport() {
		const downloadPromise = this.page
			.waitForEvent('download', { timeout: 15000 })
			.catch(() => null);
		await this.exportButton.click();
		return downloadPromise;
	}
}
