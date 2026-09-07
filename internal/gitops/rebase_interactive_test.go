package gitops

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// TestBuildTodo verifies the todo text: drops omitted, actions + shas emitted in order.
func TestBuildTodo(t *testing.T) {
	steps := []RebaseStep{
		{SHA: "aaa", Subject: "first", Action: RebasePick},
		{SHA: "bbb", Subject: "second", Action: RebaseDrop},
		{SHA: "ccc", Subject: "third", Action: RebaseFixup},
	}
	got := buildTodo(steps)
	if strings.Contains(got, "bbb") {
		t.Errorf("dropped commit should not appear: %q", got)
	}
	if !strings.Contains(got, "pick aaa first") || !strings.Contains(got, "fixup ccc third") {
		t.Errorf("todo wrong: %q", got)
	}
	// order preserved
	if strings.Index(got, "aaa") > strings.Index(got, "ccc") {
		t.Errorf("order not preserved: %q", got)
	}
}

// TestInteractiveRebaseTodo builds a repo with 3 commits and confirms the plan
// lists them oldest-first as picks.
func TestInteractiveRebaseTodo(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "1\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "c1")
	base, _ := HeadSHA(dir) // rebase onto this; commits after are the todo
	writeFile(t, dir, "a.txt", "2\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "c2")
	writeFile(t, dir, "a.txt", "3\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "c3")

	steps, err := InteractiveRebaseTodo(dir, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps (c2,c3), got %d: %+v", len(steps), steps)
	}
	if steps[0].Subject != "c2" || steps[1].Subject != "c3" {
		t.Errorf("expected oldest-first c2,c3: %+v", steps)
	}
}

// TestRunInteractiveRebase_Drop runs a REAL rebase that drops a commit, using a
// direct cp/copy-free sequence editor: we bypass self-invoke by pointing
// GIT_SEQUENCE_EDITOR at our own copy via a helper the test provides. Since a
// unit test's os.Executable() isn't gitmate, we invoke the helper logic through
// a tiny env trick: write the todo and use a POSIX/py-free approach — run the
// rebase with GIT_SEQUENCE_EDITOR set to overwrite via `git` itself is not
// possible, so we exercise the copy helper directly instead.
func TestRunInteractiveRebase_Drop(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "1\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "keep-1")
	base, _ := HeadSHA(dir)
	writeFile(t, dir, "b.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "drop-me")
	writeFile(t, dir, "c.txt", "y\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "keep-2")

	steps, err := InteractiveRebaseTodo(dir, base)
	if err != nil {
		t.Fatal(err)
	}
	// drop the middle commit ("drop-me")
	for i := range steps {
		if steps[i].Subject == "drop-me" {
			steps[i].Action = RebaseDrop
		}
	}

	// Drive the rebase with a sequence editor that copies our todo. In the real
	// app this is the self-invoked helper; in the test we emulate it with a
	// small copy using the OS's cp/copy through GIT_SEQUENCE_EDITOR set to a Go
	// helper binary is not available, so we test buildTodo + apply via a manual
	// GIT_SEQUENCE_EDITOR that uses `cp` on unix / `copy` on windows.
	todo := buildTodo(steps)
	tf, _ := os.CreateTemp("", "todo-*")
	_, _ = tf.WriteString(todo)
	tf.Close()
	defer os.Remove(tf.Name())

	editor := seqEditorCmdForTest(tf.Name())
	cmd := exec.Command("git", "rebase", "-i", base)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_SEQUENCE_EDITOR="+editor)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("rebase failed: %s", string(out))
	}

	// after dropping "drop-me", the log should have keep-1 + keep-2 only
	log, _ := GetLog(dir, 10)
	for _, c := range log {
		if c.Subject == "drop-me" {
			t.Fatalf("drop-me should have been dropped; log: %+v", log)
		}
	}
	if len(log) != 2 {
		t.Fatalf("expected 2 commits after drop, got %d: %+v", len(log), log)
	}
}

// seqEditorCmdForTest returns a GIT_SEQUENCE_EDITOR command that overwrites
// git's todo file (passed as the trailing arg) with the prepared todo at src.
// Test-only: the real app uses the self-invoked helper (MaybeRunRebaseEditor)
// instead, but a unit test's os.Executable() is the test binary, so we emulate
// the copy with the OS's native command.
func seqEditorCmdForTest(src string) string {
	// git invokes GIT_SEQUENCE_EDITOR through its bundled sh on every OS, so cp
	// is available and paths must use forward slashes (Windows backslashes get
	// mangled by the shell).
	if runtime.GOOS == "windows" {
		src = strings.ReplaceAll(src, `\`, `/`)
	}
	return `cp "` + src + `"`
}

func TestHasMessagesAndQuoteArg(t *testing.T) {
	// hasMessages: true only when a reword/squash carries a non-empty message
	none := []RebaseStep{{Action: RebasePick}, {Action: RebaseDrop}}
	if hasMessages(none) {
		t.Error("pick/drop should need no messages")
	}
	with := []RebaseStep{{Action: RebaseReword, Message: "new msg"}}
	if !hasMessages(with) {
		t.Error("reword with a message should report true")
	}
	emptyMsg := []RebaseStep{{Action: RebaseSquash, Message: "  "}}
	if hasMessages(emptyMsg) {
		t.Error("blank message should not count")
	}
	// quoteArg always wraps in quotes
	q := quoteArg("some path")
	if q[0] != '"' || q[len(q)-1] != '"' {
		t.Errorf("quoteArg should wrap in quotes: %q", q)
	}
}

func TestRunInteractiveRebase_EndToEnd(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	// Substitute the sequence editor with a cp (forward-slashed on Windows) so
	// RunInteractiveRebase's full body runs without needing the gitmate binary.
	origSeq := seqEditorCmd
	seqEditorCmd = func(todoPath string) (string, error) {
		p := todoPath
		if runtime.GOOS == "windows" {
			p = strings.ReplaceAll(p, `\`, `/`)
		}
		return `cp "` + p + `"`, nil
	}
	defer func() { seqEditorCmd = origSeq }()

	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "1\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "keep-1")
	base, _ := HeadSHA(dir)
	writeFile(t, dir, "b.txt", "x\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "drop-me")
	writeFile(t, dir, "c.txt", "y\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "keep-2")

	steps, err := InteractiveRebaseTodo(dir, base)
	if err != nil {
		t.Fatal(err)
	}
	for i := range steps {
		if steps[i].Subject == "drop-me" {
			steps[i].Action = RebaseDrop
		}
	}

	if err := RunInteractiveRebase(dir, base, steps); err != nil {
		t.Fatalf("RunInteractiveRebase: %v", err)
	}
	log, _ := GetLog(dir, 10)
	for _, c := range log {
		if c.Subject == "drop-me" {
			t.Fatalf("drop-me should be gone; log %+v", log)
		}
	}
	if len(log) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(log))
	}
}
