// static/js/src/app.js - Main entry point for bundled JS

import Alpine from 'alpinejs';
import htmx from 'htmx.org';
import { Html5Qrcode } from 'html5-qrcode';

// Make Alpine global
window.Alpine = Alpine;

// Make HTMX global
window.htmx = htmx;

// Make Html5Qrcode global for scanner.js
window.Html5Qrcode = Html5Qrcode;
window.Html5QrcodeSupportedFormats = Html5Qrcode.Html5QrcodeSupportedFormats || {
	QR_CODE: 0,
	AZTEC: 1,
	CODABAR: 2,
	CODE_39: 3,
	CODE_93: 4,
	CODE_128: 5,
	DATA_MATRIX: 6,
	MAXICODE: 7,
	ITF: 8,
	EAN_13: 9,
	EAN_8: 10,
	PDF_417: 11,
	UPC_A: 12,
	UPC_E: 13,
	UPC_EAN_EXTENSION: 14,
	ITF_14: 15,
};

// Import scanner module (will be bundled)
import './scanner.js';

// Import offline detection
import { initOfflineStore } from './offline.js';

// Import orientation-based barcode fullscreen
import { initOrientationStore } from './orientation.js';

// Import precaching for offline detail pages
import { setupPrecaching } from './precache.js';

// Initialize offline store before starting Alpine
initOfflineStore(Alpine);

// Initialize orientation store for landscape barcode fullscreen
initOrientationStore(Alpine);

// Setup precaching for detail pages
setupPrecaching();

// Start Alpine
Alpine.start();
