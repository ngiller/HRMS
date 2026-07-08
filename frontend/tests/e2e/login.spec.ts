import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
	test('should display login page', async ({ page }) => {
		await page.goto('/login');
		// Wait for the page to fully render
		await expect(page.locator('#email')).toBeVisible({ timeout: 10000 });
		await expect(page.locator('h1')).toContainText('Masuk');
		// Should have email and password fields
		await expect(page.locator('#password')).toBeVisible();
	});

	test('should login with valid admin credentials', async ({ page }) => {
		await page.goto('/login');
		// Wait for page to fully load before interacting
		await expect(page.locator('button[type="submit"]').first()).toBeVisible({ timeout: 10000 });
		// Fields are pre-filled with defaults (admin@company.com / admin123)
		// Verify the fields have values before clicking submit
		await expect(page.locator('#email')).toHaveValue('admin@company.com');
		// Click the submit button
		await page.locator('button[type="submit"]').first().click();
		// After login, should redirect to dashboard
		await expect(page).toHaveURL(/dashboard/, { timeout: 15000 });
	});

	test('should show error with empty credentials', async ({ page }) => {
		await page.goto('/login');
		await expect(page.locator('#email')).toBeVisible({ timeout: 10000 });
		// Clear the pre-filled values
		await page.locator('#email').fill('');
		await page.locator('#password').fill('');
		await page.locator('button[type="submit"]').first().click();
		// Should show validation error (required fields)
		// The browser might not submit due to required attribute
		await expect(page).toHaveURL(/login/);
	});
});
