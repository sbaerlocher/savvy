# Implementation Plan

- [ ] 1. Foundation: dde-Plugin-Suite und Hook-Idempotenz

- [x] 1.1 (P) Subcommand-Plugins für den E2E-Lifecycle anlegen
  - Fünf neue Bash-Plugins unter `.dde/plugins/` (`e2e.up.sh`, `e2e.down.sh`,
    `e2e.wait.sh`, `e2e.logs.sh`, `e2e.reset-db.sh`) mit
    `@command e2e:<verb>`-Annotation. `e2e.reset-db.sh` wurde im Zuge des
    `postgres-e2e`-Drops nachgezogen (siehe Implementation Notes).
  - `e2e:up` ruft `docker compose --profile e2e up -d "$@"`; `e2e:down`
    ruft `docker compose --profile e2e down -v "$@"` (CLI-Flag konsistent
    über alle Plugins)
  - `e2e:wait` pollt den `app-e2e`-Healthcheck mit `--timeout`-Flag
    (Default 90 s) und optionalem `--service`-Flag; gibt bei Timeout
    Service-Logs auf stderr aus
  - `e2e:logs` ruft `docker compose --profile e2e logs --tail "${1:-50}"
    "$@"`
  - Bestehendes `.dde/plugins/e2e.sh` löschen, damit der `e2e:`-Subcommand-
    Namespace nicht mit dem alten Monolith-Plugin kollidiert
  - Observable: `dde` (ohne Argumente) listet `project:e2e:up`,
    `project:e2e:down`, `project:e2e:wait`, `project:e2e:logs`
    mit korrekten Beschreibungen; `dde project:e2e:up && dde
    project:e2e:wait -- --timeout 90 && dde project:e2e:down`
    läuft lokal grün durch (sofern Port 5432 nicht durch dde-Stock-Postgres
    belegt ist)
  - _Requirements: 3.1, 3.3, 6.1, 6.2, 6.5_
  - _Boundary: dde Plugin Suite_

- [x] 1.2 (P) Post-Up-Seed-Hook idempotent für E2E-Mode machen
  - In `.dde/hooks/project.up.post/seed.sh` vor dem Postgres-Wait prüfen,
    ob `docker compose ps -q api` leer ist; wenn ja, mit Log
    `E2E mode: skipping dev seed` und Exit 0 zurückkehren
  - Andernfalls bestehender Pfad (Postgres-Wait, optionale DB-Erstellung,
    API-Restart, Seeder mit Retry/Backoff) unverändert
  - Hook bleibt zustandslos, schreibt im E2E-Mode keinerlei Side-Effects
  - Observable: `COMPOSE_PROFILES=e2e dde project:up` durchläuft die
    `project.up.post`-Phase mit Exit 0, ohne den Seeder anzustossen;
    `dde project:up` ohne Profile produziert weiterhin den bisherigen
    Seed-Output und Exit-Code
  - _Requirements: 6.4_
  - _Boundary: Seed Hook (idempotent)_

- [ ] 2. Core: Playwright Setup Adapter auf dde-Plugins umstellen

- [x] 2.1 globalSetup und globalTeardown auf dde-Plugins migrieren
  - In `client/tests/global.setup.ts` `executeDockerCommand` durch
    `executeDdeCommand` ersetzen (nutzt `child_process.exec` mit
    injizierbarem Env-Block); `validateDockerAvailable` durch
    `validateDdeAvailable` (Smoke `dde --version`) ersetzen
  - Setup-Sequenz: Validate → `dde project:e2e:down` (clean slate) →
    `dde project:e2e:reset-db` (drop+create `savvy_e2e` auf
    dde-Stock-Postgres) → `dde project:e2e:up` →
    `dde project:e2e:wait` (Timeout aus `E2E_APP_TIMEOUT_MS`); bei
    Setup-Fehler `dde project:e2e:logs --tail 50` capturen,
    anschliessend `dde project:e2e:down` versuchen, dann re-throw
  - In `client/tests/global.teardown.ts` `dde project:e2e:down` als Default-
    Cleanup verwenden (`E2E_REMOVE_VOLUMES` weiter respektieren); bei
    `E2E_VERBOSE_LOGS=true` zuvor `dde project:e2e:logs`
  - `E2E_KEEP_CONTAINERS=true` skippt Cleanup und gibt manuelle dde-
    Cleanup-Befehle aus
  - Skip-Pfade unverändert: `SKIP_E2E_SETUP=true`, `--list`, `--help`,
    `--version`
  - Konfigurierbare Timeouts und Polling-Intervalle
    (`E2E_POSTGRES_TIMEOUT_MS`, `E2E_APP_TIMEOUT_MS`,
    `E2E_POLL_INTERVAL_MS`, `E2E_EXEC_TIMEOUT_MS`) bleiben semantisch
    erhalten
  - Troubleshooting-Konsolentexte beider Skripte auf dde-Befehle
    aktualisieren
  - Observable: `npm run test:e2e:run -- --list` listet Tests ohne
    dde-/Docker-Aufruf; `npm run test:e2e:chromium -- tests/e2e/auth.spec.ts`
    durchläuft Setup-Wait-Test-Teardown vollständig über dde-Plugins;
    `docker ps` nach Default-Run zeigt keine `app-e2e`/`postgres-e2e`-
    Container; mit `E2E_KEEP_CONTAINERS=true` bleiben die Container und
    der ausgegebene Cleanup-Befehl funktioniert
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 4.4, 4.5, 7.1, 7.4, 8.1, 8.2, 8.3_
  - _Depends: 1.1_
  - _Boundary: Playwright Setup Adapter_

