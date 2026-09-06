# gitmate — Roadmap to a Full Git Client (CLI + GUI)

---

## ✅ TIER 1/2 FUNCTIONAL GAPS — CLOSED (both re-implemented after the reset)

Two features were built then lost in a `git reset --hard`. Both must be
re-implemented to make Tier 1/2 functionally complete. Doing them one at a time.

- [x] 1.7 stash apply — DONE — have save/list/pop/drop; missing `apply` (apply and keep).
- [x] 2.5 GUI cherry-pick reachability — DONE (History branch selector + target labeling) — History shows only the current branch, so
  GUI cherry-pick can only pick commits already on the branch (a no-op). Needs a
  History branch selector (log any ref) + label the cherry-pick/revert target.
  CLI cherry-pick already works.

Correctly NOT gaps (already Tier 3): remote management (3.1), clone (3.2).

---

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
- [x] GUI: Stashes sidebar (save, list, pop, apply, drop w/ confirm)

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

Plan of record (agreed): build **everything except a GitHub App** (no multi-user
auth model — no use for it). Sequenced by dependency + difficulty:
**A (REST features) → B (GraphQL) → C (Actions/workflow-dispatch) → D (Webhooks)**.
One feature at a time. Hygiene/release-engineering items live in their own section
below, not interleaved with features.

### 3.1 Remotes  [x]
- [x] engine: ListRemotes / RemoveRemote / RenameRemote (AddRemote existed)
- [x] CLI: `gitmate remote list|add|remove|rename`; GUI: Remotes sidebar view

### 3.2 Clone  [x]
- [x] engine: Clone(url, dest); CLI `gitmate clone <url> [dir]`; GUI Clone card (points app at clone)

### 3.3 / Phase A — Richer GitHub (REST)  [x]   ← DONE, merged to live
Scope decision: this is the MAIN git/GitHub tool, so features are built COMPLETE, not
minimal — half-features just send you back to the web UI. Full scope below. All via
go-github; none require GraphQL. Order: close/reopen → labels → releases → PR reviews.
Both CLI + GUI for everything.

- [x] **close / reopen** PRs and issues (state update) — CLI + GUI; + open/closed/all filter in both views
- [x] **labels — full CRUD, two levels:** DONE
      - apply: add / remove labels on a PR or issue (chips on rows: click to remove, +label to add)
      - manage definitions: Labels view — create/edit/delete repo label types (name/color/desc);
        native color picker + randomizer (no hex memorization)
      - CLI: `gitmate label list|create|edit|delete|add|remove`
- [x] **releases — full CRUD:** DONE (verified end-to-end on immutable AND non-immutable repos)
      - list / create / edit / delete + generate-notes; own Releases sidebar view
        (draft/prerelease flags, View/Edit/Delete, gen-notes button)
      - assets full CRUD — list / upload / download / delete via base64 over the JS boundary
        (no native dialog: <input type=file> upload, browser Blob download). Upload uses
        go-github official UploadReleaseAsset via temp file (hand-rolled HTTP caused HTTP/2
        PROTOCOL_ERROR).
      - IMMUTABLE releases handled first-rate: detect via raw-JSON "immutable" field; badge;
        upload control replaced with a locked note; asset-Delete + release-Delete disabled;
        handler-level guards flash a clear error if any path fires; all upload failures surface
        as toasts (no silent no-op).
      - CLI: `gitmate release list|create|delete|notes` + `release assets list|upload|download|delete`
- [x] **PR reviews — full:** DONE (merged to live)
      - whole-PR review: approve / request-changes / comment; list reviews; request/remove reviewers
      - PR diff shown in the review panel (reuses DiffView via PRDiff → ParseUnifiedDiff)
      - line-level review comments: click a diff line → pending comment; batched into one
        CreateReview submit with the verdict (ReviewDiff component; DiffView left read-only)
      - existing conversation: review-comment threads shown inline at their lines + reply;
        general PR comment stream + add-comment
      - CLI: `gitmate pr review|reviews|reviewers`
      - DEFERRED to Phase B: resolve/unresolve threads (GraphQL-only) — now built in B
- [ ] (same API family, NOT yet built — future) assignees, milestones, edit/delete comments, lock conversations

### 3.4 / Phase B — GraphQL (full client layer)  [x]   ← DONE
Scope: build a REAL GraphQL client as a peer to the REST client (not a one-off query),
because it unlocks capability, not just efficiency:
  - GraphQL-only features exist (e.g. Projects V2 has NO REST equivalent)
  - 1 nested query replaces ~11 REST calls for aggregated reads
  - separate rate-limit budget from REST (more total headroom)
