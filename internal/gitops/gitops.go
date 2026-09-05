package gitops

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FileChange is one entry from git status.
type FileChange struct {
	Staged   string // human-readable index (staged) state
	Unstaged string // human-readable working-tree state
	Path     string
	OrigPath string // set for renames/copies
}

// Status is the parsed result of `git status --porcelain=v2 --branch`.
type Status struct {
	Branch    string
	Upstream  string
	Ahead     int
	Behind    int
	Detached  bool
	Changes   []FileChange
	Untracked []string
}

// run executes git in the given dir and returns stdout, or an error
// that includes stderr so failures are legible.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

// GetStatus runs status --porcelain=v2 --branch and parses the output.
func GetStatus(dir string) (*Status, error) {
	out, err := run(dir, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return nil, err
	}

	s := &Status{}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		switch line[0] {
		case '#':
			parseBranchHeader(s, line)
		case '1':
			// Ordinary change: 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
			f := strings.SplitN(line, " ", 9)
			if len(f) == 9 {
				s.Changes = append(s.Changes, decodeXY(f[1], f[8], ""))
			}
		case '2':
			// Rename/copy: ... <Xscore> <path>\t<origPath>
			f := strings.SplitN(line, " ", 10)
			if len(f) == 10 {
				parts := strings.SplitN(f[9], "\t", 2)
				orig := ""
				if len(parts) == 2 {
					orig = parts[1]
				}
				s.Changes = append(s.Changes, decodeXY(f[1], parts[0], orig))
			}
		case '?':
			// Untracked: ? <path>
			f := strings.SplitN(line, " ", 2)
			if len(f) == 2 {
				s.Untracked = append(s.Untracked, f[1])
			}
		case 'u':
			// Unmerged (conflict): u <XY> ... <path>
			f := strings.SplitN(line, " ", 11)
			if len(f) == 11 {
				s.Changes = append(s.Changes, FileChange{
					Staged: "conflict", Unstaged: "conflict", Path: f[10],
				})
			}
		}
	}
	return s, sc.Err()
}

func parseBranchHeader(s *Status, line string) {
	// e.g. "# branch.head main", "# branch.ab +2 -1"
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return
	}
	switch fields[1] {
	case "branch.head":
		if fields[2] == "(detached)" {
			s.Detached = true
		} else {
			s.Branch = fields[2]
		}
	case "branch.upstream":
		s.Upstream = fields[2]
	case "branch.ab":
		// fields[2] = "+N", fields[3] = "-M"
		s.Ahead, _ = strconv.Atoi(strings.TrimPrefix(fields[2], "+"))
		if len(fields) > 3 {
			s.Behind, _ = strconv.Atoi(strings.TrimPrefix(fields[3], "-"))
		}
	}
}

// decodeXY turns the two-char XY code into readable states.
// X = index/staged column, Y = working-tree/unstaged column.
func decodeXY(xy, path, orig string) FileChange {
	fc := FileChange{Path: path, OrigPath: orig}
	if len(xy) == 2 {
		fc.Staged = codeName(xy[0])
		fc.Unstaged = codeName(xy[1])
	}
	return fc
}

func codeName(c byte) string {
	switch c {
	case '.':
		return ""
	case 'M':
		return "modified"
	case 'A':
		return "added"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	case 'C':
		return "copied"
	case 'U':
		return "unmerged"
	default:
		return string(c)
	}
}

// GetRemoteURL returns the URL of the named remote (e.g. "origin").
func GetRemoteURL(dir, name string) (string, error) {
	out, err := run(dir, "remote", "get-url", name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// AddRemote adds a named remote pointing at url.
func AddRemote(dir, name, url string) error {
	_, err := run(dir, "remote", "add", name, url)
	return err
}