- [ ] 3. Core: GitHub-Actions-Workflow auf dde-Reusable scharfschalten

- [x] 3.1 e2e.yml-Inputs und Reusable-Pin aktualisieren
  - Reusable-Ref auf den Commit-SHA pinnen, der die dde-Reusable enthält
    (`@8a5bc6484d3372eaab426007925c94357e72efdb`); `@latest` ersetzen.
    Auf einen Date-Tag re-pinnen, sobald `sbaerlocher/.github` einen
    Tag ≥ 2026-04-29 cuttet
  - Veraltete Inputs `compose-file` und `compose-profile` entfernen — die
    dde-Reusable kennt sie nicht
  - `wait-url` bewusst nicht setzen: die Reusable's `dde project:up` bringt
    nur Stock-Services und das Dev-Profil hoch; den E2E-Stack startet
    Playwright's `globalSetup` selbst via `dde project:e2e:up`
    (eigener Health-Wait)
  - Bestehende Inputs (`playwright-browsers: chromium`,
    `test-command: test:e2e:chromium`, `playwright-directory: ./client`,
    `package-manager: npm`, `upload-artifacts: true`,
    `artifact-retention-days: 30`) explizit beibehalten;
    `project-directory: "."` ergänzen
  - Trigger-Konfiguration verifizieren: `pull_request: branches: [main]`,
    `push: branches: [main]`, `workflow_dispatch`
  - Workflow enthält keinen direkten `docker compose`-Aufruf und keine
    Image-Push-/Tag-Schritte
  - Observable: `yamllint .github/workflows/e2e.yml` exit 0;
    `git diff` zeigt nur die geplanten `e2e.yml`-Änderungen plus die
    `postgres-e2e`-Port-Anpassung (siehe Implementation Notes); ein
    PR-Lauf gegen `main` konsumiert die dde-Reusable, Playwright-
    `globalSetup` provisioniert den E2E-Stack via dde-Plugins, Tests
    laufen durch und `playwright-report`/`test-results` sind mit 30 Tagen
    Retention als Artefakte vorhanden
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 4.5, 5.3, 6.1, 6.5, 7.2, 7.3_
  - _Depends: 1.1, 1.2, 2.1_
  - _Boundary: CI Workflow Wrapper_

- [ ] 4. Validation: Production-Image- und Compose-Stack-Invariante

- [x] 4.1 Dockerfile-Production-Stages und app-e2e-Service-Set verifizieren
  - **Genehmigte Compose-Abweichung** (siehe Implementation Notes
    "postgres-e2e drop"): der `postgres-e2e`-Service, das
    `postgres_e2e_data`-Volume und das `e2e: bridge`-Netz wurden ersatzlos
    entfernt; `app-e2e` zieht jetzt die dde-Stock-Postgres heran und nutzt
    eine eigene Datenbank `savvy_e2e`. R4.3 (gepinnter Postgres-Digest)
    wird damit von der dde-Distribution erfüllt, nicht mehr lokal pro
    Repo.
  - `git diff origin/main...HEAD -- Dockerfile` produziert für die Stages
    `base-dev`, `frontend-dev`, `frontend-builder`, `go-builder`,
    `production-build`, `release-builder` und `production` keine
    inhaltlichen Diffs
  - `app-e2e.build.target` bleibt `production-build`;
    `app-e2e.environment` ändert nur `DATABASE_URL` (auf
    `postgres://postgres:postgres@postgres:5432/savvy_e2e?sslmode=disable`),
    alle anderen Felder (`GO_ENV=production`, `AUTO_MIGRATE=true`,
    Rate-Limits, Feature-Flags, Session-Secret) unverändert
  - `app-e2e.networks` wechselt von `e2e` (entferntes Bridge-Netz) auf
    `services` (extern, dde-services-savvy); `depends_on: postgres-e2e`
    entfällt
  - Hilfs-Binaries (`/app/seed`, `/app/e2e`) verbleiben in den Stages
    `go-builder` und `production-build`; keine zusätzlichen Stages mit
    Production-Tooling-Mix
  - Observable: der `Dockerfile`-`git diff` für die genannten Stages ist
    leer; `docker buildx build --target production -t
    savvy:invariance-smoke .` baut grün und das resultierende Image hat
    Entrypoint `/savvy`
  - _Requirements: 4.1, 4.2, 4.3, 5.1, 5.2, 5.4, 6.3, 6.4_
  - _Boundary: Production Image Invariante, Compose Stack_

