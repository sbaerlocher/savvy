# Structural baseline

Safety net for the layout refactor: proves that moving pages onto a shared
layout hierarchy does not change what they render.

Two kinds of tests, different jobs:

| Suite               | Job                                                                        | State before the refactor |
| ------------------- | -------------------------------------------------------------------------- | ------------------------- |
| `structure.spec.ts` | The **target**: one `h1`, single landmarks, one shared container, equal nav | deliberately **red**      |
| `baseline.spec.ts`  | The **guard**: pixel comparison against recorded screenshots                | green (records first)     |
| `axe.spec.ts`       | Accessibility violation **count** must not rise                            | green (records first)     |

`structure.spec.ts` failing is not a bug — it is the Phase 0 inconsistency
list in executable form. Every failure names a page that still builds its own
shell.

## Running

Needs the dev stack up (`dde project:up`, serving `https://savvy.test`).

```bash
dde project:structure:test
```

Single suite or single route:

```bash
dde project:structure:test -- tests/structure/structure.spec.ts
dde project:structure:test -- --grep dashboard
```

## Recording the baseline

First time only, or after each changed image has been reviewed:

```bash
STRUCTURE_RECORD_CONFIRM=yes dde project:structure:record
```

The confirmation is deliberate. A blanket `--update-snapshots` overwrites the
reference the suite compares against — it turns a caught regression into a
silently accepted one. Approve changed images individually.

## Why the matrix is platform-based, not viewport-based

`src/lib/utils/platform.ts` derives `platform` from `navigator.userAgent` at
module load, not from a CSS breakpoint. A 390px-wide Chromium with a desktop
UA still renders the **desktop** branch. Baselining the iOS and Android
branches therefore requires setting the User-Agent; the viewport only follows
along to match each platform's real form factor.

Consequence: the matrix is 3 platforms × routes, not 3 viewports × routes.

## Route list

`routes.ts` is the single source. `scripts/check-routes.ts` fails the run when
a `+page.svelte` exists without an entry, or an entry points at a page that no
longer exists — an uncovered route would otherwise look like a passing one.

Routes with `:id` resolve at runtime: `cmd/seed/main.go` assigns random UUIDs,
so there is nothing stable to hardcode. When the seed holds no matching item
the test **skips** with a reason naming the gap, rather than passing quietly.

## Determinism

`helpers.ts:stabilise()` kills animations and transitions, hides the caret,
waits for the network to settle and for webfonts to load. Volatile content
(timestamps, relative times, monospace ids) is masked, not removed, so layout
still shows. Without this the diffs are noise and prove nothing.
