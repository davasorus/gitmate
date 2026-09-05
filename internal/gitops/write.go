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

// CurrentBranch returns the checked-out branch name.
func CurrentBranch(dir string) (string, error) {
	out, err := run(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
