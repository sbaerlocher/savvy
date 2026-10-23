import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			precompress: false,
			strict: true
		}),
		// Registered manually in $lib/pwa/register-sw.ts (availability check, update toast)
		serviceWorker: {
			register: false
		},
		alias: {
			$lib: 'src/lib',
			$components: 'src/lib/components'
		}
	}
};

export default config;
