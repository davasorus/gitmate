# gitmate frontend

The desktop GUI's React frontend. It renders engine data and calls Go service methods — it holds
**no git logic** of its own. All operations go through generated bindings to `gui/gitservice.go`.

## Architecture

```Plain Text
Go engine (internal/gitops, internal/ghapi)
        ▲
gui/gitservice.go            thin Go service — one method per operation
        ▲  (Wails generates TypeScript bindings from this)
src/context.tsx              GitmateProvider: shared state + reload(), the single data hub
        ▲
src/views/*.tsx              one view per feature tab (Changes, Branches, PullRequests, …)
src/components/*.tsx         reusable pieces (DiffView, RunFlow, CheckBadge, …)
```

Data flows one way: a view calls `service.X()` (a binding), gets engine types back, and either
renders them or calls `reload()` to refresh shared state. Views never talk to git directly.

## Key pieces

### `context.tsx` — the state hub

`GitmateProvider` owns the shared state (status, branches, commits, PRs, issues, stashes, tags,
conflict/merge/rebase flags, the current view, busy/flash UI state) and the `reload()` function
that repopulates it from the engine. Access it via the `useGit()` hook.

- `reload()` — repopulates local state (cheap, no network). Called after any mutating action.
- Auto-refresh: `reload()` runs on window focus/visibility; a periodic background `Fetch()`
  (every 5 min, with prune) keeps remote state fresh.
- `run(name, fn, okMsg)` — the standard action wrapper: sets busy, runs `fn`, flashes ok/err,
  then `reload()`s.
- `cls` — shared Tailwind class strings (buttons, inputs) for consistent styling.

### `views/*.tsx` — feature tabs

One view per operation area. Each uses `useGit()` for shared state + services, owns any
view-local state (filters, form inputs, expanded rows), and calls services through `run()` for
mutations. Examples: `Changes` (stage/commit/diff), `Branches` (local+remote, switch/merge/…),
`PullRequests` (list with CI/review rollup, review flow, threads), `Actions` (runs, logs, flowchart).

### `components/*.tsx` — reusable UI

- `DiffView` — the unified-diff renderer (file/hunk headers, +/- coloring, line numbers). The backbone
  used by Changes, History, and PR review.
- `ReviewDiff` — diff with inline PR-review comment affordances.
- `RunFlow` — SVG job-dependency flowchart for an Actions run (matrix legs expanded).
- `CheckBadge`, `StatusBadge` — status pills. `ConfirmDialog` — destructive-action confirmation.
- `LogView` — colorized Actions job-log viewer.
- `components/ui/*` — shadcn-style primitives.

## Bindings

`gui/frontend/bindings/` is **generated** by `wails3 generate bindings` from `gui/gitservice.go` —
do not edit by hand. Regenerate after changing the service's method signatures or types. The
generated TypeScript gives each Go service method a typed async function and mirrors the engine's
Go structs as TS interfaces.

## Conventions

- Views call services via `run()` for anything that mutates, so busy/flash/reload are consistent.
- Prefer reading aggregated data from a single service call (e.g. `PRsRich`, `PRDetail`) over
  multiple round-trips — the GraphQL-backed methods exist for this.
- Styling uses Tailwind + the shared `cls` helpers; keep new UI consistent with existing views.
- Format + lint before committing: `npm run format` (Prettier), `npm run lint` (oxlint), `npx tsc --noEmit`.
