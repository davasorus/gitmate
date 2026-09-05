# gitmate — Roadmap to a Full Git Client (CLI + GUI)

> Living plan. gitmate is a Cobra CLI + Wails 3 (React/Tailwind/shadcn) desktop app
> over one shared Go engine. Goal: a real, full-featured git client usable from both
> the terminal and the UI — not a toy. This doc is the reference we build against.

---

## Architecture (fixed decisions)

> SCOPE BOUNDARY (decided at end of Tier 2): gitmate is a git-OPERATIONS client, not a code editor. Anything that only makes sense inside an editor — blame annotations, inline file editing, a source file tree/viewer, syntax highlighting — is OUT OF SCOPE. Those live in the user's actual editor (VS Code + GitLens). This prevents scope creep toward 'rebuild the IDE'.

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

> STATUS: all core items shipped (1.1 diff, 1.2 switch, 1.3 branch mgmt, 1.4 staging, 1.5 discard, 1.6 fetch/pull, 1.7 stash, 1.8 show). Deferred within-tier: hunk-level staging (1.4), stash apply (1.7). Conflict *surfacing* as structured state moves to Tier 2.2.

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

### 1.6 Fetch / pull  [x]
- [x] engine: fetch
- [x] engine: pull (--rebase and merge modes)
- [ ] surface conflicts as a distinct state (feeds Tier 2 conflict UI) — errors surface as toast; structured conflict state is Tier 2.2
- [x] CLI: `gitmate fetch`, `gitmate pull [--rebase]`
- [x] GUI: fetch/pull buttons in status bar; ahead/behind refreshes on reload

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

> STATUS: TIER 2 COMPLETE — merge, conflicts, rebase, reset, cherry-pick/revert, tags, reflog, blame all shipped (CLI + GUI). Deferred within-tier: interactive rebase, per-region conflict resolution. Next: polish stage (row-action cleanup, History branch browsing, etc.) or Tier 3.

### 2.1 Merge  [x]
- [x] engine: merge a branch; detect conflicts (Merge/ConflictedFiles/MergeInProgress/MergeAbort)
- [x] CLI: `gitmate merge <branch>` + `merge-abort`
- [x] GUI: Merge button in Branches + merge-in-progress banner

### 2.2 Conflict resolution  [x] (whole-file ours/theirs + hand-edit; per-region deferred)
- [x] engine: ReadConflict (parse regions), ResolveOurs/Theirs, MarkResolved
- [x] CLI: `gitmate conflicts`, `gitmate resolve <path> [--side ours|theirs]`
- [x] GUI: Conflicts view — per-file ours/theirs preview + Take ours/theirs/Mark resolved (per-region UI deferred to polish)

### 2.3 Rebase  [x] (non-interactive; interactive rebase deferred)
- [x] engine: Rebase/RebaseContinue/RebaseAbort/RebaseInProgress
- [ ] engine: interactive rebase (reorder/squash/edit) — deferred
- [x] CLI: `gitmate rebase <base> [--continue|--abort]`
- [x] GUI: Rebase button in Branches + rebase-in-progress banner (Continue/Abort), reuses Conflicts view; interactive editor deferred

### 2.4 Reset  [x]  (dangerous — guardrails required)
- [x] engine: Reset(rev, soft|mixed|hard)
- [x] CLI: `gitmate reset <rev> [--soft|--hard]` (--hard needs --force)
- [x] GUI: Reset to here (Soft/Mixed/Hard) on each Reflog entry; Hard needs confirm dialog

### 2.5 Cherry-pick / revert  [x]
- [x] engine: CherryPick/Revert (+continue/abort, SequencerInProgress) — reuses conflict flow
- [x] CLI: `gitmate cherry-pick <rev>` / `revert <rev>` (--continue/--abort); GUI: Cherry-pick + Revert buttons on History commits, in-progress banners route to Conflicts view

### 2.6 Tags  [x]
- [x] engine: ListTags/CreateTag(lightweight+annotated)/DeleteTag/PushTag
- [x] CLI: `gitmate tag [list|create|delete|push]`
- [x] GUI: Tags sidebar view — list, create (annotated opt), delete (local or local+origin), push (triggers release), Sync tags (fetch --prune); tags now refresh on global Reload; location shown as Monitor(local)/Cloud(remote) icons; Tags promoted out of the SOON group into active nav

### 2.7 Reflog  [x]  (the safety net / undo backbone)
- [x] engine: Reflog(dir, limit) — parses selector/action/message
- [x] CLI: `gitmate reflog [-n N]`
- [x] GUI: Reflog view color-coded; per-entry Reset to here (soft/mixed/hard) wired via 2.4

