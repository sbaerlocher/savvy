// Orientation-based fullscreen barcode overlay
// When on mobile and switching to landscape, show the barcode fullscreen
// so it's easier to scan at a shop.

export function initOrientationStore (Alpine) {
  Alpine.store('barcode', {
    fullscreen: false,
    barcodeHtml: ''
  })

  // Only activate on touch devices (mobile)
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

  // Listen to orientation changes
  window.addEventListener('orientationchange', () => {
    // Small delay so innerWidth/innerHeight are updated
    setTimeout(handleOrientationChange, 100)
  })

  // Fallback: resize event (some browsers fire this instead)
  let resizeTimer = null
  window.addEventListener('resize', () => {
    if (resizeTimer) clearTimeout(resizeTimer)
    resizeTimer = setTimeout(handleOrientationChange, 150)
  })

  // Close on click/touch anywhere on the overlay
  store.close = function () {
    this.fullscreen = false
    this.barcodeHtml = ''
  }
}
