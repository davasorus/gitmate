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
