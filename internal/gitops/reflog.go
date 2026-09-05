package gitops

import "strings"

// ReflogEntry is one line of `git reflog` — where HEAD moved and why.
type ReflogEntry struct {
	Short    string // abbreviated commit hash HEAD pointed at
	Selector string // e.g. "HEAD@{0}"
	Action   string // commit, reset, checkout, merge, rebase, pull, ...
	Message  string // the rest of the description
}

// Reflog returns HEAD's reflog, newest first, up to limit entries (0 = all).
// This is the "undo safety net" — every position HEAD has held, so a bad reset
// or rebase can be traced back and recovered.
func Reflog(dir string, limit int) ([]ReflogEntry, error) {
	// Format: <short>\x00<selector>\x00<subject>
	// %gd = reflog selector (HEAD@{n}); %gs = reflog subject ("action: message")
	args := []string{"reflog", "--format=%h%x00%gd%x00%gs"}
	if limit > 0 {
		args = append(args, "-n", itoa(limit))
	}
	out, err := run(dir, args...)
	if err != nil {
		return nil, err
	}
	var entries []ReflogEntry
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.SplitN(line, "\x00", 3)
		if len(f) < 3 {
			continue
		}
		e := ReflogEntry{Short: f[0], Selector: f[1]}
		// subject is like "commit: message" or "reset: moving to HEAD~1" — split the action off
		subj := f[2]
		if i := strings.Index(subj, ": "); i >= 0 {
			e.Action = subj[:i]
			e.Message = subj[i+2:]
		} else {
			e.Message = subj
		}
		// action can have a subtype like "commit (amend)" — keep the first word for coloring
		entries = append(entries, e)
	}
	return entries, nil
}

// itoa avoids importing strconv just for one conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
