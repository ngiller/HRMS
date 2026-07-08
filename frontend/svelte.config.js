import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		warningFilter: (warning) => {
			// Suppress A11y warnings — deliberately clean UI patterns
			if (warning.code && warning.code.startsWith('a11y_')) {
				return false;
			}
			return true;
		}
	},
	kit: {
		adapter: adapter({
			fallback: 'index.html',
			pages: 'build',
			assets: 'build'
		})
	}
};

export default config;
