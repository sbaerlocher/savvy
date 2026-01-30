// rollup.config.js
import commonjs from '@rollup/plugin-commonjs';
import resolve from '@rollup/plugin-node-resolve';
import terser from '@rollup/plugin-terser';
import { visualizer } from 'rollup-plugin-visualizer';
import gzipPlugin from 'rollup-plugin-gzip';
import { brotliCompressSync } from 'zlib';
import path from 'path';

const isDev = process.env.NODE_ENV !== 'production';
const outputDir = 'internal/assets/static/js';

// Plugin to generate manifest.json for cache-busted filenames
function manifestPlugin() {
	return {
		name: 'manifest',
		generateBundle(_options, bundle) {
			const manifest = {};

			for (const [fileName, chunkInfo] of Object.entries(bundle)) {
				if (chunkInfo.type === 'chunk') {
					// Map original name to hashed name
					const originalName = chunkInfo.facadeModuleId
						? path.basename(chunkInfo.facadeModuleId, '.js')
						: chunkInfo.name;

					manifest[originalName] = fileName;
				}
			}

			// Always map 'bundle' to main entry
			const mainChunk = Object.keys(bundle).find(name =>
				bundle[name].isEntry && !name.includes('polyfills')
			);
			if (mainChunk) {
				manifest['bundle'] = mainChunk.replace('.js', '');
			}

			// Write manifest.json
			this.emitFile({
				type: 'asset',
				fileName: 'manifest.json',
				source: JSON.stringify(manifest, null, 2)
			});
		}
	};
}

export default [
	// Main Bundle: Alpine.js + HTMX + scanner-loader (html5-qrcode lazy loaded)
	{
		input: 'internal/assets/static/js/src/app.js',
		plugins: [
			resolve(),
			commonjs(),
			!isDev && terser({
				compress: {
					drop_console: false, // Keep console for debugging in production
					drop_debugger: true,
					pure_funcs: ['console.debug']
				},
				format: {
					comments: false
				}
			}),
			!isDev && manifestPlugin(),
			!isDev && visualizer({
				filename: 'bundle-stats.html',
				open: false,
				gzipSize: true,
				brotliSize: true,
				template: 'treemap' // Options: treemap, sunburst, network
			}),
			!isDev && gzipPlugin({
				filter: /\.(js|json|css|html)$/,
				additionalFiles: []
			}),
			!isDev && gzipPlugin({
				filter: /\.(js|json|css|html)$/,
				customCompression: content => brotliCompressSync(Buffer.from(content)),
				fileName: '.br'
			})
		].filter(Boolean),
		output: {
			dir: outputDir,
			format: 'es', // Use ESM for both dev and production (modern browsers support)
			entryFileNames: isDev ? 'bundle.js' : 'bundle.[hash].js',
			chunkFileNames: isDev ? '[name].js' : '[name].[hash].js',
			assetFileNames: '[name].[ext]',
			sourcemap: true,
			// Code splitting: html5-qrcode will be automatically split via dynamic import
		},
		// Watch options for development
		watch: {
			clearScreen: false,
			include: 'internal/assets/static/js/src/**'
		}
	},
];
