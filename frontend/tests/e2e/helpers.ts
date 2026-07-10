import { Page } from '@playwright/test';

const API_BASE_URL = 'http://localhost:8900';

/**
 * Login by injecting token directly via API + localStorage.
 * This bypasses form submission which is unreliable in headless Playwright.
 */
export async function loginWithToken(page: Page) {
	const result = await page.evaluate(async (apiBaseUrl) => {
		try {
			const res = await fetch(`${apiBaseUrl}/api/auth/login`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email: 'admin@company.com', password: 'admin123' }),
			});
			const data = await res.json();
			if (data.success && data.data) {
				const { access_token, refresh_token, user } = data.data;
				if (access_token) localStorage.setItem('hrms_access_token', access_token);
				if (refresh_token) localStorage.setItem('hrms_refresh_token', refresh_token);
				if (user) localStorage.setItem('hrms_user', JSON.stringify(user));
				return { success: true };
			}
			return { success: false, error: data.message };
		} catch (err) {
			return { success: false, error: String(err) };
		}
	}, API_BASE_URL);

	return result;
}
