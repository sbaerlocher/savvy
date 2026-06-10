# Research & Design Decisions

## Summary

- **Feature**: `e2e-tests-on-dde`
- **Discovery Scope**: Extension (CI- und Lokal-Provisionierungslayer)
- **Key Findings**:
  - Die externe Reusable Workflow `sbaerlocher/.github/.github/workflows/e2e-dde.yml`
    delegiert an die Composite-Action `actions/project-up@2026-04-28`, die
    `dde system:install` + `dde project:up` ausführt und einen optionalen
    `wait-url`-Health-Poll bietet.
  - dde verwaltet zwei Stock-Services (`postgres`, `mailpit`) ausserhalb von
    `docker-compose.yml`; das Projekt deklariert seine eigenen Services im
    Compose und unterscheidet `dev`-, `e2e`- und `observability`-Profile.
  - Compose-Profile lassen sich zur Laufzeit über die Umgebungsvariable
    `COMPOSE_PROFILES` aktivieren, ohne dass der Aufrufer Compose-Flags
    explizit setzen muss; dde erbt diese Variable.

## Research Log

### Reusable Workflow Contract

- **Context**: Klären, was `e2e-dde.yml@latest` automatisch leistet, damit der
  Spec nicht fälschlich Aufgaben übernimmt, die schon dort gehandhabt sind.
- **Sources Consulted**: lokal verfügbares
  `~/Developer/sbaerlocher/.github/.github/workflows/e2e-dde.yml` und der
  referenzierte Composite-Step `actions/project-up@2026-04-28`.
- **Findings**:
  - Inputs: `dde-version`, `project-directory`, `wait-url`, `wait-timeout`,
    `node-version`, `package-manager`, `playwright-directory`, `test-command`,
    `playwright-browsers`, `upload-artifacts`, `artifact-retention-days`.
  - Steps: Checkout → `project-up` (install + `dde project:up` + optional
    Health-Poll) → `dde project:ps` → Setup Node.js + PM → Install Playwright
    Browsers → Run Test → `dde project:logs` und `project:ps` als Artefakt
    bei Failure → Upload Playwright-Report + Test-Results → `dde project:down`.
  - Permissions: Job benötigt nur `contents: read`; ein separater
    `comment`-Job mit `pull-requests: write` postet PR-Kommentare.
- **Implications**:
  - Savvy-seitig genügt es, korrekte Inputs zu setzen — Health-Wait,
    Diagnose-Logs und Teardown übernimmt die Reusable Workflow.
  - Die Reusable Workflow ruft generisch `dde project:up`; jegliche
    E2E-spezifische Konfiguration muss aus dem Projekt selbst kommen
    (via `.dde/`-Konfiguration, Compose-Profile oder Env-Vars).

### dde Stock-Services und Hooks

- **Context**: Verstehen, welche Services `dde project:up` ohne weiteres
  startet und wie Hooks den Lebenszyklus erweitern.
- **Sources Consulted**: `.dde/config.yml`, `.dde/hooks/project.up.post/seed.sh`,
  `.dde/plugins/e2e.sh`, `.dde/plugins/observability.sh`.
- **Findings**:
  - `.dde/config.yml` listet `postgres` und `mailpit` als Stock-Services
    (von dde global verwaltet, nicht aus dem Compose).
  - `seed.sh` läuft als Post-Up-Hook, sucht den dev-`api`-Container und
    bricht andernfalls mit Exit 1 ab — ist also strikt dev-spezifisch.
  - Plugins sind jeweils ein Bash-Skript mit `@command`-Annotation; das
    bestehende `e2e`-Plugin wrappt `docker compose --profile e2e up`.
- **Implications**:
  - Für E2E muss der Seed-Hook idempotent werden (no-op, wenn der dev-`api`
    nicht existiert), sonst zerlegt er den E2E-Run.
  - `dde project:up` aktiviert die Compose-Default-Services. Über
    `COMPOSE_PROFILES=e2e` lassen sich zusätzlich die E2E-Services
    aktivieren, ohne dass `--profile`-Flags an dde durchgereicht werden.

### Compose-Struktur und Image-Inventar

- **Context**: Sicherstellen, dass die Production-Image-Invariante (R5)
  einhaltbar ist und die `e2e`-Profile-Services kompatibel mit
  `dde project:up` sind.
