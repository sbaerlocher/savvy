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

  // Animate in
  setTimeout(() => {
    toast.style.opacity = '1'
    toast.style.transform = 'translateY(0)'
  }, 10)

  // Animate out and remove
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
  // If navigator says offline, trust it (false is always reliable)
  if (!navigator.onLine) {
    return false
  }

  // If navigator says online, verify with real fetch
  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000) // Increased to 5s for service worker activation

    // Use health endpoint with GET (HEAD might redirect)
    // Add timestamp to prevent cache
    const response = await fetch(`/health?t=${Date.now()}`, {
      method: 'GET',
      cache: 'no-store',
      signal: controller.signal
    })

    clearTimeout(timeoutId)
    return response.ok
  } catch {
    // Network error or timeout = offline
    return false
  }
}

// Offline Detection - Alpine.js Store
export function initOfflineStore (Alpine) {
  console.log('[Alpine] Init - creating offline store')

  // Debounce timer for offline status changes
  let offlineDebounceTimer = null
  let actualOnlineStatus = navigator.onLine

  // Create store with initial state (assume online, will verify)
  Alpine.store('offline', {
    isOnline: true, // Start optimistic to prevent flash
    checking: false
  })

  // Helper to update online status with debounce
  function updateOnlineStatus (isOnline) {
    actualOnlineStatus = isOnline

    // Clear existing timer
    if (offlineDebounceTimer) {
      clearTimeout(offlineDebounceTimer)
      offlineDebounceTimer = null
    }

    // If going online, update immediately
    if (isOnline) {
      console.log('[Offline] Status changed to ONLINE (immediate)')
      Alpine.store('offline').isOnline = true
      return
    }

    // If going offline, debounce for 2 seconds to prevent flashing
    console.log('[Offline] Status may be OFFLINE, verifying...')
    offlineDebounceTimer = setTimeout(() => {
      // Verify status hasn't changed during debounce
      if (!actualOnlineStatus) {
        console.log('[Offline] Status confirmed OFFLINE after debounce')
        Alpine.store('offline').isOnline = false
      }
    }, 2000) // 2 second debounce for offline status
  }

  // Immediately verify actual connection on page load
  checkOnlineStatus().then(isOnline => {
    console.log('[Offline] Initial check - isOnline:', isOnline)
    updateOnlineStatus(isOnline)
  })

  // Listen to browser online/offline events
  window.addEventListener('online', async () => {
    console.log('[Offline] Browser online event')
    // Verify with real check
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

  // Periodic check every 10 seconds (backup for missed events, less frequent to reduce load)
  setInterval(async () => {
    const isOnline = await checkOnlineStatus()
    const storeOnline = Alpine.store('offline').isOnline

    if (isOnline !== storeOnline) {
      console.log('[Offline] Periodic check detected change:', isOnline)
      updateOnlineStatus(isOnline)
    }
  }, 10000) // Increased from 5s to 10s

  // Add manual check method to store
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
    } else {
      console.log('[Offline] Still offline')
    }
  }
}
