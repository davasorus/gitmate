package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeRunRebaseEditor_NotHandled(t *testing.T) {
	// no special subcommand → not handled
	if handled, _ := MaybeRunRebaseEditor([]string{"gitmate"}); handled {
		t.Error("bare invocation should not be handled")
	}
	if handled, _ := MaybeRunRebaseEditor([]string{"gitmate", "status"}); handled {
		t.Error("normal subcommand should not be handled")
	}
}

func TestMaybeRunRebaseEditor_SeqEditor(t *testing.T) {
	dir := t.TempDir()
	prepared := filepath.Join(dir, "prepared")
	gitTodo := filepath.Join(dir, "git-todo")
	if err := os.WriteFile(prepared, []byte("pick abc subject\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gitTodo, []byte("pick DEFAULT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// git invokes: <self> __rebase-seq-editor <prepared> <gitTodo>
	handled, code := MaybeRunRebaseEditor([]string{"gitmate", seqEditorSubcmd, prepared, gitTodo})
	if !handled || code != 0 {
		t.Fatalf("seq editor should handle + succeed, got handled=%v code=%d", handled, code)
	}
	got, _ := os.ReadFile(gitTodo)
	if !strings.Contains(string(got), "pick abc subject") {
		t.Errorf("git todo should be overwritten with prepared content, got %q", got)
	}
}

func TestMaybeRunRebaseEditor_MsgEditor(t *testing.T) {
	dir := t.TempDir()
	// two messages separated by the marker
	msgs := filepath.Join(dir, "msgs")
	if err := os.WriteFile(msgs, []byte("first msg"+"\n"+msgSep+"\n"+"second msg"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitMsg := filepath.Join(dir, "COMMIT_EDITMSG")
	_ = os.WriteFile(commitMsg, []byte("default"), 0o644)

	// first invocation pops "first msg"
	handled, code := MaybeRunRebaseEditor([]string{"gitmate", msgEditorSubcmd, msgs, commitMsg})
	if !handled || code != 0 {
		t.Fatalf("msg editor should handle+succeed")
	}
	got, _ := os.ReadFile(commitMsg)
	if string(got) != "first msg" {
		t.Errorf("expected first msg written, got %q", got)
	}
	// second invocation pops "second msg"
	MaybeRunRebaseEditor([]string{"gitmate", msgEditorSubcmd, msgs, commitMsg})
	got2, _ := os.ReadFile(commitMsg)
	if string(got2) != "second msg" {
		t.Errorf("expected second msg written, got %q", got2)
	}
}

func TestWriteMessagesFile(t *testing.T) {
	steps := []RebaseStep{
		{Action: RebaseReword, Message: "m1"},
		{Action: RebasePick},
		{Action: RebaseSquash, Message: "m2"},
	}
	path, cleanup, err := writeMessagesFile(steps)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "m1") || !strings.Contains(string(data), "m2") {
		t.Errorf("messages file should hold m1 and m2: %q", data)
	}
}
