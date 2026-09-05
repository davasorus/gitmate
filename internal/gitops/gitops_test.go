package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newTestRepo makes a fresh git repo in a temp dir and returns its path.
// t.TempDir() is auto-removed when the test finishes.
func newTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// init + identity so commits work in CI (no global config there)
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStatusUntrackedThenStaged(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello")

	// before staging: one untracked file
	s, err := GetStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Untracked) != 1 || s.Untracked[0] != "a.txt" {
		t.Fatalf("expected a.txt untracked, got %+v", s.Untracked)
	}
	if len(s.Changes) != 0 {
		t.Fatalf("expected no staged changes, got %+v", s.Changes)
	}

	// after staging: it moves into the staged column
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}
	s, err = GetStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Untracked) != 0 {
		t.Fatalf("expected nothing untracked after stage, got %+v", s.Untracked)
	}
	if len(s.Changes) != 1 || s.Changes[0].Staged != "added" {
		t.Fatalf("expected a.txt staged as added, got %+v", s.Changes)
	}
}

func TestCommitAndLog(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}

	short, err := CreateCommit(dir, "first commit")
	if err != nil {
		t.Fatal(err)
	}
	if short == "" {
		t.Fatal("expected a short hash, got empty")
	}

	commits, err := GetLog(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
	if commits[0].Subject != "first commit" {
		t.Fatalf("expected subject 'first commit', got %q", commits[0].Subject)
	}
	if commits[0].Short != short {
		t.Fatalf("log hash %q != commit hash %q", commits[0].Short, short)
	}
}

func TestCommitEmptyMessageRejected(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello")
	_ = Stage(dir)

	if _, err := CreateCommit(dir, "   "); err == nil {
		t.Fatal("expected error for empty commit message, got nil")
	}
}

func TestCurrentBranchAfterCommit(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "first"); err != nil {
		t.Fatal(err)
	}

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	// default branch name varies by git config (main/master/live); just assert non-empty
	if branch == "" {
		t.Fatal("expected a branch name, got empty")
	}
}

func TestParseTrackAheadBehind(t *testing.T) {
	// pure-function test, no repo needed
	cases := []struct {
		in            string
		ahead, behind int
	}{
		{"", 0, 0},
		{"gone", 0, 0},
		{"ahead 2", 2, 0},
		{"behind 3", 0, 3},
		{"ahead 2, behind 1", 2, 1},
	}
	for _, c := range cases {
		var b Branch
		parseTrack(&b, c.in)
		if b.Ahead != c.ahead || b.Behind != c.behind {
			t.Errorf("parseTrack(%q) = ahead %d behind %d, want ahead %d behind %d",
				c.in, b.Ahead, b.Behind, c.ahead, c.behind)
		}
	}
}

func TestStageUnstageRoundTrip(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}

	// modify + stage just this file
	writeFile(t, dir, "a.txt", "two\n")
	if err := Stage(dir, "a.txt"); err != nil {
		t.Fatal(err)
	}
	s, _ := GetStatus(dir)
	if len(s.Changes) != 1 || s.Changes[0].Staged != "modified" {
		t.Fatalf("expected a.txt staged as modified, got %+v", s.Changes)
	}

	// unstage it → change moves back to the unstaged column
	if err := Unstage(dir, "a.txt"); err != nil {
		t.Fatal(err)
	}
	s, _ = GetStatus(dir)
	if len(s.Changes) != 1 || s.Changes[0].Unstaged != "modified" {
		t.Fatalf("expected a.txt unstaged as modified, got %+v", s.Changes)
	}
}

func TestDiscardRestoresFile(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "original\n")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}

	// modify the tracked file (unstaged), then discard
	writeFile(t, dir, "a.txt", "changed\n")
	s, _ := GetStatus(dir)
	if len(s.Changes) != 1 {
		t.Fatalf("expected 1 change before discard, got %d", len(s.Changes))
	}
	if err := Discard(dir, "a.txt"); err != nil {
		t.Fatal(err)
	}
	s, _ = GetStatus(dir)
	if len(s.Changes) != 0 || len(s.Untracked) != 0 {
		t.Fatalf("expected clean tree after discard, got %+v / %+v", s.Changes, s.Untracked)
	}
}

func TestDiscardNoPathsErrors(t *testing.T) {
	if err := Discard(t.TempDir()); err == nil {
		t.Fatal("expected error when no paths given")
	}
}

func TestSwitchNewAndBack(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}

	start, err := CurrentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := SwitchNew(dir, "feature-x"); err != nil {
		t.Fatal(err)
	}
	if b, _ := CurrentBranch(dir); b != "feature-x" {
		t.Fatalf("expected feature-x, got %q", b)
	}
	if err := Switch(dir, start); err != nil {
		t.Fatal(err)
	}
	if b, _ := CurrentBranch(dir); b != start {
		t.Fatalf("expected back on %q, got %q", start, b)
	}
}

func TestSwitchNewDuplicateErrors(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchNew(dir, "dup"); err != nil {
		t.Fatal(err)
	}
	// back to a branch, then try to create "dup" again → should error
	if err := SwitchNew(dir, "dup"); err == nil {
		t.Fatal("expected error creating an existing branch")
	}
}

func TestRenameAndDeleteBranch(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}
	start, _ := CurrentBranch(dir)

	// create a branch off HEAD (no unique commits), rename it, then safe-delete it
	if err := SwitchNew(dir, "temp"); err != nil {
		t.Fatal(err)
	}
	if err := Switch(dir, start); err != nil {
		t.Fatal(err)
	}
	if err := RenameBranch(dir, "temp", "temp2"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBranch(dir, "temp2", false); err != nil {
		t.Fatalf("safe delete of merged branch should succeed: %v", err)
	}
}

func TestDeleteCurrentBranchErrors(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}
	cur, _ := CurrentBranch(dir)
	if err := DeleteBranch(dir, cur, false); err == nil {
		t.Fatal("expected error deleting the current branch")
	}
}

func TestStashSaveListPop(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}

	// make a change, stash it, tree should be clean
	writeFile(t, dir, "a.txt", "two\n")
	if err := StashSave(dir, "wip", false); err != nil {
		t.Fatal(err)
	}
	s, _ := GetStatus(dir)
	if len(s.Changes) != 0 {
		t.Fatalf("expected clean tree after stash, got %+v", s.Changes)
	}

	list, err := StashList(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 stash, got %d", len(list))
	}
	if list[0].Message != "wip" {
		t.Errorf("expected message 'wip', got %q", list[0].Message)
	}

	// pop it back, change should return
	if err := StashPop(dir, ""); err != nil {
		t.Fatal(err)
	}
	s, _ = GetStatus(dir)
	if len(s.Changes) != 1 {
		t.Fatalf("expected change restored after pop, got %+v", s.Changes)
	}
	if list, _ := StashList(dir); len(list) != 0 {
		t.Fatalf("expected empty stash list after pop, got %d", len(list))
	}
}
