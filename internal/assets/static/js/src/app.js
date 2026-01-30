// static/js/src/app.js - Main entry point for bundled JS

import Alpine from 'alpinejs';
import htmx from 'htmx.org';

// Make Alpine global
window.Alpine = Alpine;

// Make HTMX global
window.htmx = htmx;

// Import scanner loader (lazy loads html5-qrcode on demand)
import './scanner-loader.js';

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
