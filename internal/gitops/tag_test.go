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

func TestTagLocation(t *testing.T) {
	// Location() is a pure method on Tag based on Local/Remote flags
	local := Tag{Name: "v1", Local: true, Remote: false}
	both := Tag{Name: "v2", Local: true, Remote: true}
	remote := Tag{Name: "v3", Local: false, Remote: true}
	if local.Location() == "" || both.Location() == "" || remote.Location() == "" {
		t.Fatal("Location should return a non-empty label for each state")
	}
}

func TestTagPushDeleteRemoteFetch(t *testing.T) {
	dir := newRemoteRepo(t)
	if err := CreateTag(dir, "v1.0", "release"); err != nil {
		t.Fatal(err)
	}
	if err := PushTag(dir, "v1.0"); err != nil {
		t.Fatal(err)
	}
	if err := FetchTags(dir); err != nil {
		t.Fatal(err)
	}
	if err := DeleteRemoteTag(dir, "v1.0"); err != nil {
		t.Fatal(err)
	}
}

func TestSmartDeleteTag(t *testing.T) {
	dir := newRemoteRepo(t)
	if err := CreateTag(dir, "v2.0", "rel"); err != nil {
		t.Fatal(err)
	}
	if err := PushTag(dir, "v2.0"); err != nil {
		t.Fatal(err)
	}
	if _, err := SmartDeleteTag(dir, "v2.0"); err != nil {
		t.Fatal(err)
	}
}
