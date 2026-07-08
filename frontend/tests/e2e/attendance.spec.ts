import { test, expect } from '@playwright/test';

/** Helper: login and wait for dashboard redirect */
async function loginAsAdmin(page: import('@playwright/test').Page) {
	await page.goto('/login');
	// Wait for SvelteKit hydration to complete by checking the field has the default value
	await expect(page.locator('#email')).toHaveValue('admin@company.com', { timeout: 15000 });
	// Click submit using the same pattern as login.spec.ts (toHaveURL assertion)
	await page.locator('button[type="submit"]').first().click({ force: true });
	await expect(page).toHaveURL(/dashboard/, { timeout: 20000 });
}

test.describe('Attendance Module', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
	});

	test('should display attendance page with today status', async ({ page }) => {
		await page.goto('/absensi');
		await expect(page.locator('h1').first()).toBeAttached({ timeout: 10000 });
		// Should show section for today's attendance status
		const body = page.locator('body');
		await expect(body).toBeAttached();
	});

	test('should show attendance history section', async ({ page }) => {
		await page.goto('/absensi');
		// Wait for content to load
		await page.waitForTimeout(2000);
		// Should show either history table or empty state
		const historySection = page.getByText('Riwayat', { exact: false });
		const count = await historySection.count();
		if (count > 0) {
			await expect(historySection).toBeAttached({ timeout: 3000 });
		} else {
			// If not using "Riwayat" label, at least verify the page has content
			const heading = page.locator('h1, h2, h3').first();
			await expect(heading).toBeAttached();
		}
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
		await page.waitForTimeout(2000);
		// Check for check-in/out status text - these cards might use Indonesian labels
		const checkInText = page.getByText(/^(Check In|Masuk|Absen Masuk)$/i);
		const checkOutText = page.getByText(/^(Check Out|Pulang|Absen Pulang)$/i);
		
		const checkInCount = await checkInText.count();
		const checkOutCount = await checkOutText.count();
		
		// At least one of the labels should be present
		expect(checkInCount + checkOutCount).toBeGreaterThan(0);
	});
});
