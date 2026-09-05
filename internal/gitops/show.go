package gitops

import "strings"

// CommitDetail is a single commit's metadata plus its file diffs.
type CommitDetail struct {
	Hash    string
	Short   string
	Author  string
	Email   string
	Date    string
	Subject string
	Body    string
	Files   []FileDiff
}

// Show returns metadata and the diff for a single revision, using `git show`
// (which handles the root commit — no parent — correctly).
func Show(dir, rev string) (*CommitDetail, error) {
	if strings.TrimSpace(rev) == "" {
		rev = "HEAD"
	}
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
	patch, err := run(dir, "show", "--format=", rev)
	if err != nil {
		return nil, err
	}
	d.Files = parseUnifiedDiff(patch)
	return d, nil
}
