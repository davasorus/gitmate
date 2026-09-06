package gitops

import "testing"

func TestBlame(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	lines, err := Blame(dir, "a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Fatal("expected blame lines")
	}
}