### 2.8 Blame  [x]
- [x] engine: Blame(path) via git blame --porcelain
- [x] CLI: `gitmate blame <path>`
- [x] GUI: DECLINED — blame is only meaningful while reading code (chasing a bug), which is a code-editor activity. gitmate is a git-operations client, not an editor; blame has no natural context here. Kept as CLI (`gitmate blame <path>`).

---

## TIER 3 — Remotes & GitHub depth

### 3.1 Remotes  [ ]
- [ ] engine: remote remove / rename / list (add exists)
- [ ] CLI + GUI remote management

### 3.2 Clone  [ ]
- [ ] engine: clone an existing repo
- [ ] CLI: `gitmate clone <url>`
- [ ] GUI: clone dialog

### 3.3 Richer GitHub (REST)  [~]  (partial: issues list + improved PR create shipped)
- [ ] Release ships only the CLI (`gitmate.exe`), not the Wails GUI. GoReleaser builds
  `cmd/gitmate` (CLI); the desktop GUI needs `wails3 build` per-platform (embeds frontend,
  icons, manifests), which GoReleaser doesn't do natively. Options: GoReleaser custom build
  invoking wails3, or a separate release job running `wails3 build` and attaching artifacts.
  Pairs with the workflow_dispatch release item.
- [ ] Binaries are unsigned — Windows SmartScreen warns on download/run. Signing needs an
  Authenticode cert (Windows) / Apple Developer cert + notarization (macOS) stored as CI
  secrets; has real cost + setup. Defer until distributing to actual users.
- [ ] PR reviews + review comments
- [ ] labels, assignees, milestones
- [ ] releases list / create (ties to tags 2.6)
- [ ] close/reopen PRs and issues

### 3.4 GraphQL path  [ ]
- [ ] fetch a PR's full context (reviews + checks + comments) in one query
- [ ] use where REST would need many round-trips (PR detail view)

### 3.5 Webhooks (live updates)  [ ]  (only piece needing a server)
- [ ] Release ships only the CLI (`gitmate.exe`), not the Wails GUI. GoReleaser builds
  `cmd/gitmate` (CLI); the GUI needs `wails3 build` per-platform, which GoReleaser doesn't
  do natively. Options: GoReleaser custom build invoking wails3, or a separate workflow job
  running `wails3 build` and attaching the artifacts. Pairs with the workflow_dispatch
  release item. (Raised during Tier 2; deferred to Tier 3 per roadmap discipline.)
- [ ] Binaries are unsigned (Windows SmartScreen warning). Needs Authenticode cert (Windows)
  / Apple Developer cert + notarization (macOS) stored as CI secrets — cost + setup. Defer
  until distributing to real users. (Tier 3 at earliest.)
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

### UX feedback (added post-Tier-2.2)
- [ ] Branch-row actions: five equal-weight buttons (Switch/Merge/Rebase/Rename/Delete)
  feel cluttered, BUT the Integrate-menu redesign attempted mid-Tier-2 was worse
  (hidden click-to-switch, over-engineered dropdowns) and was reverted. Revisit in the
  dedicated polish stage with a light touch: visual hierarchy (primary vs secondary
  weight) only — do NOT hide state-changing actions behind clicks/menus.
- [ ] **GUI reload on window focus** — the GUI polls, it doesn't watch. Actions taken
  in the terminal (or another git tool) while the app is open aren't reflected until a
  manual Reload or a GUI action forces reload(). Reload on window-focus covers the
  common terminal↔GUI bounce cheaply (a filesystem watcher on .git is the heavier,
  fuller fix). Surfaced when a CLI-concluded merge left the banner stale.
- [ ] **GUI should set repoDir to the real repo root.** repoDir defaults to "." and the
  Wails process runs from gui/, so git commands work only because git walks *up* to
  find the repo. Path-relative filesystem commands (e.g. conflict resolve's
  `checkout --ours -- <path>`) broke because the pathspec resolved against gui/, not
  the root. ResolveOurs/Theirs/MarkResolved self-correct via `rev-parse --show-toplevel`,
  but the real fix is aligning repoDir/cwd once for all commands.
- [ ] **Merge-completion UX** — banner should auto-clear when a merge concludes, and the
  Conflicts view should offer a "Commit merge" button once conflicts hit zero (right now
  it says "commit in Changes to finish" but nothing pulls you there).
- [ ] Windows path casing (README.MD vs README.md) is a recurring gotcha — git pathspecs
  are case-sensitive even on case-insensitive filesystems. Not code-fixable; note for docs.