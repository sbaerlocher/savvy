const CACHE_VERSION = "savvy-v1.4.6";
const STATIC_CACHE = `${CACHE_VERSION}-static`;
const DYNAMIC_CACHE = `${CACHE_VERSION}-dynamic`;

// Cache Limits
const MAX_DYNAMIC_CACHE_SIZE = 100; // Max 100 Einträge im Dynamic Cache
const MAX_CACHE_AGE_DAYS = 7; // Cache-Einträge älter als 7 Tage werden gelöscht

// Debug Logging
console.log("[ServiceWorker] Script loaded, version:", CACHE_VERSION);

// Assets die sofort gecached werden (nur statische, immer verfügbare Files)
const STATIC_ASSETS = [
  "/static/js/bundle.js",
  "/static/css/bundle.css",
  "/static/manifest.json",
];

// Install Event - Cache static assets
self.addEventListener("install", (event) => {
  console.log("[ServiceWorker] Installing version:", CACHE_VERSION);
  event.waitUntil(
    Promise.all([
      // 1. Cache static assets
      caches.open(STATIC_CACHE).then((cache) => {
        console.log("[ServiceWorker] Caching static assets");
        // Cache each file individually to avoid failing the entire install
        return Promise.allSettled(
          STATIC_ASSETS.map((url) =>
            cache
              .add(new Request(url, { cache: "reload" }))
              .catch((err) =>
                console.warn("[ServiceWorker] Failed to cache:", url, err),
              ),
          ),
        );
      }),

      // 2. Proaktiv alte Caches löschen (nicht warten bis activate)
      caches.keys().then((cacheNames) => {
        const oldCaches = cacheNames.filter(
          (name) =>
            name.startsWith("savvy-") &&
            name !== STATIC_CACHE &&
            name !== DYNAMIC_CACHE,
        );

        if (oldCaches.length > 0) {
          console.log(
            "[ServiceWorker] Install: Pre-cleaning",
            oldCaches.length,
            "old caches",
          );
          return Promise.all(oldCaches.map((name) => caches.delete(name)));
        }
      }),
    ])
      .then(() => {
        console.log("[ServiceWorker] Install complete, skipping waiting");
        return self.skipWaiting();
      })
      .catch((err) => console.error("[ServiceWorker] Install failed:", err)),
  );
});

// Activate Event - Clean old caches
self.addEventListener("activate", (event) => {
  console.log("[ServiceWorker] Activating version:", CACHE_VERSION);
  event.waitUntil(
    Promise.all([
      // 1. Lösche alte Cache-Versionen
      caches.keys().then((cacheNames) => {
        const oldCaches = cacheNames.filter(
          (name) =>
            name.startsWith("savvy-") &&
            name !== STATIC_CACHE &&
            name !== DYNAMIC_CACHE,
        );

        if (oldCaches.length > 0) {
          console.log(
            "[ServiceWorker] Deleting",
            oldCaches.length,
            "old caches:",
            oldCaches,
          );
        }

        return Promise.all(oldCaches.map((name) => caches.delete(name)));
      }),

      // 2. Entferne Duplikate aus Dynamic Cache
      removeCacheDuplicates(DYNAMIC_CACHE),

      // 3. Limitiere Dynamic Cache Größe
      limitCacheSize(DYNAMIC_CACHE, MAX_DYNAMIC_CACHE_SIZE),

      // 4. Lösche alte Cache-Einträge (> 7 Tage)
      cleanOldCacheEntries(DYNAMIC_CACHE, MAX_CACHE_AGE_DAYS),
    ]).then(() => {
      console.log("[ServiceWorker] Activation complete, claiming clients");
      return self.clients.claim();
    }),
  );
});

