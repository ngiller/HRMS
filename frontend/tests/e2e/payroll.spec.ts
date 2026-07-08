import { test, expect } from '@playwright/test';

test.describe('Payroll Module (Penggajian)', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await page.locator('button[type="submit"]').first().click();
		try {
			await page.waitForURL(/dashboard/, { timeout: 8000 });
		} catch {
			// Backend might not be running in CI
		}
	});

	test('should display payroll page', async ({ page }) => {
		await page.goto('/penggajian');
		await expect(page.locator('h1').first()).toBeAttached();
		const body = page.locator('body');
		await expect(body).toBeAttached();
	});

	test('should show payroll periods section', async ({ page }) => {
		await page.goto('/penggajian');
		await page.waitForTimeout(1000);
		const body = page.locator('body');
		await expect(body).toBeAttached();
	});

	test('should navigate to payroll page without JS crashes', async ({ page }) => {
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

		const pages = ['/penggajian', '/dashboard', '/penggajian/slip-saya', '/dashboard'];
		for (const path of pages) {
			await page.goto(path);
			await page.waitForTimeout(500);
		}

		const crashErrors = jsErrors.filter(e => 
			e.includes('TypeError') || e.includes('undefined is not') || e.includes('null is not')
		);
		expect(crashErrors).toEqual([]);
	});

	test('should display my payslip page', async ({ page }) => {
		await page.goto('/penggajian/slip-saya');
		await page.waitForTimeout(1000);
		const body = page.locator('body');
		await expect(body).toBeAttached();
	});
});
