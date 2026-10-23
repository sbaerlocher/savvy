import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	define: {
		__APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.8.1')
	},
	build: {
		chunkSizeWarningLimit: 1000,
		// Enable minification and tree-shaking optimizations
		minify: 'esbuild',
		// Reduce build time by disabling source maps in production
		sourcemap: false,
		// Target modern browsers for smaller output
		target: 'es2020'
	},
	// Include WASM files as assets (needed for barcode-detector WASM polyfill)
	assetsInclude: ['**/*.wasm'],
	optimizeDeps: {
		// Exclude WASM modules from optimization to preserve dynamic imports
		exclude: ['barcode-detector']
	},
	plugins: [tailwindcss(), sveltekit()],
	server: {
		port: 5173,
		// Leading dot covers subdomains: dde >= v2.0.0-beta.2 serves git
		// worktrees on `<suffix>.savvy.test`.
		allowedHosts: ['.savvy.test'],
		proxy: {
			'/api': {
				target: 'http://api:8080', // Docker: api container on port 8080
				changeOrigin: true
			},
			'/auth': {
				target: 'http://api:8080', // OAuth routes
				changeOrigin: true
			},
			'/health': {
				target: 'http://api:8080', // Health check
				changeOrigin: true
			},
			'/ready': {
				target: 'http://api:8080', // Readiness check
				changeOrigin: true
			},
			'/metrics': {
				target: 'http://api:8080', // Prometheus metrics
				changeOrigin: true
			}
		}
	}
});
