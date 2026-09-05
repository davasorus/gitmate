package gitops

import (
	"os"
	"path/filepath"
	"strings"
)

// LastCommitSubject returns the subject line of the most recent commit on the
// given branch (or HEAD if branch is empty). Used to default a PR title.
func LastCommitSubject(dir, branch string) (string, error) {
	ref := branch
	if ref == "" {
		ref = "HEAD"
	}
	out, err := run(dir, "log", "-1", "--format=%s", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ReadPRTemplate returns the repo's PR template contents, or "" if none.
func ReadPRTemplate(dir string) string {
	candidates := []string{
		".github/PULL_REQUEST_TEMPLATE.md",
		".github/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
		"docs/PULL_REQUEST_TEMPLATE.md",
		"docs/pull_request_template.md",
	}
	for _, c := range candidates {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(c)))
		if err == nil {
			return string(b)
		}
	}
	return ""
}
