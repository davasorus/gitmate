package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RebaseAction is what to do with a commit in an interactive rebase.
type RebaseAction string

// Interactive-rebase actions.
const (
	RebasePick   RebaseAction = "pick"
	RebaseReword RebaseAction = "reword"
	RebaseSquash RebaseAction = "squash"
	RebaseFixup  RebaseAction = "fixup"
	RebaseDrop   RebaseAction = "drop"
)

// RebaseStep is one line of the interactive-rebase plan: a commit + the action
// to take, plus (for reword/squash) an optional replacement message.
type RebaseStep struct {
	SHA     string       // full or short commit SHA
	Subject string       // commit subject (for display)
	Action  RebaseAction // pick/reword/squash/fixup/drop
	Message string       // replacement/combined message for reword & squash; empty otherwise
}

// InteractiveRebaseTodo returns the commits from base..HEAD (oldest first, the
// order a rebase todo uses) as a default plan of "pick" steps, ready to be
// reordered/edited by the caller.
func InteractiveRebaseTodo(dir, base string) ([]RebaseStep, error) {
	// oldest→newest, one "sha<TAB>subject" per line
	out, err := run(dir, "log", "--reverse", "--format=%H%x09%s", base+"..HEAD")
	if err != nil {
		return nil, err
	}
	var steps []RebaseStep
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimRight(ln, "\r")
		if ln == "" {
			continue
		}
		parts := strings.SplitN(ln, "\t", 2)
		sha := parts[0]
		subj := ""
		if len(parts) == 2 {
			subj = parts[1]
		}
		steps = append(steps, RebaseStep{SHA: sha, Subject: subj, Action: RebasePick})
	}
	return steps, nil
}

// buildTodo renders the rebase todo file text from a plan. Dropped steps are
// omitted. squash/fixup fold into the preceding kept commit (git requires the
// first line to be pick/reword, never squash/fixup).
func buildTodo(steps []RebaseStep) string {
	var b strings.Builder
	for _, s := range steps {
		if s.Action == RebaseDrop {
			continue
		}
		// use short-form action + sha; subject is a comment git ignores
		b.WriteString(string(s.Action) + " " + s.SHA + " " + s.Subject + "\n")
	}
	return b.String()
}

// hasMessages reports whether any step supplies a replacement message (reword
// or squash), which means we also need to drive the commit-message editor.
func hasMessages(steps []RebaseStep) bool {
	for _, s := range steps {
		if (s.Action == RebaseReword || s.Action == RebaseSquash) && strings.TrimSpace(s.Message) != "" {
			return true
		}
	}
	return false
}

// seqEditorCmd builds the GIT_SEQUENCE_EDITOR command that overwrites git's todo
// with the prepared file at todoPath. It's a package var so tests can substitute
// a shell copy (cp) instead of the self-invoke helper, letting RunInteractiveRebase
// run end-to-end under test. Production default: this program's own helper.
var seqEditorCmd = func(todoPath string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot locate self for rebase editor: %w", err)
	}
	return quoteArg(self) + " " + seqEditorSubcmd + " " + quoteArg(todoPath), nil
}

// msgEditorCmd builds the GIT_EDITOR command that feeds reword/squash messages.
// Also a package var for test substitution.
var msgEditorCmd = func(msgPath string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	return quoteArg(self) + " " + msgEditorSubcmd + " " + quoteArg(msgPath), nil
}

// RunInteractiveRebase performs an interactive rebase onto base, applying the
// given plan (reorder/drop/squash/fixup/reword) by driving git non-interactively
// via the injectable sequence/message editors.
func RunInteractiveRebase(dir, base string, steps []RebaseStep) error {
	// write our todo to a temp file the sequence editor will copy over git's todo
	todoFile, err := os.CreateTemp("", "gitmate-rebase-todo-*")
	if err != nil {
		return err
	}
	todoPath := todoFile.Name()
	defer func() { _ = os.Remove(todoPath) }()
	if _, err := todoFile.WriteString(buildTodo(steps)); err != nil {
		_ = todoFile.Close()
		return err
	}
	_ = todoFile.Close()

	seqEd, err := seqEditorCmd(todoPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "rebase", "-i", "--autostash", base)
	cmd.Dir = dir
	hideWindow(cmd)
	env := os.Environ()
	env = append(env, "GIT_SEQUENCE_EDITOR="+seqEd)

	// messages for reword/squash, if any
	if hasMessages(steps) {
		mp, cleanup, mErr := writeMessagesFile(steps)
		if mErr != nil {
			return mErr
		}
		defer cleanup()
		msgEd, mErr := msgEditorCmd(mp)
		if mErr != nil {
			return mErr
		}
		env = append(env, "GIT_EDITOR="+msgEd)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("interactive rebase: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// quoteArg wraps a path in double quotes for the GIT_*_EDITOR shell string,
// escaping embedded quotes. Works for POSIX sh and Windows cmd.
func quoteArg(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
