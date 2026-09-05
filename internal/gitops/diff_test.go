package gitops

import "testing"

func TestDiffUnstaged(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "line1\nline2\nline3\n")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}

	// modify a tracked file — this is an unstaged change
	writeFile(t, dir, "a.txt", "line1\nCHANGED\nline3\n")

	files, err := Diff(dir, DiffOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(files))
	}
	f := files[0]
	if f.NewPath != "a.txt" {
		t.Fatalf("expected path a.txt, got %q", f.NewPath)
	}
	if len(f.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
	}

	var adds, removes int
	for _, ln := range f.Hunks[0].Lines {
		switch ln.Kind {
		case LineAdd:
			adds++
			if ln.Content != "CHANGED" {
				t.Errorf("unexpected added content %q", ln.Content)
			}
		case LineRemove:
			removes++
			if ln.Content != "line2" {
				t.Errorf("unexpected removed content %q", ln.Content)
			}
		}
	}
	if adds != 1 || removes != 1 {
		t.Fatalf("expected 1 add + 1 remove, got %d add %d remove", adds, removes)
	}
}

func TestDiffStagedVsUnstaged(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "one\n")
	_ = Stage(dir)
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}

	// change + stage it → shows in --staged, not in plain diff
	writeFile(t, dir, "a.txt", "two\n")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}

	staged, err := Diff(dir, DiffOptions{Staged: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 1 {
		t.Fatalf("expected 1 staged file diff, got %d", len(staged))
	}

	unstaged, err := Diff(dir, DiffOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(unstaged) != 0 {
		t.Fatalf("expected 0 unstaged diffs (all staged), got %d", len(unstaged))
	}
}
