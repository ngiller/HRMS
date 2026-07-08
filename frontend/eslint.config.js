import js from '@eslint/js';
import ts from 'typescript-eslint';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';

export default ts.config(
	// 1. Base ESLint recommended rules
	js.configs.recommended,

	// 2. TypeScript recommended rules
	...ts.configs.recommended,

	// 3. Svelte recommended rules
	...svelte.configs['flat/recommended'],

	// 4. Global settings
	{
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node,
			},
		},
	},

	// 5. Svelte + TypeScript config
	{
		files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
		languageOptions: {
			parserOptions: {
				projectService: true,
				extraFileExtensions: ['.svelte'],
				parser: ts.parser,
			},
		},
	},

	// 6. Ignore build output and node_modules
	{
		ignores: ['build/', '.svelte-kit/', 'node_modules/'],
	},

	// 7. Custom rule overrides
	{
		rules: {
			// Allow console.log for debugging
			'no-console': 'off',
			// TypeScript-specific overrides
			'@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
			'@typescript-eslint/no-explicit-any': 'warn',
		},
	},
);
