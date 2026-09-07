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
	IsLocal     bool   // has a local refs/heads/ ref
	IsRemote    bool   // exists on a remote (refs/remotes/)
	Remote      string // remote name for remote-only branches (e.g. "origin")
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
	// Local branches (refs/heads/).
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
	index := map[string]int{} // short branch name -> index in branches
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
			IsLocal:     true,
			Name:        f[1],
			Upstream:    f[2],
			LastHash:    f[4],
			LastSubject: f[6],
		}
		parseTrack(&b, f[3])
		if ts, err := strconv.ParseInt(strings.TrimSpace(f[5]), 10, 64); err == nil {
			b.LastWhen = time.Unix(ts, 0)
		}
		index[b.Name] = len(branches)
		branches = append(branches, b)
	}

	// Remote branches (refs/remotes/). Mark existing local ones as also-remote;
	// add remote-only branches as new entries the user can check out.
	rout, err := run(dir,
		"for-each-ref",
		"--sort=-committerdate",
		"--format="+branchFormat,
		"refs/remotes/",
	)
	if err == nil {
		for _, line := range strings.Split(rout, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			f := strings.Split(line, "\x00")
			if len(f) != 7 {
				continue
			}
			full := f[1] // e.g. "origin/dev-Branch" or "origin/HEAD"
			slash := strings.IndexByte(full, '/')
			if slash < 0 {
				continue
			}
			remote := full[:slash]
			short := full[slash+1:]
			if short == "HEAD" { // skip the origin/HEAD symbolic ref
				continue
			}
			if idx, ok := index[short]; ok {
				// local branch of the same name already exists → mark also-remote
				branches[idx].IsRemote = true
				if branches[idx].Remote == "" {
					branches[idx].Remote = remote
				}
				continue
			}
			// remote-only branch: surface it so it's visible + checkout-able
			b := Branch{
				Name:        short,
				IsRemote:    true,
				Remote:      remote,
				LastHash:    f[4],
				LastSubject: f[6],
			}
			if ts, err := strconv.ParseInt(strings.TrimSpace(f[5]), 10, 64); err == nil {
				b.LastWhen = time.Unix(ts, 0)
			}
			index[short] = len(branches)
			branches = append(branches, b)
		}
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
