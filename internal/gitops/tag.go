package gitops

import (
	"errors"
	"strings"
)

var errTagName = errors.New("tag name cannot be empty")

// Tag is one entry from the tag list, with where it exists.
type Tag struct {
	Name    string
	Subject string // annotation message or the tagged commit's subject (local only)
	Local   bool   // exists in the local repo
	Remote  bool   // exists on origin
}

// Location returns a human label: "local only", "remote only", or "both".
func (t Tag) Location() string {
	switch {
	case t.Local && t.Remote:
		return "both"
	case t.Local:
		return "local only"
	case t.Remote:
		return "remote only"
	default:
		return ""
	}
}

// ListTags returns tags, newest first (by creation), with a one-line subject.
func ListTags(dir string) ([]Tag, error) {
	// Local tags with their subject line.
	out, err := run(dir, "for-each-ref", "--sort=-creatordate",
		"--format=%(refname:short)%00%(contents:subject)", "refs/tags/")
	if err != nil {
		return nil, err
	}
	byName := map[string]*Tag{}
	var order []string
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.SplitN(line, "\x00", 2)
		t := &Tag{Name: f[0], Local: true}
		if len(f) == 2 {
			t.Subject = f[1]
		}
		byName[t.Name] = t
		order = append(order, t.Name)
	}

	// Remote tags (best-effort — network may fail; then we just show local).
	if rout, rerr := run(dir, "ls-remote", "--tags", "origin"); rerr == nil {
		for _, line := range strings.Split(rout, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// lines look like: <sha>\trefs/tags/<name>   (and <name>^{} for annotated)
			i := strings.Index(line, "refs/tags/")
			if i < 0 {
				continue
			}
			name := strings.TrimSuffix(line[i+len("refs/tags/"):], "^{}")
			if t, ok := byName[name]; ok {
				t.Remote = true
			} else {
				byName[name] = &Tag{Name: name, Remote: true}
				order = append(order, name)
			}
		}
	}

	tags := make([]Tag, 0, len(order))
	seen := map[string]bool{}
	for _, n := range order {
		if seen[n] {
			continue
		}
		seen[n] = true
		tags = append(tags, *byName[n])
	}
	return tags, nil
}

// CreateTag makes a tag at HEAD. If message is non-empty it's an annotated tag
// (-a -m); otherwise a lightweight tag. Fails if the tag already exists.
func CreateTag(dir, name, message string) error {
	if strings.TrimSpace(name) == "" {
		return errTagName
	}
	var args []string
	if strings.TrimSpace(message) != "" {
		args = []string{"tag", "-a", name, "-m", message}
	} else {
		args = []string{"tag", name}
	}
	_, err := run(dir, args...)
	return err
}

// DeleteTag removes a local tag.
func DeleteTag(dir, name string) error {
	_, err := run(dir, "tag", "-d", name)
	return err
}

// PushTag pushes a single tag to origin. This is what triggers a tag-based
// release workflow — tags do NOT travel with a normal `git push`.
func PushTag(dir, name string) error {
	_, err := run(dir, "push", "origin", "refs/tags/"+name)
	return err
}

// DeleteRemoteTag deletes a tag from origin (git push origin --delete <tag>).
// Local deletion is separate (DeleteTag) — this only removes it from the remote.
func DeleteRemoteTag(dir, name string) error {
	_, err := run(dir, "push", "origin", "--delete", "refs/tags/"+name)
	return err
}

// SmartDeleteTag deletes a tag wherever it exists: local, remote, or both.
// It checks presence first so it never errors trying to delete a side that
// isn't there. Returns a short description of what was removed.
func SmartDeleteTag(dir, name string) (string, error) {
	// determine current location
	var local, remote bool
	if _, err := run(dir, "rev-parse", "--verify", "--quiet", "refs/tags/"+name); err == nil {
		local = true
	}
	if out, err := run(dir, "ls-remote", "--tags", "origin", "refs/tags/"+name); err == nil && strings.TrimSpace(out) != "" {
		remote = true
	}
	switch {
	case local && remote:
		if err := DeleteTag(dir, name); err != nil {
			return "", err
		}
		if err := DeleteRemoteTag(dir, name); err != nil {
			return "", err
		}
		return "local + origin", nil
	case local:
		if err := DeleteTag(dir, name); err != nil {
			return "", err
		}
		return "local", nil
	case remote:
		if err := DeleteRemoteTag(dir, name); err != nil {
			return "", err
		}
		return "origin", nil
	default:
		return "", errTagName // nothing to delete
	}
}

// FetchTags syncs tags from origin, pruning local tags that were deleted on the
// remote (so ListTags reflects the remote after e.g. a browser-side deletion).
func FetchTags(dir string) error {
	_, err := run(dir, "fetch", "--tags", "--prune", "--prune-tags", "origin")
	return err
}
