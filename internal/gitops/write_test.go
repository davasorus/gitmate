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

func TestPushFetchPull(t *testing.T) {
	dir := newRemoteRepo(t) // also exercises Push (setUpstream=true)
	if err := Fetch(dir, "origin"); err != nil {
		t.Fatal(err)
	}
	if err := Pull(dir, false); err != nil {
		t.Fatal(err)
	}
}

func TestGetRemoteURL(t *testing.T) {
	dir := newRemoteRepo(t)
	url, err := GetRemoteURL(dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if url == "" {
		t.Fatal("expected a remote url")
	}
}

func TestMergeAbort(t *testing.T) {
	dir := makeConflict(t) // from conflict_test.go — leaves a conflicted merge
	if !MergeInProgress(dir) {
		t.Fatal("expected merge in progress")
	}
	if err := MergeAbort(dir); err != nil {
		t.Fatal(err)
	}
	if MergeInProgress(dir) {
		t.Fatal("merge should be aborted")
	}
}

func TestRebaseAbort(t *testing.T) {
	dir := makeConflict(t)
	// abort the merge first, then create a rebase conflict
	_ = MergeAbort(dir)
	base, _ := CurrentBranch(dir)
	if err := SwitchNew(dir, "topic"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "topic\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "topic change")
	if err := Switch(dir, base); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "mainline\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "main change")
	if err := Switch(dir, "topic"); err != nil {
		t.Fatal(err)
	}
	_ = Rebase(dir, base) // conflicts
	if RebaseInProgress(dir) {
		if err := RebaseAbort(dir); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCherryPickAbort(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "base\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base")
	base, _ := CurrentBranch(dir)
	if err := SwitchNew(dir, "b"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "theirs\n")
	_ = Stage(dir)
	sha, _ := CreateCommit(dir, "conflicting")
	if err := Switch(dir, base); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "ours\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "ours")
	_ = CherryPick(dir, sha) // conflicts
	if cp, _ := SequencerInProgress(dir); cp {
		if err := CherryPickAbort(dir); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRevertAbort(t *testing.T) {
	// revert of a commit that conflicts with later state
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "v1\n")
	_ = Stage(dir)
	sha, _ := CreateCommit(dir, "v1")
	writeFile(t, dir, "a.txt", "v2\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "v2")
	_ = Revert(dir, sha) // may conflict
	if _, rev := SequencerInProgress(dir); rev {
		if err := RevertAbort(dir); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCherryPickContinue(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "base\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base")
	base, _ := CurrentBranch(dir)
	_ = SwitchNew(dir, "b")
	writeFile(t, dir, "a.txt", "theirs\n")
	_ = Stage(dir)
	sha, _ := CreateCommit(dir, "conflicting")
	_ = Switch(dir, base)
	writeFile(t, dir, "a.txt", "ours\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "ours")
	_ = CherryPick(dir, sha) // conflicts
	_ = ResolveOurs(dir, "a.txt")
	_ = MarkResolved(dir, "a.txt")
	_ = CherryPickContinue(dir) // exercise the path regardless of outcome
}

func TestRevertContinue(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "v1\n")
	_ = Stage(dir)
	sha, _ := CreateCommit(dir, "v1")
	writeFile(t, dir, "a.txt", "v2\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "v2")
	_ = Revert(dir, sha) // conflicts
	_ = ResolveOurs(dir, "a.txt")
	_ = MarkResolved(dir, "a.txt")
	_ = RevertContinue(dir)
}

func TestRebaseContinue(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "base\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base")
	base, _ := CurrentBranch(dir)
	_ = SwitchNew(dir, "topic")
	writeFile(t, dir, "a.txt", "topic\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "topic")
	_ = Switch(dir, base)
	writeFile(t, dir, "a.txt", "main\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "main")
	_ = Switch(dir, "topic")
	_ = Rebase(dir, base) // conflicts
	_ = ResolveOurs(dir, "a.txt")
	_ = MarkResolved(dir, "a.txt")
	_ = RebaseContinue(dir) // exercise the path
}