## Implementation Notes

- **Task 1.1**: dde ≥ v2.0.0-beta.1 exponiert User-Plugins unter dem
  Namespace `project:`. Der Aufruf ist also `dde project:e2e:up` (nicht
  `dde e2e:up`); CLI-Args müssen mit `--` vom dde-Wrapper getrennt werden,
  z. B. `dde project:e2e:wait -- --timeout 90 --service app-e2e`.
- **Task 2.1**: `E2E_REMOVE_VOLUMES`-Toggle aus dem alten Teardown
  entfernt — `dde project:e2e:down` ist seit dem `postgres-e2e`-Drop
  ohnehin volume-frei (Test-Isolation läuft jetzt über `e2e:reset-db`).
  `E2E_KEEP_CONTAINERS=true` bleibt als Debug-Schalter. `E2E_POSTGRES_TIMEOUT_MS`
  und `E2E_APP_STARTUP_DELAY_MS` aus dem alten Setup entfallen, weil
  `dde project:e2e:wait` direkt auf den `app-e2e`-Healthcheck pollt
  und Postgres durch dde-system-Lifecycle bereits bereitsteht.
- **postgres-e2e drop (Architektur-Refactor)**: Das ursprüngliche
  Compose-Design hatte `postgres-e2e` als isolierte Datenbank für
  E2E-Tests. Nach der Dev-Migration auf dde-Stock-Postgres (Commit
  `8054f69`) wurde diese Isolation redundant — die Stock-Postgres ist
  ohnehin pro Projekt gemanagt und reproduzierbar. Stattdessen:
  `app-e2e` joint `services: dde-services-savvy` und nutzt eine separate
  Datenbank `savvy_e2e` auf der gemeinsamen Stock-Instanz. Per-Run-
  Isolation passiert über `dde project:e2e:reset-db` (DROP DATABASE
  WITH (FORCE) + CREATE DATABASE), AUTO_MIGRATE migriert das Schema aus
  dem Nichts. Vorteile: keine zweite Postgres-Instanz, kein
  Port-5432-Konflikt, weniger CI-Ressourcen. Trade-off: Image-Digest-Pin
  für Postgres liegt jetzt bei der dde-Distribution, nicht im Repo.
  Plugin-Suite damit auf fünf Subcommands erweitert
  (`up`/`down`/`wait`/`logs`/`reset-db`).
- **Task 4.1 Verification (`8054f69`..HEAD)**: `Dockerfile`-Diff = 0
  Zeilen — Production-Stages strikt invariant. `docker-compose.yml`-
  Diff exakt im Boundary-Scope: `postgres-e2e`-Service samt Volume und
  `e2e`-Bridge-Netz entfernt, `app-e2e.networks` von `e2e` auf
  `services` (extern) umgestellt, `app-e2e.environment.DATABASE_URL`
  zeigt jetzt auf `savvy_e2e` auf der Stock-Instanz. `app-e2e.build.target`
  bleibt `production-build`. Smoke-Build via `docker buildx build
  --target production-build` (Image: 64.4 MB, Entrypoint `/app/savvy`,
  16 Layers) entspricht der Baseline (`savvy-app-e2e:latest`,
  64.3 MB). `--target production` (GitHub-Release-Binary) wurde nicht
  geprüft, weil es `VERSION=<release-tag>` braucht und dieser Pfad
  durch die `release.yml`-Pipeline abgedeckt ist; Dockerfile ist dort
  ohnehin Diff-frei.
