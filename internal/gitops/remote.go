package gitops

import "strings"

// Remote is a named remote and its fetch URL.
type Remote struct {
	Name string
	URL  string
}

// ListRemotes returns the configured remotes with their fetch URLs.
func ListRemotes(dir string) ([]Remote, error) {
	// `git remote -v` prints two lines per remote (fetch + push); dedupe by name,
	// keeping the fetch URL.
	out, err := run(dir, "remote", "-v")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var remotes []Remote
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// format: "<name>\t<url> (fetch|push)"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name, url := fields[0], fields[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		remotes = append(remotes, Remote{Name: name, URL: url})
	}
	return remotes, nil
}

// RemoveRemote deletes a remote reference (does not touch commits).
func RemoveRemote(dir, name string) error {
	_, err := run(dir, "remote", "remove", name)
	return err
}

// RenameRemote renames a remote (and its remote-tracking refs).
func RenameRemote(dir, oldName, newName string) error {
	_, err := run(dir, "remote", "rename", oldName, newName)
	return err
}
