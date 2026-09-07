# gitmate

A git operations client available as both a command-line tool and a desktop GUI,
sharing one Go engine underneath.

gitmate is a *git-operations* client (branches, commits, merges, PRs, releases, CI), not a
code editor. It wraps `git` and the GitHub API behind a single engine that both front-ends call,
so the CLI and GUI always behave identically.

## Features

**Local git**

- Status, staging, commit, discard
- Branches (local **and** remote-tracking), switch/checkout, merge, rebase, rename, delete
- History log, commit detail, blame, reflog, reset
- Stash (save/list/pop/apply/drop)
- Merge / rebase / cherry-pick / revert with conflict resolution (take ours/theirs, mark resolved, commit merge)
- Tags (annotated, push, delete local/remote), remotes (add/rename/remove, clone)

**GitHub (REST + GraphQL)**

- Pull requests: list with review-decision + CI status rollup, create, review (approve / request-changes / comment),
  inline + review-thread comments, resolve/unresolve threads, merge, close/reopen
- Issues: list with labels + assignees, create, label, close/reopen
- Labels: full CRUD
- Releases: full CRUD, asset upload/download, generate release notes
- **GitHub Actions**: workflows, runs (grouped, status-filtered), job logs (colorized), cancel/rerun/rerun-failed,
  `workflow_dispatch` with inputs, and a run **dependency flowchart** (including matrix legs)

Reads that aggregate nested data (PR detail, PR list, issue list) use GraphQL; simple reads and
all writes use REST — mixed intentionally.

## Install

### Download a release

Grab the latest build from the [Releases](https://github.com/davasorus/gitmate/releases) page:
- **CLI**: `gitmate` (Linux/macOS) or `gitmate.exe` (Windows) — a single static binary
- **GUI**: the desktop app for your platform

Binaries are **not code-signed**. On first run your OS may warn you:

- **Windows**: "More info" → "Run anyway"
- **macOS**: right-click → Open (or clear the quarantine attribute)
- Verify the download against the release `checksums` file if you prefer, or build from source (below).

### Build from source

Requires **Go 1.25+**. The GUI additionally requires [Wails v3](https://v3alpha.wails.io/) and,
on Linux, GTK4/WebKit dev packages.

```bash
# CLI
go build -o gitmate ./cmd/gitmate

# GUI (from gui/)
cd gui
wails3 build
```

## Usage

### Authentication

gitmate reads a GitHub token from the `GITHUB_TOKEN` environment variable (a `.env` file in the
repo is also honored). The token needs the `repo` scope; to cancel/rerun/dispatch Actions it also
needs the **`workflow`** scope.

### CLI

```bash
gitmate status
gitmate commit -m "message"
gitmate branch
gitmate log

# GitHub
gitmate pr list
gitmate pr checks <number>
gitmate issue list
gitmate release list
gitmate actions            # workflow runs
gitmate actions dispatch <workflow.yml> key=value
```

Run `gitmate <command> --help` for the full flag set of any command.

### GUI

```bash
cd gui
wails3 dev      # development
```

The desktop app presents the same operations as tabbed views: Changes, Branches, History,
Conflicts, Stashes, Reflog, Tags, Remotes, Pull Requests, Issues, Labels, Releases, and Actions.
It auto-refreshes on window focus and periodically fetches (with prune) so remote state stays current.

## Architecture

One shared engine, two front-ends:

```Plain Text
internal/gitops/   git via os/exec         ┐
internal/ghapi/    GitHub REST + GraphQL    ├─ shared engine
                                            │
cmd/gitmate/       Cobra CLI  ──────────────┤
gui/               Wails v3 desktop app  ───┘  (React + Tailwind, calls a thin service layer)
```

The CLI and GUI never reimplement logic — they both call `internal/gitops` and `internal/ghapi`.
See [ARCHITECTURE.md](ARCHITECTURE.md) for the design decisions and [CONTRIBUTING.md](CONTRIBUTING.md)
for the development workflow.

## License

MIT — see [LICENSE](LICENSE).