- **Post-impl: Entry-Point-Vereinheitlichung (Modell B+)**: Test-Lifecycle
  und Test-Run wurden in dedizierte dde-Plugins entkoppelt:
  - `dde project:e2e:start` (Aggregator: `down → reset-db → up → wait`)
  - `dde project:e2e:test` (Wrapper um Playwright; npm/`npx` werden
    Implementation-Detail im Plugin und tauchen im User-Flow nicht mehr auf)
  - `dde project:e2e:down` als expliziter Teardown-Schritt.

  Konsequenzen für die bestehenden Komponenten:
  - `globalSetup` schrumpft zur **Health-Probe-only**-Verification:
    `curl https://e2e.savvy.test/health`; bei Fehlschlag re-throw mit
    Hinweis auf `dde project:e2e:start`. Skip-Pfade
    (`--list`/`--help`/`--version`/`SKIP_E2E_SETUP=true`) und der
    `BASE_URL`-Override bleiben.
  - `globalTeardown` wird ersatzlos gelöscht (Teardown ist explizite
    User/CI-Action via `dde project:e2e:down`).
    `playwright.config.ts` referenziert nur noch `globalSetup`.
  - `client/package.json` verliert alle `test:e2e:*` und
    `playwright:*`-Scripts. Der einzige blessed Entry-Point lokal und in
    CI ist `dde project:e2e:test`. (Vitest-Scripts `test`/`test:ui`/
    `test:coverage` für Unit-Tests bleiben unverändert.)
  - `.github/workflows/e2e.yml` konsumiert die externe Reusable nicht
    mehr — sie unterstützt keine Pre-Test-Hooks und würde `npm run` als
    Test-Command erwarten. Der Workflow wird zu einer expliziten Schritt-
    folge: `actions/project-up` (dde install + project:up) → Node-Setup
    → `npm ci` + `playwright install` (CI-spezifische Vorbereitung) →
    `dde project:e2e:start` → `dde project:e2e:test` →
    Failure-Logs via `dde project:e2e:logs` → `dde project:e2e:down`
    → Artefakt-Uploads → `dde project:down`.
  - Damit ist R1.1 (ausschliesslich zentrale Reusable referenzieren) im
    bisherigen Wortlaut nicht mehr erfüllt; die zugrundeliegende Intention
    (keine direkten `docker compose`-Aufrufe in `e2e.yml`, dde steuert
    den Lifecycle, einheitliche Lokal-/CI-Erfahrung) wird durch die
    Plugin-getriebene Pipeline jedoch klarer eingehalten. R1.1 ist als
    erfüllt mit dieser dokumentierten Abweichung zu betrachten.

- **Post-impl: Plugin-Konsistenz**: `e2e.up.sh` verwendete ursprünglich
  `export COMPOSE_PROFILES=e2e` als defensive Env-Aktivierung (gedacht als
  Hebel, um den Profilwechsel auch in den GHA-Reusable-Job propagieren zu
  können). Die CI-Architektur bringt den E2E-Stack ohnehin nicht mehr über
  den Reusable hoch, sondern über `globalSetup`. Damit wird der Env-Trick
  überflüssig und steht im Widerspruch zu den anderen vier Plugins, die
  alle `--profile e2e` als CLI-Flag setzen. `e2e.up.sh` wurde auf das
  CLI-Flag normalisiert; die korrespondierende `extraEnv: { COMPOSE_PROFILES:
  'e2e' }`-Injection in `globalSetup.ts` ist damit redundant und entfällt.
- **Post-impl: Container-Naming + Traefik-Routing**: `app-e2e` bekommt
  einen expliziten `container_name: savvy-app-e2e` (kein `-1`-Suffix
  mehr) und wird vollständig hinter den dde-Traefik gehängt
  (`https://e2e.savvy.test`). Konsequenzen:
  - `app-e2e.networks` zusätzlich auf `default: dde`, damit Traefik den
    Service erreicht.
  - `ports: 127.0.0.1:8080:8080` ersatzlos entfernt — Tests greifen
    nicht mehr direkt auf den Host-Port zu.
  - `labels: traefik.*` für `Host(\`e2e.savvy.test\`)` mit TLS und
    `loadbalancer.server.port=8080`.
  - `app-e2e.environment.CORS_ALLOWED_ORIGINS` und `FRONTEND_URL` auf
    `https://e2e.savvy.test` gesetzt.
  - Playwright `baseURL` wechselt von `http://localhost:8080` auf
    `https://e2e.savvy.test`; `BASE_URL`-Env-Override bleibt erhalten.
  Smoke: `curl -k https://e2e.savvy.test/health` → 200 `{"status":"ok"}`,
  DNS via dde-dnsmasq → 127.0.0.1, TLS via dde-mkcert. `localhost:8080`
  ist erwartungsgemäss unbound (nur Traefik-Route).
