import { defineConfig } from 'vitest/config';
import { defaultClientConditions } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { resolve } from 'path';

export default defineConfig({
	plugins: [svelte()],
	test: {
		globals: true,
		environment: 'happy-dom',
		setupFiles: ['./src/tests/setup.ts'],
		include: ['src/**/*.{test,spec}.{js,ts}'],
		coverage: {
			provider: 'v8',
			reporter: ['text', 'json', 'html', 'lcov'],
			exclude: [
				'node_modules/',
				'src/tests/',
				'**/*.d.ts',
				'**/*.config.*',
				'**/mockData/',
				'build/',
				'.svelte-kit/'
			]
		}
	},
	resolve: {
		// Component tests mount into happy-dom, so Svelte must resolve to its
		// client build — without this condition it picks index-server.js and
		// every `render()` dies with "mount(...) is not available on the server".
		// Appended to Vite's defaults, not replacing them: a bare ['browser']
		// would drop `module`/`development|production` for every dependency.
		conditions: [...defaultClientConditions, 'browser'],
		alias: {
			$lib: resolve('./src/lib'),
			$app: resolve('./src/tests/mocks/$app')
		}
	}
});
