package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCloneDeriveDest(t *testing.T) {
	// make a bare "remote" with a commit to clone from
	src := newRemoteRepo(t) // working repo pushed to a bare origin
	origin, err := GetRemoteURL(src, "origin")
	if err != nil {
		t.Fatal(err)
	}

	// clone into a temp cwd with empty dest → Clone derives the name from the url
	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	dest, err := Clone(origin, "")
	if err != nil {
		t.Fatal(err)
	}
	if dest == "" {
		t.Fatal("expected a derived dest path")
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err != nil {
		t.Fatalf("cloned repo missing .git: %v", err)
	}
}
