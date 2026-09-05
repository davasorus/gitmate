# gitmate — Roadmap to a Full Git Client (CLI + GUI)

> Living plan. gitmate is a Cobra CLI + Wails 3 (React/Tailwind/shadcn) desktop app
> over one shared Go engine. Goal: a real, full-featured git client usable from both
> the terminal and the UI — not a toy. This doc is the reference we build against.

---

## Architecture (fixed decisions)

- **Shared engine, two frontends.** All git/GitHub logic lives in `internal/`.
  The CLI (`cmd/gitmate`) and GUI (`gui/`) are thin layers that call the same engine.
  A bug fixed in the engine fixes both frontends at once. This is non-negotiable —
  never put git logic in a command handler or a React component.
- **`internal/gitops`** — local git, shells out to the real `git` binary via `os/exec`,
  parses machine-readable output (porcelain v2, null-delimited pretty-format,
  for-each-ref). Rationale: exact git compatibility + keeps the plumbing visible.
- **`internal/ghapi`** — GitHub, via `google/go-github` v66 + oauth2. Token resolved
  from the `GITHUB_TOKEN` Windows user env var (persistent).
- **Engine strategy:** stay on `os/exec` for correctness and full git compatibility.
  Only reach for `go-git` (or go hybrid) if a specific operation makes text-parsing
  genuinely unworkable. Revisit per-feature, not wholesale.
- **GUI:** the `GitService` wrapper exposes engine methods to the frontend; run
  `wails3 generate bindings` after adding/changing service methods.

### Cross-cutting concerns every feature must respect
- **Errors surface, never crash.** Every GUI action goes through the `run()` wrapper
  (busy state + toast + reload). A thrown error must show a red toast, not blank the app.
- **Nil-safe rendering.** Go nil slices serialize to JSON `null`; guard every
  `.length`/`.map` in React with `?? []`.
- **Working-dir discipline.** CLI runs from module root; `wails3` runs from `gui/`.
- **Destructive ops need guardrails** (confirmation; ideally reflog-backed undo).
- **Conflicts are a first-class state** — merge/rebase/pull can all conflict; the
  client must detect, display, and help resolve, not just error out.

---

## Status legend
- [x] done
- [~] in progress
- [ ] not started

---

## DONE (baseline as of Sep 2026)

### Engine — gitops (local)
- [x] status (porcelain v2, staged/unstaged/untracked, ahead/behind)
- [x] log (null-delimited pretty-format)
- [x] branches (for-each-ref, sorted, upstream tracking)
- [x] stage (all)
- [x] commit
- [x] push (with -u / set-upstream)
- [x] current branch
- [x] remote get-url / add

### Engine — ghapi (GitHub)
- [x] auth (PAT via GITHUB_TOKEN), whoami, rate limit
- [x] list PRs (IsPullRequest filter), list issues
- [x] repo create + parse owner/repo from remote URL
- [x] PR create, merge, comment
- [x] PR checks (check runs on head SHA)
- [x] issue create

### CLI
- [x] status, log, branches, whoami, prs, issues
- [x] init-remote, stage, commit, push
- [x] pr create / merge / comment / checks, issue create

### GUI
- [x] React + Tailwind + shadcn dashboard, system light/dark
- [x] read panels: status, branches, recent commits, open PRs
- [x] action forms: commit+push, new PR, new issue
- [x] per-PR merge button + on-demand checks (color-coded)
- [x] toasts + loading states + reload-on-success

### Infra
- [x] table tests for gitops (temp-repo integration)
- [x] CI: vet + test + golangci-lint (v2.13.2, Go 1.25, scoped to cmd/ + internal/)
- [x] GoReleaser + release-on-tag workflow
- [x] Dependabot (gomod, npm, github-actions)
- [x] repo hygiene: node_modules/bindings/build output untracked + gitignored
- [x] token resolution via persistent Windows user env var

---

