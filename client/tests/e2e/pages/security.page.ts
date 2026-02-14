import type { Locator, Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { BasePage } from './base.page';

export class SecurityPage extends BasePage {
	constructor(page: Page) {
		super(page);
	}

	// --- Page heading ---
	get heading(): Locator {
		return this.page
			.locator('h1, h2')
			.filter({ hasText: /Sicherheit|Security|Sécurité/i })
			.first();
	}

	// --- Password section ---
	get currentPasswordInput(): Locator {
		return this.page.locator('#currentPassword');
	}

	get newPasswordInput(): Locator {
		return this.page.locator('#newPassword');
	}

	get confirmPasswordInput(): Locator {
		return this.page.locator('#confirmPassword');
	}

	get changePasswordButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Passwort ändern"), button:has-text("Change Password"), button:has-text("Ändern")'
			)
			.first();
	}

	// --- 2FA section ---
	get twoFactorSection(): Locator {
		return this.page
			.locator('text=/Zwei-Faktor|Two-Factor|2FA|Authenticator/i')
			.first();
	}

	get enable2FAButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Aktivieren"), button:has-text("Enable"), button:has-text("2FA einrichten"), button:has-text("Setup 2FA"), [data-testid="enable-2fa"]'
			)
			.first();
	}

	get disable2FAButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Deaktivieren"), button:has-text("Disable"), [data-testid="disable-2fa"]'
			)
			.first();
	}

	get setupModal(): Locator {
		return this.page
			.locator('[role="dialog"], .modal, [data-testid="2fa-setup-modal"]')
			.first();
	}

	get totpCodeInput(): Locator {
		return this.page
			.locator('input[type="text"], input#code, input[name="code"]')
			.first();
	}

	get verifyButton(): Locator {
		return this.page
			.locator(
				'button[type="submit"]:has-text("Verifizieren"), button[type="submit"]:has-text("Verify"), button[type="submit"]:has-text("Bestätigen"), button[type="submit"]:has-text("Confirm")'
			)
			.first();
	}

	get cancelButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Abbrechen"), button:has-text("Cancel"), [data-testid="modal-cancel"]'
			)
			.first();
	}

	// --- Sessions section ---
	get sessionsSection(): Locator {
		return this.page
			.locator('text=/Aktive Sitzungen|Active Sessions|Sessions actives/i')
			.first();
	}

	get revokeOthersButton(): Locator {
		return this.page
			.locator(
				'button:has-text("Alle anderen abmelden"), button:has-text("Sign out all others"), button:has-text("Déconnecter toutes les autres")'
			)
			.first();
	}

	get currentSessionBadge(): Locator {
		return this.page
			.locator('.rounded-full')
			.filter({ hasText: /Aktuell|Current|Actuelle/i })
			.first();
	}

	// --- Navigation ---
	async goto() {
		const profileResponse = this.page
			.waitForResponse((resp) => resp.url().includes('/api/v1/profile'), {
				timeout: 15000
			})
			.catch(() => null);
		await this.page.goto('/security');
		await this.page.waitForURL(/\/security/, { timeout: 10000 });
		await profileResponse;
		await this.waitForPageReady();
	}

	async expectSecurityPage() {
		await expect(this.heading).toBeVisible({ timeout: 5000 });
	}

	async changePassword(current: string, newPass: string, confirm: string) {
		await expect(this.currentPasswordInput).toBeVisible({ timeout: 5000 });
		await this.currentPasswordInput.fill(current);
		await this.newPasswordInput.fill(newPass);
		await this.confirmPasswordInput.fill(confirm);

		const passwordResponse = this.page
			.waitForResponse(
				(resp) =>
					resp.url().includes('/change-password') &&
					resp.request().method() === 'POST',
				{ timeout: 10000 }
			)
			.catch(() => null);
		await this.changePasswordButton.click();
		await passwordResponse;
	}

	async open2FASetup() {
		await expect(this.enable2FAButton).toBeVisible({ timeout: 5000 });
		await this.enable2FAButton.click();
		await expect(this.setupModal).toBeVisible({ timeout: 5000 });
	}

	async close2FAModal() {
		if (
			await this.cancelButton.isVisible({ timeout: 3000 }).catch(() => false)
		) {
			await this.cancelButton.click();
		} else {
			await this.page.keyboard.press('Escape');
		}
		await expect(this.setupModal).not.toBeVisible({ timeout: 3000 });
	}
}
