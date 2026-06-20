import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import svelteParser from 'svelte-eslint-parser';
import tseslint from 'typescript-eslint';

export default [
	js.configs.recommended,
	...tseslint.configs.recommended,
	...svelte.configs['flat/recommended'],
	{
		languageOptions: {
			globals: { ...globals.browser, ...globals.node }
		},
		rules: {
			'@typescript-eslint/no-unused-vars': [
				'error',
				{ argsIgnorePattern: '^_' }
			],
			'@typescript-eslint/no-explicit-any': 'warn'
		}
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
		languageOptions: {
			parser: svelteParser,
			parserOptions: { parser: tseslint.parser }
		}
	},
	{
		// Playwright fixtures are injected by destructuring the test args
		// (e.g. `async ({ authenticatedPage }) => …`), which runs the fixture's
		// setup even when the value isn't referenced in the body. Don't flag
		// those as unused, but still catch genuinely unused locals.
		files: ['tests/**/*.ts'],
		rules: {
			'@typescript-eslint/no-unused-vars': [
				'error',
				{ args: 'none', varsIgnorePattern: '^_' }
			]
		}
	},
	{
		ignores: ['build/', '.svelte-kit/', 'dist/', 'node_modules/', 'static/']
	}
];
