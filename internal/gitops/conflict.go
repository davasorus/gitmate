package gitops

import (
	"os"
	"path/filepath"
	"strings"
)

// ConflictHunk is one conflict region: the "ours" and "theirs" content.
type ConflictHunk struct {
	Ours   []string // stage 2 — current branch (HEAD)
	Theirs []string // stage 3 — the merged-in branch
}

// ConflictFile is a conflicted file's regions plus the raw marked-up content.
type ConflictFile struct {
	Path  string
	Hunks []ConflictHunk
	Raw   string
}

// repoFilePath resolves a repo-relative path to an absolute one under dir,
// so os file I/O works regardless of the process's working directory.
func repoFilePath(dir, path string) (string, error) {
	// Ask git for the repo's top level so writes land in the right place even
	// when dir is "." and the process cwd differs (e.g. the Wails GUI runs from gui/).
	top, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		// fall back to joining dir directly
		return filepath.Join(dir, filepath.FromSlash(path)), nil
	}
	return filepath.Join(strings.TrimSpace(top), filepath.FromSlash(path)), nil
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// ReadConflict returns the conflict for a file, reading the authoritative
// "ours" (index stage 2) and "theirs" (stage 3) blobs via git — reliable
// regardless of marker formatting or the process working directory.
func ReadConflict(dir, path string) (*ConflictFile, error) {
	cf := &ConflictFile{Path: path}

	// raw working-tree file (with markers) for the hand-edit fallback
	if abs, err := repoFilePath(dir, path); err == nil {
		if b, rerr := os.ReadFile(abs); rerr == nil {
			cf.Raw = string(b)
		}
	}

	ours, oerr := run(dir, "show", ":2:"+path)   // ours = HEAD
	theirs, terr := run(dir, "show", ":3:"+path) // theirs = merged-in
	if oerr == nil || terr == nil {
		cf.Hunks = []ConflictHunk{{Ours: splitLines(ours), Theirs: splitLines(theirs)}}
		return cf, nil
	}
	// fall back to parsing markers from the raw content
	cf.Hunks = parseConflictHunks(cf.Raw)
	return cf, nil
}

// parseConflictHunks extracts ours/theirs from <<<<<<< ======= >>>>>>> markers.
func parseConflictHunks(content string) []ConflictHunk {
	var hunks []ConflictHunk
	var cur *ConflictHunk
	side := 0
	for _, ln := range strings.Split(content, "\n") {
		switch {
		case strings.HasPrefix(ln, "<<<<<<<"):
			cur = &ConflictHunk{}
			side = 1
		case strings.HasPrefix(ln, "=======") && cur != nil:
			side = 2
		case strings.HasPrefix(ln, ">>>>>>>") && cur != nil:
			hunks = append(hunks, *cur)
			cur = nil
			side = 0
		default:
			if cur != nil && side == 1 {
				cur.Ours = append(cur.Ours, ln)
			} else if cur != nil && side == 2 {
				cur.Theirs = append(cur.Theirs, ln)
			}
		}
	}
	return hunks
}

// repoRoot returns the repository's top-level directory, so path-based git
// commands resolve correctly even when the process cwd differs (the Wails GUI
// runs from gui/, not the repo root).
func repoRoot(dir string) string {
	top, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return dir
	}
	return strings.TrimSpace(top)
}

// ResolveOurs resolves a file by taking our side entirely (git checkout --ours).
func ResolveOurs(dir, path string) error {
	root := repoRoot(dir)
	if _, err := run(root, "checkout", "--ours", "--", path); err != nil {
		return err
	}
	_, err := run(root, "add", "--", path)
	return err
}

// ResolveTheirs resolves a file by taking their side entirely (git checkout --theirs).
func ResolveTheirs(dir, path string) error {
	root := repoRoot(dir)
	if _, err := run(root, "checkout", "--theirs", "--", path); err != nil {
		return err
	}
	_, err := run(root, "add", "--", path)
	return err
}

// MarkResolved stages a (hand-edited) conflicted file, marking it resolved.
func MarkResolved(dir, path string) error {
	root := repoRoot(dir)
	_, err := run(root, "add", "--", path)
	return err
}
