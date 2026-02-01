/**
 * Orientation-based barcode fullscreen
 * Shows barcode fullscreen on mobile landscape for easier scanning at shops
 */
export function initOrientationStore (Alpine) {
  Alpine.store('barcode', {
    fullscreen: false,
    barcodeHtml: ''
  })

  if (!('ontouchstart' in window)) {
    return
  }

  const store = Alpine.store('barcode')

  function getBarcodeSectionHtml () {
    const section = document.getElementById('barcode-section')
    if (!section) return ''
    return section.innerHTML
  }

  function handleOrientationChange () {
    const isLandscape =
      window.screen.orientation?.type?.includes('landscape') ||
      window.innerWidth > window.innerHeight

    const section = document.getElementById('barcode-section')
    if (!section) {
      store.fullscreen = false
      return
    }

    if (isLandscape) {
      store.barcodeHtml = getBarcodeSectionHtml()
      store.fullscreen = true
    } else {
      store.fullscreen = false
      store.barcodeHtml = ''
    }
  }

  window.addEventListener('orientationchange', () => {
    setTimeout(handleOrientationChange, 100)
  })

  let resizeTimer = null
  window.addEventListener('resize', () => {
    if (resizeTimer) clearTimeout(resizeTimer)
    resizeTimer = setTimeout(handleOrientationChange, 150)
  })

  store.close = function () {
    this.fullscreen = false
    this.barcodeHtml = ''
  }
}
