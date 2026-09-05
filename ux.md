# gitmate — UX & UI Plan

> Companion to ROADMAP.md. This governs how the GUI is organized, how you navigate,
> how git state is communicated, and how it looks. The roadmap says WHAT features
> exist; this says WHERE they live and HOW they feel. Build features into this
> structure — do not bolt new panels onto the old stacked-card dashboard.

---

## Locked decisions

1. **Layout: sidebar + main area** (VS Code Source Control shape).
   Persistent left sidebar for navigation; one large main area that swaps content
   based on the selected sidebar item. NOT full three-pane / commit-graph-as-hero.
   Rationale: lighter, scales cleanly (every new feature is a sidebar section or a
   main view), doesn't force history to be the star.

2. **Home = working changes.** Default main view is the working tree
   (staged/unstaged files + commit box + push). History is a peer view, one click
   away in the sidebar — present, not primary. If "feel" later favors history,
   flip the default; no restructure needed.

3. **Dense display, approachable interaction.** Two different axes, both turned up:
   - Information density HIGH (rich file lists, git state everywhere, keyboard shortcuts).
   - Interaction friendliness HIGH (labeled actions, discoverable, confirms on
     destructive ops, no hidden magic).
   This is why VS Code Source Control serves both power users and beginners. We copy
   that stance: power-user density + beginner-safe interaction.

---

## Information Architecture

### App shell
```
┌───────────────┬─────────────────────────────────────────────┐
│  SIDEBAR      │  MAIN AREA                                    │
│  (nav +       │  (swaps based on sidebar selection)           │
│   git state)  │                                               │
│               │                                               │
│  repo header  │   [ working changes | history | branch |      │
│  ───────────  │     PR detail | diff | conflicts | ... ]      │
│  CHANGES      │                                               │
│  BRANCHES     │                                               │
│  REMOTES      │                                               │
│  STASHES      │                                               │
│  TAGS         │                                               │
│  PULL REQS    │                                               │
│  ───────────  │                                               │
│  status bar   │                                               │
└───────────────┴─────────────────────────────────────────────┘
```

