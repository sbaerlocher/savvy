# PWA (Progressive Web App) Features

**Status**: ✅ Implemented (v1.1.0)
**Last Updated**: 2026-01-26

---

## 📱 Overview

Savvy System ist eine Progressive Web App (PWA), die offline funktioniert und auf dem Homescreen installiert werden kann.

### Key Features

- ✅ **Offline Viewing** - Karten, Gutscheine und Geschenkkarten offline ansehen
- ✅ **Service Worker** - Automatisches Caching für schnelleren Zugriff
- ✅ **Installierbar** - Als App auf dem Homescreen (iOS, Android, Desktop)
- ✅ **Offline-Erkennung** - Visuelles Feedback wenn Verbindung verloren geht
- ✅ **Smart Caching** - Network-First Strategie mit Cache-Fallback

---

## 🚀 Installation

### iOS (Safari)

1. Öffne Savvy in Safari
2. Tippe auf **Teilen-Button** (Rechteck mit Pfeil nach oben)
3. Scrolle runter und wähle **"Zum Home-Bildschirm"**
4. Tippe **"Hinzufügen"**

### Android (Chrome)

1. Öffne Savvy in Chrome
2. Tippe auf **Menü** (3 Punkte)
3. Wähle **"Zum Startbildschirm hinzufügen"**
4. Bestätige mit **"Hinzufügen"**

### Desktop (Chrome/Edge)

1. Öffne Savvy in Chrome/Edge
2. Klicke auf **Install-Icon** in der Adressleiste (➕)
3. Oder: Menü → **"App installieren"**

---

## 📖 Offline-Funktionalität

### Was funktioniert OFFLINE?

✅ **Anzeigen (Read-Only)**:
- Alle eigenen Karten/Gutscheine/Geschenkkarten
- Geteilte Items von anderen Benutzern
- Favoriten durchsuchen
- Barcode-Details ansehen
- Dashboard mit Statistiken (gecachte Daten)

✅ **Features**:
- Filter & Sortierung (client-side)
- Suchfunktion (client-side)
- Navigation zwischen Seiten
- Barcode-Scanning (Camera API)

### Was funktioniert NICHT offline?

❌ **Schreibvorgänge**:
- Neue Items erstellen
- Bestehende Items bearbeiten
- Items löschen
- Sharing verwalten
- Favoriten hinzufügen/entfernen
- Transaktionen (Gift Cards)

❌ **Server-Abhängige Features**:
- Neue Shares empfangen
- Synchronisation mit anderen Geräten
- Echtzeit-Updates

---

## 🔧 Technische Details

### Service Worker

**Datei**: `static/service-worker.js`
**Strategie**: Network First, Cache Fallback
**Cache-Version**: `savvy-v1.0.0`

#### Gecachte Routes

**Statisch** (sofort beim Install):
```
- / (Home)
- /static/css/styles.css
- /static/js/bundle.js
- /offline (Fallback Page)
```

**Dynamisch** (beim ersten Besuch):
```
- /dashboard
- /cards
- /cards/:id
- /vouchers
- /vouchers/:id
- /gift-cards
- /gift-cards/:id
- /favorites
```

### Cache-Verhalten

```javascript
// 1. Versuche Network Request
fetch(request)
  .then(response => {
    // Cache die Response für später
    cache.put(request, response.clone());
    return response;
  })
  .catch(() => {
    // 2. Falls offline, nutze Cache
    return caches.match(request);
  });
```

### Offline-Erkennung

**Alpine.js Component** in `layout.templ`:

```javascript
window.offlineDetector = () => ({
  isOnline: navigator.onLine,

  init() {
    window.addEventListener('online', () => this.isOnline = true);
    window.addEventListener('offline', () => this.isOnline = false);
  }
});
```

**UI-Feedback**:
- Gelbes Banner oben bei Offline-Status
- Buttons deaktiviert mit Tooltip
- "Erneut versuchen" Button

---

## 🎨 UI-Anpassungen

### Offline-Banner

```html
<div x-show="!isOnline" class="bg-yellow-50 border-b border-yellow-200">
  <p>Offline-Modus</p>
  <button @click="checkConnection()">Erneut versuchen</button>
</div>
```

### Deaktivierte Buttons

