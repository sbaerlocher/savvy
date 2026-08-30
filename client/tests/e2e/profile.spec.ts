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

			// Profile fields should be visible (Android discloses them via the
			// name row first).
			await profilePage.openNameEditor();
			await expect(profilePage.firstNameInput).toBeVisible({ timeout: 5000 });
			await expect(profilePage.lastNameInput).toBeVisible({ timeout: 5000 });

			// Email: a read-only input on the inline layouts, a plain M3 row on
			// Android — wait for either presentation, then assert the one shown.
			await expect(
				profilePage.emailInput
					.or(profilePage.page.getByText(TEST_USERS.regular.email))
					.first()
			).toBeVisible({ timeout: 10000 });
			if (await profilePage.emailInput.isVisible().catch(() => false)) {
				await expect(profilePage.emailInput).toBeDisabled();
				const emailValue = await profilePage.emailInput.inputValue();
				expect(emailValue).toContain(TEST_USERS.regular.email);
			} else {
				await expect(
					profilePage.page.getByText(TEST_USERS.regular.email).first()
				).toBeVisible({ timeout: 5000 });
			}
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

			// Verify the update stuck. Android closes the editor on save and shows
			// the name on the settings row; the inline layouts keep the fields.
			// Probe the text first — the closing editor makes an input probe racy.
			const updatedName = profilePage.page
				.getByText(`${newFirstName} ${newLastName}`)
				.first();
			if (await updatedName.isVisible({ timeout: 2000 }).catch(() => false)) {
				await expect(updatedName).toBeVisible();
				// The editor collapses on save — wait that out before the restore
				// below re-opens it, or the toggle races the close animation.
				await expect(profilePage.firstNameInput).toBeHidden({
					timeout: 5000
				});
			} else {
				expect(await profilePage.firstNameInput.inputValue()).toBe(
					newFirstName
				);
				expect(await profilePage.lastNameInput.inputValue()).toBe(newLastName);
			}

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

			// Should show the email address: a read-only #email input on the
			// inline layouts, a plain M3 row on Android. Wait for either
			// presentation, then assert the one shown.
			await expect(
				profilePage.emailInput
					.or(profilePage.page.getByText(TEST_USERS.regular.email))
					.first()
			).toBeVisible({ timeout: 10000 });
			if (await profilePage.emailInput.isVisible().catch(() => false)) {
				await expect(profilePage.emailInput).toHaveValue(
					TEST_USERS.regular.email
				);
			} else {
				await expect(
					profilePage.page.getByText(TEST_USERS.regular.email).first()
				).toBeVisible({ timeout: 5000 });
			}

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

		// GDPR / data-portability: the export is advertised as containing all of a
		// user's data. This asserts the payload actually carries every resource
		// category with real content (not just that a download happens), so a
		// regression that drops a category from the export is caught. The regular
		// seed user (Anna) owns cards, vouchers, gift cards with transactions, and
		// favorites, which keeps the assertions non-vacuous.
		test('export contains all user data categories', async ({
			authenticatedProfilePage,
			profilePage
		}) => {
			await profilePage.waitForPageReady();

			// Reuse the authenticated page's session cookie for the API call.
			const response = await profilePage.page.request.get('/api/v1/export');
			expect(response.ok()).toBeTruthy();

			const data = await response.json();

			// Top-level structure
			expect(data).toHaveProperty('exported_at');
			expect(data.user?.email).toBe(TEST_USERS.regular.email);

			// Every resource category present and non-empty for the seed user
			expect(Array.isArray(data.cards)).toBe(true);
			expect(data.cards.length).toBeGreaterThan(0);
			expect(Array.isArray(data.vouchers)).toBe(true);
			expect(data.vouchers.length).toBeGreaterThan(0);
			expect(Array.isArray(data.gift_cards)).toBe(true);
			expect(data.gift_cards.length).toBeGreaterThan(0);
			expect(Array.isArray(data.favorites)).toBe(true);
			expect(data.favorites.length).toBeGreaterThan(0);

			// Card content is real, not placeholder rows
			expect(data.cards[0]).toHaveProperty('merchant_name');
			expect(data.cards[0].merchant_name).toBeTruthy();
			expect(data.cards[0]).toHaveProperty('card_number');

			// Gift card transactions are nested and at least one gift card carries them
			const giftCardWithTx = data.gift_cards.find(
				(gc: { transactions?: unknown[] }) =>
					Array.isArray(gc.transactions) && gc.transactions.length > 0
			);
			expect(giftCardWithTx).toBeTruthy();
			expect(giftCardWithTx.transactions[0]).toHaveProperty('amount');

			// Favorites reference resources by type + id
			expect(data.favorites[0]).toHaveProperty('resource_type');
			expect(data.favorites[0]).toHaveProperty('resource_id');
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

			// For local auth users, password is also required
			const passwordInput = profilePage.deletePasswordInput;
			if (await passwordInput.isVisible({ timeout: 1000 }).catch(() => false)) {
				await passwordInput.fill(TEST_USERS.regular.password);
			}

			// Button should now be enabled (or at least not disabled)
			await expect(confirmBtn).toBeEnabled({ timeout: 3000 });

			// Close without actually deleting
			await profilePage.page.keyboard.press('Escape');
		});
	});
});
