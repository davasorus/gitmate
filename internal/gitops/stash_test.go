package gitops

import "testing"

func TestStashApplyAndDrop(t *testing.T) {
	dir := newTestRepo(t)
	writeFile(t, dir, "a.txt", "hello\n")
	_ = Stage(dir)
	_, _ = CreateCommit(dir, "init")

	writeFile(t, dir, "a.txt", "modified\n")
	if err := StashSave(dir, "wip", false); err != nil {
		t.Fatal(err)
	}
	list, err := StashList(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 stash, got %d", len(list))
	}
	// apply keeps the stash
	if err := StashApply(dir, "stash@{0}"); err != nil {
		t.Fatal(err)
	}
	if list, _ = StashList(dir); len(list) != 1 {
		t.Fatalf("apply should keep stash, got %d", len(list))
	}
	// drop removes it
	if err := StashDrop(dir, "stash@{0}"); err != nil {
		t.Fatal(err)
	}
	if list, _ = StashList(dir); len(list) != 0 {
		t.Fatalf("drop should remove stash, got %d", len(list))
	}
}
