# Contributing

Thanks for working on gitmate. This document covers the build, the branching model, commit
conventions, and how to check your work locally before pushing.

## Prerequisites

- **Go 1.25+**
- For the GUI: **[Wails v3](https://v3alpha.wails.io/)** (`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`),
  **Node 20+**, and on Linux the GTK4/WebKit dev packages (`libgtk-4-dev libwebkitgtk-6.0-dev`).
- A GitHub token in `GITHUB_TOKEN` (or a `.env` file) for any GitHub feature — `repo` scope, plus
  `workflow` scope if you touch Actions cancel/rerun/dispatch.

## Build & run

```bash
# CLI
go build -o gitmate ./cmd/gitmate
./gitmate status

# GUI
cd gui
wails3 generate bindings   # regenerate TS bindings after changing gui/gitservice.go
wails3 dev                 # run the desktop app in dev mode
wails3 build               # production build
```

Regenerate bindings whenever you change the signatures/types exposed by `gui/gitservice.go` —
the frontend calls the generated TypeScript.

## Project layout

See [ARCHITECTURE.md](ARCHITECTURE.md). In short: `internal/gitops` + `internal/ghapi` are the
shared engine; `cmd/gitmate` is the CLI; `gui/` is the desktop app. **Build features in the engine**,
not in the CLI or GUI layers.

## Branching model (two trunks)

```plain Text
feat/your-thing  ──PR──►  dev-Branch  ──(accumulate)──►  PR  ──►  live  (releases cut here)
```

- Branch off **`dev-Branch`** for normal work: `feat/...`, `fix/...`.
- Open a PR into `dev-Branch`. CI must be green.
- `live` is the release trunk. Changes reach it via a PR from `dev-Branch`.
- Do not push directly to the trunks.

## Commit conventions

gitmate uses **conventional commits** — semantic-release parses them to compute the next version
and group the release notes:

- `feat: ...` → minor bump, **Features**
- `fix: ...` → patch bump, **Fixes**
- `perf:`, `refactor:`, `docs:`, `build:`, `ci:` → grouped accordingly
- `feat!:` / a `BREAKING CHANGE:` footer → major bump

Adoption is **soft** — semantic-release parses what it can; commit-lint is not enforced. But the
PR title / commit subject drives the release, so use the types where it matters.

## Check your work locally (matches CI)

Run these before pushing — they mirror what CI enforces.

**Go:**

```bash
gofmt -w cmd internal                       # format (CI fails on unformatted code)
go vet ./cmd/... ./internal/...
go build ./cmd/... ./internal/...
go test ./internal/...                       # add -race locally if you have cgo
go test -coverprofile=coverage.out -covermode=count ./internal/...
go tool cover -func=coverage.out | tail -1   # total must be >= 85%
golangci-lint run ./cmd/... ./internal/...
```

**Frontend (from `gui/frontend`):**

```bash
npm run format        # prettier --write (CI runs format:check)
npm run lint          # oxlint
npx tsc --noEmit      # typecheck
```

CI runs the same steps: matrix build (Linux/Windows/macOS), gofmt gate, `go vet`, race tests,
a **coverage gate (≥ 85% on `internal/...`)**, golangci-lint, and the frontend lint/format/typecheck.
Coverage is measured without `-race` so the number is reproducible locally.

## Tests

- `internal/gitops`: build a throwaway repo in a temp dir and run real git against it (see the
  `newTestRepo` / `newRemoteRepo` helpers).
- `internal/ghapi`: point the REST/GraphQL client at an `httptest` mock server (see the helpers in
  `testutil_test.go`).
- Name test files after the source file they cover (`write_test.go` tests `write.go`).

## Releases

Releases are cut from `live` through one gated pipeline — see the release workflow. Two entry points:

1. **Automated** — a push to `live` runs semantic-release (dry-run) to propose the next version +
   grouped notes.
2. **Manual tag** — pushing a `v*` tag releases at that exact version.

Both converge on: **propose → approval gate (environment reviewer) → build (CLI + GUI) → publish one
release** with the tag, grouped notes, and all platform assets. Nothing is built or published until
the proposed version + notes are reviewed and approved.