## TIER 1 — Core daily git (a client isn't a client without these)

**Build order within tier: diff FIRST** (everything downstream renders diffs).

### 1.1 Diff  [x]
- [x] engine: working-tree diff (unstaged)
- [x] engine: staged diff (`--cached`)
- [x] engine: commit-to-commit diff
- [x] engine: single-file diff
- [x] parse unified-diff / patch into structured hunks (file, +/- lines, headers)
- [x] CLI: `gitmate diff [--staged] [<path>] [--rev <rev>]`
- [x] GUI: DiffView component (+/- line coloring, line-number gutters, hunk headers)
- [x] GUI: click a file in Changes → show its diff (inline)
- [x] GUI: click a commit in History → show its diff (via 1.8 show)

### 1.2 Checkout / switch  [x]
- [x] engine: switch branch
- [x] engine: restore file(s) from index/HEAD (see 1.5 discard)
- [x] handle dirty-tree refusal (surfaced as error/toast)
- [x] CLI: `gitmate switch <branch>` (-c to create)
- [x] GUI: branch switcher (Branches view, click to switch)

### 1.3 Branch management  [x]
- [x] engine: create / delete / rename branch
- [x] CLI: `gitmate switch -c` (create), `gitmate branch delete|rename`, `gitmate branch list`
- [x] GUI: new-branch input + rename + delete (confirm dialog) on branch list

### 1.4 Granular staging  [x]
- [x] engine: stage / unstage a specific file
- [ ] engine: stage a hunk (advanced — patch application) — deferred
- [x] CLI: `gitmate stage <path>`, `gitmate unstage <path>`
- [x] GUI: per-file stage/unstage in Changes (Staged/Unstaged split); per-hunk deferred

### 1.5 Discard / restore changes  [x]
- [x] engine: discard working-tree changes to a file (guarded — destructive)
- [x] CLI: `gitmate discard <path> --force`
- [x] GUI: discard button per file (confirm dialog)

### 1.6 Fetch / pull  [ ]
- [ ] engine: fetch
- [ ] engine: pull (--rebase and merge modes)
- [ ] surface conflicts as a distinct state (feeds Tier 2 conflict UI)
- [ ] CLI: `gitmate fetch`, `gitmate pull [--rebase]`
- [ ] GUI: fetch/pull buttons; show ahead/behind refresh

### 1.7 Stash  [x]
- [x] engine: stash save / list / pop / drop (apply deferred)
- [x] CLI: `gitmate stash [save|list|pop|drop]`
- [x] GUI: Stashes sidebar section (save, list, pop, drop w/ confirm)

### 1.8 Show (inspect one commit)  [x]
- [x] engine: commit metadata + diff for a rev (root-commit safe)
- [x] CLI: `gitmate show <rev>`
- [x] GUI: commit detail in History (click commit → metadata + DiffView)

---

## TIER 2 — History & repair (git gets sharp; guardrails matter)

### 2.1 Merge  [ ]
- [ ] engine: merge a branch; detect conflicts
- [ ] CLI: `gitmate merge <branch>`
- [ ] GUI: merge action + conflict surfacing

### 2.2 Conflict resolution  [ ]  (shared by merge/rebase/pull)
- [ ] engine: list conflicted files; read conflict markers; mark resolved (add)
- [ ] CLI: `gitmate conflicts`, resolve flow
- [ ] GUI: conflict view — per-file, choose ours/theirs/edited, mark resolved

### 2.3 Rebase  [ ]
- [ ] engine: rebase onto; continue/abort
- [ ] engine: interactive rebase (reorder/squash/edit) — hard to model
- [ ] CLI: `gitmate rebase <base> [--continue|--abort]`
- [ ] GUI: (stretch) interactive rebase editor

