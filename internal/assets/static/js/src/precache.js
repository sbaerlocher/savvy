/**
 * Precache setup for offline access
 * Caches all detail pages and associated resources for offline browsing
 */
let precacheInitialized = false
const cachedUrls = new Set()

export function setupPrecaching () {
  if (precacheInitialized || !('serviceWorker' in navigator)) {
    return
  }
  precacheInitialized = true

  console.log('[Precache] Initializing...')

  function cacheUrl (url) {
    if (cachedUrls.has(url)) {
      return
    }
    cachedUrls.add(url)

    fetch(url, {
      method: 'GET',
      credentials: 'same-origin',
      cache: 'no-store'
    }).catch(() => {
      cachedUrls.delete(url)
    })
  }

  async function cacheDetailPageWithBarcodes (detailUrl) {
    try {
      const response = await fetch(detailUrl, {
        method: 'GET',
        credentials: 'same-origin',
        cache: 'no-store'
      })

      if (!response.ok) return

      cachedUrls.add(detailUrl)

      const html = await response.text()
      const parser = new DOMParser()
      const doc = parser.parseFromString(html, 'text/html')

      const barcodeImages = doc.querySelectorAll('img[src^="/barcode"]')
      barcodeImages.forEach(img => {
        const barcodeUrl = img.getAttribute('src')
        if (barcodeUrl) {
          console.log('[Precache] Caching barcode:', barcodeUrl)
          cacheUrl(barcodeUrl)
        }
      })
    } catch (err) {
      console.warn('[Precache] Failed to cache detail page:', detailUrl, err)
    }
  }

  async function cacheListPageAndDetails (listUrl) {
    console.log('[Precache] Fetching list page:', listUrl)
    try {
      const response = await fetch(listUrl, {
        method: 'GET',
        credentials: 'same-origin',
        cache: 'no-store'
      })

      if (!response.ok) {
        console.warn('[Precache] Failed to fetch list page:', listUrl, response.status)
        return
      }

      cachedUrls.add(listUrl)
      console.log('[Precache] Cached list page:', listUrl)

      const html = await response.text()
      const parser = new DOMParser()
      const doc = parser.parseFromString(html, 'text/html')

      const links = doc.querySelectorAll('a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]')
      console.log(`[Precache] Found ${links.length} links in ${listUrl}`)

      links.forEach(link => {
        const url = link.getAttribute('href')
        if (/^\/(?:cards|vouchers|gift-cards)\/[0-9a-f-]{36}/.test(url)) {
          console.log('[Precache] Caching detail page with barcodes:', url)
          cacheDetailPageWithBarcodes(url)
        }
      })
    } catch (err) {
      console.warn('[Precache] Failed to cache list page:', listUrl, err)
    }
  }

  async function cacheAllPages () {
    cacheUrl('/')

    const listPages = ['/cards', '/vouchers', '/gift-cards']
    for (const page of listPages) {
      await cacheListPageAndDetails(page)
    }

    console.log('[Precache] Initial caching complete')
  }

  cacheAllPages()

  function scanForDetailPages () {
    const links = document.querySelectorAll('a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]')

    links.forEach(link => {
      const url = link.getAttribute('href')
      if (/^\/(?:cards|vouchers|gift-cards)\/[0-9a-f-]{36}/.test(url)) {
        cacheUrl(url)
      }
    })
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', scanForDetailPages)
  } else {
    scanForDetailPages()
  }

  if (typeof htmx !== 'undefined') {
    document.body.addEventListener('htmx:afterSwap', () => {
      setTimeout(scanForDetailPages, 100)
    })
  }
}
