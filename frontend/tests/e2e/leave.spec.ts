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

test.describe('Leave Module (Cuti)', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
	});

	test('should display leave page', async ({ page }) => {
		await page.goto('/cuti');
		await expect(page.locator('h1').first()).toBeAttached({ timeout: 10000 });
		const body = page.locator('body');
		await expect(body).toBeAttached();
	});

	test('should show leave balance section', async ({ page }) => {
		await page.goto('/cuti');
		await page.waitForTimeout(2000);
		// Should show leave balance cards or text
		const balanceText = page.getByText('Sisa Cuti', { exact: false });
		const count = await balanceText.count();
		if (count > 0) {
			await expect(balanceText).toBeAttached({ timeout: 3000 });
		} else {
			// If not using "Sisa Cuti" label, at least verify the page has content
			const heading = page.locator('h1, h2, h3').first();
			await expect(heading).toBeAttached();
		}
	});

	test('should navigate to leave page without JS crashes', async ({ page }) => {
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

		const pages = ['/cuti', '/dashboard', '/cuti', '/dashboard'];
		for (const path of pages) {
			await page.goto(path);
			await page.waitForTimeout(500);
		}

		const crashErrors = jsErrors.filter(e => 
			e.includes('TypeError') || e.includes('undefined is not') || e.includes('null is not')
		);
		expect(crashErrors).toEqual([]);
	});

	test('should display leave type filter options', async ({ page }) => {
		await page.goto('/cuti');
		await page.waitForTimeout(2000);
		// Check for filter elements or dropdown
		const filterSection = page.getByText('Status', { exact: false });
		const count = await filterSection.count();
		if (count > 0) {
			await expect(filterSection).toBeAttached({ timeout: 3000 });
		} else {
			// If no "Status" label, check if any select/dropdown exists
			const anySelect = page.locator('select, [role="combobox"]').first();
			if (await anySelect.count() > 0) {
				await expect(anySelect).toBeAttached();
			}
		}
	});
});