// Fetch Event - Network First, Cache Fallback
self.addEventListener("fetch", (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Nur same-origin requests cachen
  if (url.origin !== location.origin) {
    return;
  }

  // Ignore non-GET requests
  if (request.method !== "GET") {
    return;
  }

  // NEVER cache /health endpoint (used for online detection)
  if (url.pathname === "/health") {
    return;
  }

  // Strategy: Network First, Cache Fallback
  event.respondWith(
    fetch(request)
      .then((response) => {
        // Nur erfolgreiche Responses cachen (200 OK)
        // Keine Redirects (3xx) cachen, außer es ist die finale URL nach Redirect
        if (!response || response.type === "error") {
          return response;
        }

        // Bei Redirects: Response zurückgeben, aber nicht cachen
        if (
          response.redirected ||
          (response.status >= 300 && response.status < 400)
        ) {
          return response;
        }

        // Nur 200 OK cachen
        if (response.status !== 200) {
          return response;
        }

        // Clone response (kann nur einmal gelesen werden)
        const responseClone = response.clone();

        // Bestimme Cache basierend auf URL
        const cacheName = shouldCacheDynamically(url.pathname)
          ? DYNAMIC_CACHE
          : STATIC_CACHE;

        // Cache die Response asynchron (nur für cacheable routes)
        if (
          shouldCacheDynamically(url.pathname) ||
          STATIC_ASSETS.includes(url.pathname)
        ) {
          caches.open(cacheName).then((cache) => {
            // Verwende nur URL (ohne Query-Parameter) als Cache-Key um Duplikate zu vermeiden
            const cacheUrl = new URL(request.url);
            cacheUrl.search = ""; // Entferne Query-Parameter
            const cacheRequest = new Request(cacheUrl.toString(), {
              method: "GET",
              headers: { Accept: request.headers.get("Accept") || "*/*" },
            });

            cache
              .put(cacheRequest, responseClone)
              .then(() => {
                // Nach jedem neuen Cache-Eintrag: Limitiere Größe
                if (cacheName === DYNAMIC_CACHE) {
                  limitCacheSize(DYNAMIC_CACHE, MAX_DYNAMIC_CACHE_SIZE);
                }
              })
              .catch((err) => {
                console.warn(
                  "[ServiceWorker] Failed to cache:",
                  url.pathname,
                  err,
                );
              });
          });
        }

        return response;
      })
      .catch(() => {
        // Network fehlgeschlagen, versuche Cache
        // Normalisiere Request für Cache-Lookup (ohne Query-Parameter)
        const cacheUrl = new URL(request.url);
        cacheUrl.search = ""; // Entferne Query-Parameter
        const cacheRequest = new Request(cacheUrl.toString(), {
          method: "GET",
          headers: { Accept: request.headers.get("Accept") || "*/*" },
        });

        return caches.match(cacheRequest).then((cachedResponse) => {
          if (cachedResponse) {
            console.log(
              "[ServiceWorker] Serving from cache:",
              cacheUrl.toString(),
            );
            return cachedResponse;
          }

          // Fallback auf offline page für HTML requests
          if (
            request.headers.get("accept") &&
            request.headers.get("accept").includes("text/html")
          ) {
            // Versuche offline page aus cache, sonst fetch
            return caches.match("/offline").then((offlinePage) => {
              if (offlinePage) {
                return offlinePage;
              }
              // Offline page nicht gecached, versuche zu fetchen (sollte klappen da public)
              return fetch("/offline").catch(() => {
                // Fallback: Simple HTML
                return new Response(
                  `
                    <!DOCTYPE html>
                    <html>
                    <head><title>Offline</title></head>
                    <body style="font-family: sans-serif; text-align: center; padding: 50px;">
                      <h1>📡 Offline</h1>
                      <p>Keine Internetverbindung</p>
                      <button onclick="location.reload()">Erneut versuchen</button>
                    </body>
                    </html>
                  `,
                  {
                    headers: { "Content-Type": "text/html" },
                  },
                );
              });
            });
          }

          // Für andere requests: leere Response
          return new Response("Offline - Resource not available", {
            status: 503,
            statusText: "Service Unavailable",
            headers: new Headers({
              "Content-Type": "text/plain",
            }),
          });
        });
      }),
  );
});

// Bestimme welche URLs dynamisch gecached werden sollen
function shouldCacheDynamically(pathname) {
  // Cache alle wichtigen Routes für Offline-Viewing
  const dynamicRoutes = [
    "/",
    "/dashboard",
    "/cards",
    "/vouchers",
    "/gift-cards",
    "/favorites",
    "/offline",
  ];

  // Cache Detail-Pages (UUID Pattern)
  const uuidRegex =
    /^\/(?:cards|vouchers|gift-cards)\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

  // Cache Barcode Images
  const isBarcodeUrl = pathname.startsWith("/barcode");

  return (
    dynamicRoutes.some((route) => pathname.startsWith(route)) ||
    uuidRegex.test(pathname) ||
    isBarcodeUrl
  );
}

