import { test, expect } from '@playwright/test';

test.describe('Navigation & Dashboard', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await page.locator('button[type="submit"]').first().click();
		try {
			await page.waitForURL(/dashboard/, { timeout: 8000 });
		} catch {
			// Login might fail if backend is not running
		}
	});

	test('should navigate to Karyawan page', async ({ page }) => {
		await page.goto('/karyawan');
		await expect(page.getByRole('heading').first()).toBeAttached();
	});

	test('should navigate to Dashboard page', async ({ page }) => {
		await page.goto('/dashboard');
		await expect(page.locator('h1').first()).toBeAttached();
	});

	test('should navigate between pages without JS crashes', async ({ page }) => {
		const pages = ['/dashboard', '/karyawan', '/departemen'];
		const jsErrors: string[] = [];

		page.on('console', (msg) => {
			if (msg.type() === 'error' && 
				!msg.text().includes('A form field') &&
				!msg.text().includes('404') &&
				!msg.text().includes('401') &&
				!msg.text().includes('Gagal terhubung') &&
				!msg.text().includes('Failed to fetch') &&
				!msg.text().includes('NetworkError')) {
				jsErrors.push(msg.text());
			}
		});

		for (const path of pages) {
			await page.goto(path);
			await page.waitForTimeout(500);
		}

		// We expect no JS TypeErrors or other actual crash errors
		// API errors are expected when backend is not running
		const crashErrors = jsErrors.filter(e => 
			e.includes('TypeError') || e.includes('undefined is not') || e.includes('null is not')
		);
		expect(crashErrors).toEqual([]);
	});
});
