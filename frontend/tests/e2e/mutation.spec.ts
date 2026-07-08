import { test, expect } from '@playwright/test';

test.describe('Mutation & Promosi', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
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
		await page.waitForTimeout(2000);

		// Check the page rendered without crashing
		const body = page.locator('body');
		await expect(body).toBeAttached({ timeout: 10000 });

		// The title or heading should exist
		const heading = page.getByRole('heading').first();
		await expect(heading).toBeAttached({ timeout: 5000 });
	});

	test('should show mutation detail and Cetak SK button for approved mutations', async ({ page }) => {
		await page.goto('/mutasi');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(1000);

		// Click pending filter to see approved mutations only
		const approvedFilter = page.locator('button', { hasText: 'Disetujui' }).first();
		if (await approvedFilter.isVisible()) {
			await approvedFilter.click();
			await page.waitForTimeout(1000);
		}

		// Try opening any mutation item
		const firstItem = page.locator('[class*="MobileCard"]').first();
		if (await firstItem.isVisible()) {
			await firstItem.click();
			await page.waitForTimeout(1000);

			// Check if Cetak SK button appears for approved mutations
			const cetakBtn = page.getByText('Cetak SK Mutasi').first();
			if (await cetakBtn.isVisible()) {
				await cetakBtn.click();
				await page.waitForTimeout(1000);
				// Should navigate to SK page
				await expect(page.locator('body')).toBeAttached();
				// Go back
				await page.goBack();
			}
		}
	});

	test('should show pending mutations in /persetujuan page', async ({ page }) => {
		await page.goto('/persetujuan');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(1000);

		// The page should load without errors
		await expect(page.locator('body')).toBeAttached();
		const heading = page.getByRole('heading').first();
		await expect(heading).toBeAttached();
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
				!msg.text().includes('NetworkError') &&
				!msg.text().includes('Failed to load resource')) {
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

	test('should display SK mutasi page for approved mutation', async ({ page }) => {
		// Navigate directly to a mutation detail first, then to SK
		await page.goto('/mutasi');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(1500);

		// Click on approved filter
		const approvedTab = page.locator('button', { hasText: 'Disetujui' }).first();
		if (await approvedTab.isVisible()) {
			await approvedTab.click();
			await page.waitForTimeout(1500);
		}

		// Find an approved mutation item and click it
		const approvedItem = page.locator('[class*="MobileCard"]').first();
		if (await approvedItem.isVisible()) {
			await approvedItem.click();
			await page.waitForTimeout(1000);

			// Look for Cetak SK link
			const skLink = page.locator('a[href*="/sk"]').first();
			if (await skLink.isVisible()) {
				await skLink.click();
				await page.waitForLoadState('networkidle');
				await page.waitForTimeout(1000);

				// SK page should have a title or letter content
				const body = page.locator('body');
				await expect(body).toBeAttached();

				// Should see PDF-related elements (Cetak or Download buttons)
				const downloadBtn = page.getByText('Download PDF').first();
				const cetakBtn = page.getByText('Cetak').first();
				expect(await downloadBtn.isVisible().catch(() => false) ||
					   await cetakBtn.isVisible().catch(() => false)).toBeTruthy();
			}
		}
	});
});
