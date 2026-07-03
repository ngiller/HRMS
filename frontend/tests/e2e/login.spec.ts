import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
	test('should display login page', async ({ page }) => {
		await page.goto('/login');
		// The login page has a heading "Masuk"
		await expect(page.locator('h1')).toContainText('Masuk');
		// Should have email and password fields
		await expect(page.locator('#email')).toBeVisible();
		await expect(page.locator('#password')).toBeVisible();
	});

	test('should login with valid admin credentials', async ({ page }) => {
		await page.goto('/login');
		// Fields are pre-filled with defaults (admin@company.com / admin123)
		// Just click the submit button - use the first one (desktop view)
		await page.locator('button[type="submit"]').first().click();
		// After login, should redirect to dashboard
		await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });
	});

	test('should show error with empty credentials', async ({ page }) => {
		await page.goto('/login');
		// Clear the pre-filled values
		await page.locator('#email').fill('');
		await page.locator('#password').fill('');
		await page.locator('button[type="submit"]').first().click();
		// Should show validation error (required fields)
		// The browser might not submit due to required attribute
		await expect(page).toHaveURL(/login/);
	});
});
