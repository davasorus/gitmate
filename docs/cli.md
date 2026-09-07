# gitmate CLI

gitmate is a git and GitHub companion tool available as both a desktop GUI and a
command-line tool. This document covers the **CLI** — the `gitmate` command. Both
front-ends share the same engine, so the CLI does everything the GUI does, from a
terminal.

## Install

### `go install` (recommended)

Requires **Go 1.25+**:

```bash
go install github.com/davasorus/gitmate/cmd/gitmate@latest
```

This puts the `gitmate` binary in `$(go env GOPATH)/bin` (usually `~/go/bin`).
Make sure that directory is on your `PATH`.

### Download a release binary

Grab `gitmate` (Linux/macOS) or `gitmate.exe` (Windows) from the
[Releases](https://github.com/davasorus/gitmate/releases) page and put it on your
`PATH`. Binaries are not code-signed — on first run your OS may warn you:

- **Windows:** "More info" → "Run anyway"
- **macOS:** right-click → Open, or clear the quarantine attribute

Verify the download against the release `checksums` file if you prefer.

### Build from source

```bash
git clone https://github.com/davasorus/gitmate.git
cd gitmate
go build -o gitmate ./cmd/gitmate
```

## Authentication

GitHub features (PRs, issues, releases, Actions) need a token. gitmate reads it
from the `GITHUB_TOKEN` environment variable, or from a `.env` file in the repo:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxx      # macOS/Linux
$env:GITHUB_TOKEN = "ghp_xxxxxxxxxxxxxxxx"    # PowerShell
```

Scopes:

- **`repo`** — required for most GitHub operations
- **`workflow`** — additionally required to cancel, rerun, or dispatch Actions

Verify auth any time:

```bash
gitmate whoami
```

Purely local git operations (status, commit, branch, diff, stash, …) need no token.

## Usage

gitmate runs against the git repository in the current directory. Run any command
with `--help` for its full flags:

```bash
gitmate <command> --help
```

### Local git

| Command | What it does |
| --- | --- |
| `gitmate status` | working-tree status |
| `gitmate diff [rev]` | changes (working tree, `--staged`, or a revision) |
| `gitmate stage [paths…]` | stage changes (all if no paths) |
| `gitmate unstage [paths…]` | unstage changes |
| `gitmate discard <paths…>` | discard working-tree changes (destructive) |
| `gitmate commit -m "msg"` | commit staged changes |
| `gitmate log` | commit history |
| `gitmate show [rev]` | a commit's metadata + diff (default HEAD) |
| `gitmate blame <path>` | line-by-line authorship |
| `gitmate branch` | list / delete / rename branches |
| `gitmate switch [-c] <name>` | switch to a branch (`-c` to create) |
| `gitmate merge <branch>` · `merge-abort` | merge / abort |
| `gitmate rebase <base>` | rebase (`--continue` / `--abort` mid-rebase) |
| `gitmate cherry-pick <rev>` · `revert <rev>` | apply / undo a commit |
| `gitmate reset <rev>` | move HEAD (`--soft` / `--hard`; default mixed) |
| `gitmate reflog` | where HEAD has been (the undo safety net) |
| `gitmate conflicts` · `resolve --side ours\|theirs <file>` | list / resolve conflicts |
| `gitmate stash` | save / list / pop / drop stashes |
| `gitmate tag` | list / create / delete / push tags |

### Remotes & sync

| Command | What it does |
| --- | --- |
| `gitmate clone <url> [dir]` | clone a repository |
| `gitmate remote` | list / add / remove / rename remotes |
| `gitmate fetch` | download remote changes without merging |
| `gitmate pull [--rebase]` | fetch + integrate (`--rebase` for linear history) |
| `gitmate push` | push the current branch |
| `gitmate init-remote` | create a GitHub repo + add it as origin |

### GitHub

| Command | What it does |
| --- | --- |
| `gitmate whoami` | verify auth, show rate limit |
| `gitmate prs` · `pr …` | list PRs / work with a PR (create, review, merge, checks, …) |
| `gitmate issues` · `issue …` | list issues / work with an issue |
| `gitmate label` | manage labels (definitions + apply) |
| `gitmate release` | manage GitHub releases (assets, notes) |
| `gitmate actions` | list + inspect workflow runs (logs, cancel, rerun, dispatch) |

### Shell completion

```bash
gitmate completion --help    # bash / zsh / fish / powershell
```

## Tips

- gitmate operates on the repo in your current directory — `cd` into the repo first.
- Everything the GUI does, the CLI does — they share one engine, so behavior is identical.
- For the desktop app instead, see the main [README](../README.md) and
  [CONTRIBUTING](../CONTRIBUTING.md).
