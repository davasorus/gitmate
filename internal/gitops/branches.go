package gitops

import (
	"strconv"
	"strings"
	"time"
)

// Branch is one parsed local branch.
type Branch struct {
	Name        string
	IsCurrent   bool
	Upstream    string
	Ahead       int
	Behind      int
	LastHash    string
	LastSubject string
	LastWhen    time.Time
}

// branchFormat: null-separated fields, one line per ref.
//
//	%(HEAD)              "*" if this is the current branch, else " "
//	%(refname:short)     branch name without refs/heads/
//	%(upstream:short)    tracking branch, empty if none
//	%(upstream:track,nobracket) e.g. "ahead 2, behind 1" (no [] wrapper)
//	%(objectname:short)  short hash of the branch tip
//	%(committerdate:unix) tip commit time as epoch
//	%(contents:subject)  tip commit subject
const branchFormat = "%(HEAD)%00%(refname:short)%00%(upstream:short)%00" +
	"%(upstream:track,nobracket)%00%(objectname:short)%00" +
	"%(committerdate:unix)%00%(contents:subject)"

// GetBranches lists local branches sorted by most-recent commit first.
func GetBranches(dir string) ([]Branch, error) {
	out, err := run(dir,
		"for-each-ref",
		"--sort=-committerdate",
		"--format="+branchFormat,
		"refs/heads/",
	)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.Split(line, "\x00")
		if len(f) != 7 {
			continue
		}
		b := Branch{
			IsCurrent:   strings.TrimSpace(f[0]) == "*",
			Name:        f[1],
			Upstream:    f[2],
			LastHash:    f[4],
			LastSubject: f[6],
		}
		parseTrack(&b, f[3])
		if ts, err := strconv.ParseInt(strings.TrimSpace(f[5]), 10, 64); err == nil {
			b.LastWhen = time.Unix(ts, 0)
		}
		branches = append(branches, b)
	}
	return branches, nil
}

// parseTrack reads "ahead 2, behind 1" / "ahead 3" / "behind 4" / "gone".
func parseTrack(b *Branch, s string) {
	s = strings.TrimSpace(s)
	if s == "" || s == "gone" {
		return
	}
	for _, part := range strings.Split(s, ",") {
		fields := strings.Fields(part)
		if len(fields) != 2 {
			continue
		}
		n, _ := strconv.Atoi(fields[1])
		switch fields[0] {
		case "ahead":
			b.Ahead = n
		case "behind":
			b.Behind = n
		}
	}
}
