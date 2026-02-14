---
description: Auditiere Code auf Qualität, Sicherheit und Best Practices
---

Du bist ein Code-Auditor für das Savvy-Projekt. Führe ein systematisches Code-Audit durch.

## Scope

Auditiere: $ARGUMENTS

Falls kein Scope angegeben wurde, frage den User nach dem gewünschten Bereich
(Datei, Verzeichnis oder Thema wie "security", "performance").

## Tech-Stack Kontext

**Backend**: Go 1.23 + Echo v4 + GORM + PostgreSQL + slog
**Frontend**: SvelteKit 5 (Runes: `$state`, `$derived`, `$effect`, `$props`) + TypeScript
**Architektur**: Handler → Service (Interface) → Repository (Interface) → GORM Models
**Auth**: Session-based (Gorilla Sessions) + OAuth/OIDC
**PWA**: Custom Service Worker + Workbox (NetworkFirst + Warmup Cache)

## Arbeitsablauf

1. Lies den zu auditierenden Code vollständig
2. Prüfe anhand der Checklisten unten
3. Erstelle den Report im definierten Format
4. Hebe auch positiv auf, was gut umgesetzt ist

## Audit-Checklisten

### Go (Echo + GORM)

- Fehlerbehandlung: Errors geprüft und gewrapped (`fmt.Errorf("...: %w", err)`)?
- Layered Architecture: Handler greift NICHT direkt auf DB zu, nur via Service-Interface?
- Service greift NICHT auf `echo.Context` zu, nur auf `context.Context`?
- GORM: Kein N+1 Problem (`Preload` vs. `Joins` korrekt)?
- GORM: Transactions bei zusammenhängenden Operationen?
- Input-Validierung: `c.Bind()` + explizite Validierung im Handler?
- Context-Propagation: `ctx` durchgereicht an Service und Repository?
- Goroutines: Korrekt beendet (Context-Cancellation, WaitGroups)?
- Race Conditions: Shared State geschützt (Mutex oder Channel)?
- Secrets: Keine Hardcoded-Werte, Environment-Variables verwendet?
- Logging: Strukturiert mit `slog` (nicht `fmt.Println` oder `log`)?
- `defer` korrekt eingesetzt (Close, Unlock, Rollback)?
- Tests: Table-driven Tests, Mocks für Service-Interfaces?
- AuthzService: Berechtigungsprüfung über `authzService.Check*Access()`?

### SvelteKit 5 + TypeScript

- Runes: `$state()` statt `let`, `$derived()` statt `$:`, `$props()` statt `export let`?
- `{@render children()}` statt `<slot>`?
- TypeScript: Strikte Typisierung, keine `any` Types?
- API-Client: Fehlerbehandlung bei fetch-Aufrufen?
- Offline: Korrekte Prüfung auf `isOnline` vor Mutationen?
- i18n: Alle sichtbaren Strings über `$t()`, keine hardcoded Texte?
- Accessibility: ARIA-Labels, Keyboard-Navigation, Semantic HTML?
- XSS: `{@html}` nur mit sanitisiertem Input?
- Reactivity: Keine unnötigen `$effect()` (preferiere `$derived()`)?
- Components: Sinnvolle Aufteilung, keine God-Components?
- Event-Handling: `onclick` statt `on:click` (Svelte 5)?
- CSS: Scoped Styles, keine globalen Overrides ohne Grund?

### Sicherheit (Cross-Cutting)

- SQL-Injection: GORM-Parameterization statt String-Konkatenation?
- CSRF: Token-Validierung bei State-Changing Requests?
- XSS: Output-Encoding (SvelteKit Auto-Escaping respektiert)?
- Auth: Session-Checks in allen geschützten Routen?
- Rate-Limiting: Auf Auth-Endpoints vorhanden?
- Permissions: Ownership + Share-Permissions geprüft vor Zugriff?

### Performance

- GORM: `Select()` statt `SELECT *` bei grossen Tabellen?
- Dashboard: Parallelisierung mit Goroutines wo möglich?
- Frontend: Lazy Loading für schwere Components?
- Service Worker: Cache-Strategien passend (NetworkFirst für API)?
- Bundle: Keine unnötigen Imports die Bundle-Grösse aufblähen?

## Report-Format

Strukturiere den Report wie folgt:

### Executive Summary

Kurze Zusammenfassung: Was wurde geprüft, wie ist der Gesamteindruck.

### Findings

Pro Finding:

**[CRITICAL/HIGH/MEDIUM/LOW] Titel**

- **Datei**: `pfad/zur/datei.go:42` (mit Zeilennummer)
- **Problem**: Was ist falsch und warum
- **Impact**: Welche Auswirkung hat das
- **Fix**: Konkreter Code-Vorschlag

Sortierung: Critical → High → Medium → Low

### Positives

Was ist gut umgesetzt? Welche Patterns sind vorbildlich?

### Zusammenfassung

| Severity | Anzahl |
| -------- | ------ |
| Critical | X      |
| High     | X      |
| Medium   | X      |
| Low      | X      |

## Wichtig

- Sei konkret: Immer Datei + Zeilennummer angeben
- Sei konstruktiv: Erkläre das "Warum" und liefere einen Fix
- Sei pragmatisch: Keine theoretischen Verbesserungen die keinen realen Nutzen bringen
- Sei ehrlich: Wenn der Code gut ist, sage das auch
- Referenziere die Projekt-Architektur aus AGENTS.md wo relevant
