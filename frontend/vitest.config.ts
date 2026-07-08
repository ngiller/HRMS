import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';

export default defineConfig({
	plugins: [svelte({ hot: !process.env.VITEST })],
	test: {
		environment: 'jsdom',
		globals: true,
		include: ['src/**/*.{test,spec}.{js,ts}'],
		exclude: ['tests/e2e/**', 'node_modules/**'],
		setupFiles: [],
	},
	resolve: {
		conditions: ['browser'],
		alias: {
			$lib: path.resolve(__dirname, 'src/lib'),
		},
	},
});
