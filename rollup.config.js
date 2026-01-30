// rollup.config.js
import commonjs from '@rollup/plugin-commonjs';
import resolve from '@rollup/plugin-node-resolve';
import terser from '@rollup/plugin-terser';

const isDev = process.env.NODE_ENV !== 'production';

export default [
	// Main Bundle: Alpine.js + HTMX + html5-qrcode + scanner logic
	{
		input: 'internal/assets/static/js/src/app.js',
		plugins: [
			resolve(),
			commonjs(),
			!isDev && terser(),
		].filter(Boolean),
		output: {
			file: 'internal/assets/static/js/bundle.js',
			format: 'iife',
			sourcemap: isDev,
		},
	},
];