### Sidebar sections (top → bottom)
- **Repo header** — repo name, current branch, ahead/behind, dirty indicator.
  Repo path switcher lives here (replaces today's free-text path box).
- **CHANGES** — count badge of modified/staged files. Click → working-changes view (home).
- **BRANCHES** — list; current marked; click a branch → branch view; actions:
  new, switch, delete, rename. (Roadmap 1.2, 1.3)
- **REMOTES** — list; add/remove/rename. (Roadmap 3.1)
- **STASHES** — list; save/pop/drop. (Roadmap 1.7)
- **TAGS** — list; create/delete. (Roadmap 2.6)
- **PULL REQUESTS** — open PRs; click → PR detail in main. (existing + 3.3)
- **Status bar (bottom)** — global state line: branch · ahead/behind · sync buttons
  (fetch/pull/push) · background-op spinner.

### Main-area views (one at a time, chosen by sidebar)
- **Working changes (home)** — unstaged list, staged list, commit box, commit+push.
  Selecting a file shows its diff in the same view (split or stacked). (1.1, 1.4, 1.5)
- **History** — commit list for current branch; select a commit → its diff/metadata. (1.8)
- **Branch view** — that branch's commits + ahead/behind + actions.
- **Diff view** — the backbone renderer (see Component Inventory). Used inline in
  working-changes and history, and standalone for commit-to-commit.
- **PR detail** — title/body, checks (color-coded), reviews/comments, merge action. (3.3, 3.4)
- **Conflicts** — per-file conflict resolution when merge/rebase/pull conflicts. (2.2)

---

## Navigation flows (the core loops)

- **Daily commit:** CHANGES → stage files (per-file or all) → write message → Commit → Push.
- **Inspect history:** BRANCHES/HISTORY → click commit → read diff → (optional) act.
- **Branch work:** BRANCHES → switch / new / delete; dirty-tree warns before switch.
- **PR loop:** PULL REQUESTS → PR detail → view checks → comment → merge.
- **Conflict:** a failed merge/rebase/pull routes to the Conflicts view automatically;
  resolve per file → mark resolved → continue.

Principle: **selection in the sidebar drives the main area.** One thing selected,
one main view. No modal stacking, no card pileups.

---

## State communication (git has many states — surfacing them IS the UX)

Every relevant git state must be legible at a glance, tied to color/iconography:
- **File status** — added / modified / deleted / renamed / untracked / conflicted.
  Consistent letter+color badge, same everywhere the file appears.
- **Branch state** — current, ahead N, behind N, no-upstream, gone, detached HEAD.
- **Sync state** — clean / dirty / operation-in-progress.
- **Check state** — success / failure / pending / neutral (already color-coded; keep).
- **Conflict state** — loud and unmissable; blocks "continue" until resolved.

Rule: never make the user run a command to learn state the UI could show. The status
bar + sidebar badges should answer "where am I / what's the situation" without a click.

---

## Component inventory (build once, reuse everywhere)

- [ ] **DiffView** — unified/split diff renderer: file header, hunk headers, +/- line
  coloring, line numbers. THE backbone (ROADMAP C.1). Reused by working-changes,
  history, commit detail, conflict preview, PR review.
- [ ] **FileStatusRow** — path + status badge + per-file actions (stage/unstage/discard).
  Used in working-changes and commit detail.
- [ ] **CommitRow** — short hash + subject + author + relative time. Used in history,
  branch view, PR commits.
- [ ] **BranchItem** — name + current marker + ahead/behind + row actions.
- [ ] **CheckBadge** — check name + colored status. (exists inline; extract it)
- [ ] **StatusBadge** — the added/modified/deleted/conflict letter+color chip.
- [ ] **ActionButton / ConfirmButton** — standard button; ConfirmButton for destructive
  ops (reset --hard, force push, branch/stash delete). (ROADMAP C.3)
- [ ] **Toast** — success/error (exists; keep).
- [ ] **EmptyState** — consistent "nothing here yet" for empty lists.

---

## Visual system

### Point of view (the "de-AI-ify" brief)
The current UI reads as generic-LLM output: three identical shadcn cards stacked
vertically, default styling, no opinion. Fixes, concrete:
- **Monospace-forward developer aesthetic.** Lean into JetBrains Mono (already in use).
  A git client is a developer tool; it should look like one, not like a generic SaaS
  dashboard. Mono for code/paths/hashes/diffs; a clean UI sans only for chrome if needed.
- **Kill the stacked-cards pattern.** Replace the vertical card pile with the
  sidebar+main shell. Cards, if used at all, are for genuinely grouped content — not
  the primary layout device.
- **Color = meaning, not decoration.** Color is reserved for git semantics:
  added (green), removed (red), modified (amber/blue), conflict (loud red/orange),
  ahead/behind, check pass/fail. Chrome stays neutral so the semantic colors read.
- **Intentional density.** Tight, information-rich rows (like a terminal / VS Code SCM),
  not airy marketing whitespace. Density is the aesthetic.
- **Restraint.** No gratuitous gradients, shadows, rounded-everything. Flat, precise,
  legible. The diff and the file list are the stars; chrome recedes.

### Tokens (define once, in style.css theme)
- [ ] spacing scale (tight baseline)
- [ ] type scale (mono primary; sizes for chrome / body / code)
- [ ] semantic color tokens:
  `--added --removed --modified --conflict --ahead --behind --check-pass --check-fail --check-pending`
- [ ] neutral chrome tokens (bg / panel / border / muted / fg), light + dark
  (system-aware already in place; extend, don't replace)

### Visual identity caveat
Structure and interaction (this doc) are solid to build now. A distinctive *visual
identity* — the thing that makes it look designed rather than defaulted — benefits
from a deliberate pass (or a designer's eye) once the shell exists. Don't block
building on it; do budget a polish pass after the sidebar+main shell + DiffView land.

---

## Build sequencing (UI side, interleaved with ROADMAP tiers)

1. **Shell first.** Build the sidebar + main-area shell and move today's panels into it:
   - CHANGES → working-changes home view
   - existing branches/commits/PRs → sidebar sections + their main views
   - status bar with sync buttons
   This replaces the stacked-card App.tsx. Do it alongside/just before ROADMAP 1.1.
2. **DiffView (C.1)** — the backbone component; lands with ROADMAP 1.1 (diff).
3. Extract shared components (FileStatusRow, CommitRow, BranchItem, StatusBadge) as
   the features that need them land.
4. Semantic color tokens + mono-forward restyle — the "de-AI-ify" pass — once the
   shell + DiffView exist and there's real structure to style.
5. Visual-identity polish pass — deliberate, after the above.

## Working agreement (UI)
- New feature UI goes into the shell (a sidebar section and/or a main view), never as
  another stacked card.
- Reuse the component inventory; extract a shared component the second time a pattern
  repeats.
- Destructive actions always use ConfirmButton.
- Keep the errors-surface-as-toast + nil-safe rendering rules from ROADMAP.
- Update checkboxes here as components/tokens land.