- **Sources Consulted**: `Dockerfile`, `docker-compose.yml`.
- **Findings**:
  - Stages: `base-dev` → `backend-dev` (Air), `frontend-dev` (Vite),
    `frontend-builder`, `go-builder`, `production-build` (Distroless,
    bündelt `savvy`, `seed`, `e2e`-Hilfs-Binary), `release-builder`,
    `production` (Distroless aus GitHub-Release).
  - Compose-Services im `e2e`-Profil:
    - `app-e2e` baut `production-build` mit `VERSION=e2e-test` und
      Entrypoint `/app/e2e`.
    - `postgres-e2e` ist Postgres 18 mit gepinntem Digest und eigenem
      `postgres_e2e_data`-Volume.
  - Compose-Netzwerke: `default` (extern, `dde`), `services` (extern,
    `dde-services-savvy`), `e2e` (intern, `bridge`).
- **Implications**:
  - `production`- und `production-build`-Stages bleiben unangetastet; das
    `e2e`-Hilfsbinary ist bereits Teil dieser Distroless-Stage und kein
    Anlass für Production-Image-Drift.
  - Die `e2e`-Bridge ist isoliert von dde-Netzwerken — `app-e2e` erreicht
    weder `postgres` noch `mailpit` von dde. Das ist gewünscht, weil
    E2E gegen `postgres-e2e` läuft (volume-isoliert, pro Run zurücksetzbar).

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| Eigene CI-Workflow ohne Reusable | savvy schreibt eigene `dde system:install` + Compose-Schritte | Volle Kontrolle | Duplikation, Drift gegenüber anderen Projekten | Verstösst gegen Workspace-Standard (`STANDARDS.md`) |
| Reusable + Compose-Profil per Env-Var | `e2e-dde.yml@latest` konsumieren, `COMPOSE_PROFILES=e2e` als Job-Env, `wait-url` für Health | Minimal-invasive Konsumation, keine Forks | Reihenfolge zwischen dde-Stock-Services und E2E-Services nicht steuerbar | **Gewählt** |
| Reusable + dde-Plugin als zusätzlichem Step | Reusable bleibt unverändert; Plugin `dde e2e:up` läuft als zusätzlicher Step zwischen `project:up` und Tests | Klare Trennung Stock vs. E2E | Erfordert Forken/Erweitern der Reusable, weil keine Post-Up-Step-Hooks existieren | Verworfen — gegen die Reusable arbeiten erhöht Maintenance |

## Design Decisions

### Decision: Compose-Profil-Aktivierung über `COMPOSE_PROFILES` statt Compose-Restruktur

- **Context**: `dde project:up` aktiviert per Default kein Compose-Profil; die
  E2E-Services liegen aktuell hinter dem Profil `e2e`.
- **Alternatives Considered**:
  1. Compose so umbauen, dass `app-e2e`/`postgres-e2e` zum Default werden
     und der Dev-Stack hinter `dev` rückt — bricht das gerade frisch
     migrierte Lokal-Dev-Setup.
  2. Über die Umgebungsvariable `COMPOSE_PROFILES=e2e` aktivieren — dde
     erbt Env-Vars aus dem Aufrufer (Job/Shell), Compose erkennt sie nativ.
- **Selected Approach**: Option 2. Im CI-Workflow als `env`-Block, im
  Playwright-`globalSetup` programmatisch vor dem `child_process.exec`-Aufruf.
- **Rationale**: Erfüllt R6.1 (kein zusätzliches Flag am `dde`-Aufruf), ist
  reversibel und tangiert weder den Dev-Stack noch die Production-Stages.
- **Trade-offs**: dde bringt zusätzlich seine Stock-Services (`postgres`,
  `mailpit`) hoch, die in E2E weitgehend ungenutzt bleiben — geringe
  Ressourcen-Mehrkosten gegenüber einem vollständigen Stock-Bypass.
- **Follow-up**: `seed.sh`-Hook muss idempotent werden (skip, wenn der
  dev-`api`-Container nicht läuft), sonst scheitert `project:up.post`
  in E2E.

### Decision: Health-Wait über Reusable `wait-url`-Input statt eigener Polling-Logik

