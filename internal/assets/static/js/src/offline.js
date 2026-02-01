// Toast notification helper
function showToast (message, type = 'info') {
  const toast = document.createElement('div')
  toast.className = `fixed bottom-4 right-4 px-6 py-3 rounded-lg shadow-lg text-white font-medium z-50 transition-all duration-300 ${
    type === 'success'
? 'bg-green-600'
    : type === 'warning'
? 'bg-yellow-600'
    : type === 'error'
? 'bg-red-600'
    : 'bg-blue-600'
  }`
  toast.textContent = message
  toast.style.opacity = '0'
  toast.style.transform = 'translateY(20px)'

  document.body.appendChild(toast)

  setTimeout(() => {
    toast.style.opacity = '1'
    toast.style.transform = 'translateY(0)'
  }, 10)

  setTimeout(() => {
    toast.style.opacity = '0'
    toast.style.transform = 'translateY(20px)'
    setTimeout(() => toast.remove(), 300)
  }, 3000)
}

/**
 * Reliable connection check
 * Based on: https://dev.to/maxmonteil/is-your-app-online-here-s-how-to-reliably-know-in-just-10-lines-of-js-guide-3in7
 *
 * navigator.onLine is unreliable:
 * - false = always reliable (you are offline)
 * - true = unreliable (only means "connected to network", not "has internet")
 */
async function checkOnlineStatus () {
  if (!navigator.onLine) {
    return false
  }

  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000)

    const response = await fetch(`/health?t=${Date.now()}`, {
      method: 'GET',
      cache: 'no-store',
      signal: controller.signal
    })

    clearTimeout(timeoutId)
    return response.ok
  } catch {
    return false
  }
}

/**
 * Initialize offline detection store in Alpine.js
 * Provides reliable offline status with debouncing and localStorage persistence
 */
export function initOfflineStore (Alpine) {
  console.log('[Alpine] Init - creating offline store')

  let offlineDebounceTimer = null
  let actualOnlineStatus = navigator.onLine

  let lastKnownStatus = true
  try {
    const stored = localStorage.getItem('savvy:online-status')
    if (stored !== null) {
      lastKnownStatus = stored === 'true'
      console.log('[Offline] Restored status from localStorage:', lastKnownStatus)
    }
  } catch (e) {
  }

  Alpine.store('offline', {
    isOnline: lastKnownStatus,
    checking: false
  })

  function updateOnlineStatus (isOnline) {
    actualOnlineStatus = isOnline

    if (offlineDebounceTimer) {
      clearTimeout(offlineDebounceTimer)
      offlineDebounceTimer = null
    }

    if (isOnline) {
      console.log('[Offline] Status changed to ONLINE (immediate)')
      Alpine.store('offline').isOnline = true
      try {
        localStorage.setItem('savvy:online-status', 'true')
      } catch (e) {}
      return
    }

    console.log('[Offline] Status may be OFFLINE, verifying...')
    offlineDebounceTimer = setTimeout(() => {
      if (!actualOnlineStatus) {
        console.log('[Offline] Status confirmed OFFLINE after debounce')
        Alpine.store('offline').isOnline = false
        try {
          localStorage.setItem('savvy:online-status', 'false')
        } catch (e) {}
      }
    }, 2000)
  }

  const shouldVerify = !localStorage.getItem('savvy:online-status') ||
                       !lastKnownStatus ||
                       navigator.onLine !== lastKnownStatus

  if (shouldVerify) {
    console.log('[Offline] Verifying connection on page load...')
    checkOnlineStatus().then(isOnline => {
      console.log('[Offline] Initial check - isOnline:', isOnline)
      updateOnlineStatus(isOnline)
    })
  } else {
    console.log('[Offline] Skipping initial check (using cached status:', lastKnownStatus, ')')
  }

  window.addEventListener('online', async () => {
    console.log('[Offline] Browser online event')
    const isOnline = await checkOnlineStatus()
    updateOnlineStatus(isOnline)
    if (isOnline) {
      showToast('✅ Verbindung wiederhergestellt', 'success')
    }
  })

  window.addEventListener('offline', () => {
    console.log('[Offline] Browser offline event')
    updateOnlineStatus(false)
  })

  setInterval(async () => {
    const isOnline = await checkOnlineStatus()
    const storeOnline = Alpine.store('offline').isOnline

    if (isOnline !== storeOnline) {
      console.log('[Offline] Periodic check detected change:', isOnline)
      updateOnlineStatus(isOnline)
    }
  }, 10000)

  const store = Alpine.store('offline')
  store.checkConnection = async function () {
    this.checking = true
    console.log('[Offline] Manual check...')

    const isOnline = await checkOnlineStatus()
    this.isOnline = isOnline
    this.checking = false

    if (isOnline) {
      console.log('[Offline] Server reachable - reloading...')
      setTimeout(() => window.location.reload(), 300)
    }
  }
}

/**
 * Handler for standalone offline page (offline.templ)
 * Simpler alternative without Alpine.js store for dedicated offline page
 */
export function offlineHandler () {
  return {
    checking: false,

    init () {
      window.addEventListener('online', () => {
        window.location.reload()
      })
    },

    async checkConnection () {
      this.checking = true

      try {
        const response = await fetch('/health', {
          method: 'HEAD',
          cache: 'no-cache'
        })

        if (response.ok) {
          window.location.reload()
        } else {
          this.showError()
        }
      } catch (error) {
        this.showError()
      } finally {
        this.checking = false
      }
    },

    showError () {
      console.log('Still offline')
    }
  }
}
