package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLastCommitSubject(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	cur, _ := CurrentBranch(dir)
	sub, err := LastCommitSubject(dir, cur)
	if err != nil {
		t.Fatal(err)
	}
	if sub != "init" {
		t.Fatalf("got %q", sub)
	}
}

func TestReadPRTemplate(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	_ = ReadPRTemplate(dir) // no template → empty, no panic
	if err := os.MkdirAll(filepath.Join(dir, ".github"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, ".github/pull_request_template.md", "## Summary\n")
	if tmpl := ReadPRTemplate(dir); !strings.Contains(tmpl, "Summary") {
		t.Fatalf("template not read: %q", tmpl)
	}
}
