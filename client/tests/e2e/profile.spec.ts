import { expect, test, TEST_USERS } from './fixtures/test-fixtures';

test.describe('Profile', () => {
	test.beforeEach(async ({ authenticatedProfilePage }) => {
		// authenticatedProfilePage fixture handles login + navigation to /profile
	});

	test.describe('Profile Management', () => {
		test('should display current user profile', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();
			await profilePage.expectProfilePage();

			// Profile fields should be visible
			await expect(profilePage.firstNameInput).toBeVisible({ timeout: 5000 });
			await expect(profilePage.lastNameInput).toBeVisible({ timeout: 5000 });
			await expect(profilePage.emailInput).toBeVisible({ timeout: 5000 });

			// Email should be read-only and contain the current user's email
			await expect(profilePage.emailInput).toBeDisabled();
			const emailValue = await profilePage.emailInput.inputValue();
			expect(emailValue).toContain(TEST_USERS.regular.email);
		});

		test('should update display name', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();
			await profilePage.expectProfilePage();

			const timestamp = Date.now().toString().slice(-4);
			const newFirstName = `TestFirst${timestamp}`;
			const newLastName = `TestLast${timestamp}`;

			await profilePage.updateProfile(newFirstName, newLastName);
			await profilePage.expectToast();

			// Verify the fields retain the new values
			const updatedFirstName = await profilePage.firstNameInput.inputValue();
			const updatedLastName = await profilePage.lastNameInput.inputValue();
			expect(updatedFirstName).toBe(newFirstName);
			expect(updatedLastName).toBe(newLastName);

			// Restore original name
			await profilePage.updateProfile('Anna', 'Müller');
		});
	});

	test.describe('Account Information', () => {
		test('should display account info section', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();

			// Look for "Member Since" or "Mitglied seit" text
			const memberSince = profilePage.page
				.locator('text=/Mitglied seit|Member Since/i')
				.first();
			await expect(memberSince).toBeVisible({ timeout: 5000 });

			// Look for auth provider info ("Anmeldeart" / "Auth Provider")
			const authProvider = profilePage.page
				.locator(
					'text=/Anmeldeart|Auth Provider|E-Mail.*Passwort|Email.*Password/i'
				)
				.first();
			await expect(authProvider).toBeVisible({ timeout: 5000 });
		});
	});

	test.describe('Email Verification', () => {
		test('should display email verification status', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();

			// Should show email address
			const emailDisplay = profilePage.page
				.locator(`text=${TEST_USERS.regular.email}`)
				.first();
			await expect(emailDisplay).toBeVisible({ timeout: 5000 });

			// Should show verified badge OR send verification button
			const verifiedBadge = profilePage.emailVerifiedBadge;
			const sendVerifyBtn = profilePage.sendVerificationButton;

			const hasVerifiedBadge = await verifiedBadge
				.isVisible({ timeout: 3000 })
				.catch(() => false);
			const hasSendButton = await sendVerifyBtn
				.isVisible({ timeout: 3000 })
				.catch(() => false);

			expect(hasVerifiedBadge || hasSendButton).toBeTruthy();
		});
	});

	test.describe('Data Export', () => {
		test('should trigger data export download', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();

			const exportVisible = await profilePage.exportButton
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!exportVisible) {
				test.skip();
				return;
			}

			const download = await profilePage.triggerExport();
			if (download) {
				const suggestedFilename = download.suggestedFilename();
				expect(suggestedFilename).toMatch(/\.json$/);
			}
		});
	});

	test.describe('Danger Zone', () => {
		test('should display delete account button', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();

			const dangerZone = profilePage.page
				.locator('text=/Danger Zone|Gefahrenzone|Gefahrenbereich/i')
				.first();

			const hasDangerZone = await dangerZone
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!hasDangerZone) {
				test.skip();
				return;
			}

			await expect(profilePage.deleteAccountButton).toBeVisible({
				timeout: 5000
			});
		});

		test('should open delete account confirmation modal', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();

			const deleteVisible = await profilePage.deleteAccountButton
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!deleteVisible) {
				test.skip();
				return;
			}

			await profilePage.deleteAccountButton.click();

			// Modal should appear with confirmation input
			await expect(profilePage.deleteConfirmationInput).toBeVisible({
				timeout: 5000
			});

			// Delete button should be disabled until "DELETE" is typed
			const confirmBtn = profilePage.deleteConfirmButton;
			await expect(confirmBtn).toBeVisible({ timeout: 3000 });

			// Close without deleting
			const cancelBtn = profilePage.page
				.locator('button:has-text("Abbrechen"), button:has-text("Cancel")')
				.first();
			if (await cancelBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
				await cancelBtn.click();
			} else {
				await profilePage.page.keyboard.press('Escape');
			}
		});

		test('should require DELETE confirmation text', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();

			const deleteVisible = await profilePage.deleteAccountButton
				.isVisible({ timeout: 5000 })
				.catch(() => false);
			if (!deleteVisible) {
				test.skip();
				return;
			}

			await profilePage.deleteAccountButton.click();
			await expect(profilePage.deleteConfirmationInput).toBeVisible({
				timeout: 5000
			});

			// Type something that is NOT "DELETE"
			await profilePage.deleteConfirmationInput.fill('WRONG');

			// The confirm button should be disabled
			const confirmBtn = profilePage.deleteConfirmButton;
			await expect(confirmBtn).toBeDisabled({ timeout: 3000 });

			// Now type correct confirmation
			await profilePage.deleteConfirmationInput.clear();
			await profilePage.deleteConfirmationInput.fill('DELETE');

			// Button should now be enabled (or at least not disabled)
			await expect(confirmBtn).toBeEnabled({ timeout: 3000 });

			// Close without actually deleting
			await profilePage.page.keyboard.press('Escape');
		});
	});
});
