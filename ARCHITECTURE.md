# Architecture

gitmate is built around one principle: **a single shared engine, two front-ends.** The
command-line tool and the desktop GUI never reimplement logic — they both call the same Go
packages, so their behavior is identical by construction.

## Layout

```Plain Text
internal/gitops/   local git operations, via os/exec
internal/ghapi/    GitHub operations, via REST (go-github) + GraphQL (githubv4)
cmd/gitmate/       Cobra CLI — calls the engine directly
gui/               Wails v3 desktop app
  gitservice.go    thin service layer — wraps the engine for the frontend
  frontend/        React + Tailwind (+ shadcn) UI, generated bindings to the service
```

Both `cmd/gitmate` and `gui` import `internal/gitops` and `internal/ghapi`. Nothing in the UI
layers duplicates engine logic.

## Fixed decisions

### 1. Shared engine, two front-ends

The engine (`internal/`) is UI-agnostic. The CLI is a thin Cobra wrapper; the GUI is a thin Wails
service wrapper. A feature is built once in the engine and both surfaces get it. This is the core
constraint — resist putting logic in `cmd/` or `gui/`.

### 2. `internal/gitops` shells out to `git` via os/exec

Rather than a pure-Go git implementation (go-git), gitmate runs the real `git` binary. Reasons:

- Exact fidelity with the user's git (config, hooks, credential helpers, worktrees).
- Every git feature is available without reimplementing it.
- Output parsing is the cost; the engine centralizes it (porcelain formats, null-separated fields).

All commands run through one `run(dir, args...)` helper. `repoDir` is pinned to the repository
**root** (`rev-parse --show-toplevel`) so path-relative commands work regardless of the process's
working directory.

### 3. `internal/ghapi` — REST + GraphQL, mixed on purpose

- **REST** (google/go-github v88): simple single-resource reads and all writes/mutations.
- **GraphQL** (shurcooL/githubv4): nested/aggregated reads where one query beats N REST round-trips —
  PR detail, the PR list (review decision + CI check rollup), the issue list (labels + assignees).
- Both clients share one oauth2-authenticated HTTP client. GitHub supports mixing them; node IDs
  bridge the two when needed.
- GitHub Actions is REST-only (GraphQL has no Actions API), so all Actions code is REST.

### 4. The GUI service layer

`gui/gitservice.go` is a deliberately thin adapter: each method resolves the repo/owner, calls the
engine, and returns engine types. Wails generates TypeScript bindings from this service, so the
frontend calls Go methods as typed async functions. The frontend holds no git logic — it renders
engine data and calls service methods.

### 5. Auth

A GitHub token is read from `GITHUB_TOKEN` (or a `.env` file). `repo` scope covers most operations;
the `workflow` scope is additionally required to cancel/rerun/dispatch Actions (the engine returns a
clear error when a 403 indicates the scope is missing).

## Data flow

```Plain Text
user action
   │
   ├─ CLI:  cmd/gitmate  ─┐
   │                      ├─►  internal/gitops  ──►  git (os/exec)
   └─ GUI:  frontend      │    internal/ghapi   ──►  GitHub REST + GraphQL
            → gitservice ─┘
```

## Testing

The engine is unit-tested directly. `internal/gitops` tests build throwaway repositories in temp
dirs and run real git against them; `internal/ghapi` tests point the REST/GraphQL clients at an
`httptest` mock server. CI enforces a coverage threshold on `internal/...`.

## CI/CD

Two trunks — `dev-Branch` (integration) and `live` (release). CI runs on every push and PR
(matrix build across Linux/Windows/macOS, lint, format, gofmt, coverage gate). Releases are cut
from `live` through a single gated pipeline: propose (compute version + notes, create nothing) →
approve (environment reviewer) → build (CLI + GUI) → publish one release. See CONTRIBUTING.md.
