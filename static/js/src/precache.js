// Precache - Cache all detail pages for offline access
let precacheInitialized = false;
const cachedUrls = new Set();

export function setupPrecaching() {
	// Only run once
	if (precacheInitialized || !('serviceWorker' in navigator)) {
		return;
	}
	precacheInitialized = true;

	console.log('[Precache] Initializing...');

	// Helper to cache a URL (deduplicated)
	function cacheUrl(url) {
		if (cachedUrls.has(url)) {
			return;
		}
		cachedUrls.add(url);

		fetch(url, {
			method: 'GET',
			credentials: 'same-origin',
			cache: 'no-store'
		}).catch(() => {
			// Silently fail, remove from set to retry later
			cachedUrls.delete(url);
		});
	}

	// Fetch detail page and cache its barcode images
	async function cacheDetailPageWithBarcodes(detailUrl) {
		try {
			const response = await fetch(detailUrl, {
				method: 'GET',
				credentials: 'same-origin',
				cache: 'no-store'
			});

			if (!response.ok) return;

			cachedUrls.add(detailUrl);

			// Parse HTML to find barcode images
			const html = await response.text();
			const parser = new DOMParser();
			const doc = parser.parseFromString(html, 'text/html');

			// Find barcode images
			const barcodeImages = doc.querySelectorAll('img[src^="/barcode"]');
			barcodeImages.forEach(img => {
				const barcodeUrl = img.getAttribute('src');
				if (barcodeUrl) {
					console.log('[Precache] Caching barcode:', barcodeUrl);
					cacheUrl(barcodeUrl);
				}
			});
		} catch (err) {
			console.warn('[Precache] Failed to cache detail page:', detailUrl, err);
		}
	}

	// Fetch and parse list pages to find all detail links
	async function cacheListPageAndDetails(listUrl) {
		console.log('[Precache] Fetching list page:', listUrl);
		try {
			// First cache the list page itself
			const response = await fetch(listUrl, {
				method: 'GET',
				credentials: 'same-origin',
				cache: 'no-store'
			});

			if (!response.ok) {
				console.warn('[Precache] Failed to fetch list page:', listUrl, response.status);
				return;
			}

			cachedUrls.add(listUrl);
			console.log('[Precache] Cached list page:', listUrl);

			// Parse HTML to find detail page links
			const html = await response.text();
			const parser = new DOMParser();
			const doc = parser.parseFromString(html, 'text/html');

			// Find all detail page links in this list page
			const links = doc.querySelectorAll('a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]');
			console.log(`[Precache] Found ${links.length} links in ${listUrl}`);

			links.forEach(link => {
				const url = link.getAttribute('href');
				if (/^\/(?:cards|vouchers|gift-cards)\/[0-9a-f-]{36}/.test(url)) {
					console.log('[Precache] Caching detail page with barcodes:', url);
					// Cache detail page AND its barcodes
					cacheDetailPageWithBarcodes(url);
				}
			});
		} catch (err) {
			console.warn('[Precache] Failed to cache list page:', listUrl, err);
		}
	}

	// Cache dashboard and all three list pages with their details
	async function cacheAllPages() {
		// First cache dashboard
		cacheUrl('/');

		// Then cache all list pages and their details
		const listPages = ['/cards', '/vouchers', '/gift-cards'];
		for (const page of listPages) {
			await cacheListPageAndDetails(page);
		}

		console.log('[Precache] Initial caching complete');
	}

	// Start caching
	cacheAllPages();

	// Function to scan for detail pages in current DOM
	function scanForDetailPages() {
		const links = document.querySelectorAll('a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]');

		links.forEach(link => {
			const url = link.getAttribute('href');
			if (/^\/(?:cards|vouchers|gift-cards)\/[0-9a-f-]{36}/.test(url)) {
				cacheUrl(url);
			}
		});
	}

	// Initial scan after page load
	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', scanForDetailPages);
	} else {
		scanForDetailPages();
	}

	// Re-scan after HTMX swaps
	if (typeof htmx !== 'undefined') {
		document.body.addEventListener('htmx:afterSwap', () => {
			setTimeout(scanForDetailPages, 100);
		});
	}
}
