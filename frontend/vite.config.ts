import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import Icons from 'unplugin-icons/vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		Icons({
			compiler: 'svelte',
			iconCustomizer(_collection, _icon, props) {
				// Default all icons to 20px solid variant
				props.width = '20px';
				props.height = '20px';
			}
		})
	]
});
