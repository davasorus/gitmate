package gitops

import "testing"

func TestGetLogAndLogRef(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	commits, err := GetLog(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Subject != "init" {
		t.Fatalf("log wrong: %+v", commits)
	}
	cur, _ := CurrentBranch(dir)
	ref, err := GetLogRef(dir, cur, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(ref) != 1 {
		t.Fatalf("logref wrong: %+v", ref)
	}
}
