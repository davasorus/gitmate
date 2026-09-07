package gitops

import (
	"strings"
	"testing"
)

func TestStageHunk(t *testing.T) {
	dir := newTestRepo(t)
	// a file with several lines, committed
	writeFile(t, dir, "f.txt", "l1\nl2\nl3\nl4\nl5\nl6\nl7\nl8\nl9\nl10\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "base")

	// change line 1 (top) and line 10 (bottom) → two separate hunks
	writeFile(t, dir, "f.txt", "CHANGED1\nl2\nl3\nl4\nl5\nl6\nl7\nl8\nl9\nCHANGED10\n")

	// parse the unstaged diff
	files, err := Diff(dir, DiffOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var fd *FileDiff
	for i := range files {
		if files[i].NewPath == "f.txt" || files[i].OldPath == "f.txt" {
			fd = &files[i]
		}
	}
	if fd == nil {
		t.Fatal("no diff for f.txt")
	}
	if len(fd.Hunks) < 2 {
		t.Fatalf("expected 2 hunks (top + bottom), got %d", len(fd.Hunks))
	}

	// stage ONLY the first hunk (the top change)
	if err := StageHunk(dir, "f.txt", fd.Hunks[0]); err != nil {
		t.Fatalf("StageHunk: %v", err)
	}

	// the staged diff should contain CHANGED1 but NOT CHANGED10
	staged, err := Diff(dir, DiffOptions{Staged: true})
	if err != nil {
		t.Fatal(err)
	}
	var stagedText strings.Builder
	for _, f := range staged {
		for _, h := range f.Hunks {
			for _, ln := range h.Lines {
				stagedText.WriteString(ln.Content + "\n")
			}
		}
	}
	got := stagedText.String()
	if !strings.Contains(got, "CHANGED1") {
		t.Errorf("expected CHANGED1 staged, staged diff:\n%s", got)
	}
	if strings.Contains(got, "CHANGED10") {
		t.Errorf("did NOT expect CHANGED10 staged (only hunk 0 was staged):\n%s", got)
	}

	// now unstage that hunk → nothing staged
	if err := UnstageHunk(dir, "f.txt", fd.Hunks[0]); err != nil {
		t.Fatalf("UnstageHunk: %v", err)
	}
	staged2, _ := Diff(dir, DiffOptions{Staged: true})
	if len(staged2) != 0 {
		t.Errorf("expected nothing staged after unstage, got %+v", staged2)
	}
}

func TestBuildHunkPatch(t *testing.T) {
	h := Hunk{
		Header: "@@ -1,3 +1,3 @@",
		Lines: []Line{
			{Kind: LineContext, Content: "ctx"},
			{Kind: LineRemove, Content: "old"},
			{Kind: LineAdd, Content: "new"},
		},
	}
	patch := buildHunkPatch("f.txt", h)
	// must have file headers, the hunk header, and correctly-prefixed lines
	for _, want := range []string{
		"diff --git a/f.txt b/f.txt",
		"--- a/f.txt",
		"+++ b/f.txt",
		"@@ -1,3 +1,3 @@",
		" ctx",
		"-old",
		"+new",
	} {
		if !strings.Contains(patch, want) {
			t.Errorf("patch missing %q:\n%s", want, patch)
		}
	}
}