**Buttons**:
```html
<button
  hx-delete="/cards/123"
  :disabled="!$root.isOnline"
  :class="!$root.isOnline ? 'opacity-50 cursor-not-allowed' : ''"
  :title="!$root.isOnline ? 'Löschen nur online möglich' : ''"
>
  Löschen
</button>
```

**Links**:
```html
<a
  href="/cards/new"
  @click="if (!$root.isOnline) {
    $event.preventDefault();
    alert('Erstellen nur online möglich');
  }"
  :class="!$root.isOnline ? 'opacity-50 cursor-not-allowed' : ''"
>
  Neue Karte
</a>
```

---

## 🐛 Troubleshooting

### Service Worker wird nicht registriert

**Problem**: Console zeigt keine Service Worker Registration

**Lösung**:
```bash
# 1. Browser-Cache leeren
# 2. Service Worker de-registrieren
# Developer Tools → Application → Service Workers → Unregister

# 3. Hard Reload
Cmd+Shift+R (macOS) / Ctrl+Shift+R (Windows)
```

### Alte Inhalte werden angezeigt

**Problem**: Änderungen werden nicht angezeigt

**Lösung**:
1. Cache-Version in `service-worker.js` erhöhen:
   ```javascript
   const CACHE_VERSION = 'savvy-v1.0.1';  // Increment
   ```
2. Deploy → Service Worker aktualisiert automatisch

### Offline-Banner erscheint nicht

**Problem**: `$root.isOnline` ist undefined

**Lösung**:
```html
<!-- Body muss Alpine.js data haben -->
<body x-data="offlineDetector()" x-init="init()">
```

---

## 📊 PWA Manifest

**Datei**: `static/manifest.json`

```json
{
  "name": "Savvy - Card Management System",
  "short_name": "Savvy",
  "start_url": "/",
  "display": "standalone",
  "theme_color": "#4F46E5",
  "icons": [
    { "src": "/static/icons/icon-192.png", "sizes": "192x192" },
    { "src": "/static/icons/icon-512.png", "sizes": "512x512" }
  ]
}
```

### Display Modes

- **`standalone`**: Looks like native app (no browser UI)
- **`fullscreen`**: Full-screen (for games)
- **`minimal-ui`**: Minimal browser UI
- **`browser`**: Regular browser tab

---

## 🔐 Security Considerations

### HTTPS Required

**PWA features benötigen HTTPS**:
- ❌ Service Worker auf HTTP (außer localhost)
- ❌ Camera API (Barcode-Scanner)
- ❌ Push Notifications (future)

**Exception**: `localhost` für Development

### Cache Poisoning

**Risiko**: Manipulierte Responses im Cache

**Mitigation**:
```javascript
// Nur 200 OK Responses cachen
if (response.status !== 200) {
  return response; // Don't cache errors
}
```

---

## 📈 Monitoring

### Service Worker Status

**Chrome DevTools**:
```
Application → Service Workers
- Status: Activated and running
- Update on reload: Checkbox für Development
```

### Cache Inspection

**Chrome DevTools**:
```
Application → Cache Storage
- savvy-v1.2.9-static (Static assets)
- savvy-v1.2.9-dynamic (Pages)
```

---

## 🧹 Cache Management (v1.3.1+)

### Automatisches Cache-Cleanup

Der Service Worker bereinigt automatisch alte Caches, um Speicherplatz zu sparen:

**1. Alte Cache-Versionen löschen**
- **Wann**: Bei jedem Service Worker Update (Aktivierung)
- **Was**: Alle `savvy-*` Caches außer aktuelle Version
- **Beispiel**:
  ```
  ✅ savvy-v1.3.1-static   (aktuell)
  ✅ savvy-v1.3.1-dynamic  (aktuell)
  ❌ savvy-v1.3.0-static   (gelöscht)
  ❌ savvy-v1.3.0-dynamic  (gelöscht)
  ```

**2. Duplikate entfernen** ⭐ NEU in v1.3.1
- **Problem**: Gleiche URL mit unterschiedlichen Request-Varianten (Headers, Query-Parameter)
- **Lösung**: Nur eine Version pro URL wird behalten
- **Wann**: Bei Service Worker Aktivierung
- **Beispiel**:
  ```
  ✅ /cards/123 (behalten)
  ❌ /cards/123?timestamp=1234 (gelöscht - Duplikat)
  ❌ /cards/123 (andere Headers) (gelöscht - Duplikat)
  ```

