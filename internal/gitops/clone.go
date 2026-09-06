package gitops

import (
	"os"
	"path/filepath"
	"strings"
)

// Clone clones url into dest. dest is the target directory for the working copy
// (e.g. C:\repos\gitmate). If dest is empty, git derives it from the URL's repo
// name, placed under the current directory. Returns the absolute path of the
// resulting repo directory.
//
// Unlike other operations, clone runs OUTSIDE any existing repo — git is invoked
// in dest's parent directory (created if needed) so the clone lands correctly.
func Clone(url, dest string) (string, error) {
	url = strings.TrimSpace(url)
	dest = strings.TrimSpace(dest)

	if dest == "" {
		// derive "repo" from ".../repo.git" or ".../repo"
		name := url
		if i := strings.LastIndexAny(name, "/:"); i >= 0 {
			name = name[i+1:]
		}
		name = strings.TrimSuffix(name, ".git")
		if name == "" {
			name = "repo"
		}
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dest = filepath.Join(cwd, name)
	}

	abs, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	parent := filepath.Dir(abs)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}

	// run git clone in the parent, targeting the leaf dir name
	if _, err := run(parent, "clone", url, abs); err != nil {
		return "", err
	}
	return abs, nil
}
