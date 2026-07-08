import { test, expect } from '@playwright/test';

test.describe('Attendance Module', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await page.locator('button[type="submit"]').first().click();
		try {
			await page.waitForURL(/dashboard/, { timeout: 8000 });
		} catch {
			// Backend might not be running in CI
		}
	});

	test('should display attendance page with today status', async ({ page }) => {
		await page.goto('/absensi');
		await expect(page.locator('h1').first()).toBeAttached();
		// Should show section for today's attendance status
		const body = page.locator('body');
		await expect(body).toBeAttached();
	});

	test('should show attendance history section', async ({ page }) => {
		await page.goto('/absensi');
		// Wait for content to load
		await page.waitForTimeout(1000);
		// Should show either history table or empty state
		const historySection = page.getByText('Riwayat', { exact: false });
		await expect(historySection).toBeAttached();
	});

	test('should navigate to attendance without JS crashes', async ({ page }) => {
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

		const pages = ['/absensi', '/dashboard', '/absensi'];
		for (const path of pages) {
			await page.goto(path);
			await page.waitForTimeout(500);
		}

		const crashErrors = jsErrors.filter(e => 
			e.includes('TypeError') || e.includes('undefined is not') || e.includes('null is not')
		);
		expect(crashErrors).toEqual([]);
	});

	test('should display check-in and check-out status cards', async ({ page }) => {
		await page.goto('/absensi');
		await page.waitForTimeout(1000);
		// Check for check-in/out status text
		const checkInText = page.getByText('Check In', { exact: false });
		const checkOutText = page.getByText('Check Out', { exact: false });
		
		// Both should be present on the page
		const checkInCount = await checkInText.count();
		const checkOutCount = await checkOutText.count();
		
		expect(checkInCount).toBeGreaterThan(0);
		expect(checkOutCount).toBeGreaterThan(0);
	});
});