**3. Größenlimit für Dynamic Cache**
- **Limit**: 100 Einträge (konfigurierbar via `MAX_DYNAMIC_CACHE_SIZE`)
- **Strategie**: FIFO (First In, First Out) - älteste Einträge werden zuerst gelöscht
- **Wann**: Nach jedem neuen Cache-Eintrag
- **Beispiel**: Bei 101 Einträgen wird Eintrag #1 gelöscht

**4. Alte Cache-Einträge entfernen**
- **Max-Alter**: 7 Tage (konfigurierbar via `MAX_CACHE_AGE_DAYS`)
- **Wann**: Bei Service Worker Aktivierung
- **Prüfung**: Basiert auf `Date` Header der Response
- **Beispiel**: Einträge vom 2026-01-20 werden am 2026-01-27 gelöscht

### Cache-Konfiguration

**Datei**: `static/service-worker.js`

```javascript
const CACHE_VERSION = "savvy-v1.3.2";
const MAX_DYNAMIC_CACHE_SIZE = 100;    // Max Anzahl Einträge
const MAX_CACHE_AGE_DAYS = 7;          // Max Alter in Tagen
```

### Cache-Key Normalisierung (v1.3.1+)

Um Duplikate zu vermeiden, werden Cache-Keys normalisiert:

**Beim Speichern (cache.put)**:
```javascript
// Vor v1.3.1: Vollständiger Request mit Query-Parametern
cache.put(request, response);  // /cards/123?timestamp=1234

// Ab v1.3.1: Nur URL ohne Query-Parameter
const cacheUrl = new URL(request.url);
cacheUrl.search = '';  // Entferne Query-Parameter
const cacheRequest = new Request(cacheUrl.toString(), {
  method: 'GET',
  headers: { 'Accept': request.headers.get('Accept') || '*/*' }
});
cache.put(cacheRequest, response);  // /cards/123
```

**Beim Abrufen (cache.match)**:
```javascript
// Gleiches Normalisierungs-Pattern für Cache-Lookup
const cacheUrl = new URL(request.url);
cacheUrl.search = '';
const cacheRequest = new Request(cacheUrl.toString(), {
  method: 'GET',
  headers: { 'Accept': request.headers.get('Accept') || '*/*' }
});
const response = await cache.match(cacheRequest);
```

**Vorteile**:
- ✅ Keine Duplikate durch Query-Parameter
- ✅ Keine Duplikate durch verschiedene Headers
- ✅ Weniger Speicherverbrauch
- ✅ Konsistentes Caching-Verhalten

### Manuelles Cache-Löschen

**Option 1: Via Browser DevTools**
```
Chrome DevTools → Application → Cache Storage
→ Rechtsklick auf Cache → Delete
```

**Option 2: Via Service Worker Message**
```javascript
// Alle Caches löschen
navigator.serviceWorker.controller.postMessage({
  type: 'CLEAR_CACHE'
});
```

**Option 3: Service Worker Deregistrieren**
```javascript
// Service Worker komplett entfernen
navigator.serviceWorker.getRegistrations().then(registrations => {
  registrations.forEach(registration => registration.unregister());
});
```

### Update-Benachrichtigung

Benutzer werden automatisch benachrichtigt, wenn eine neue Version verfügbar ist:

- **Banner**: Erscheint unten rechts (Desktop) oder unten (Mobile)
- **Actions**:
  - "Jetzt aktualisieren" → Lädt Seite neu
  - "Später" → Versteckt Banner
- **Auto-Update**: Neuer Service Worker wird automatisch aktiviert (via `skipWaiting()`)

**Beispiel-Banner**:
```
🔄 Neue Version verfügbar
Eine aktualisierte Version der App ist bereit.
[Jetzt aktualisieren] [Später]
```

### Logging

Alle Cache-Operationen werden in der Browser-Konsole geloggt:

```javascript
[ServiceWorker] Activating version: savvy-v1.3.2
[ServiceWorker] Deleting 2 old caches: ["savvy-v1.3.1-static", "savvy-v1.3.1-dynamic"]
[ServiceWorker] Removed 47 duplicate entries from savvy-v1.3.2-dynamic
[ServiceWorker] Cache savvy-v1.3.2-dynamic exceeds limit (101/100), deleting 1 oldest entries
[ServiceWorker] Deleted 15 entries older than 7 days from savvy-v1.3.2-dynamic
[ServiceWorker] Activation complete, claiming clients
```

