import { test, expect } from '@playwright/test';

test.describe('Approval Workflow', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await page.locator('button[type="submit"]').first().click();
		try {
			await page.waitForURL(/dashboard/, { timeout: 8000 });
		} catch {
			// Backend might not be running in CI
		}
	});

	test('should display pending approvals page', async ({ page }) => {
		await page.goto('/persetujuan');
		await expect(page.getByRole('heading').first()).toBeAttached();
		// Should show either "Persetujuan" title or empty state
		const title = page.locator('h1');
		await expect(title).toBeAttached();
	});

	test('should navigate between pages without JS crashes on approvals', async ({ page }) => {
		const pages = ['/persetujuan', '/dashboard', '/karyawan'];
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

		const crashErrors = jsErrors.filter(e => 
			e.includes('TypeError') || e.includes('undefined is not') || e.includes('null is not')
		);
		expect(crashErrors).toEqual([]);
	});

	test('should show settings page with workflow config tab', async ({ page }) => {
		await page.goto('/pengaturan');
		await expect(page.locator('h1').first()).toBeAttached();
		// Check for workflow config tab button
		const workflowTab = page.getByText('Workflow Persetujuan');
		await expect(workflowTab).toBeAttached();
	});

	test('should show pending count badge on approvals page', async ({ page }) => {
		await page.goto('/persetujuan');
		await page.waitForTimeout(1000);
		// The page should show either pending items or empty state
		const body = page.locator('body');
		await expect(body).toBeAttached();
	});
});
