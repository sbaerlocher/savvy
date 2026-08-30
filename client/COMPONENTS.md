# Layout components

How a page gets its frame, and where to put a new one.

## The layers, outside in

| Layer          | Lives in                                              | Owns                                                                  |
| -------------- | ----------------------------------------------------- | --------------------------------------------------------------------- |
| Root layout    | `src/routes/+layout.svelte`                           | HTML shell, global styles, providers, the `<main>` container           |
| App shell      | `src/lib/components/shell/`, `MobileNav`, `Toast`, …  | Navigation, footer, toasts, offline banner, the global "New" dialog    |
| **Page shell** | `src/lib/components/layout/PageShell.svelte`          | **Content width, horizontal padding, page title**                      |
| Sections       | `ResourceDetail`, `WalletView`, `SettingsTabs`, …     | A screen's inner content structure                                         |
| Atoms          | `src/lib/components/ui/`                              | `Button`, `EmptyState`, `Skeleton`, `Modal`, …                         |

## Writing a new page

```svelte
<script lang="ts">
	import PageShell from '$lib/components/layout/PageShell.svelte';
	import { t } from '$lib/stores/i18n';
</script>

<svelte:head>
	<title>{$t('thing.title')} - {$t('common.appName')}</title>
</svelte:head>

<PageShell title={$t('thing.title')}>
	<!-- content only -->
</PageShell>
```

That is the whole contract. Do **not** add `max-w-*`, `mx-auto`, `px-*` or
`pb-20 md:pb-4` — the root layout's `<main>` already sets
`max-w-7xl mx-auto pt-4 pb-6 px-4 sm:px-6 lg:px-8`, and `.main-with-mobile-nav`
in `app.css` handles the mobile nav clearance with `!important`. A page that
re-states any of those stacks a second container inside the first, and the
padding applies twice. That was the state this refactor removed: five competing
container dialects, iOS `/dashboard` rendering 311px of content where the shell
defines 343px.

The structure tests fail on exactly this, so a regression shows up as a failing
test rather than as a slightly-off screen.

The same goes for vertical rhythm under the header: the gap is `mb-8`, defined
once in `PageHeader`/`PageShell`. Form pages that need a back-to-overview line
pass it as the `back` snippet instead of wrapping it in their own `mb-6` div —
that mix of `mb-6` and `mb-8` was the last page-owned spacing left after the
container cleanup.

## PageShell props

| Prop            | Default     | Purpose                                                    |
| --------------- | ----------- | ---------------------------------------------------------- |
| `title`         | –           | The page's single `<h1>`, rendered through `PageHeader`     |
| `subtitle`      | –           | Supporting line under the title                            |
| `eyebrow`       | –           | Small line above the title (greeting, section kicker)      |
| `width`         | `'default'` | `default` · `narrow` · `full` · `bleed` — see below        |
| `header`        | `true`      | `false` when a section component renders the header itself |
| `mobileActions` | `true`      | Bell + New on the title row; `false` on sub-screens        |
| `actions`       | –           | Snippet: trailing controls on the title row                |
| `onBack`        | –           | Back chevron left of the title                             |
| `back`          | –           | Snippet: back-to-overview line under the title (form pages) |

### `width`

- `default` — the shared container. Adds nothing: `<main>` already is it.
- `narrow` — reading-width column for forms and settings.
- `full` — cancels the shell's horizontal padding, for list screens whose rows
  bleed to the edge (Android settings).
- `bleed` — centres content in the viewport with the auth cards' own `px-5`
  gutter. Only for the Android auth screens, which the root layout already
  renders without shell padding or footer.

### `header={false}`

Use it when the header is not the first thing in the flow, or when a section
component owns it:

- `dashboard` — a grid places the header in column 1 and the stat tiles in
  column 2 of the same row, so `PageHeader` must stay inside its grid cell.
- `wallet`, `merchants/[id]` — `WalletView` renders the header through its own
  snippets.
- the resource detail pages — `ResourceDetail` renders its own header per
  platform.
- `admin/users`, `admin/system-health` — four mutually exclusive platform
  branches, each with its header inside an elevated panel.

## StateScreen

`src/lib/components/layout/StateScreen.svelte` — full-screen state pages
(offline, 404, server error, back online). These render outside the app shell,
so they are not PageShell's job. Pass a `tone`, an `icon` path and the copy;
optional `children` sit between text and actions.

## Platform variants

`src/lib/utils/platform.ts` derives `platform` from the **User-Agent** at
module load, not from a CSS breakpoint. A narrow desktop window is still
`'other'`. Platform-specific chrome therefore lives in the components
(`PageHeader`, `WalletView`, `ResourceDetail`), and PageShell stays
platform-neutral: it contributes the container, the same on all three.

The iOS, Android and desktop mockups are separate designs — never derive one
platform's layout from another's.

## Verifying a change

```bash
dde project:structure:test
```

See `tests/structure/README.md` for what the suite asserts, how to record a
baseline, and why the matrix is platform-based rather than viewport-based.

## Known follow-ups

Found during the refactor, deliberately left alone:

- `admin/merchants/new` and `admin/merchants/[id]/edit` link back to
  `/merchants` (the public overview) instead of `/admin/merchants`. The same
  wrong target also appears in their cancel links and post-submit `goto()`
  calls, so a fix should sweep all of them.
- Both `admin/users` and `admin/system-health` carry an `{:else}` fallback
  branch that no value of `platform` can reach. It still uses the old
  `bg-white shadow rounded-lg` panel styling.
- `--spacing-screen` (17px) in `tokens.css` has no remaining users since
  `WalletView` moved to the shell's 16px padding.
