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
