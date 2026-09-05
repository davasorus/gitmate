package gitops

import (
	"strconv"
	"strings"
)

// Stash is one entry from `git stash list`.
type Stash struct {
	Ref     string // e.g. "stash@{0}"
	Index   int    // 0-based position
	Branch  string // branch the stash was made on (best-effort parse)
	Message string // the stash description
}

// StashSave stashes tracked working-tree and index changes. If message is
// non-empty it's used as the stash description. includeUntracked also stashes
// untracked files (git stash push -u). Returns nil even if there was nothing
// to stash — callers can re-check status.
func StashSave(dir, message string, includeUntracked bool) error {
	args := []string{"stash", "push"}
	if includeUntracked {
		args = append(args, "-u")
	}
	if strings.TrimSpace(message) != "" {
		args = append(args, "-m", message)
	}
	_, err := run(dir, args...)
	return err
}

// StashList returns the stash entries, newest first (stash@{0} is newest).
func StashList(dir string) ([]Stash, error) {
	// Format: "stash@{0}: On branch: message" or "stash@{0}: WIP on branch: ..."
	out, err := run(dir, "stash", "list")
	if err != nil {
		return nil, err
	}
	var stashes []Stash
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		st := Stash{}
		// split off the "stash@{N}" ref before the first ": "
		if i := strings.Index(line, ": "); i >= 0 {
			st.Ref = line[:i]
			rest := line[i+2:]
			st.Index = parseStashIndex(st.Ref)
			// rest is like "On main: msg" or "WIP on main: msg"
			if j := strings.Index(rest, ": "); j >= 0 {
				head := rest[:j]
				st.Message = rest[j+2:]
				head = strings.TrimPrefix(head, "WIP on ")
				head = strings.TrimPrefix(head, "On ")
				st.Branch = head
			} else {
				st.Message = rest
			}
		} else {
			st.Ref = line
		}
		stashes = append(stashes, st)
	}
	return stashes, nil
}

// StashPop applies the stash at the given ref and removes it from the stash
// list. An empty ref pops the most recent (stash@{0}).
func StashPop(dir, ref string) error {
	args := []string{"stash", "pop"}
	if ref != "" {
		args = append(args, ref)
	}
	_, err := run(dir, args...)
	return err
}

// StashDrop discards a stash entry without applying it. An empty ref drops the
// most recent. Destructive — the stashed changes are lost.
func StashDrop(dir, ref string) error {
	args := []string{"stash", "drop"}
	if ref != "" {
		args = append(args, ref)
	}
	_, err := run(dir, args...)
	return err
}

// parseStashIndex pulls N out of "stash@{N}".
func parseStashIndex(ref string) int {
	l := strings.Index(ref, "{")
	r := strings.Index(ref, "}")
	if l >= 0 && r > l {
		if n, err := strconv.Atoi(ref[l+1 : r]); err == nil {
			return n
		}
	}
	return 0
}
