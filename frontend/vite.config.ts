import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import Icons from 'unplugin-icons/vite';
export default defineConfig(async () => {
	const plugins = [
		tailwindcss(),
		sveltekit(),
		Icons({
			compiler: 'svelte',
			iconCustomizer(_collection, _icon, props) {
				// Default all icons to 20px solid variant
				props.width = '20px';
				props.height = '20px';
			}
		}),
	];

	if (process.env.VITE_ANALYZE) {
		const { visualizer } = await import('rollup-plugin-visualizer');
		plugins.push(visualizer({
			filename: 'bundle-stats.html',
			open: false,
			gzipSize: true,
			brotliSize: true,
			template: 'treemap',
		}));
	}

	return { 
		plugins,
		server: {
			host: true,
		}
	};
});