### 2.4 Reset  [ ]  (dangerous — guardrails required)
- [ ] engine: reset soft / mixed / hard
- [ ] CLI: `gitmate reset [--soft|--mixed|--hard] <rev>` (confirm on hard)
- [ ] GUI: reset with explicit mode choice + confirm

### 2.5 Cherry-pick / revert  [ ]
- [ ] engine: cherry-pick <rev>, revert <rev>
- [ ] CLI + GUI actions

### 2.6 Tags  [ ]
- [ ] engine: create / list / delete tags (lightweight + annotated)
- [ ] CLI: `gitmate tag [create|list|delete]`
- [ ] GUI: tag list + create (ties to GoReleaser release flow)

### 2.7 Reflog  [ ]  (the safety net / undo backbone)
- [ ] engine: read reflog
- [ ] CLI: `gitmate reflog`
- [ ] GUI: reflog view; (stretch) one-click restore to a reflog entry

### 2.8 Blame  [ ]
- [ ] engine: line-by-line authorship for a file
- [ ] CLI: `gitmate blame <path>`
- [ ] GUI: blame gutter in file view

---

## TIER 3 — Remotes & GitHub depth

### 3.1 Remotes  [ ]
- [ ] engine: remote remove / rename / list (add exists)
- [ ] CLI + GUI remote management

### 3.2 Clone  [ ]
- [ ] engine: clone an existing repo
- [ ] CLI: `gitmate clone <url>`
- [ ] GUI: clone dialog

### 3.3 Richer GitHub (REST)  [ ]
- [ ] PR reviews + review comments
- [ ] labels, assignees, milestones
- [ ] releases list / create (ties to tags 2.6)
- [ ] close/reopen PRs and issues

### 3.4 GraphQL path  [ ]
- [ ] fetch a PR's full context (reviews + checks + comments) in one query
- [ ] use where REST would need many round-trips (PR detail view)

### 3.5 Webhooks (live updates)  [ ]  (only piece needing a server)
- [ ] small HTTP server to receive events (HMAC-SHA256 verify, constant-time compare)
- [ ] push events → GUI updates live instead of polling
- [ ] biggest new concept; largest infra commitment

---

## CROSS-CUTTING BUILDS (not single commands — whole subsystems)

### C.1 Diff renderer (GUI)  [ ]
- The backbone view. +/- line coloring, hunk headers, file headers.
- Reused by: file diff, commit diff, merge/rebase conflict preview, PR review.
- Do this properly as part of 1.1; everything leans on it.

### C.2 Navigation model (GUI)  [ ]
- Move from flat dashboard → app shell: branch list → commit list → diff.
- Panes/routing; this is where a component library fully earns its place.
- Rethink when Tier 1 diff + log-detail land.

### C.3 Safety & undo  [ ]
- Confirmation dialogs for destructive ops (reset --hard, force push, branch delete).
- Reflog-backed "undo" where possible (leans on 2.7).

### C.4 Interactive input  [ ]
- Hunk-level staging (1.4), interactive rebase (2.3), conflict resolution (2.2)
  all need richer UI than buttons — shared interaction patterns.

---

## Sequencing (recommended)

1. **Tier 1, diff first** (1.1 + C.1) — the backbone.
2. Rest of Tier 1: 1.2 → 1.3 → 1.4 → 1.5 → 1.6 → 1.7 → 1.8.
   Somewhere in here, tackle C.2 (navigation) once diff + log-detail exist.
3. **Tier 2** — merge (2.1) + conflict resolution (2.2) together, since merge
   without conflict handling is half a feature. Then reset/rebase/etc. with C.3 safety.
4. **Tier 3** — remotes/clone (cheap), then GitHub depth, then GraphQL, then webhooks last.

## Working agreement
- Build per feature: engine method(s) → CLI command → GUI wiring → verify, in one batch.
- Keep CI green; run `golangci-lint run ./cmd/... ./internal/...` before pushing.
- Add a gitops table test for each new engine operation.
- Update this doc's checkboxes as things land.