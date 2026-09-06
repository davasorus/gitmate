package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestShowCommit(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "first commit"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "one\ntwo\n")
	_ = Stage(dir)
	short, err := CreateCommit(dir, "second commit")
	if err != nil {
		t.Fatal(err)
	}

	d, err := Show(dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if d.Subject != "second commit" {
		t.Errorf("expected subject 'second commit', got %q", d.Subject)
	}
	if d.Short != short {
		t.Errorf("expected short %q, got %q", short, d.Short)
	}
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file in the commit, got %d", len(d.Files))
	}
	var adds int
	for _, h := range d.Files[0].Hunks {
		for _, ln := range h.Lines {
			if ln.Kind == LineAdd {
				adds++
			}
		}
	}
	if adds != 1 {
		t.Errorf("expected 1 added line, got %d", adds)
	}
}

func TestShowRootCommit(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}
	d, err := Show(dir, "HEAD")
	if err != nil {
		t.Fatalf("show on root commit errored: %v", err)
	}
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file in root commit, got %d", len(d.Files))
	}
}

func TestFetchNoRemoteErrors(t *testing.T) {
	// a fresh repo with no remote configured should error on fetch
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}
	if err := Fetch(dir, "origin"); err == nil {
		t.Fatal("expected error fetching with no origin remote")
	}
}

func TestConflictResolveOurs(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "base\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "base"); err != nil {
		t.Fatal(err)
	}
	main, _ := CurrentBranch(dir)

	if err := SwitchNew(dir, "other"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "theirs\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "theirs"); err != nil {
		t.Fatal(err)
	}
	if err := Switch(dir, main); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "ours\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "ours"); err != nil {
		t.Fatal(err)
	}

	_ = Merge(dir, "other") // conflicts
	files, _ := ConflictedFiles(dir)
	if len(files) != 1 {
		t.Fatalf("expected 1 conflict, got %v", files)
	}

	cf, err := ReadConflict(dir, "a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(cf.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(cf.Hunks))
	}

	if err := ResolveOurs(dir, "a.txt"); err != nil {
		t.Fatal(err)
	}
	if files, _ := ConflictedFiles(dir); len(files) != 0 {
		t.Fatalf("expected resolved, still conflicted: %v", files)
	}
	// after taking ours, the file's content should be our side ("ours")
	got, _ := run(dir, "show", "HEAD:a.txt") // committed base still there; check working file instead
	_ = got
	wt, _ := run(dir, "cat-file", "-p", ":0:a.txt") // stage 0 = resolved/staged content
	if !strings.Contains(wt, "ours") || strings.Contains(wt, "theirs") {
		t.Fatalf("expected ours content after ResolveOurs, got %q", wt)
	}
}

func TestClone(t *testing.T) {
	// make a source repo with one commit
	src := newTestRepo(t)
	writeFile(t, src, "a.txt", "hello\n")
	_ = Stage(src)
	if _, err := CreateCommit(src, "init"); err != nil {
		t.Fatal(err)
	}

	// clone it into a fresh dest under a temp dir
	dest := t.TempDir() + string(os.PathSeparator) + "cloned"
	got, err := Clone(src, dest)
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	// the cloned repo should have the committed file
	if _, err := run(got, "cat-file", "-e", "HEAD:a.txt"); err != nil {
		t.Fatalf("expected a.txt in clone: %v", err)
	}
}

// newRemoteRepo makes a bare repo (acting as origin) + a working clone wired to
// it, with one pushed commit on the current branch. Returns the working dir.
func newRemoteRepo(t *testing.T) string {
	t.Helper()
	bare := t.TempDir()
	if _, err := run(bare, "init", "--bare"); err != nil {
		t.Fatal(err)
	}
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello\n")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}
	if err := AddRemote(dir, "origin", bare); err != nil {
		t.Fatal(err)
	}
	// push the ACTUAL current branch (newTestRepo's default may be master, not main)
	br, err := CurrentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := Push(dir, "origin", br, true); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestStatusVariedStates(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "keep.txt", "keep\n")
	writeFile(t, dir, "del.txt", "gone\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	// added (new staged file), deleted (removed staged), modified (changed)
	writeFile(t, dir, "new.txt", "added\n")
	writeFile(t, dir, "keep.txt", "changed\n")
	if _, err := run(dir, "rm", "del.txt"); err != nil {
		t.Fatal(err)
	}
	_ = Stage(dir)

	st, err := GetStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil {
		t.Fatal("expected status")
	}
	// exercise the parse paths; expect staged file changes present
	if len(st.Changes) == 0 {
		t.Fatalf("expected file changes, got %+v", st)
	}
}

func TestStatusUpstreamAheadBehind(t *testing.T) {
	// working repo with an origin → gives branch.upstream + branch.ab header lines
	dir := newRemoteRepo(t)
	// make a local commit so we're ahead of origin
	writeFile(t, dir, "ahead.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "ahead commit")

	st, err := GetStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if st.Upstream == "" {
		t.Fatalf("expected an upstream to be parsed, got %+v", st)
	}
	if st.Ahead < 1 {
		t.Fatalf("expected ahead>=1, got %d", st.Ahead)
	}
}

func TestStatusDetachedHead(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	_ = Stage(dir)
	sha, _ := CreateCommit(dir, "one")
	writeFile(t, dir, "a.txt", "two\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "two")

	// detach HEAD at the first commit
	if _, err := run(dir, "checkout", sha); err != nil {
		t.Fatal(err)
	}
	st, err := GetStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Detached {
		t.Fatalf("expected detached HEAD, got %+v", st)
	}
}

func TestCodeName(t *testing.T) {
	cases := map[byte]string{
		'.': "", 'M': "modified", 'A': "added", 'D': "deleted",
		'R': "renamed", 'C': "copied", 'U': "unmerged",
	}
	for c, want := range cases {
		if got := codeName(c); got != want {
			t.Errorf("codeName(%q) = %q, want %q", c, got, want)
		}
	}
	// default branch: unknown code returns the char as string
	if got := codeName('X'); got != "X" {
		t.Errorf("codeName('X') = %q, want X", got)
	}
}

func TestUnstageAndDeleteBranchErrors(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	// unstage a specific path (covers the path-arg branch)
	writeFile(t, dir, "b.txt", "new\n")
	_ = Stage(dir)
	if err := Unstage(dir, "b.txt"); err != nil {
		t.Fatal(err)
	}

	// deleting a non-existent branch should error (covers the error branch)
	if err := DeleteBranch(dir, "does-not-exist", false); err == nil {
		t.Fatal("expected error deleting non-existent branch")
	}
}
