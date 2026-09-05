package gitops

import (
	"errors"
	"strings"
)

var errTagName = errors.New("tag name cannot be empty")

// Tag is one entry from the tag list.
type Tag struct {
	Name    string
	Subject string // annotation message or the tagged commit's subject
}

// ListTags returns tags, newest first (by creation), with a one-line subject.
func ListTags(dir string) ([]Tag, error) {
	// %(refname:short) = tag name; %(contents:subject) = annotation or commit subject
	out, err := run(dir, "for-each-ref", "--sort=-creatordate",
		"--format=%(refname:short)%00%(contents:subject)", "refs/tags/")
	if err != nil {
		return nil, err
	}
	var tags []Tag
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.SplitN(line, "\x00", 2)
		t := Tag{Name: f[0]}
		if len(f) == 2 {
			t.Subject = f[1]
		}
		tags = append(tags, t)
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
