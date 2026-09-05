package gitops

import (
	"errors"
	"strings"
)

// Stage runs `git add`. With no paths, stages everything (new, modified, deleted).
func Stage(dir string, paths ...string) error {
	args := []string{"add"}
	if len(paths) == 0 {
		args = append(args, "-A")
	} else {
		args = append(args, paths...)
	}
	_, err := run(dir, args...)
	return err
}

// Unstage removes paths from the index (staged -> unstaged), leaving the
// working tree untouched. With no paths, unstages everything.
func Unstage(dir string, paths ...string) error {
	args := []string{"restore", "--staged"}
	if len(paths) == 0 {
		args = append(args, ".")
	} else {
		args = append(args, paths...)
	}
	_, err := run(dir, args...)
	return err
}

// Discard throws away working-tree changes to tracked paths, restoring them to
// their staged/HEAD state. This is DESTRUCTIVE — uncommitted edits are lost and
// cannot be recovered. It does not touch untracked files (git restore ignores
// them); callers should guard against passing untracked paths.
func Discard(dir string, paths ...string) error {
	if len(paths) == 0 {
		return errors.New("discard requires at least one path")
	}
	args := append([]string{"restore"}, paths...)
	_, err := run(dir, args...)
	return err
}

// Switch changes the checked-out branch. Git refuses if uncommitted changes
// would be overwritten; that refusal is returned as an error for the caller
// to surface (commit, discard, or — once available — stash first).
func Switch(dir, branch string) error {
	_, err := run(dir, "switch", branch)
	return err
}

// SwitchNew creates a new branch from the current HEAD and switches to it
// (git switch -c). Fails if the branch already exists.
func SwitchNew(dir, branch string) error {
	_, err := run(dir, "switch", "-c", branch)
	return err
}

// DeleteBranch deletes a branch. With force=false it uses safe delete (-d),
// which refuses to delete a branch holding commits not merged elsewhere. With
// force=true it uses -D, deleting regardless (may orphan commits). Git refuses
// to delete the currently checked-out branch either way.
func DeleteBranch(dir, name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := run(dir, "branch", flag, name)
	return err
}

// RenameBranch renames a branch (git branch -m). Renaming the current branch
// is allowed.
func RenameBranch(dir, oldName, newName string) error {
	_, err := run(dir, "branch", "-m", oldName, newName)
	return err
}

// CreateCommit creates a commit with the given message and returns the new short hash.
func CreateCommit(dir, message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", errors.New("commit message cannot be empty")
	}
	if _, err := run(dir, "commit", "-m", message); err != nil {
		return "", err
	}
	out, err := run(dir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Push pushes a branch to a remote. setUpstream adds -u to link the local
// branch to remote/<branch> (needed the first time a branch is pushed).
func Push(dir, remote, branch string, setUpstream bool) error {
	args := []string{"push"}
	if setUpstream {
		args = append(args, "-u")
	}
	args = append(args, remote, branch)
	_, err := run(dir, args...)
	return err
}

// Fetch downloads objects and refs from a remote without merging. Empty remote
// defaults to "origin". Updates remote-tracking branches (so ahead/behind
// reflect the remote) but does not touch the working tree.
func Fetch(dir, remote string) error {
	if remote == "" {
		remote = "origin"
	}
	_, err := run(dir, "fetch", remote)
	return err
}

// Pull integrates remote changes into the current branch. rebase=true replays
// local commits on top of upstream (linear); otherwise merges. Either mode can
// conflict — returned as an error for the caller to surface.
func Pull(dir string, rebase bool) error {
	args := []string{"pull"}
	if rebase {
		args = append(args, "--rebase")
	}
	_, err := run(dir, args...)
	return err
}

// CurrentBranch returns the checked-out branch name.
func CurrentBranch(dir string) (string, error) {
	out, err := run(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Merge merges the named branch into the current branch. On conflicts git exits
// non-zero and leaves conflict markers; that error is returned to surface.
func Merge(dir, branch string) error {
	_, err := run(dir, "merge", branch)
	return err
}

// MergeAbort aborts an in-progress merge, restoring the pre-merge state.
func MergeAbort(dir string) error {
	_, err := run(dir, "merge", "--abort")
	return err
}

// ConflictedFiles returns paths currently in a conflicted (unmerged) state.
func ConflictedFiles(dir string) ([]string, error) {
	out, err := run(dir, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// MergeInProgress reports whether a merge is underway (MERGE_HEAD exists).
func MergeInProgress(dir string) bool {
	_, err := run(dir, "rev-parse", "--verify", "--quiet", "MERGE_HEAD")
	return err == nil
}

// Rebase replays the current branch's commits on top of base. Like merge it can
// conflict — but per replayed commit, so it may stop repeatedly. On conflict git
// exits non-zero; resolve then RebaseContinue, or RebaseAbort to bail.
func Rebase(dir, base string) error {
	_, err := run(dir, "rebase", base)
	return err
}

// RebaseContinue resumes a rebase after conflicts are resolved and staged.
func RebaseContinue(dir string) error {
	// -c core.editor=true skips the commit-message editor prompt.
	_, err := run(dir, "-c", "core.editor=true", "rebase", "--continue")
	return err
}

// RebaseAbort aborts an in-progress rebase, restoring the pre-rebase state.
func RebaseAbort(dir string) error {
	_, err := run(dir, "rebase", "--abort")
	return err
}

// RebaseInProgress reports whether a rebase is underway (the rebase-merge or
// rebase-apply state dir exists under .git).
func RebaseInProgress(dir string) bool {
	// `git rev-parse --git-path` gives the path; existence check via status.
	out, err := run(dir, "status")
	if err != nil {
		return false
	}
	return strings.Contains(out, "rebase in progress") || strings.Contains(out, "interactive rebase in progress")
}
