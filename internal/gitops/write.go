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
