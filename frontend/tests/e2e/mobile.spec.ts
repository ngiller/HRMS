import { test, expect } from '@playwright/test';
import { loginWithToken } from './helpers.ts';

test.describe('Mobile Viewport (375px)', () => {
	test.beforeEach(async ({ page }) => {
		await page.setViewportSize({ width: 375, height: 812 });
		await loginWithToken(page);
	});

	test('bottom tab bar renders correctly', async ({ page }) => {
		await page.goto('/dashboard');
		await page.waitForLoadState('networkidle');

		// Tab bar should have 5 tabs on mobile
		const tabs = page.locator('nav[aria-label="Navigasi utama"] button');
		await expect(tabs.first()).toBeVisible();
		const tabCount = await tabs.count();
		expect(tabCount).toBeGreaterThanOrEqual(4); // At least 4 tabs (dashboard, absensi, pengajuan, menu)
	});

	test('navigation via bottom tab bar works', async ({ page }) => {
		await page.goto('/dashboard');
		await page.waitForLoadState('networkidle');

		// Click Attendance tab
		const tabButtons = page.locator('nav[aria-label="Navigasi utama"] button');
		const attendanceTab = tabButtons.filter({ hasText: 'Absensi' });
		if (await attendanceTab.count() > 0) {
			await attendanceTab.first().click();
			await page.waitForLoadState('networkidle');
			expect(page.url()).toContain('/absensi');
		}
	});

	test('menu drawer opens and shows items', async ({ page }) => {
		await page.goto('/dashboard');
		await page.waitForLoadState('networkidle');

		// Click Menu tab (last tab with grid dots)
		const menuButton = page.locator('nav[aria-label="Navigasi utama"] button[aria-label="Menu"]');
		await expect(menuButton).toBeVisible();
		await menuButton.click();

		// Menu drawer should appear
		const drawer = page.locator('text=Semua Menu');
		await expect(drawer).toBeVisible({ timeout: 3000 });

		// Should have menu items
		const menuItems = page.locator('a[href]');
		const itemCount = await menuItems.count();
		expect(itemCount).toBeGreaterThan(0);
	});

	test('leave page loads without crash on mobile', async ({ page }) => {
		await page.goto('/cuti');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText('Cuti', { timeout: 10000 });
	});

	test('overtime page loads without crash on mobile', async ({ page }) => {
		await page.goto('/lembur');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText('Lembur', { timeout: 10000 });
	});

	test('reimbursement page loads without crash on mobile', async ({ page }) => {
		await page.goto('/reimbursement');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText('Reimbursement', { timeout: 10000 });
	});

	test('manual attendance page loads without crash on mobile', async ({ page }) => {
		await page.goto('/absensi-manual');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText('Absensi Manual', { timeout: 10000 });
	});

	test('daily journal page loads without crash on mobile', async ({ page }) => {
		await page.goto('/jurnal-harian');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText('Jurnal Harian', { timeout: 10000 });
	});

	test('responsive layout — desktop hidden elements not shown on mobile', async ({ page }) => {
		await page.goto('/cuti');
		await page.waitForLoadState('networkidle');

		// AG Grid should be hidden on mobile (hidden md:block)
		const agGrid = page.locator('.ag-theme-quartz');
		await expect(agGrid).not.toBeVisible();
	});

	test('bottom tab bar menu closes drawer on item click', async ({ page }) => {
		await page.goto('/dashboard');
		await page.waitForLoadState('networkidle');

		// Open menu drawer
		const menuButton = page.locator('nav[aria-label="Navigasi utama"] button[aria-label="Menu"]');
		await menuButton.click();

		// Wait for drawer to appear
		const drawer = page.locator('text=Semua Menu');
		await expect(drawer).toBeVisible({ timeout: 3000 });

		// Click a menu item (e.g., Absensi or first available)
		const firstMenuItem = page.locator('a[href]').first();
		await firstMenuItem.click();

		// Drawer should close
		await expect(drawer).not.toBeVisible({ timeout: 3000 });
	});
});
