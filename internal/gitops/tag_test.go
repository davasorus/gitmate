package gitops

import "testing"

func tagRepo(t *testing.T) string {
	t.Helper()
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello\n")
	if err := Stage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCommit(dir, "init"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestTagsLifecycle(t *testing.T) {
	dir := tagRepo(t)
	if err := CreateTag(dir, "v1.0", "first release"); err != nil {
		t.Fatal(err)
	}
	tags, err := ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0].Name != "v1.0" {
		t.Fatalf("tags wrong: %+v", tags)
	}
	if err := DeleteTag(dir, "v1.0"); err != nil {
		t.Fatal(err)
	}
	if tags, _ = ListTags(dir); len(tags) != 0 {
		t.Fatalf("expected no tags after delete, got %+v", tags)
	}
}
