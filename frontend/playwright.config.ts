import { defineConfig } from '@playwright/test';

export default defineConfig({
	testDir: './tests',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 1,
	workers: process.env.CI ? 1 : 2,
	reporter: 'html',
	use: {
		baseURL: 'http://localhost:5177',
		trace: 'on-first-retry',
	},
	projects: [
		{
			name: 'chromium',
			use: { browserName: 'chromium' },
		},
	],
	webServer: {
		command: 'npm run dev',
		url: 'http://localhost:5177',
		reuseExistingServer: !process.env.CI,
		timeout: 30000,
	},
});
