package gitops

import "testing"

func TestGetBranches(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	brs, err := GetBranches(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(brs) == 0 {
		t.Fatal("expected at least one branch")
	}
}

func TestGetBranchesIncludesRemote(t *testing.T) {
	// working repo with an origin that has a branch the local repo doesn't
	dir := newRemoteRepo(t) // pushes current branch to a bare origin
	origin, err := GetRemoteURL(dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	// create a branch in a second clone and push it, so origin has a branch
	// our 'dir' repo has no local ref for.
	other := t.TempDir()
	if _, err := run(other, "clone", origin, "."); err != nil {
		t.Fatal(err)
	}
	if err := SwitchNew(other, "feature-x"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, other, "f.txt", "x\n")
	_ = Stage(other)
	_, _ = CreateCommit(other, "feat")
	br, _ := CurrentBranch(other)
	if err := Push(other, "origin", br, true); err != nil {
		t.Fatal(err)
	}
	// fetch in dir so origin/feature-x exists as a remote-tracking ref
	if err := Fetch(dir, "origin"); err != nil {
		t.Fatal(err)
	}

	brs, err := GetBranches(dir)
	if err != nil {
		t.Fatal(err)
	}
	var remoteOnly *Branch
	for i := range brs {
		if brs[i].Name == "feature-x" {
			remoteOnly = &brs[i]
		}
	}
	if remoteOnly == nil {
		t.Fatalf("remote-only branch feature-x not surfaced: %+v", brs)
	}
	if !remoteOnly.IsRemote || remoteOnly.IsLocal {
		t.Errorf("feature-x should be remote-only: %+v", *remoteOnly)
	}
	// and switching to it should create a local tracking branch
	if err := Switch(dir, "feature-x"); err != nil {
		t.Fatalf("checkout of remote-only branch failed: %v", err)
	}
}
