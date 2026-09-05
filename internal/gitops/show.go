package gitops

import "strings"

// CommitDetail is a single commit's metadata plus its file diffs.
type CommitDetail struct {
	Hash    string
	Short   string
	Author  string
	Email   string
	Date    string // committer date, human-readable
	Subject string
	Body    string // full message body after the subject (may be empty)
	Files   []FileDiff
}

// Show returns metadata and the diff for a single revision. It uses `git show`,
// which correctly handles the root commit (no parent) by treating it as all
// additions. rev may be a sha, short sha, "HEAD", "HEAD~2", a tag, etc.
func Show(dir, rev string) (*CommitDetail, error) {
	if strings.TrimSpace(rev) == "" {
		rev = "HEAD"
	}

	// Metadata via a null-delimited format (same discipline as log).
	const fmtStr = "%H%x00%h%x00%an%x00%ae%x00%cd%x00%s%x00%b"
	out, err := run(dir, "show", "--no-patch", "--date=format:%Y-%m-%d %H:%M", "--format="+fmtStr, rev)
	if err != nil {
		return nil, err
	}
	f := strings.Split(strings.TrimRight(out, "\n"), "\x00")
	d := &CommitDetail{}
	if len(f) >= 7 {
		d.Hash, d.Short, d.Author, d.Email, d.Date, d.Subject, d.Body =
			f[0], f[1], f[2], f[3], f[4], f[5], strings.TrimSpace(f[6])
	}

	// The patch itself: `git show <rev>` prints the same unified-diff format
	// our parser already understands. Use --first-parent-ish plain show.
	patch, err := run(dir, "show", "--format=", rev)
	if err != nil {
		return nil, err
	}
	d.Files = parseUnifiedDiff(patch)
	return d, nil
}
