package gitops

import (
	"strconv"
	"strings"
	"time"
)

// Commit is one parsed log entry.
type Commit struct {
	Hash    string
	Short   string
	Author  string
	Email   string
	When    time.Time
	Subject string
}

// logFormat maps to --pretty placeholders, null-separated:
//
//	%H  full hash
//	%h  short hash
//	%an author name
//	%ae author email
//	%at author date, UNIX timestamp
//	%s  subject
//
// %x1e (record separator) ends each commit so we can split cleanly.
const logFormat = "%H%x00%h%x00%an%x00%ae%x00%at%x00%s%x1e"

// GetLog runs git log with a fixed format and parses up to `limit` commits.
// GetLog logs the current branch (HEAD).
func GetLog(dir string, limit int) ([]Commit, error) {
	return GetLogRef(dir, "", limit)
}

// GetLogRef logs a specific ref (branch, tag, commit-ish). Empty ref = HEAD.
func GetLogRef(dir, ref string, limit int) ([]Commit, error) {
	args := []string{"log", "--pretty=format:" + logFormat}
	if limit > 0 {
		args = append(args, "-n", strconv.Itoa(limit))
	}
	if ref != "" {
		args = append(args, ref)
	}
	out, err := run(dir, args...)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	// Records are separated by the RS byte (\x1e). Trailing empty split is skipped.
	for _, rec := range strings.Split(out, "\x1e") {
		rec = strings.TrimLeft(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(rec, "\x00")
		if len(f) != 6 {
			continue
		}
		ts, _ := strconv.ParseInt(f[4], 10, 64)
		commits = append(commits, Commit{
			Hash:    f[0],
			Short:   f[1],
			Author:  f[2],
			Email:   f[3],
			When:    time.Unix(ts, 0),
			Subject: f[5],
		})
	}
	return commits, nil
}
