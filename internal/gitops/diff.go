package gitops

import "strings"

// LineKind classifies a diff line.
type LineKind string

const (
	LineContext LineKind = "context"
	LineAdd     LineKind = "add"
	LineRemove  LineKind = "remove"
)

// Line is one line within a hunk.
type Line struct {
	Kind    LineKind
	Content string
	OldNum  int
	NewNum  int
}

// Hunk is one @@ ... @@ block.
type Hunk struct {
	Header string
	Lines  []Line
}

// FileDiff is the diff for a single file.
type FileDiff struct {
	OldPath string
	NewPath string
	Binary  bool
	Hunks   []Hunk
}

// DiffOptions selects what to diff.
type DiffOptions struct {
	Staged bool
	Rev    string
	Path   string
}

// Diff returns structured diffs for the given options.
func Diff(dir string, opts DiffOptions) ([]FileDiff, error) {
	args := []string{"diff"}
	if opts.Rev != "" {
		args = append(args, opts.Rev)
	} else if opts.Staged {
		args = append(args, "--cached")
	}
	if opts.Path != "" {
		args = append(args, "--", opts.Path)
	}
	out, err := run(dir, args...)
	if err != nil {
		return nil, err
	}
	return ParseUnifiedDiff(out), nil
}

// ParseUnifiedDiff turns `git diff` output into []FileDiff.
func ParseUnifiedDiff(out string) []FileDiff {
	var files []FileDiff
	var cur *FileDiff
	var hunk *Hunk
	oldNum, newNum := 0, 0

	flushHunk := func() {
		if cur != nil && hunk != nil {
			cur.Hunks = append(cur.Hunks, *hunk)
			hunk = nil
		}
	}
	flushFile := func() {
		flushHunk()
		if cur != nil {
			files = append(files, *cur)
			cur = nil
		}
	}

	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flushFile()
			cur = &FileDiff{}
		case strings.HasPrefix(line, "Binary files "):
			if cur != nil {
				cur.Binary = true
			}
		case strings.HasPrefix(line, "--- "):
			if cur != nil {
				cur.OldPath = strings.TrimPrefix(strings.TrimPrefix(line, "--- "), "a/")
			}
		case strings.HasPrefix(line, "+++ "):
			if cur != nil {
				cur.NewPath = strings.TrimPrefix(strings.TrimPrefix(line, "+++ "), "b/")
			}
		case strings.HasPrefix(line, "@@"):
			flushHunk()
			hunk = &Hunk{Header: line}
			oldNum, newNum = hunkStart(line)
		case hunk != nil && strings.HasPrefix(line, "+"):
			hunk.Lines = append(hunk.Lines, Line{Kind: LineAdd, Content: line[1:], NewNum: newNum})
			newNum++
		case hunk != nil && strings.HasPrefix(line, "-"):
			hunk.Lines = append(hunk.Lines, Line{Kind: LineRemove, Content: line[1:], OldNum: oldNum})
			oldNum++
		case hunk != nil && strings.HasPrefix(line, " "):
			hunk.Lines = append(hunk.Lines, Line{Kind: LineContext, Content: line[1:], OldNum: oldNum, NewNum: newNum})
			oldNum++
			newNum++
		}
	}
	flushFile()
	return files
}

// hunkStart parses the starting old/new line numbers from a hunk header.
func hunkStart(header string) (oldStart, newStart int) {
	fields := strings.Fields(header)
	for _, f := range fields {
		if strings.HasPrefix(f, "-") {
			oldStart = leadingInt(strings.TrimPrefix(f, "-"))
		} else if strings.HasPrefix(f, "+") {
			newStart = leadingInt(strings.TrimPrefix(f, "+"))
		}
	}
	return oldStart, newStart
}

// leadingInt reads the integer before an optional ",count" suffix.
func leadingInt(s string) int {
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = s[:i]
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