- **Context**: Die Reusable Workflow akzeptiert `wait-url` und pollt bis
  `2xx/3xx` (TLS-Fehler ignoriert) oder `wait-timeout` Sekunden.
- **Alternatives Considered**:
  1. Eigene Polling-Logik im savvy-Workflow oder in einem dde-Plugin.
  2. `wait-url: http://localhost:8080/health` an die Reusable übergeben,
     `wait-timeout` defensiv setzen.
- **Selected Approach**: Option 2.
- **Rationale**: Vorhandene Mechanik nutzen; weniger Code, identische
  Semantik wie andere Projekte, die die Reusable konsumieren.
- **Trade-offs**: Reusable kennt nur einen `wait-url`. Reicht hier, weil
  `app-e2e:8080/health` der einzige relevante Indikator ist (`postgres-e2e`
  ist über dessen Healthcheck via `app-e2e.depends_on` abgesichert).
- **Follow-up**: Lokales `globalSetup` muss eine eigene Health-Wartelogik
  behalten, da `wait-url` eine GitHub-Actions-Funktion ist.

### Decision: Lokale Playwright-Provisionierung über dde-Plugin-Subcommands

- **Context**: `globalSetup`/`globalTeardown` rufen heute direkt `docker
  compose --profile e2e ...`. Für Parität (R3) sollen dde-Kommandos genutzt
  werden.
- **Alternatives Considered**:
  1. `globalSetup` ruft `docker compose ...` weiter und setzt nur die
     Service-Namen um (Status quo des Branches).
  2. `globalSetup` ruft `dde e2e:up`/`dde e2e:down`/`dde e2e:wait`-Plugins,
     die intern Compose mit `--profile e2e` wrappen.
  3. `globalSetup` ruft direkt `dde project:up`/`dde project:down` mit
     `COMPOSE_PROFILES=e2e` und delegiert Health-Wait an einen separaten
     dde-Plugin oder eine TS-Wartelogik.
- **Selected Approach**: Option 2 — eigenständige `e2e:up`, `e2e:down`,
  `e2e:wait`, `e2e:logs`-Plugins, weil sie den Lebenszyklus klar benennen
  und auch ad-hoc per CLI nutzbar sind.
- **Rationale**: Halt am dde-Idiom (Plugins als Kommandos), bessere
  Diagnose-Pfade, kein Verstecken des E2E-Lifecycle hinter
  `dde project:up`-Semantik, die in CI ohnehin von der Reusable kontrolliert
  wird.
- **Trade-offs**: CI und Lokal benutzen unterschiedliche Entry-Points
  (`dde project:up + COMPOSE_PROFILES=e2e` in CI vs. `dde e2e:up` lokal).
  Beide aktivieren denselben Compose-Profile-Satz und bauen dasselbe
  Image, also bleibt die Test-Umgebung-Parität (R4) erhalten.
- **Follow-up**: Plugin-Skripte müssen klar dokumentiert werden, damit
  Entwickler ohne Plugin-Wissen den Lifecycle nachvollziehen können.

### Decision: Production-Image-Invariante über Verhaltens-Vergleich

- **Context**: R5 fordert, dass die Stages `production-build` und
  `production` unverändert bleiben.
- **Alternatives Considered**:
  1. Hash der Image-Layer pinnen und im Workflow vergleichen.
  2. Implizit: bei Reviews nur sicherstellen, dass keine Änderung an den
     Production-Stages passiert; CI bricht bei Image-Diff nicht ab.
  3. Stage-Diff per Snapshot-Test validieren (Image-Größe, Entrypoint,
     Layer-Anzahl).
- **Selected Approach**: Option 3 als manueller Verifizierungspunkt im
  PR-Review, kein automatisierter Gate. Begründet durch geringe Tooling-
  Kosten und niedrige Drift-Wahrscheinlichkeit.
- **Rationale**: Volle Hash-Pin-Verifizierung wäre überengineered; ein
  expliziter Review-Punkt reicht hier, zumal die Stages nur kosmetisch
  anfasst werden dürfen.
- **Trade-offs**: Manuelles Review statt automatisierter Check.

### Decision: `postgres-e2e`-Service entfernt zugunsten der dde-Stock-Postgres

