package gitops

import (
	"os"
	"path/filepath"
	"strings"
)

// ConflictHunk is one <<<<<<< ======= >>>>>>> region within a conflicted file.
type ConflictHunk struct {
	Ours   []string // lines from our side (current branch, HEAD)
	Theirs []string // lines from their side (the merged-in branch)
}

// ConflictFile is a conflicted file's regions plus the raw marked-up content.
type ConflictFile struct {
	Path  string
	Hunks []ConflictHunk
	Raw   string // full file including markers, for hand-editing
}

// readRepoFile / writeRepoFile do plain file I/O relative to the repo dir.
func readRepoFile(dir, path string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeRepoFile(dir, path, content string) error {
	return os.WriteFile(filepath.Join(dir, filepath.FromSlash(path)), []byte(content), 0o644)
}

// ReadConflict parses the conflict regions of a single conflicted file.
func ReadConflict(dir, path string) (*ConflictFile, error) {
	content, err := readRepoFile(dir, path)
	if err != nil {
		return nil, err
	}
	return &ConflictFile{Path: path, Raw: content, Hunks: parseConflictHunks(content)}, nil
}

func parseConflictHunks(content string) []ConflictHunk {
	var hunks []ConflictHunk
	var cur *ConflictHunk
	side := 0 // 1 = ours, 2 = theirs
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

// ResolveOurs resolves a file by taking our side for every conflict region.
func ResolveOurs(dir, path string) error { return resolveSide(dir, path, true) }

// ResolveTheirs resolves a file by taking their side for every conflict region.
func ResolveTheirs(dir, path string) error { return resolveSide(dir, path, false) }

func resolveSide(dir, path string, ours bool) error {
	content, err := readRepoFile(dir, path)
	if err != nil {
		return err
	}
	if err := writeRepoFile(dir, path, applySide(content, ours)); err != nil {
		return err
	}
	_, err = run(dir, "add", "--", path) // stage → mark resolved
	return err
}

func applySide(content string, ours bool) string {
	var out []string
	side := 0
	for _, ln := range strings.Split(content, "\n") {
		switch {
		case strings.HasPrefix(ln, "<<<<<<<"):
			side = 1
		case strings.HasPrefix(ln, "======="):
			side = 2
		case strings.HasPrefix(ln, ">>>>>>>"):
			side = 0
		default:
			if side == 0 || (side == 1 && ours) || (side == 2 && !ours) {
				out = append(out, ln)
			}
		}
	}
	return strings.Join(out, "\n")
}

// MarkResolved stages a (hand-edited) conflicted file, marking it resolved.
func MarkResolved(dir, path string) error {
	_, err := run(dir, "add", "--", path)
	return err
}
