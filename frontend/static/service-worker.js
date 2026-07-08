// HRMS PWA Service Worker
const CACHE_NAME = 'hrms-cache-v1';
const STATIC_ASSETS = [
  '/',
  '/manifest.json',
  '/icons/icon-192.svg',
  '/icons/icon-512.svg',
  '/icons/icon-maskable.svg'
];

// Assets with versioned URLs (build output) — cache-first
const VERSIONED_CACHE = 'hrms-build-v1';

// API responses — network-first with cache fallback
const API_CACHE = 'hrms-api-v1';

// ==========================================
// INSTALL — pre-cache critical static assets
// ==========================================
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS);
    })
  );
  // Activate immediately without waiting for reload
  self.skipWaiting();
});

// ==========================================
// ACTIVATE — clean up old caches
// ==========================================
self.addEventListener('activate', (event) => {
  const validCaches = [CACHE_NAME, VERSIONED_CACHE, API_CACHE];
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => !validCaches.includes(name))
          .map((name) => caches.delete(name))
      );
    }).then(() => {
      // Enable navigation preload for faster navigation
      if (self.registration.navigationPreload) {
        return self.registration.navigationPreload.enable().then(() => {
          return self.clients.claim();
        });
      }
      // Take control of all pages immediately
      return self.clients.claim();
    })
  );
});

// ==========================================
// FETCH — routing strategy
// ==========================================
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests
  if (request.method !== 'GET') return;

  // Skip chrome-extension and other protocols
  if (!url.protocol.startsWith('http')) return;

  // Skip Vite dev server internal URLs (@vite/client, @fs/, node_modules/, etc.)
  if (url.hostname === 'localhost' || url.hostname === '127.0.0.1') {
    if (url.pathname.includes('@vite') || url.pathname.includes('@fs') || url.pathname.includes('/node_modules/')) {
      return;
    }
  }

  // ======================================
  // Strategy 1: Cache-First for build assets
  // (files with hash in URL or from /_app/)
  // ======================================
  if (url.pathname.startsWith('/_app/') || url.pathname.includes('/_app/')) {
    event.respondWith(
      caches.open(VERSIONED_CACHE).then((cache) => {
        return cache.match(request).then((cached) => {
          if (cached) return cached;
          return fetch(request).then((response) => {
            if (response.ok) {
              cache.put(request, response.clone());
            }
            return response;
          });
        });
      })
    );
    return;
  }

  // ======================================
  // Strategy 2: Cache-First for static assets
  // (icons, manifest, fonts)
  // ======================================
  if (
    url.pathname.startsWith('/icons/') ||
    url.pathname === '/manifest.json' ||
    url.pathname.startsWith('/fonts/')
  ) {
    event.respondWith(
      caches.open(CACHE_NAME).then((cache) => {
        return cache.match(request).then((cached) => {
          if (cached) return cached;
          return fetch(request).then((response) => {
            if (response.ok) {
              cache.put(request, response.clone());
            }
            return response;
          });
        });
      })
    );
    return;
  }

  // ======================================
  // Strategy 3: Network-First for API calls
  // with offline fallback
  // ======================================
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(
      fetch(request)
        .then((response) => {
          if (response.ok) {
            const clone = response.clone();
            caches.open(API_CACHE).then((cache) => {
              cache.put(request, clone);
              trimCache(API_CACHE, 50);
            });
          }
          return response;
        })
        .catch(() => {
          // Offline — return cached response directly if available
          return caches.open(API_CACHE).then((cache) => {
            return cache.match(request).then((cached) => {
              if (cached) {
                // Return the actual cached response with offline header
                const headers = new Headers(cached.headers);
                headers.set('X-Offline', 'true');
                return new Response(cached.body, {
                  status: cached.status,
                  statusText: cached.statusText,
                  headers
                });
              }
              // No cached data — return offline error
              return new Response(
                JSON.stringify({
                  success: false,
                  offline: true,
                  message: 'Anda sedang offline. Data tidak tersedia di cache.'
                }),
                {
                  status: 503,
                  headers: { 'Content-Type': 'application/json' }
                }
              );
            });
          });
        })
    );
    return;
  }

  // ======================================
  // Strategy 4: Network-First for navigation
  // with navigationPreload + offline HTML fallback
  // ======================================
  if (request.mode === 'navigate') {
    event.respondWith(
      (async () => {
        try {
          // Try the preload response first
          const preloadResponse = await event.preloadResponse;
          if (preloadResponse) {
            const clone = preloadResponse.clone();
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(request, clone);
            });
            return preloadResponse;
          }
          // Fallback to normal fetch
          const response = await fetch(request);
          if (response.ok) {
            const clone = response.clone();
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(request, clone);
            });
          }
          return response;
        } catch {
          const cached = await caches.match('/');
          return cached || new Response('Offline', { status: 503 });
        }
      })()
    );
    return;
  }

  // ======================================
  // Strategy 5: Stale-While-Revalidate for
  // everything else (images, fonts, etc.)
  // ======================================
  event.respondWith(
    caches.match(request).then((cached) => {
      const fetchPromise = fetch(request).then((response) => {
        if (response.ok) {
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(request, response.clone());
          });
        }
        return response;
      });
      return cached || fetchPromise;
    })
  );
});

// ==========================================
// HELPER: Trim cache to max entries
// ==========================================
function trimCache(cacheName, maxItems) {
  caches.open(cacheName).then((cache) => {
    cache.keys().then((keys) => {
      if (keys.length > maxItems) {
        // Delete oldest entries
        const toDelete = keys.slice(0, keys.length - maxItems);
        toDelete.forEach((key) => cache.delete(key));
      }
    });
  });
}
