package gitops

import (
	"os"
	"strings"
)

// Hidden subcommand names used when this program invokes ITSELF as git's
// sequence editor / commit-message editor during an interactive rebase.
const (
	seqEditorSubcmd = "__rebase-seq-editor"
	msgEditorSubcmd = "__rebase-msg-editor"
)

// MaybeRunRebaseEditor checks os.Args for the hidden rebase-editor subcommands
// and, if present, performs the edit and returns true (the caller should then
// exit). Both the CLI and GUI mains call this at the very top of main() so that
// when git invokes "<self> __rebase-seq-editor <prepared> <gitTodo>" the process
// does the copy and exits — no external script, identical on every OS.
//
// Returns (handled, exitCode). When handled is true, main must exit with
// exitCode. When false, main proceeds normally.
func MaybeRunRebaseEditor(args []string) (handled bool, exitCode int) {
	if len(args) < 2 {
		return false, 0
	}
	switch args[1] {
	case seqEditorSubcmd:
		// args: [self, subcmd, preparedTodo, gitTodoFile]
		if len(args) < 4 {
			return true, 1
		}
		if err := copyFile(args[2], args[3]); err != nil {
			return true, 1
		}
		return true, 0
	case msgEditorSubcmd:
		// args: [self, subcmd, messagesFile, gitCommitMsgFile]
		// Pop the next message from the messages file (one message per record,
		// separated by a NUL line marker) into the commit-message file git passed.
		if len(args) < 4 {
			return true, 1
		}
		if err := popMessage(args[2], args[3]); err != nil {
			return true, 1
		}
		return true, 0
	}
	return false, 0
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// messages are stored one-per-record separated by a line containing only the
// marker below. popMessage writes the first remaining record to dst and rewrites
// the source without it, so successive reword/squash editor invocations each get
// the next message in plan order.
const msgSep = "\x00---gitmate-msg---\x00"

// writeMessagesFile serializes reword/squash messages (in plan order) to a temp
// file for the message-editor helper to consume one at a time.
func writeMessagesFile(steps []RebaseStep) (path string, cleanup func(), err error) {
	var msgs []string
	for _, s := range steps {
		if (s.Action == RebaseReword || s.Action == RebaseSquash) && strings.TrimSpace(s.Message) != "" {
			msgs = append(msgs, s.Message)
		}
	}
	f, err := os.CreateTemp("", "gitmate-rebase-msgs-*")
	if err != nil {
		return "", func() {}, err
	}
	_, _ = f.WriteString(strings.Join(msgs, "\n"+msgSep+"\n"))
	_ = f.Close()
	return f.Name(), func() { _ = os.Remove(f.Name()) }, nil
}

func popMessage(msgsFile, dst string) error {
	data, err := os.ReadFile(msgsFile)
	if err != nil {
		return err
	}
	records := strings.Split(string(data), "\n"+msgSep+"\n")
	if len(records) == 0 {
		return os.WriteFile(dst, []byte{}, 0o644)
	}
	first := records[0]
	rest := records[1:]
	if err := os.WriteFile(dst, []byte(first), 0o644); err != nil {
		return err
	}
	// rewrite the messages file without the consumed record
	return os.WriteFile(msgsFile, []byte(strings.Join(rest, "\n"+msgSep+"\n")), 0o644)
}
