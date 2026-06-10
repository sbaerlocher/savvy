# Requirements Document

## Project Description (Input)

Bestehende E2E-Tests sollen in der GitHub Action gegen eine mit dde
(<https://dde.sh>) hochgefahrene Umgebung laufen, statt direkt auf dem Runner.
Ziel: gleiche Test-Umgebung lokal wie in CI. Production-Images werden
unverändert verwendet. dde-Setup (system:install + project:up) wird als
Setup-Phase in den Workflow integriert.

## Introduction

E2E-Tests von Savvy laufen heute in GitHub Actions, indem Playwrights
`globalSetup` direkt `docker compose --profile e2e ...` auf dem Runner aufruft.
Lokale Entwicklung ist parallel auf **dde** (<https://dde.sh>) umgestellt
worden. Damit lokal und CI dieselbe Provisionierung nutzen, soll auch die
CI-Test-Umgebung über dde hochgefahren werden, und die lokale
Playwright-Provisionierung soll auf dde-Kommandos umgestellt werden. Der
Workflow erhält eine explizite Setup-Phase mit `dde system:install` gefolgt
von `dde project:up`. Soweit nötig, werden `Dockerfile`,
`docker-compose.yml` und `.dde/`-Konfiguration angepasst, ohne dass sich
Build-Logik oder Output der produktiven Runtime-Images ändern.

## Boundary Context

- **In scope**:
  - `.github/workflows/e2e.yml` (savvy-seitige Konsumation der zentralen
    Reusable Workflow)
  - `client/tests/global.setup.ts` und `client/tests/global.teardown.ts`
    (Provisionierung über dde)
  - Anpassungen an `Dockerfile` und `docker-compose.yml`, soweit nötig,
    damit `dde project:up` den E2E-Stack deterministisch hochfährt
  - Anpassungen an `.dde/`-Konfiguration (`config.yml`, Plugins, Hooks) für
    den E2E-Lifecycle

- **Out of scope**:
  - Die zentrale Reusable Workflow
    `sbaerlocher/.github/.github/workflows/e2e-dde.yml` selbst (externe
    Abhängigkeit)
  - Inhalte und Assertions der Playwright-Spezifikationen unter
    `client/tests/e2e/`
  - Helm-Chart und Kubernetes-Deployment unter `deploy/`
  - Goreleaser- und Release-Pipeline (`.goreleaser.yaml`, `release.yml`)
  - Lokale Nicht-E2E-Entwicklung (bereits auf dde umgestellt)

- **Adjacent expectations**:
  - Die externe Reusable Workflow stellt einen dde-fähigen GitHub-Runner
    bereit, akzeptiert Inputs für Browser-Auswahl, Test-Command,
    Artefakt-Upload und Aufbewahrungsdauer und kümmert sich um
    Layer-Caching ausserhalb dieses Specs.
  - dde (`dde.sh`) ist auf dem GitHub-Runner installierbar (über
    `dde system:install`).
  - Bestehende GitHub-Secrets und OAuth-/SMTP-Konfigurationen müssen für die
    E2E-Pipeline nicht erweitert werden.

## Requirements

### Requirement 1: E2E-Workflow nutzt dde-basierte Provisionierung

**Objective:** As CI/CD-Verantwortlicher, I want, dass die
GitHub-Actions-E2E-Pipeline die Test-Umgebung ausschliesslich über dde
hochfährt, so that die CI-Provisionierung mit der lokalen Provisionierung
übereinstimmt und keine Runner-spezifischen `docker compose`-Aufrufe mehr im
Workflow stehen.

#### Acceptance Criteria

1. The E2E-Workflow shall ausschliesslich die zentrale Reusable Workflow für
   dde-basierte E2E-Tests referenzieren und keinen direkten
   `docker compose`-Aufruf in `e2e.yml` enthalten.
2. When ein Pull Request gegen `main` geöffnet oder aktualisiert wird, the
   E2E-Workflow shall die dde-basierte Reusable Workflow auslösen.
3. When ein Push auf `main` erfolgt, the E2E-Workflow shall die dde-basierte
   Reusable Workflow auslösen.
4. Where ein manueller Trigger gewünscht ist, the E2E-Workflow shall ein
   `workflow_dispatch`-Event akzeptieren.
5. The E2E-Workflow shall der Reusable Workflow alle Inputs übergeben, die
   für Browser-Auswahl, Test-Command, Artefakt-Upload und Aufbewahrungsdauer
   der Artefakte nötig sind, ohne dass die Reusable Workflow Default-Werte
   raten muss.

### Requirement 2: dde-Setup-Phase im CI-Job

**Objective:** As CI/CD-Verantwortlicher, I want eine explizite Setup-Phase
im E2E-Job, die `dde system:install` und `dde project:up` ausführt, so that
die E2E-Tests gegen eine vollständig durch dde provisionierte Umgebung laufen
statt gegen einen ad-hoc auf dem Runner gestarteten Stack.

#### Acceptance Criteria

1. The E2E-Job shall vor dem Test-Run `dde system:install` ausführen, sodass
   dde auf dem Runner verfügbar ist.
2. When `dde system:install` erfolgreich abgeschlossen ist, the E2E-Job shall
   `dde project:up` ausführen, um den Test-Stack zu starten.
3. When `dde project:up` startet, the E2E-Job shall warten, bis alle für E2E
   erforderlichen Services ihren `healthy`-Zustand erreicht haben, bevor
   Playwright gestartet wird.
4. If `dde system:install` oder `dde project:up` fehlschlägt, then the
   E2E-Job shall den Workflow als fehlgeschlagen markieren und keine Tests
   starten.
5. When der Test-Run abgeschlossen ist, the E2E-Job shall den dde-Stack
   einschliesslich Volumes wieder herunterfahren, ausser ein dokumentierter
   Debug-Schalter überschreibt dieses Verhalten.

### Requirement 3: Lokale Playwright-Provisionierung über dde

**Objective:** As Entwickler, I want, dass die lokale Playwright-Ausführung
denselben dde-Lifecycle nutzt wie der CI-Job, so that lokale Test-Ergebnisse
die CI-Ausführung verlässlich vorhersagen.

#### Acceptance Criteria

1. The Playwright-`globalSetup` shall die Test-Umgebung über dde-Kommandos
   starten und keinen direkten `docker compose ...`-Aufruf für
   Provisionierung mehr enthalten.
2. While Playwright lokal startet, the Playwright-`globalSetup` shall prüfen,
   dass dde verfügbar ist; if dde nicht verfügbar ist, then the
   Playwright-`globalSetup` shall mit einer eindeutigen Fehlermeldung
   abbrechen, die den nächsten Schritt für den Entwickler nennt.
3. The Playwright-`globalTeardown` shall den über dde gestarteten Stack über
   entsprechende dde-Kommandos einschliesslich Datenbank-Volumes wieder
   herunterfahren.
4. Where die Umgebungsvariable `E2E_KEEP_CONTAINERS=true` gesetzt ist, the
   Playwright-`globalTeardown` shall den Stack stehen lassen und einen
   klaren manuellen Cleanup-Befehl ausgeben.
5. Where Playwright im Info-Modus läuft (`--list`, `--help`, `--version`)
   oder `SKIP_E2E_SETUP=true` gesetzt ist, the Playwright-`globalSetup` shall
   keine Provisionierung durchführen und unmittelbar zurückkehren.

### Requirement 4: Test-Umgebungs-Parität zwischen lokal und CI

**Objective:** As Entwickler, I want, dass die per dde provisionierte
Test-Umgebung lokal und in CI dieselben Services, Image-Versionen und
Konfiguration verwendet, so that "works on my machine"-Abweichungen
entfallen.

#### Acceptance Criteria

1. The dde-Konfiguration shall denselben Service-Satz für E2E lokal und in
   CI definieren, ohne CI-spezifische Service-Definitionen.
2. The Test-Anwendung shall in beiden Umgebungen aus derselben
   Dockerfile-Stage gebaut werden.
3. The Datenbank- und weitere abhängige Services shall in beiden Umgebungen
   denselben gepinnten Image-Digest verwenden.
4. The E2E-relevanten Umgebungsvariablen (Feature-Flags, Rate-Limit-Werte,
   OTel-Status) shall in beiden Umgebungen identisch gesetzt sein, sofern
   sie das Test-Verhalten beeinflussen.
5. The Playwright-Test-Command shall in beiden Umgebungen dasselbe
   npm-Script aufrufen.

### Requirement 5: Production-Runtime-Image bleibt invariant

**Objective:** As Release-Verantwortlicher, I want, dass die produktiven
Runtime-Images (lokales `production-build` und Release-Binary-basiertes
`production`) unverändert bleiben, so that Container-Konsumenten in
Production keinerlei Verhaltens- oder Build-Änderung wahrnehmen.

#### Acceptance Criteria

1. The Dockerfile shall die Stages `production-build` und `production` so
   belassen, dass Image-Layout, Entrypoint und Größe gegenüber dem Stand vor
   Umsetzung dieses Specs unverändert bleiben.
2. The CI- und lokalen Build-Pfade für das Production-Image shall denselben
   Output liefern wie vor Umsetzung dieses Specs (verifizierbar über
   Image-Größe, Layer-Anzahl und Entrypoint).
3. The E2E-Pipeline shall keine produktiven Image-Tags überschreiben oder
   neu nach `:latest` oder `:v<x>` taggen.
4. Where für E2E ein zusätzliches Hilfs-Binary oder eine zusätzliche Stage
   benötigt wird, the Dockerfile shall dieses ausschliesslich in
   Nicht-Production-Stages bündeln.

### Requirement 6: Anpassbarkeit von Dockerfile und Compose für dde-Integration

**Objective:** As Maintainer, I want, dass `Dockerfile` und
`docker-compose.yml` so strukturiert sind, dass `dde project:up` den
E2E-Stack deterministisch und vollständig hochfährt, so that ausserhalb des
dde-Lifecycles keine zusätzlichen `docker compose --profile`-Schritte nötig
sind.

#### Acceptance Criteria

1. The `docker-compose.yml` shall die für E2E benötigten Services so
   deklarieren, dass `dde project:up` sie lokal und in CI vollständig
   startet, ohne dass der Aufrufer ein Profil oder weitere Flags zusätzlich
   angeben muss.
2. The `.dde/`-Konfiguration shall die für E2E benötigten Services in
   `config.yml` oder einem dde-Plugin so registrieren, dass dde deren
   Lifecycle steuern kann.
3. The Dockerfile shall die Stages für Frontend-Build, Go-Build und
   Hilfs-Binaries so trennen, dass eine reine Production-Build-Reihenfolge
   möglich bleibt, ohne E2E-spezifische Stages zu durchlaufen.
4. While E2E läuft, the Anwendung shall im selben Build-Modus wie in
   Production betrieben werden (statisches Frontend eingebettet,
   `GO_ENV=production`, `AUTO_MIGRATE=true`).
5. If ein abhängiger Service nicht innerhalb des konfigurierten
   Health-Check-Timeouts gesund wird, then the dde-Bring-up-Schritt shall
   fehlschlagen und einen Fehler liefern, der die Service-Logs enthält oder
   darauf verweist.

### Requirement 7: Diagnose und Artefakte bei Fehlschlägen

**Objective:** As Entwickler oder CI-Reviewer, I want, dass fehlgeschlagene
E2E-Runs Logs und Playwright-Artefakte liefern, so that ich Ursachen ohne
lokale Reproduktion nachvollziehen kann.

#### Acceptance Criteria

1. If ein Service-Healthcheck während des Setups scheitert, then the
   Setup-Schritt shall die letzten Logs der betroffenen dde-Services in der
   Job-Ausgabe sichtbar machen.
2. If der Test-Run scheitert, then the E2E-Workflow shall den
   Playwright-HTML-Report und Playwright-Traces als Artefakt hochladen.
3. The hochgeladenen Artefakte shall mindestens 30 Tage aufbewahrt werden.
4. Where `E2E_VERBOSE_LOGS=true` gesetzt ist, the Playwright-`globalTeardown`
   shall Service-Logs der dde-Stack-Container in die Konsole schreiben.

### Requirement 8: E2E-Test-Spezifikationen bleiben unverändert

**Objective:** As Test-Autor, I want, dass die Migration auf dde keine
Änderungen an den Playwright-Test-Specs erzwingt, so that die fachliche
Test-Coverage stabil bleibt und Reviews fokussiert sind.

#### Acceptance Criteria

1. The Migration shall keine Test-Datei unter `client/tests/e2e/` inhaltlich
   verändern.
2. The Migration shall keine bestehenden Selektoren, Test-IDs oder Fixtures
   invalidieren.
3. The npm-Scripts unter `client/package.json` für E2E (`test:e2e`,
   `test:e2e:chromium`, `test:e2e:run`, `test:e2e:firefox`, `test:e2e:mobile`,
   `test:e2e:debug`, `test:e2e:headed`, `test:e2e:ui`) shall weiterhin wie
   bisher aufrufbar sein und denselben Lifecycle auslösen.
