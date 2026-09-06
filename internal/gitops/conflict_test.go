package gitops

import "testing"

// makeConflict creates a repo with a merge conflict on a.txt and returns the dir.
func makeConflict(t *testing.T) string {
	t.Helper()
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "base\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base")
	base, _ := CurrentBranch(dir)

	if err := SwitchNew(dir, "other"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "theirs\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "other change")

	if err := Switch(dir, base); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, "a.txt", "ours\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "our change")

	_ = Merge(dir, "other") // conflicts
	return dir
}

func TestReadConflictAndResolveTheirs(t *testing.T) {
	dir := makeConflict(t)
	files, err := ConflictedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "a.txt" {
		t.Fatalf("expected a.txt conflicted, got %v", files)
	}
	cf, err := ReadConflict(dir, "a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(cf.Hunks) == 0 {
		t.Fatal("expected conflict hunks")
	}
	if err := ResolveTheirs(dir, "a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := MarkResolved(dir, "a.txt"); err != nil {
		t.Fatal(err)
	}
	files, _ = ConflictedFiles(dir)
	if len(files) != 0 {
		t.Fatalf("expected no conflicts after resolve, got %v", files)
	}
}