// Message Event - für manuelle Cache-Updates vom Client
self.addEventListener("message", (event) => {
  if (event.data && event.data.type === "SKIP_WAITING") {
    console.log("[ServiceWorker] Received SKIP_WAITING message");
    self.skipWaiting();
  }

  if (event.data && event.data.type === "CLEAR_CACHE") {
    console.log("[ServiceWorker] Received CLEAR_CACHE message");
    event.waitUntil(
      caches.keys().then((cacheNames) => {
        console.log("[ServiceWorker] Clearing all caches:", cacheNames);
        return Promise.all(
          cacheNames.map((cacheName) => caches.delete(cacheName)),
        );
      }),
    );
  }

  if (event.data && event.data.type === "CLEANUP_OLD_CACHES") {
    console.log("[ServiceWorker] Received CLEANUP_OLD_CACHES message");
    event.waitUntil(
      caches.keys().then((cacheNames) => {
        const oldCaches = cacheNames.filter(
          (name) =>
            name.startsWith("savvy-") &&
            name !== STATIC_CACHE &&
            name !== DYNAMIC_CACHE,
        );

        if (oldCaches.length > 0) {
          console.log(
            "[ServiceWorker] Manual cleanup: Deleting",
            oldCaches.length,
            "old caches:",
            oldCaches,
          );
          return Promise.all(oldCaches.map((name) => caches.delete(name)));
        } else {
          console.log("[ServiceWorker] No old caches to delete");
        }
      }),
    );
  }
});

// Helper: Entferne Duplikate aus Cache (gleiche URL, unterschiedliche Request-Varianten)
async function removeCacheDuplicates(cacheName) {
  const cache = await caches.open(cacheName);
  const keys = await cache.keys();
  const seenUrls = new Map(); // URL -> Request
  let duplicatesCount = 0;

  for (const request of keys) {
    const url = new URL(request.url);
    // Normalisiere URL (ohne Query-Parameter und Hash)
    url.search = "";
    url.hash = "";
    const normalizedUrl = url.toString();

    if (seenUrls.has(normalizedUrl)) {
      // Duplikat gefunden - lösche es
      await cache.delete(request);
      duplicatesCount++;
    } else {
      // Erste Instanz dieser URL - behalte sie
      seenUrls.set(normalizedUrl, request);
    }
  }

  if (duplicatesCount > 0) {
    console.log(
      `[ServiceWorker] Removed ${duplicatesCount} duplicate entries from ${cacheName}`,
    );
  }

  return duplicatesCount;
}

// Helper: Limitiere Cache-Größe (FIFO - First In, First Out)
async function limitCacheSize(cacheName, maxSize) {
  const cache = await caches.open(cacheName);
  const keys = await cache.keys();

  if (keys.length > maxSize) {
    const deleteCount = keys.length - maxSize;
    console.log(
      `[ServiceWorker] Cache ${cacheName} exceeds limit (${keys.length}/${maxSize}), deleting ${deleteCount} oldest entries`,
    );

    // Lösche die ältesten Einträge (FIFO)
    for (let i = 0; i < deleteCount; i++) {
      await cache.delete(keys[i]);
    }
  }
}

// Helper: Lösche Cache-Einträge älter als X Tage
async function cleanOldCacheEntries(cacheName, maxAgeDays) {
  const cache = await caches.open(cacheName);
  const keys = await cache.keys();
  const now = Date.now();
  const maxAge = 1; // Tage in Millisekunden

  let deletedCount = 0;

  for (const request of keys) {
    const response = await cache.match(request);
    if (!response) continue;

    // Prüfe Date Header (Server Response Time)
    const dateHeader = response.headers.get("date");
    if (dateHeader) {
      const cacheDate = new Date(dateHeader).getTime();
      const age = now - cacheDate;

      if (age > maxAge) {
        await cache.delete(request);
        deletedCount++;
      }
    }
  }

  if (deletedCount > 0) {
    console.log(
      `[ServiceWorker] Deleted ${deletedCount} entries older than ${maxAgeDays} days from ${cacheName}`,
    );
  }
}
