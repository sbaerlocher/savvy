import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class ProfilePage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	// --- Page heading ---
	// Desktop renders the merged settings page ("Einstellungen") with a profile
	// tab; iOS and Android keep the standalone "Mein Profil" screen. Both
	// headings are valid depending on the platform the test runs as.
	get heading(): Locator {
		return this.page
			.locator('h1, h2')
			.filter({
				hasText:
					/Mein Profil|My Profile|Mon profil|Einstellungen|Settings|Paramètres/i
			})
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

	// Android (and iOS) disclose the name form behind an M3 settings row;
	// the accessible name starts with the row title "Name".
	get nameRowButton(): Locator {
		return this.page.getByRole('button', { name: /\bName\b/ }).first();
	}

	/** Opens the name editor where it is disclosed; a no-op on the inline
	 *  layouts that render the fields directly. */
	async openNameEditor() {
		// The Android settings screen loads the profile async and the row is a
		// TOGGLE, so a probe racing the closing editor can click it shut again —
		// retry once when the form does not materialise.
		for (let attempt = 0; attempt < 2; attempt++) {
			if (await this.firstNameInput.isVisible().catch(() => false)) return;
			await this.nameRowButton
				.waitFor({ state: 'visible', timeout: 5000 })
				.catch(() => {});
			if (!(await this.nameRowButton.isVisible().catch(() => false))) return;
			await this.nameRowButton.click();
			const opened = await this.firstNameInput
				.waitFor({ state: 'visible', timeout: 3000 })
				.then(() => true)
				.catch(() => false);
			if (opened) return;
		}
		await expect(this.firstNameInput).toBeVisible({ timeout: 5000 });
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
		// Desktop uses rounded-full, the Android M3 badge rounded-m3-full.
		return this.page
			.locator('[class*="rounded"]')
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
		// One id per platform arm: desktop, Android M3 dialog, iOS sheet.
		return this.page.locator(
			'#deleteConfirmation, #androidDeleteConfirmation, #ios-deleteConfirmation'
		);
	}

	get deletePasswordInput(): Locator {
		return this.page.locator(
			'#deletePassword, #androidDeletePassword, #ios-deletePassword'
		);
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
		await this.openNameEditor();
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
