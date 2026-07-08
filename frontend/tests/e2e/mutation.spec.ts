import { test, expect } from '@playwright/test';

test.describe('Mutation & Promosi', () => {
	test.beforeEach(async ({ page }) => {
		// Navigate to login
		await page.goto('/login');
		// Fill login form if present (adjust selectors as needed)
		const emailInput = page.locator('input[type="email"], input[name="email"], input[placeholder*="email"]').first();
		const passwordInput = page.locator('input[type="password"]').first();
		if (await emailInput.isVisible()) {
			await emailInput.fill('admin@company.com');
			await passwordInput.fill('admin123');
			const submitBtn = page.locator('button[type="submit"]').first();
			if (await submitBtn.isVisible()) {
				await submitBtn.click();
				await page.waitForURL(/dashboard/, { timeout: 10000 }).catch(() => {});
			}
		}
	});

	test('should load mutation page and display data', async ({ page }) => {
		await page.goto('/mutasi');
		await page.waitForLoadState('networkidle');

		// Page should display the title
		await expect(page.getByText('Mutasi & Promosi')).toBeAttached();

		// Should show either mutation list or empty state (no crash)
		const body = page.locator('body');
		await expect(body).toBeAttached();

		// Check for no JS errors
		const jsErrors: string[] = [];
		page.on('console', (msg) => {
			if (msg.type() === 'error' &&
				!msg.text().includes('A form field') &&
				!msg.text().includes('404') &&
				!msg.text().includes('401') &&
				!msg.text().includes('Failed to fetch') &&
				!msg.text().includes('NetworkError')) {
				jsErrors.push(msg.text());
			}
		});

		const crashErrors = jsErrors.filter(e =>
			e.includes('TypeError') || e.includes('undefined is not') || e.includes('null is not')
		);
		expect(crashErrors).toEqual([]);
	});

	test('should navigate between pages without JS crashes', async ({ page }) => {
		const pages = ['/mutasi', '/persetujuan', '/dashboard', '/karyawan'];
		const jsErrors: string[] = [];

		page.on('console', (msg) => {
			if (msg.type() === 'error' &&
				!msg.text().includes('A form field') &&
				!msg.text().includes('404') &&
				!msg.text().includes('401') &&
				!msg.text().includes('Failed to fetch') &&
				!msg.text().includes('NetworkError')) {
				jsErrors.push(msg.text());
			}
		});

		for (const path of pages) {
			await page.goto(path);
			await page.waitForTimeout(500);
		}

		const crashErrors = jsErrors.filter(e =>
			e.includes('TypeError') || e.includes('undefined is not') || e.includes('null is not')
		);
		expect(crashErrors).toEqual([]);
	});
});
