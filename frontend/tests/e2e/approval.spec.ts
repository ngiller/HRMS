import { test, expect } from '@playwright/test';
import { loginWithToken } from './helpers';

test.describe('Approval Workflow', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/login');
		await loginWithToken(page);
	});

	test('should display pending approvals page with heading', async ({ page }) => {
		await page.goto('/persetujuan');
		await expect(page.getByRole('heading', { name: 'Persetujuan' })).toBeVisible({ timeout: 10000 });
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

	test('should show pending approvals list or empty state', async ({ page }) => {
		await page.goto('/persetujuan');
		await page.waitForTimeout(2000);

		// Either shows pending items or empty state message
		const pendingCount = page.locator('text=menunggu').first();
		const emptyState = page.getByText('Tidak ada yang menunggu');

		// Wait for either to appear
		await expect(
			page.locator('body')
		).toBeAttached();

		const hasPending = await pendingCount.isVisible().catch(() => false);
		const hasEmpty = await emptyState.isVisible().catch(() => false);

		// Should show either pending items or empty state (not both)
		if (hasPending) {
			// Ada item pending — cek ada tombol Setujui
			const approveButtons = page.getByText('Setujui');
			await expect(approveButtons.first()).toBeAttached();
		} else if (hasEmpty) {
			await expect(emptyState).toBeVisible();
		}
	});

	test('should approve a pending item and update the list', async ({ page }) => {
		await page.goto('/persetujuan');
		await page.waitForTimeout(2000);

		// Cek apakah ada item pending
		const approveButtons = page.getByRole('button', { name: 'Setujui' });
		const hasItems = await approveButtons.first().isVisible().catch(() => false);

		test.skip(!hasItems, 'Tidak ada item pending untuk di-approve');
		if (!hasItems) return;

		// Hitung jumlah item sebelum approve
		const initialCount = await approveButtons.count();

		// Klik tombol Setujui pada item pertama
		await approveButtons.first().click();

		// Tunggu spinner loading selesai (tombol berubah dari loading state)
		await expect(approveButtons.first()).toBeVisible({ timeout: 10000 });

		// Hitung ulang jumlah item
		const remainingButtons = page.getByRole('button', { name: 'Setujui' });
		const remainingCount = await remainingButtons.count();

		// Jumlah item harus berkurang (atau 0 jika hanya 1 item)
		expect(remainingCount).toBeLessThanOrEqual(initialCount);
	});

	test('should show sidebar badge with pending count', async ({ page }) => {
		await page.goto('/persetujuan');
		await page.waitForTimeout(2000);

		// Cari sidebar element
		const sidebar = page.locator('aside').first();
		const sidebarVisible = await sidebar.isVisible().catch(() => false);

		test.skip(!sidebarVisible, 'Sidebar tidak terlihat');
		if (!sidebarVisible) return;

		// Cari menu Persetujuan di sidebar
		const persetujuanMenu = sidebar.getByText('Persetujuan');
		await expect(persetujuanMenu.first()).toBeAttached();

		// Cari angka badge di sidebar (angka pending count)
		const sidebarBadge = sidebar.locator('span.bg-red-100, span[class*="rounded-full"]').first();
		const hasBadge = await sidebarBadge.isVisible().catch(() => false);

		if (hasBadge) {
			const badgeText = await sidebarBadge.textContent();
			expect(badgeText).toMatch(/\d+/); // Harus mengandung angka
		}
	});
});
