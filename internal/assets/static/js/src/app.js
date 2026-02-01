// static/js/src/app.js - Main entry point for bundled JS

import Alpine from 'alpinejs'
import htmx from 'htmx.org'

import './scanner-loader.js'
import { initOfflineStore, offlineHandler } from './offline.js'
import { initOrientationStore } from './orientation.js'
import { setupPrecaching } from './precache.js'

window.Alpine = Alpine
window.htmx = htmx
window.offlineHandler = offlineHandler

initOfflineStore(Alpine)
initOrientationStore(Alpine)
setupPrecaching()

Alpine.start()