- **Context**: Während Task 3.1 stellte sich heraus, dass die GitHub-
  Actions-Reusable-Workflow keinen Job-Level `env:`-Block akzeptiert (GHA-
  Limitation für reusable-workflow-Jobs) und die Composite-Action
  `actions/project-up` `dde project:up` ohne Profil-Steuerung aufruft.
  Gleichzeitig wurde sichtbar, dass `postgres-e2e` und der dde-Stock-
  Postgres beide Port 5432 binden, was lokal und in CI zu Konflikten
  führt.
- **Alternatives Considered**:
  1. `postgres-e2e.ports` droppen (intern bleibt erreichbar) — löst nur
     den Port-Konflikt, hinterlässt aber zwei parallele Postgres-Instanzen.
  2. Reusable-Workflow um einen `compose-profiles`-Input erweitern —
     externer Repo-Change, ausserhalb der Spec-Boundary.
  3. `postgres-e2e` ersatzlos entfernen, `app-e2e` joint
     `dde-services-savvy` und nutzt eine dedizierte Datenbank
     `savvy_e2e` auf der Stock-Instanz; Per-Run-Isolation via
     `dde project:e2e:reset-db` (DROP DATABASE WITH (FORCE) +
     CREATE DATABASE).
- **Selected Approach**: Option 3.
- **Rationale**: Konsistent mit dem Architektur-Schwenk in
  Commit `8054f69` (Dev nutzt dde-Stock-Postgres). Eine einzige
  Postgres-Instanz pro Projekt, weniger Container-Footprint, kein
  Port-Konflikt mehr. DB-Reset via DROP/CREATE ist mit Postgres 13+
  (dde liefert 18.x) genauso schnell wie ein Volume-Reset.
- **Trade-offs**: Image-Digest-Pin für Postgres liegt jetzt bei der
  dde-Distribution statt im Repo (R4.3 wird damit von dde erfüllt, nicht
  mehr lokal). Test-Isolation hängt davon ab, dass `e2e:reset-db`
  zuverlässig läuft — getestet via Smoke (`drop+create`, `e2e:up`,
  `e2e:wait`, `curl /health`).
- **Follow-up**: `e2e.down.sh` behält `-v` als Defensive (kein Volume
  vorhanden, aber zukunftssicher). `tasks.md`-Task 4.1 erfasst diese
  Compose-Abweichung als genehmigte Ausnahme.

## Risks & Mitigations

- **dde-Stock-Postgres und postgres-e2e koexistieren in CI** — keine
  fachliche Kollision, da `app-e2e` per Compose-Netzwerk an `postgres-e2e`
  angebunden ist; Mehrkosten = ein zusätzlicher leichter Container.
- **`seed.sh`-Hook bricht E2E-Lauf** — Hook idempotent machen: Erkennen,
  dass kein dev-`api`-Container existiert, und mit Exit 0 zurückkehren.
- **Branch-Schutz: Reusable Workflow `@latest`** — Verweis auf `@latest`
  bricht potenziell still bei Breaking Changes. Mitigation: nach
  Stabilisierung auf einen datierten Tag pinnen (z. B. `@2026-04-28`),
  analog zu anderen Workflow-Refs in den `sbaerlocher/*`-Repos.
- **`wait-url` über `localhost:8080`** — funktioniert nur, wenn `app-e2e`
  diesen Port auf den Runner publishrt. Compose tut dies bereits
  (`127.0.0.1:8080:8080`). Falls die Compose-Definition geändert wird,
  muss `wait-url` mitgepflegt werden.

## References

- `~/Developer/sbaerlocher/.github/.github/workflows/e2e-dde.yml` — externe
  Reusable Workflow (read-only Quelle in dieser Discovery)
- `~/Developer/sbaerlocher/.github/.github/actions/project-up@2026-04-28` —
  Composite Action für `dde system:install` + `project:up`
- `.dde/plugins/e2e.sh`, `.dde/hooks/project.up.post/seed.sh` — bestehender
  dde-Lifecycle im savvy-Projekt
- `Dockerfile` Multi-Stage-Definition (Stages: `base-dev`, `backend-dev`,
  `frontend-dev`, `frontend-builder`, `go-builder`, `production-build`,
  `release-builder`, `production`)
- `docker-compose.yml` mit Profilen `dev`, `e2e`, `observability`