- [x] add GraphQL client (shurcooL/githubv4) alongside go-github in internal/ghapi — done: gql field on Client, same oauth2 token
- [x] engine+service: PRDetailGraphQL — ONE query returns reviews + review threads (with IDs +
      resolve state) + checks + labels + assignees; ResolveThread/UnresolveThread mutations
      (GraphQL-ONLY — no REST equivalent). Service: PRDetail / ResolveThread / UnresolveThread.
- [x] GUI wiring: DONE — review panel loads reviews/threads/labels/assignees via the single
      PRDetail query (replaced 3 REST read calls: ListReviews/ListRequestedReviewers/ListReviewComments);
      diff stays REST (PRDiff); writes stay REST (3.7 migration deferred); Review-threads section
      with Resolve/Unresolve buttons (GraphQL-only mutation) + resolve state badges.
- [x] depends on Phase A: REST write-actions exist (Phase A merged) ✓
- [ ] leaves the door open for GraphQL-only features later (Projects V2, etc.)

### 3.7 / Later — REST→GraphQL migration pass (Phase 4 or late Phase 3)  [ ]
Deliberate per-endpoint pass AFTER Phase B exists (don't migrate against a moving target,
don't rip-and-replace). Decision rule:
  - USE GRAPHQL for: nested/aggregated reads (PR detail, dashboards), GraphQL-only features
  - KEEP REST for: simple single-resource reads, and most writes/mutations (clearer, already built)
- [ ] audit each ghapi call against the rule; migrate the ones where GraphQL genuinely wins
- [ ] mixed REST+GraphQL is expected and fine (GitHub explicitly supports it; node IDs bridge them)

### 3.5 / Phase C — GitHub Actions (first-class)  [x]   ← DONE (C-1 data+polling, C-1.5 grouping, C-2 controls, C-3 logs+color, C-4 flowchart)
Actions as a first-class citizen: watch, control, view. Live feel via SMART POLLING
(only the active run, only while viewing, back off when idle/done) — the standard
approach for desktop git tools (VS Code/GitHub Desktop poll too). Both CLI + GUI.
Build order: (1) list+run-tree+status+polling → (2) controls → (3) logs → (4) flowchart.

- [x] data: ListWorkflows/ListRuns/GetRun/ListRunJobs (engine); ListRuns/RunJobs/GetRun (service); CLI `actions runs|view`; GUI Actions view — runs list + jobs/steps tree, status color-coded
- [x] smart polling: fixed 5s timer, ONLY the open+active run, stops on completion/unmount, timer-driven not render-driven (applied the Phase A infinite-fetch lesson)
      ⚠️ LESSON (Phase A close/reopen): a useEffect→useCallback loop with unstable deps
      caused an infinite fetch that tripped GitHub's SECONDARY RATE LIMIT in <1s. Any
      auto-fetch MUST use a fixed controlled cadence (timer), never tied to render cycles.
      Effect deps = only the thing that should trigger a refetch.
- [x] controls: cancel run; re-run (all/failed-only); trigger workflow_dispatch with DYNAMIC inputs (parse workflow YAML → render a field per declared input) — engine/service/CLI/GUI
- [x] logs: download + display per-step logs AFTER completion — JobLogs (per-job, on-demand); best-effort per-step split via ##[group] markers, whole-job Raw fallback; CLI `actions logs <jobID>`; GUI Logs toggle per completed job; colored log output (LogView): GitHub ##[error]/##[warning]/##[notice] markers + general failure-word heuristic, universal (any language/action), markers+timestamps stripped
      (public API has no live per-step log streaming — accepted limit)
- [x] view 1: run-tree (indented jobs/steps list) — DONE
- [x] **C-1.5 organize the run list** DONE (was under-scoped in C-1; a flat 30-run list is
  noise for a multi-workflow repo). Scope: group runs BY WORKFLOW; filter by STATUS
  (all/success/failure/in-progress); show richer per-run detail (trigger/event, branch,
  duration, run #). Engine adds WorkflowName + timing to the run; frontend groups + filters.
- [x] view 2: flowchart DONE — RunJobGraph parses jobs+needs from workflow YAML; RunFlow renders depth-column DAG with SVG dependency arrows, boxes colored by live job status; Tree/Flowchart toggle in the run panel
- [~] token needs `workflow` scope for dispatch/cancel — cancel/rerun/dispatch will 403 without it; regenerate token with workflow scope

### 3.6 — Webhooks  [DECLINED]
Considered for live updates; declined. Webhooks require GitHub to POST to a public URL,
i.e. a hosted server + app↔server channel — wrong fit for a personal desktop tool, and
comparable tools (VS Code, GitHub Desktop) don't use them; they poll. Smart polling (in
Phase C) covers the live-update need with zero infrastructure. Revisit only if gitmate
ever gains a server component for another reason.

---

## TIER 3 / Phase D — hygiene & release engineering  [x]   ← DONE

1. [x] **GUI in releases** — DONE: release.yml now has a `gui` matrix job (windows/macos/ubuntu)
   that installs Wails, builds the desktop app (`wails3 build`), and attaches binaries to the
   release via softprops/action-gh-release, alongside the GoReleaser CLI job. (Linux installs GTK/WebKit deps.)
2. [x] **Signing** — DECIDED: do NOT sign (free cert routes all need token/pipeline friction).
   Ship SHA256 checksums (GoReleaser already generates them) + document "unsigned, Run anyway,
   verify checksums, or build from source." (Doc part lands in Phase F.) Optional later:
   SignPath Foundation free OSS signing if the SmartScreen warning ever matters.
3. [x] **Bindings → generate-in-CI** — DONE: /gui/frontend/bindings/ re-ignored; ci.yml frontend
   job installs Wails + runs `wails3 generate bindings` before npm ci/tsc/vite. No more committed
   generated code or drift. (Must `git rm -r --cached gui/frontend/bindings` once to untrack.)
4. [x] **.gitattributes** — DONE — normalize line endings to kill the CRLF churn (core.autocrlf
   caused whole-tree phantom diffs). Commit a .gitattributes with text=auto + eol rules.
5. [x] **Bump go-github** — DONE (v66→v88; WithAuthToken option API; typed GetImmutable) — v66 predates the typed `Immutable` release field (added Sept
   2025); we currently read it from raw JSON. Bump for the native field + a year of fixes.

---

## TIER 3 / Phase E — CI/CD efficacy & GitHub-flow optimization  [ ]
Big phase — its own mini-roadmap, built in ordered sub-steps. Organizing principle:
a TWO-TRUNK model (`live` = production, `dev` = development) that everything else
configures around. Covers CI quality, CD/release maturity, and process/governance.

### E-0 — Two-trunk foundation (do FIRST)
- [ ] create `dev` branch; make it the default PR target
- [ ] flow: feat/* → PR → dev → (accumulate) → PR dev → live (release). Emergencies:
      hotfix → live, then back-merge live → dev to prevent drift.
- [ ] branch protection on BOTH trunks (require green CI + PR review); no direct pushes
- [ ] enforce: production commits only via live; development only via dev
- [ ] Dependabot retargeted to `dev` (never straight to production trunk)

### E-1 — CI quality (after E-0)  [x] DONE
- [x] triggers: CI on all pushes + PRs; gate skips push run when branch has open PR; dorny/paths-filter skips go/frontend by change area; concurrency dedupes
- [x] caching: Go module+build cache, npm cache, wails3 binary cache
- [x] composite action .github/actions/setup-go (checkout+setup-go+cache); used by go+frontend jobs. NOTE: release.yml not yet migrated to it — finish in end-of-E cleanup
- [x] matrix: go job runs ubuntu/windows/macos (fail-fast:false); lint/gofmt/coverage Linux-only
- [x] lint+format: OXLINT (not ESLint — typescript-eslint hard-locks TS<6.1, unusable on our TS7; oxlint is tsgo/TS7-native) + Prettier; both enforced in CI
- [x] gofmt gate: fails build if cmd/internal not gofmt-clean (Linux-only)
- [x] coverage: -coverprofile, prints total, uploads artifact; THRESHOLD via workflow env COVERAGE_MIN=85 (PR-gated to change), fails under. NOTE: don't make it a required check until actually at 85%

### E-2 — CD / release pipeline (the capstone)
- [ ] GUI in releases via a MATRIX `wails3 build` job (reuses E-1 matrix); attach desktop artifacts alongside CLI
- [ ] releases cut from `live` ONLY (enforced)
- [ ] TWO release paths:
      (a) manual tag — restricted to authorized users via a TAG PROTECTION RULESET on `v*`
          (note: CODEOWNERS does NOT gate tags; tag protection rulesets do)
      (b) automated: semantic-release on live → auto-version + release notes GROUPED BY
          conventional-commit type (feat/fix/breaking → Features/Fixes/…)
- [ ] gated approval before publish: GitHub Actions ENVIRONMENT with required reviewer
- [ ] conventional commits: SOFT adoption — semantic-release parses what it can; NOT hard-
      enforced by commit-lint (not all commits will follow the format, and that's accepted)

### E-3 — process/governance (light)
- [ ] required status checks tied to branch protection (green CI is a GATE, not a suggestion)
- [ ] Dependabot auto-merge on green minor/patch (optional)
- [ ] PR templates / CODEOWNERS (borders on Phase F docs)

## TIER 3 / Phase F — documentation sweep  [ ]  (LAST — after A–E settle)
Document the FINISHED system, not the moving one (that is why F is last). Written as if a
reviewer will read it (public repo), even though the real audience is future-me. Living, not
one-time.

WRITING STANDARD: **ASD-STE100 Simplified Technical English** for ALL documentation — every
doc AND every code comment. Rules we follow: approved/controlled vocabulary (one word = one
meaning, no synonyms), short sentences (~20 words procedures / ~25 descriptions), active
voice, present tense, one instruction per sentence, no ambiguity/jargon. STE is a WRITING
DISCIPLINE we apply by hand — we do NOT buy an STE checker tool (paying to heuristically lint
prose is not worth it).

ENFORCEMENT (machine): only that documentation EXISTS. CI fails if exported code lacks
comments. STE *style* is human-followed, not machine-validated.

### Front 1 — GitHub / repo docs
- [ ] README (what it is, features, screenshots, install/build, usage) — the front door
- [ ] CONTRIBUTING (build, two-trunk flow, commit conventions, local CI)
- [ ] install/release docs incl. the Phase-D unsigned-binary note (Run anyway / verify checksums / build from source)
- [ ] LICENSE (confirm one exists — public repo)
- [ ] issue/PR templates, CODEOWNERS (overlaps E-3 governance)

### Front 2 — Code docs (everything, enforced-to-exist)
- [ ] godoc comment on EVERY exported Go symbol (internal/gitops, internal/ghapi, gui service)
- [ ] frontend documented (components/props, context, shared-engine→service→views architecture)
- [ ] ARCHITECTURE.md (extract the fixed decisions: shared engine, os/exec strategy, two frontends)
- [ ] CI ENFORCES existence: linter rule requiring comments on exported symbols (e.g. revive /
      golangci-lint exported-comment rule) — undocumented exports FAIL CI

### Front 3 — Process docs
- [ ] build/dev setup (wails3 dev, GITHUB_TOKEN, generate-bindings, gotchas: Windows path casing, working-dir)
- [ ] release process (E-2 two-path model: manual tag vs semantic-release; how to cut a release)
- [ ] CI/CD explainer (what the workflows do, the two-trunk model)

### Living
- [ ] docs stay current: CI existence-check + PR expectation that docs update with code

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
- [ ] **Branch listing shows LOCAL branches only** — GetBranches lists refs/heads only, so
  branches that exist on the remote (e.g. `dev` created on origin) are invisible in the
  Branches tab, the PR base dropdown, AND the History branch selector. Fix at the engine:
  GetBranches should also surface remote-tracking branches (refs/remotes) marked
  local/remote/both, with a checkout action to create a local tracking branch. Prereq: a
  fetch must have run to populate origin/* refs (pairs with the fetch --prune auto-refresh
  item). Blocks the two-trunk dev workflow — fix early (branch-management pass).
- [ ] **Auto-refresh (kill the manual Reload button for normal use).** Two parts:
  (1) reload local state on window focus, so returning to the app after terminal/
  browser work reflects reality without a manual Reload; (2) periodic background
  fetch (every few minutes and/or on focus) so ahead/behind + remote state stay
  fresh — NOT fetch-on-every-reload (that's slow/chatty/offline-fragile: reload is
  cheap+local, fetch is network). After this, Reload becomes a rarely-needed manual
  override, not the primary refresh mechanism. Supersedes the earlier bare
  "reload on window focus" note.
  (3) **Prune stale remote-tracking branches:** the periodic/focus fetch must use
  `fetch --prune` so branches deleted on origin (origin/foo) stop showing locally.
  Currently plain fetch leaves stale origin/* refs — deleted-on-remote branches
  linger in the branch list until a manual prune. Folding --prune into the
  auto-refresh fetch fixes this for free. (Reported: deleted remote branches not
  reflected locally.)
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