### Troubleshooting: Mehrere Cache-Namen mit gleicher Version

**Problem**: Browser DevTools zeigt mehrfach den gleichen Cache-Namen

![Beispiel](https://via.placeholder.com/400x200?text=savvy-v1.3.1-static+%2812x%29)

**Ursache**:
- Service Worker wurde mehrfach installiert (z.B. bei Development)
- Alte Caches wurden nicht korrekt gelöscht beim `activate` Event
- Browser-Timing: `activate` läuft manchmal nach `install`

**Lösung** (ab v1.3.3):

1. ✅ **Proaktives Cleanup beim Install**: Alte Caches werden bereits beim Install gelöscht, nicht erst beim Activate
2. ✅ **Manuelles Cleanup beim Page Load**: App sendet `CLEANUP_OLD_CACHES` Message
3. ✅ **Doppeltes Cleanup**: Sowohl in Install als auch in Activate Event

**Manuelle Bereinigung**:

**Option 1: Via Browser Console (Alle Caches löschen)**
```javascript
// ACHTUNG: Löscht ALLE Caches inkl. aktuelle Version
caches.keys().then(keys => {
  console.log('Deleting', keys.length, 'caches');
  keys.forEach(key => caches.delete(key));
  location.reload();
});
```

**Option 2: Via Service Worker Message (Nur alte Caches)**
```javascript
// Löscht nur alte Versionen, behält aktuelle
if (navigator.serviceWorker.controller) {
  navigator.serviceWorker.controller.postMessage({
    type: 'CLEANUP_OLD_CACHES'
  });
  setTimeout(() => location.reload(), 1000);
}
```

**Option 3: Browser DevTools (Manuell)**
```
1. Chrome DevTools → Application → Cache Storage
2. Rechtsklick auf alten Cache → Delete
3. Wiederhole für alle alten Versionen
4. Behalte nur aktuellste Version (z.B. savvy-v1.3.3-static/dynamic)
```

### Troubleshooting: Duplikate innerhalb eines Caches

**Problem**: Gleiche URLs erscheinen mehrfach **innerhalb** eines Caches

**Ursache** (vor v1.3.1):
- Query-Parameter erstellen separate Einträge: `/cards/123?t=1`, `/cards/123?t=2`
- Unterschiedliche Request-Headers erstellen Duplikate
- Browser sendet verschiedene Request-Varianten (z.B. bei HTMX)

**Lösung** (ab v1.3.1):
1. ✅ Cache-Key Normalisierung (Query-Parameter entfernt)
2. ✅ Automatische Duplikat-Entfernung beim Service Worker Update
3. ✅ Konsistente GET-Requests mit standardisierten Headers

### Network Tab

**Offline Testing**:
```
Network → Throttling → Offline
```

---

## 🚀 Future Enhancements

### Phase 1 (Current) ✅
- ✅ Service Worker mit Network-First
- ✅ Offline-Erkennung & UI-Feedback
- ✅ PWA Manifest
- ✅ Installierbar

### Phase 2 (Planned)
- ⏳ Background Sync (änderungen synchronisieren wenn online)
- ⏳ Push Notifications (neue Shares, Transaktionen)
- ⏳ Offline-Queue für Änderungen

### Phase 3 (Future)
- 🔮 IndexedDB für strukturierte Offline-Daten
- 🔮 Conflict Resolution für gleichzeitige Änderungen
- 🔮 Periodic Background Sync

---

## 📚 Resources

- [PWA Checklist](https://web.dev/pwa-checklist/)
- [Service Worker API](https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API)
- [Web App Manifest](https://developer.mozilla.org/en-US/docs/Web/Manifest)
- [Workbox (Google)](https://developers.google.com/web/tools/workbox) - Advanced SW library

---

## 🤝 Contributing

Bei Bugs oder Feature-Requests:
1. Erstelle Issue in GitHub
2. Beschreibe Problem + Browser + OS
3. Console Logs beifügen (DevTools → Console)

---

**Version**: 1.1.0
**Author**: Simon Bärlocher (@sbaerlocher)
**License**: MIT
