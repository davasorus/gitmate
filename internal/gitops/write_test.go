package gitops

import "testing"

func TestMergeNoConflict(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	base, _ := CurrentBranch(dir)
	if err := SwitchNew(dir, "feature"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "b.txt", "feature\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "feat"); err != nil {
		t.Fatal(err)
	}
	if err := Switch(dir, base); err != nil {
		t.Fatal(err)
	}
	if err := Merge(dir, "feature"); err != nil {
		t.Fatal(err)
	}
	if MergeInProgress(dir) {
		t.Fatal("expected clean merge, but merge still in progress")
	}
}

func TestResetModes(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	_ = Stage(dir)
	first, _ := CreateCommit(dir, "first")
	writeFile(t, dir, "a.txt", "two\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "second")

	// soft reset back to first: HEAD moves, changes stay staged
	if err := Reset(dir, first, ResetSoft); err != nil {
		t.Fatal(err)
	}
	log, _ := GetLog(dir, 10)
	if len(log) != 1 || log[0].Subject != "first" {
		t.Fatalf("soft reset log wrong: %+v", log)
	}
}

func TestResetHard(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	_ = Stage(dir)
	first, _ := CreateCommit(dir, "first")
	writeFile(t, dir, "a.txt", "two\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "second")

	if err := Reset(dir, first, ResetHard); err != nil {
		t.Fatal(err)
	}
	log, _ := GetLog(dir, 10)
	if len(log) != 1 {
		t.Fatalf("hard reset should leave 1 commit, got %d", len(log))
	}
}

func TestCherryPickClean(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "base\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base")
	base, _ := CurrentBranch(dir)

	// make a commit on a feature branch to cherry-pick back
	if err := SwitchNew(dir, "feature"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "b.txt", "feature\n")
	_ = Stage(dir)
	sha, _ := CreateCommit(dir, "add b")

	if err := Switch(dir, base); err != nil {
		t.Fatal(err)
	}
	if err := CherryPick(dir, sha); err != nil {
		t.Fatal(err)
	}
	cp, rev := SequencerInProgress(dir)
	if cp || rev {
		t.Fatalf("clean cherry-pick should not leave sequencer in progress (cp=%v rev=%v)", cp, rev)
	}
	log, _ := GetLog(dir, 10)
	if len(log) != 2 {
		t.Fatalf("expected 2 commits after cherry-pick, got %d", len(log))
	}
}

func TestRevertClean(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "first")
	writeFile(t, dir, "b.txt", "two\n")
	_ = Stage(dir)
	sha, _ := CreateCommit(dir, "add b")

	if err := Revert(dir, sha); err != nil {
		t.Fatal(err)
	}
	_, rev := SequencerInProgress(dir)
	if rev {
		t.Fatal("clean revert should not leave sequencer in progress")
	}
	log, _ := GetLog(dir, 10)
	if len(log) != 3 {
		t.Fatalf("expected 3 commits after revert, got %d", len(log))
	}
}

func TestRebaseClean(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "base\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base")
	base, _ := CurrentBranch(dir)

	if err := SwitchNew(dir, "feature"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "f.txt", "feat\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "feat commit")

	// advance base with a non-conflicting change
	if err := Switch(dir, base); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "c.txt", "base2\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base advance")

	if err := Switch(dir, "feature"); err != nil {
		t.Fatal(err)
	}
	if err := Rebase(dir, base); err != nil {
		t.Fatal(err)
	}
	if RebaseInProgress(dir) {
		t.Fatal("clean rebase should not be in progress")
	}
}
